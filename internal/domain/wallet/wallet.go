package wallet

import (
	"errors"

	"github.com/shopspring/decimal"
)

var (
	ErrWalletNotFound            = errors.New("wallet not found")
	ErrWalletInsufficientBalance = errors.New("wallet insufficient balance")
)

type (
	TransactionType   string
	TransactionStatus string
)

const (
	Deposit  TransactionType = "deposit"
	Withdraw TransactionType = "withdraw"
	Transfer TransactionType = "transfer"

	Success TransactionStatus = "success"
	Failed  TransactionStatus = "failed"
)

type Wallet struct {
	ID        string `db:"id"`
	UserID    string `db:"user_id"`
	Balance   uint64 `db:"balance"`
	CreatedAt string `db:"created_at"`
}

type Transaction struct {
	ID                    string            `db:"id"`
	InitiatorWalletUserId string            `db:"initiator_wallet_user_id"`
	Type                  TransactionType   `db:"type"`
	Status                TransactionStatus `db:"status"`
	Amount                uint64            `db:"amount"`
	RecipientWalletUserId *string           `db:"recipient_wallet_user_id"`
	CreatedAt             string            `db:"created_at"`
}

// ConvertFromCentsToDollarsString used for displaying dollars amount in string
func ConvertFromCentsToDollarsString(cents uint64) string {
	amount := decimal.NewFromUint64(cents).Div(decimal.NewFromInt(100))
	return amount.StringFixed(2)
}
