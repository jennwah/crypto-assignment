package wallet

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domainwallet "github.com/jennwah/crypto-assignment/internal/domain/wallet"
)

func (r *Repository) GetWallet(ctx context.Context, userID string) (domainwallet.Wallet, error) {
	const query = `
		SELECT id, user_id, balance, created_at
		FROM wallets
		WHERE user_id = $1
		LIMIT 1;
	`

	var dst domainwallet.Wallet
	err := r.db.GetContext(ctx, &dst, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainwallet.Wallet{}, domainwallet.ErrWalletNotFound
		}
		return domainwallet.Wallet{}, fmt.Errorf("failed to get wallet by userID: %w", err)
	}

	return dst, nil
}

func (r *Repository) GetWalletTransactionsHistory(
	ctx context.Context,
	userID string,
	offset, pageSize int,
) ([]domainwallet.Transaction, int, error) {
	var walletID string
	err := r.db.GetContext(ctx, &walletID, `SELECT id FROM wallets WHERE user_id = $1`, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, fmt.Errorf(
				"failed get user wallet: %w",
				domainwallet.ErrWalletNotFound,
			)
		}
		return nil, 0, fmt.Errorf("failed to get wallet for user %s: %w", userID, err)
	}

	var total int
	const countQuery = `
		SELECT COUNT(*) FROM transactions t
		JOIN wallets iw ON t.initiator_wallet_id = iw.id
		LEFT JOIN wallets rw ON t.recipient_wallet_id = rw.id
		WHERE iw.user_id = $1 OR rw.user_id = $1;`

	err = r.db.GetContext(ctx, &total, countQuery, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count transactions: %w", err)
	}
	if total == 0 {
		return nil, 0, nil
	}

	const query = `
		SELECT 
			t.id,
			iw.user_id AS initiator_wallet_user_id,
			t.type,
			t.status,
			t.amount,
			rw.user_id AS recipient_wallet_user_id,
			t.created_at
		FROM transactions t
		JOIN wallets iw ON t.initiator_wallet_id = iw.id
		LEFT JOIN wallets rw ON t.recipient_wallet_id = rw.id
		WHERE iw.user_id = $1 OR rw.user_id = $1
		ORDER BY t.created_at DESC
		OFFSET $2 LIMIT $3;
	`

	var transactions []domainwallet.Transaction
	err = r.db.SelectContext(ctx, &transactions, query, userID, offset, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch transactions: %w", err)
	}

	return transactions, total, nil
}
