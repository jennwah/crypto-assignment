package wallet

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	domainwallet "github.com/jennwah/crypto-assignment/internal/domain/wallet"
	"github.com/jennwah/crypto-assignment/internal/handler/models"
)

type DepositWalletRequest struct {
	Amount uint64 `json:"amount" binding:"required"`
}

type DepositWalletResponse struct {
	TransactionID string `json:"transaction_id"`
}

// DepositWallet godoc
// @Summary      Deposit to wallet
// @Description  Deposit a specific amount (in cents) to the user's wallet
// @Tags         Wallet
// @Accept       json
// @Produce      json
// @Param        X-USER-ID header string true "User ID (UUID)"
// @Param        X-IDEMPOTENCY-KEY header string true "Idempotency Key (UUID)"
// @Param        request body DepositWalletRequest true "Deposit amount in cents"
// @Success      200 {object} DepositWalletResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      404 {object} models.ErrorResponse
// @Failure      500 {object} models.ErrorResponse
// @Router       /api/v1/wallet/deposit [post]
func (h *Handler) DepositWallet(c *gin.Context) {
	userID := c.GetHeader(models.UserIDHeader)
	if err := uuid.Validate(userID); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "invalid user id",
		})
		return
	}

	idempotencyKey := c.GetHeader(models.IdempotencyKeyHeader)
	if err := uuid.Validate(idempotencyKey); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "invalid idempotency key",
		})
		return
	}

	var reqBody DepositWalletRequest
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "invalid request",
		})
		return
	}

	transactionID, err := h.walletService.DepositWallet(c, userID, idempotencyKey, reqBody.Amount)
	if err != nil {
		if errors.Is(err, domainwallet.ErrWalletNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, models.ErrorResponse{
				Message: domainwallet.ErrWalletNotFound.Error(),
			})
			return
		}

		h.logger.Error("deposit wallet handler err", slog.Any("error", err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "internal server error",
		})
		return
	}

	c.AbortWithStatusJSON(http.StatusOK, DepositWalletResponse{
		TransactionID: transactionID,
	})
}
