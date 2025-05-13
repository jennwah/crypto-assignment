package wallet

import (
	"context"

	"github.com/jennwah/crypto-assignment/internal/domain/wallet"
)

type IWalletRepository interface {
	GetWallet(ctx context.Context, userID string) (wallet.Wallet, error)
	GetWalletTransactionsHistory(
		ctx context.Context, userID string, offset, pageSize int,
	) ([]wallet.Transaction, int, error)
	DepositWallet(ctx context.Context, userID, idempotencyKey string, amount uint64) (string, error)
	WithdrawWallet(
		ctx context.Context,
		userID, idempotencyKey string,
		amount uint64,
	) (string, error)
	Transfer(
		ctx context.Context, initiatorUserID, recipientUserID, idempotencyKey string, amount uint64,
	) (string, error)
}
