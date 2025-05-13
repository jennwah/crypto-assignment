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

type TransferRequest struct {
	RecipientUserID string `json:"recipient_user_id" binding:"required,uuid"`
	Amount          uint64 `json:"amount"            binding:"required,gt=0"`
}

type TransferResponse struct {
	TransactionID string `json:"transaction_id"`
}

// Transfer godoc
// @Summary      Transfer money to another user
// @Description  Transfers money from the initiator user to the recipient user.
// @Tags         Wallet
// @Accept       json
// @Produce      json
// @Param        X-USER-ID header string true "Initiator's User ID (UUID)"
// @Param        X-IDEMPOTENCY-KEY header string true "Idempotency Key (UUID)"
// @Param        transferRequest body TransferRequest true "Transfer request payload"
// @Success      200 {object} TransferResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      404 {object} models.ErrorResponse
// @Failure      422 {object} models.ErrorResponse
// @Failure      500 {object} models.ErrorResponse
// @Router       /api/v1/wallet/transfer [post]
func (h *Handler) Transfer(c *gin.Context) {
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

	var reqBody TransferRequest
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "invalid request",
		})
		return
	}

	txID, err := h.walletService.Transfer(
		c,
		userID,
		reqBody.RecipientUserID,
		idempotencyKey,
		reqBody.Amount,
	)
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

		h.logger.Error("transfer handler err", slog.Any("error", err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "internal server error",
		})
		return
	}

	c.AbortWithStatusJSON(http.StatusOK, TransferResponse{
		TransactionID: txID,
	})
}
