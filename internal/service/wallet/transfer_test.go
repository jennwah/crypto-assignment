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

func TestTransfer(t *testing.T) {
	type args struct {
		initiatorUserID string
		recipientUserID string
		idempotencyKey  string
		amount          uint64
	}
	tests := []struct {
		name          string
		args          args
		mockBehavior  func(m *mocks.MockIWalletRepository)
		expectedTxID  string
		expectedError error
	}{
		{
			name: "happy case - valid transfer",
			args: args{
				initiatorUserID: "user123",
				recipientUserID: "user456",
				idempotencyKey:  "unique-key",
				amount:          1000,
			},
			mockBehavior: func(m *mocks.MockIWalletRepository) {
				m.EXPECT().
					Transfer(gomock.Any(), "user123", "user456", "unique-key", uint64(1000)).
					Return("tx123", nil)
			},
			expectedTxID:  "tx123",
			expectedError: nil,
		},
		{
			name: "repo error",
			args: args{
				initiatorUserID: "user789",
				recipientUserID: "user321",
				idempotencyKey:  "unique-key",
				amount:          500,
			},
			mockBehavior: func(m *mocks.MockIWalletRepository) {
				m.EXPECT().
					Transfer(gomock.Any(), "user789", "user321", "unique-key", uint64(500)).
					Return("", errors.New("db connection error"))
			},
			expectedTxID:  "",
			expectedError: errors.New("repo transfer err: db connection error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockIWalletRepository(ctrl)
			tt.mockBehavior(mockRepo)

			service := wallet.New(mockRepo)

			txID, err := service.Transfer(
				context.Background(),
				tt.args.initiatorUserID,
				tt.args.recipientUserID,
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
