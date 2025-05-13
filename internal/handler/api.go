package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	_ "github.com/jennwah/crypto-assignment/docs"
	"github.com/jennwah/crypto-assignment/internal/handler/wallet"
	walletrepo "github.com/jennwah/crypto-assignment/internal/repository/wallet"
	walletsrv "github.com/jennwah/crypto-assignment/internal/service/wallet"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupHandlers(router *gin.Engine, logger *slog.Logger, db *sqlx.DB, cache *redis.Client) {
	walletRepo := walletrepo.New(db, cache, logger)
	walletService := walletsrv.New(walletRepo)
	walletHandler := wallet.New(logger, walletService)

	// v1
	v1 := router.Group("/api/v1")
	{
		v1Wallet := v1.Group("/wallet")
		{
			v1Wallet.GET("/", walletHandler.GetWallet)
			v1Wallet.GET("/transactions", walletHandler.GetTransactions)
			v1Wallet.POST("/deposit", walletHandler.DepositWallet)
			v1Wallet.POST("/withdraw", walletHandler.WithdrawWallet)
			v1Wallet.POST("/transfer", walletHandler.Transfer)
		}
	}

	// setup Swagger docs
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}
