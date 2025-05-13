package wallet

import (
	"context"
	"fmt"
)

func (s *Service) WithdrawWallet(
	ctx context.Context,
	userID, idempotencyKey string,
	amount uint64,
) (string, error) {
	txID, err := s.walletRepo.WithdrawWallet(ctx, userID, idempotencyKey, amount)
	if err != nil {
		return "", fmt.Errorf("withdraw wallet repo err: %w", err)
	}

	return txID, nil
}
