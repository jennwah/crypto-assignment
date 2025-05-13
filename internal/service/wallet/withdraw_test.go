package wallet_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jennwah/crypto-assignment/internal/repository/wallet/mocks"
	"github.com/jennwah/crypto-assignment/internal/service/wallet"
	"github.com/stretchr/testify/assert"
)

func TestWithdrawWallet(t *testing.T) {
	type args struct {
		userID         string
		idempotencyKey string
		amount         uint64
	}
	tests := []struct {
		name          string
		args          args
		mockBehavior  func(m *mocks.MockIWalletRepository)
		expectedTxID  string
		expectedError error
	}{
		{
			name: "success - valid withdrawal",
			args: args{
				userID:         "user123",
				idempotencyKey: "withdraw-key-1",
				amount:         750,
			},
			mockBehavior: func(m *mocks.MockIWalletRepository) {
				m.EXPECT().
					WithdrawWallet(gomock.Any(), "user123", "withdraw-key-1", uint64(750)).
					Return("tx789", nil)
			},
			expectedTxID:  "tx789",
			expectedError: nil,
		},
		{
			name: "error - insufficient funds",
			args: args{
				userID:         "user999",
				idempotencyKey: "withdraw-fail",
				amount:         5000,
			},
			mockBehavior: func(m *mocks.MockIWalletRepository) {
				m.EXPECT().
					WithdrawWallet(gomock.Any(), "user999", "withdraw-fail", uint64(5000)).
					Return("", errors.New("insufficient funds"))
			},
			expectedTxID:  "",
			expectedError: errors.New("withdraw wallet repo err: insufficient funds"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockIWalletRepository(ctrl)
			tt.mockBehavior(mockRepo)

			service := wallet.New(mockRepo)

			txID, err := service.WithdrawWallet(
				context.Background(),
				tt.args.userID,
				tt.args.idempotencyKey,
				tt.args.amount,
			)

			assert.Equal(t, tt.expectedTxID, txID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
