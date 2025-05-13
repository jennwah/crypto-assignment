package wallet

import (
	"context"
	"fmt"
)

func (s *Service) Transfer(
	ctx context.Context,
	initiatorUserID, recipientUserID, idempotencyKey string,
	amount uint64,
) (string, error) {
	txID, err := s.walletRepo.Transfer(
		ctx,
		initiatorUserID,
		recipientUserID,
		idempotencyKey,
		amount,
	)
	if err != nil {
		return "", fmt.Errorf("repo transfer err: %w", err)
	}

	return txID, nil
}
