package wallet_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	redismock "github.com/go-redis/redismock/v9"
	domainwallet "github.com/jennwah/crypto-assignment/internal/domain/wallet"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jennwah/crypto-assignment/internal/repository/wallet"
)

func TestDepositWallet(t *testing.T) {
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
				redisMock.ExpectGet("deposit-user123-idem123").SetVal("tx-already")
			},
			prepareSQL:    func() {},
			expectTxnID:   "tx-already",
			expectedError: nil,
		},
		{
			name:           "wallet not found",
			userID:         "user123",
			idempotencyKey: "idem124",
			amount:         200,
			prepareRedis: func() {
				redisMock.ExpectGet("deposit-user123-idem124").RedisNil()
			},
			prepareSQL: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT id FROM wallets WHERE user_id = \$1 FOR UPDATE`).
					WithArgs("user123").
					WillReturnError(sql.ErrNoRows)
			},
			expectedError: fmt.Errorf("wallet not found: %w", domainwallet.ErrWalletNotFound),
		},
		{
			name:           "successful deposit",
			userID:         "user456",
			idempotencyKey: "idem456",
			amount:         500,
			prepareRedis: func() {
				redisMock.ExpectGet("deposit-user456-idem456").RedisNil()
			},
			prepareSQL: func() {
				mock.ExpectBegin()

				mock.ExpectQuery(`SELECT id FROM wallets WHERE user_id = \$1 FOR UPDATE`).
					WithArgs("user456").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("wallet456"))

				mock.ExpectExec(`UPDATE wallets SET balance = balance \+ \$1 WHERE id = \$2`).
					WithArgs(500, "wallet456").
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectQuery(`INSERT INTO transactions .* RETURNING id`).
					WithArgs("wallet456", "deposit", "success", 500).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("tx456"))

				mock.ExpectCommit()
			},
			expectTxnID:   "tx456",
			expectedError: nil,
		},
		{
			name:           "get idempotency key cache error",
			userID:         "user789",
			idempotencyKey: "idem789",
			amount:         300,
			prepareRedis: func() {
				redisMock.ExpectGet("deposit-user789-idem789").SetErr(errors.New("redis down"))
			},
			prepareSQL:    func() {},
			expectedError: fmt.Errorf("redis get failed: %w", errors.New("redis down")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepareRedis()
			tt.prepareSQL()

			txID, err := repo.DepositWallet(context.Background(), tt.userID, tt.idempotencyKey, tt.amount)

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
