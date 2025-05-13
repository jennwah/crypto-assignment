package wallet

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	domainwallet "github.com/jennwah/crypto-assignment/internal/domain/wallet"
	"github.com/redis/go-redis/v9"
)

const depositCacheKey = `deposit-%s-%s` // deposit-userID-idempotencyKey

// DepositWallet does the following:
// 1. Check from redis cache on key = deposit-{userID}-{idempotencyKey}, if exists we just return nil error
// 2. If not, proceed with deposit amount into user wallet
// 3. Cache if successful and return appriopriate errors
func (r *Repository) DepositWallet(
	ctx context.Context,
	userID, idempotencyKey string,
	amount uint64,
) (string, error) {
	cacheKey := fmt.Sprintf(depositCacheKey, userID, idempotencyKey)
	cachedTxId, err := r.cache.Get(ctx, cacheKey).Result()
	// Idempotent: already processed
	if err == nil {
		return cachedTxId, nil
	}

	if err != nil && err != redis.Nil {
		return "", fmt.Errorf("redis get failed: %w", err)
	}

	// it's a new deposit operation
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to begin database tx: %w", err)
	}
	defer tx.Rollback()

	// Hold row-level lock on user wallet
	var walletID string
	query := `SELECT id FROM wallets WHERE user_id = $1 FOR UPDATE`
	err = tx.GetContext(ctx, &walletID, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("wallet not found: %w", domainwallet.ErrWalletNotFound)
		}
		return "", fmt.Errorf(
			"failed to hold row-level lock on wallet: %w, userID: %s",
			err,
			userID,
		)
	}

	// Update balance
	update := `UPDATE wallets SET balance = balance + $1 WHERE id = $2`
	_, err = tx.ExecContext(ctx, update, amount, walletID)
	if err != nil {
		return "", fmt.Errorf("failed to update balance: %w", err)
	}

	// Insert transaction record
	var transactionID string
	insertTxn := `
		INSERT INTO transactions (initiator_wallet_id, type, status, amount, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id
	`
	err = tx.GetContext(
		ctx,
		&transactionID,
		insertTxn,
		walletID,
		domainwallet.Deposit,
		domainwallet.Success,
		amount,
	)
	if err != nil {
		return "", fmt.Errorf("failed to insert transaction record: %w", err)
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf("failed to commit tx: %w", err)
	}

	// Cache idempotency Key
	if err := r.cache.Set(ctx, cacheKey, transactionID, ttl).Err(); err != nil {
		r.logger.Error(
			"failed to cache idempotency key on DepositWallet",
			slog.String("userID", userID),
			slog.String("idempotencyKey", idempotencyKey),
			slog.String("transactionID", transactionID),
		)
	}

	return transactionID, nil
}
