package wallet_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	redismock "github.com/go-redis/redismock/v9"
	domainwallet "github.com/jennwah/crypto-assignment/internal/domain/wallet"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jennwah/crypto-assignment/internal/repository/wallet"
)

func TestTransferWallet(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "postgres")

	redisClient, redisMock := redismock.NewClientMock()
	logger := slog.Default()
	repo := wallet.New(sqlxDB, redisClient, logger)

	tests := []struct {
		name            string
		initiatorUserID string
		recipientUserID string
		idempotencyKey  string
		amount          uint64
		prepareRedis    func()
		prepareSQL      func()
		expectedError   error
		expectedTxnID   string
	}{
		{
			name:            "idempotency key already processed",
			initiatorUserID: "user1",
			recipientUserID: "user2",
			idempotencyKey:  "idem001",
			amount:          100,
			prepareRedis: func() {
				redisMock.ExpectGet("transfer-user1-idem001").SetVal("tx-already")
			},
			prepareSQL:    func() {},
			expectedTxnID: "tx-already",
			expectedError: nil,
		},
		{
			name:            "redis get error",
			initiatorUserID: "user3",
			recipientUserID: "user4",
			idempotencyKey:  "idem002",
			amount:          50,
			prepareRedis: func() {
				redisMock.ExpectGet("transfer-user3-idem002").SetErr(errors.New("redis down"))
			},
			prepareSQL:    func() {},
			expectedError: fmt.Errorf("redis get transferCacheKey failed: %w", errors.New("redis down")),
		},
		{
			name:            "wallet not found for initiator",
			initiatorUserID: "user5",
			recipientUserID: "user6",
			idempotencyKey:  "idem003",
			amount:          20,
			prepareRedis: func() {
				redisMock.ExpectGet("transfer-user5-idem003").RedisNil()
			},
			prepareSQL: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT id, balance FROM wallets WHERE user_id = \$1 FOR UPDATE`).
					WithArgs("user5").
					WillReturnError(sql.ErrNoRows)
			},
			expectedError: fmt.Errorf("wallet not found: %w", domainwallet.ErrWalletNotFound),
		},
		{
			name:            "insufficient balance",
			initiatorUserID: "user7",
			recipientUserID: "user8",
			idempotencyKey:  "idem004",
			amount:          1000,
			prepareRedis: func() {
				redisMock.ExpectGet("transfer-user7-idem004").RedisNil()
			},
			prepareSQL: func() {
				mock.ExpectBegin()

				mock.ExpectQuery(`SELECT id, balance FROM wallets WHERE user_id = \$1 FOR UPDATE`).
					WithArgs("user7").
					WillReturnRows(sqlmock.NewRows([]string{"id", "balance"}).AddRow("wallet7", 100))

				mock.ExpectQuery(`SELECT id, balance FROM wallets WHERE user_id = \$1 FOR UPDATE`).
					WithArgs("user8").
					WillReturnRows(sqlmock.NewRows([]string{"id", "balance"}).AddRow("wallet8", 200))
			},
			expectedError: fmt.Errorf("insufficient balance to transfer: %w", domainwallet.ErrWalletInsufficientBalance),
		},
		{
			name:            "successful transfer",
			initiatorUserID: "user9",
			recipientUserID: "user10",
			idempotencyKey:  "idem005",
			amount:          500,
			prepareRedis: func() {
				redisMock.ExpectGet("transfer-user9-idem005").RedisNil()
				redisMock.ExpectSet("transfer-user9-idem005", "tx999", 24*time.Hour).SetVal("OK")
			},
			prepareSQL: func() {
				mock.ExpectBegin()

				mock.ExpectQuery(`SELECT id, balance FROM wallets WHERE user_id = \$1 FOR UPDATE`).
					WithArgs("user9").
					WillReturnRows(sqlmock.NewRows([]string{"id", "balance"}).AddRow("wallet9", 1000))

				mock.ExpectQuery(`SELECT id, balance FROM wallets WHERE user_id = \$1 FOR UPDATE`).
					WithArgs("user10").
					WillReturnRows(sqlmock.NewRows([]string{"id", "balance"}).AddRow("wallet10", 250))

				mock.ExpectExec(`UPDATE wallets SET balance = balance - \$1 WHERE id = \$2`).
					WithArgs(500, "wallet9").
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec(`UPDATE wallets SET balance = balance \+ \$1 WHERE id = \$2`).
					WithArgs(500, "wallet10").
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectQuery(`INSERT INTO transactions .* RETURNING id`).
					WithArgs("wallet9", "wallet10", "transfer", "success", 500).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("tx999"))

				mock.ExpectCommit()
			},
			expectedTxnID: "tx999",
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepareRedis()
			tt.prepareSQL()

			txID, err := repo.Transfer(context.Background(), tt.initiatorUserID, tt.recipientUserID, tt.idempotencyKey, tt.amount)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedTxnID, txID)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
			assert.NoError(t, redisMock.ExpectationsWereMet())
		})
	}
}
