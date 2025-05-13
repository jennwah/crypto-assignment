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

func TestDepositWallet(t *testing.T) {
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
			name: "success - valid deposit",
			args: args{
				userID:         "user123",
				idempotencyKey: "deposit-key-1",
				amount:         1500,
			},
			mockBehavior: func(m *mocks.MockIWalletRepository) {
				m.EXPECT().
					DepositWallet(gomock.Any(), "user123", "deposit-key-1", uint64(1500)).
					Return("tx456", nil)
			},
			expectedTxID:  "tx456",
			expectedError: nil,
		},
		{
			name: "error - repository failure",
			args: args{
				userID:         "user999",
				idempotencyKey: "deposit-fail",
				amount:         100,
			},
			mockBehavior: func(m *mocks.MockIWalletRepository) {
				m.EXPECT().
					DepositWallet(gomock.Any(), "user999", "deposit-fail", uint64(100)).
					Return("", errors.New("db write error"))
			},
			expectedTxID:  "",
			expectedError: errors.New("deposit wallet repo err: db write error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockIWalletRepository(ctrl)
			tt.mockBehavior(mockRepo)

			service := wallet.New(mockRepo)

			txID, err := service.DepositWallet(
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
