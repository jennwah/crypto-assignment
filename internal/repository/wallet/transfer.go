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

const transferCacheKey = `transfer-%s-%s` // transfer-initiatorUserID-idempotencyKey

// Transfer does the following:
// 1. Check from redis cache on key = transfer-{initiatorUserID}-{idempotencyKey}, if exists we just return cached transactionID and nil error
// 2. If not, proceed with transfer amount from initiatorUser wallet to recipientUser wallet
// 3. Cache if successful and return appriopriate errors (insufficient balance)
func (r *Repository) Transfer(
	ctx context.Context,
	initiatorUserID, recipientUserID, idempotencyKey string,
	amount uint64,
) (string, error) {
	cacheKey := fmt.Sprintf(transferCacheKey, initiatorUserID, idempotencyKey)
	cachedTxID, err := r.cache.Get(ctx, cacheKey).Result()
	// Idempotent: already processed such transfer
	if err == nil {
		return cachedTxID, nil
	}

	if err != nil && err != redis.Nil {
		return "", fmt.Errorf("redis get transferCacheKey failed: %w", err)
	}

	// it's a new transfer operation
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to begin database tx: %w", err)
	}
	defer tx.Rollback()

	// Hold row-level lock on both initiator and recipient user wallets
	query := `SELECT id, balance FROM wallets WHERE user_id = $1 FOR UPDATE`

	var dbInitiatorWallet userWallet
	err = tx.GetContext(ctx, &dbInitiatorWallet, query, initiatorUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("wallet not found: %w", domainwallet.ErrWalletNotFound)
		}

		return "", fmt.Errorf(
			"failed to hold row-level lock on wallet: %w, userID: %s",
			err,
			initiatorUserID,
		)
	}

	var dbRecipientWallet userWallet
	err = tx.GetContext(ctx, &dbRecipientWallet, query, recipientUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("wallet not found: %w", domainwallet.ErrWalletNotFound)
		}

		return "", fmt.Errorf(
			"failed to hold row-level lock on wallet: %w, userID: %s",
			err,
			recipientUserID,
		)
	}

	// Check Initiator User wallet balance
	if dbInitiatorWallet.Balance < amount {
		return "", fmt.Errorf(
			"insufficient balance to transfer: %w",
			domainwallet.ErrWalletInsufficientBalance,
		)
	}

	// Update balance for both wallets
	deductQuery := `UPDATE wallets SET balance = balance - $1 WHERE id = $2`
	_, err = tx.ExecContext(ctx, deductQuery, amount, dbInitiatorWallet.ID)
	if err != nil {
		return "", fmt.Errorf("failed to update balance: %w", err)
	}

	addQuery := `UPDATE wallets SET balance = balance + $1 WHERE id = $2`
	_, err = tx.ExecContext(ctx, addQuery, amount, dbRecipientWallet.ID)
	if err != nil {
		return "", fmt.Errorf("failed to update balance: %w", err)
	}

	// Insert transaction record
	var transactionID string
	insertTxn := `
			INSERT INTO transactions (initiator_wallet_id, recipient_wallet_id, type, status, amount, created_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
			RETURNING id
		`
	err = tx.GetContext(
		ctx,
		&transactionID,
		insertTxn,
		dbInitiatorWallet.ID,
		dbRecipientWallet.ID,
		domainwallet.Transfer,
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
			"failed to cache idempotency key on Transfer",
			slog.String("initiatorUserID", initiatorUserID),
			slog.String("recipientUserID", recipientUserID),
			slog.String("idempotencyKey", idempotencyKey),
			slog.String("transactionID", transactionID),
		)
	}

	return transactionID, nil
}
