package wallet

import (
	"context"
	"fmt"
)

func (s *Service) DepositWallet(
	ctx context.Context,
	userID, idempotencyKey string,
	amount uint64,
) (string, error) {
	txID, err := s.walletRepo.DepositWallet(ctx, userID, idempotencyKey, amount)
	if err != nil {
		return "", fmt.Errorf("deposit wallet repo err: %w", err)
	}

	return txID, nil
}
