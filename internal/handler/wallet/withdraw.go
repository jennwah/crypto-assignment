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

type WithdrawWalletRequest struct {
	Amount uint64 `json:"amount" binding:"required"`
}

type WithdrawWalletResponse struct {
	TransactionID string `json:"transaction_id"`
}

// WithdrawWallet godoc
// @Summary      Withdraw from wallet
// @Description  Withdraw a specific amount (in cents) from the user's wallet
// @Tags         Wallet
// @Accept       json
// @Produce      json
// @Param        X-USER-ID header string true "User ID (UUID)"
// @Param        X-IDEMPOTENCY-KEY header string true "Idempotency Key (UUID)"
// @Param        request body WithdrawWalletRequest true "Withdraw amount in cents"
// @Success      200 {object} WithdrawWalletResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      404 {object} models.ErrorResponse
// @Failure      422 {object} models.ErrorResponse
// @Failure      500 {object} models.ErrorResponse
// @Router       /api/v1/wallet/withdraw [post]
func (h *Handler) WithdrawWallet(c *gin.Context) {
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

	var reqBody WithdrawWalletRequest
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "invalid request",
		})
		return
	}

	transactionID, err := h.walletService.WithdrawWallet(c, userID, idempotencyKey, reqBody.Amount)
	if err != nil {
		if errors.Is(err, domainwallet.ErrWalletNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, models.ErrorResponse{
				Message: domainwallet.ErrWalletNotFound.Error(),
			})
			return
		}

		if errors.Is(err, domainwallet.ErrWalletInsufficientBalance) {
			c.AbortWithStatusJSON(http.StatusUnprocessableEntity, models.ErrorResponse{
				Message: domainwallet.ErrWalletInsufficientBalance.Error(),
			})
			return
		}

		h.logger.Error("withdraw wallet handler err", slog.Any("error", err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "internal server error",
		})
		return
	}

	c.AbortWithStatusJSON(http.StatusOK, WithdrawWalletResponse{
		TransactionID: transactionID,
	})
}
