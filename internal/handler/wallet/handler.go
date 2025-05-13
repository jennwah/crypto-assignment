package wallet

import (
	"log/slog"

	"github.com/jennwah/crypto-assignment/internal/service/wallet"
)

type Handler struct {
	logger        *slog.Logger
	walletService wallet.IWalletService
}

func New(logger *slog.Logger, walletService wallet.IWalletService) *Handler {
	return &Handler{
		logger:        logger,
		walletService: walletService,
	}
}
