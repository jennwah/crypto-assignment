package wallet_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domainwallet "github.com/jennwah/crypto-assignment/internal/domain/wallet"
	"github.com/jennwah/crypto-assignment/internal/repository/wallet"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestGetWallet(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	r := wallet.New(sqlxDB, nil, nil)

	tests := []struct {
		name          string
		prepareMock   func()
		expected      domainwallet.Wallet
		expectedError error
	}{
		{
			name: "wallet found",
			prepareMock: func() {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, balance, created_at FROM wallets WHERE user_id = $1 LIMIT 1;`)).
					WithArgs("user123").
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "balance", "created_at"}).
						AddRow("wallet-1", "user123", 1000, time.Now()))
			},
			expected: domainwallet.Wallet{
				ID:      "wallet-1",
				UserID:  "user123",
				Balance: 1000,
			},
			expectedError: nil,
		},
		{
			name: "wallet not found",
			prepareMock: func() {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, balance, created_at FROM wallets WHERE user_id = $1 LIMIT 1;`)).
					WithArgs("").
					WillReturnError(sql.ErrNoRows)
			},
			expected:      domainwallet.Wallet{},
			expectedError: domainwallet.ErrWalletNotFound,
		},
		{
			name: "db error",
			prepareMock: func() {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, balance, created_at FROM wallets WHERE user_id = $1 LIMIT 1;`)).
					WithArgs("").
					WillReturnError(errors.New("db error"))
			},
			expected:      domainwallet.Wallet{},
			expectedError: errors.New("failed to get wallet by userID: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepareMock()
			wallet, err := r.GetWallet(context.Background(), tt.expected.UserID)
			assert.Equal(t, tt.expected.ID, wallet.ID)
			assert.Equal(t, tt.expected.UserID, wallet.UserID)
			assert.Equal(t, tt.expected.Balance, wallet.Balance)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetWalletTransactionsHistory(t *testing.T) {
	testTime := time.Now()
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	r := wallet.New(sqlxDB, nil, nil)

	tests := []struct {
		name          string
		prepareMock   func()
		expectedTxs   []domainwallet.Transaction
		expectedTotal int
		expectedError error
	}{
		{
			name: "successful history fetch",
			prepareMock: func() {
				mock.ExpectQuery(`SELECT id FROM wallets WHERE user_id = \$1`).
					WithArgs("user123").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("wallet-1"))

				mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM transactions t JOIN wallets iw ON t.initiator_wallet_id = iw.id LEFT JOIN wallets rw ON t.recipient_wallet_id = rw.id WHERE iw.user_id = $1 OR rw.user_id = $1;`)).
					WithArgs("user123").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

				mock.ExpectQuery(regexp.QuoteMeta(`SELECT t.id, iw.user_id AS initiator_wallet_user_id, t.type, t.status, t.amount, rw.user_id AS recipient_wallet_user_id, t.created_at FROM transactions t JOIN wallets iw ON t.initiator_wallet_id = iw.id LEFT JOIN wallets rw ON t.recipient_wallet_id = rw.id WHERE iw.user_id = $1 OR rw.user_id = $1 ORDER BY t.created_at DESC OFFSET $2 LIMIT $3;`)).
					WithArgs("user123", 0, 10).
					WillReturnRows(sqlmock.NewRows([]string{"id", "initiator_wallet_user_id", "type", "status", "amount", "recipient_wallet_user_id", "created_at"}).
						AddRow("tx1", "user123", "deposit", "success", 100, nil, testTime.String()))
			},
			expectedTxs: []domainwallet.Transaction{
				{
					ID:                    "tx1",
					InitiatorWalletUserId: "user123",
					Type:                  "deposit",
					Status:                "success",
					Amount:                100,
					RecipientWalletUserId: nil,
					CreatedAt:             testTime.String(),
				},
			},
			expectedTotal: 1,
			expectedError: nil,
		},
		{
			name: "wallet not found",
			prepareMock: func() {
				mock.ExpectQuery(`SELECT id FROM wallets WHERE user_id = \$1`).
					WithArgs("user123").
					WillReturnError(sql.ErrNoRows)
			},
			expectedTxs:   nil,
			expectedTotal: 0,
			expectedError: domainwallet.ErrWalletNotFound,
		},
		{
			name: "error selecting wallet",
			prepareMock: func() {
				mock.ExpectQuery(`SELECT id FROM wallets WHERE user_id = \$1`).
					WithArgs("user123").
					WillReturnError(fmt.Errorf("db error"))
			},
			expectedTxs:   nil,
			expectedTotal: 0,
			expectedError: fmt.Errorf("db error"),
		},
		{
			name: "error counting transactions",
			prepareMock: func() {
				mock.ExpectQuery(`SELECT id FROM wallets WHERE user_id = \$1`).
					WithArgs("user123").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("wallet-1"))

				mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM transactions t JOIN wallets iw ON t.initiator_wallet_id = iw.id LEFT JOIN wallets rw ON t.recipient_wallet_id = rw.id WHERE iw.user_id = $1 OR rw.user_id = $1;`)).
					WithArgs("user123").
					WillReturnError(fmt.Errorf("count error"))
			},
			expectedTxs:   nil,
			expectedTotal: 0,
			expectedError: fmt.Errorf("count error"),
		},
		{
			name: "no transactions",
			prepareMock: func() {
				mock.ExpectQuery(`SELECT id FROM wallets WHERE user_id = \$1`).
					WithArgs("user123").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("wallet-1"))

				mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM transactions t JOIN wallets iw ON t.initiator_wallet_id = iw.id LEFT JOIN wallets rw ON t.recipient_wallet_id = rw.id WHERE iw.user_id = $1 OR rw.user_id = $1;`)).
					WithArgs("user123").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			expectedTxs:   nil,
			expectedTotal: 0,
			expectedError: nil,
		},
		{
			name: "error fetching transactions after count",
			prepareMock: func() {
				mock.ExpectQuery(`SELECT id FROM wallets WHERE user_id = \$1`).
					WithArgs("user123").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("wallet-1"))

				mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM transactions t JOIN wallets iw ON t.initiator_wallet_id = iw.id LEFT JOIN wallets rw ON t.recipient_wallet_id = rw.id WHERE iw.user_id = $1 OR rw.user_id = $1;`)).
					WithArgs("user123").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

				mock.ExpectQuery(regexp.QuoteMeta(`SELECT t.id, iw.user_id AS initiator_wallet_user_id, t.type, t.status, t.amount, rw.user_id AS recipient_wallet_user_id, t.created_at FROM transactions t JOIN wallets iw ON t.initiator_wallet_id = iw.id LEFT JOIN wallets rw ON t.recipient_wallet_id = rw.id WHERE iw.user_id = $1 OR rw.user_id = $1 ORDER BY t.created_at DESC OFFSET $2 LIMIT $3;`)).
					WithArgs("user123", 0, 10).
					WillReturnError(fmt.Errorf("fetch error"))
			},
			expectedTxs:   nil,
			expectedTotal: 0,
			expectedError: fmt.Errorf("fetch error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepareMock()
			txs, total, err := r.GetWalletTransactionsHistory(context.Background(), "user123", 0, 10)
			assert.Equal(t, tt.expectedTxs, txs)
			assert.Equal(t, tt.expectedTotal, total)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
