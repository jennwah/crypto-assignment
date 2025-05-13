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

func TestWithdrawWallet(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "postgres")

	redisClient, redisMock := redismock.NewClientMock()
	logger := slog.Default()
	repo := wallet.New(sqlxDB, redisClient, logger)

	tests := []struct {
		name           string
		userID         string
		idempotencyKey string
		amount         uint64
		prepareRedis   func()
		prepareSQL     func()
		expectedError  error
		expectTxnID    string
	}{
		{
			name:           "already processed idempotent request",
			userID:         "user123",
			idempotencyKey: "idem123",
			amount:         100,
			prepareRedis: func() {
				redisMock.ExpectGet("withdraw-user123-idem123").SetVal("tx-already")
			},
			prepareSQL:    func() {},
			expectTxnID:   "tx-already",
			expectedError: nil,
		},
		{
			name:           "wallet not found",
			userID:         "user124",
			idempotencyKey: "idem124",
			amount:         100,
			prepareRedis: func() {
				redisMock.ExpectGet("withdraw-user124-idem124").RedisNil()
			},
			prepareSQL: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT id, balance FROM wallets WHERE user_id = \$1 FOR UPDATE`).
					WithArgs("user124").
					WillReturnError(sql.ErrNoRows)
			},
			expectedError: fmt.Errorf("wallet not found: %w", domainwallet.ErrWalletNotFound),
		},
		{
			name:           "insufficient balance",
			userID:         "user125",
			idempotencyKey: "idem125",
			amount:         500,
			prepareRedis: func() {
				redisMock.ExpectGet("withdraw-user125-idem125").RedisNil()
			},
			prepareSQL: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT id, balance FROM wallets WHERE user_id = \$1 FOR UPDATE`).
					WithArgs("user125").
					WillReturnRows(sqlmock.NewRows([]string{"id", "balance"}).AddRow("wallet125", 100))
				mock.ExpectRollback()
			},
			expectedError: fmt.Errorf("insufficient balance to deduct: %w", domainwallet.ErrWalletInsufficientBalance),
		},
		{
			name:           "successful withdraw",
			userID:         "user126",
			idempotencyKey: "idem126",
			amount:         200,
			prepareRedis: func() {
				redisMock.ExpectGet("withdraw-user126-idem126").RedisNil()
				redisMock.ExpectSet("withdraw-user126-idem126", "tx126", time.Hour*24).SetVal("OK")
			},
			prepareSQL: func() {
				mock.ExpectBegin()

				mock.ExpectQuery(`SELECT id, balance FROM wallets WHERE user_id = \$1 FOR UPDATE`).
					WithArgs("user126").
					WillReturnRows(sqlmock.NewRows([]string{"id", "balance"}).AddRow("wallet126", 1000))

				mock.ExpectExec(`UPDATE wallets SET balance = balance - \$1 WHERE id = \$2`).
					WithArgs(200, "wallet126").
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectQuery(`INSERT INTO transactions .* RETURNING id`).
					WithArgs("wallet126", "withdraw", "success", 200).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("tx126"))

				mock.ExpectCommit()
			},
			expectTxnID:   "tx126",
			expectedError: nil,
		},
		{
			name:           "redis get failure",
			userID:         "user127",
			idempotencyKey: "idem127",
			amount:         100,
			prepareRedis: func() {
				redisMock.ExpectGet("withdraw-user127-idem127").SetErr(errors.New("redis failure"))
			},
			prepareSQL:    func() {},
			expectedError: fmt.Errorf("redis get failed: %w", errors.New("redis failure")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepareRedis()
			tt.prepareSQL()

			txID, err := repo.WithdrawWallet(context.Background(), tt.userID, tt.idempotencyKey, tt.amount)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectTxnID, txID)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
			assert.NoError(t, redisMock.ExpectationsWereMet())
		})
	}
}
