package wallet

import (
	"context"
	"fmt"

	domainwallet "github.com/jennwah/crypto-assignment/internal/domain/wallet"
)

func (s *Service) GetWallet(ctx context.Context, userID string) (domainwallet.Wallet, error) {
	wallet, err := s.walletRepo.GetWallet(ctx, userID)
	if err != nil {
		return domainwallet.Wallet{}, fmt.Errorf("wallet service get wallet err: %w", err)
	}

	return wallet, nil
}

func (s *Service) GetWalletTransactionsHistory(
	ctx context.Context,
	userID string,
	offset, pageSize int,
) ([]domainwallet.Transaction, int, error) {
	transactions, total, err := s.walletRepo.GetWalletTransactionsHistory(
		ctx,
		userID,
		offset,
		pageSize,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("wallet service get wallet transactions history err: %w", err)
	}

	return transactions, total, nil
}
