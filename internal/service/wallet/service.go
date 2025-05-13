package wallet

import "github.com/jennwah/crypto-assignment/internal/repository/wallet"

type Service struct {
	walletRepo wallet.IWalletRepository
}

func New(walletRepo wallet.IWalletRepository) *Service {
	return &Service{
		walletRepo: walletRepo,
	}
}
