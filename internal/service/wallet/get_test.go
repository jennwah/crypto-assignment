package wallet_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jennwah/crypto-assignment/internal/domain/wallet"
	"github.com/jennwah/crypto-assignment/internal/repository/wallet/mocks"
	servicewallet "github.com/jennwah/crypto-assignment/internal/service/wallet"
	"github.com/stretchr/testify/assert"
)

func TestGetWallet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIWalletRepository(ctrl)
	svc := servicewallet.New(mockRepo)

	testCases := []struct {
		name        string
		userID      string
		mockResult  wallet.Wallet
		mockError   error
		expectError bool
	}{
		{
			name:   "happy case",
			userID: "user123",
			mockResult: wallet.Wallet{
				ID:        "id-1",
				UserID:    "user123",
				Balance:   10000,
				CreatedAt: "2016-06-01T14:46:22.001Z",
			},
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "repo error",
			userID:      "user456",
			mockResult:  wallet.Wallet{},
			mockError:   errors.New("db error"),
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo.
				EXPECT().
				GetWallet(gomock.Any(), tc.userID).
				Return(tc.mockResult, tc.mockError)

			result, err := svc.GetWallet(context.Background(), tc.userID)

			if tc.expectError {
				assert.Error(t, err)
				assert.Equal(t, wallet.Wallet{}, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.mockResult, result)
			}
		})
	}
}

func TestGetWalletTransactionsHistory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIWalletRepository(ctrl)
	svc := servicewallet.New(mockRepo)

	testCases := []struct {
		name           string
		userID         string
		offset         int
		pageSize       int
		mockTxs        []wallet.Transaction
		mockTotal      int
		mockError      error
		expectError    bool
		expectedTxsLen int
	}{
		{
			name:     "happy case",
			userID:   "user789",
			offset:   0,
			pageSize: 2,
			mockTxs: []wallet.Transaction{
				{ID: "tx1", Amount: 500, Type: wallet.Deposit, Status: wallet.Success, InitiatorWalletUserId: "user789"},
				{ID: "tx2", Amount: 700, Type: wallet.Withdraw, Status: wallet.Success, InitiatorWalletUserId: "user789"},
			},
			mockTotal:      10,
			mockError:      nil,
			expectError:    false,
			expectedTxsLen: 2,
		},
		{
			name:           "repo error",
			userID:         "user123",
			offset:         0,
			pageSize:       2,
			mockTxs:        nil,
			mockTotal:      0,
			mockError:      errors.New("query failed"),
			expectError:    true,
			expectedTxsLen: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo.
				EXPECT().
				GetWalletTransactionsHistory(gomock.Any(), tc.userID, tc.offset, tc.pageSize).
				Return(tc.mockTxs, tc.mockTotal, tc.mockError)

			txs, total, err := svc.GetWalletTransactionsHistory(context.Background(), tc.userID, tc.offset, tc.pageSize)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, txs)
				assert.Equal(t, 0, total)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.mockTxs, txs)
				assert.Equal(t, tc.mockTotal, total)
			}
		})
	}
}
