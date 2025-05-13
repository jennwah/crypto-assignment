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

type GetWalletResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Balance   string `json:"balance"`
	CreatedAt string `json:"created_at"`
}

// GetWallet godoc
// @Summary      Get wallet
// @Description  Retrieves the wallet details of the current user
// @Tags         Wallet
// @Accept       json
// @Produce      json
// @Param        X-USER-ID header string true "User ID (UUID)"
// @Success      200 {object} GetWalletResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      404 {object} models.ErrorResponse
// @Failure      500 {object} models.ErrorResponse
// @Router       /api/v1/wallet [get]
func (h *Handler) GetWallet(c *gin.Context) {
	userID := c.GetHeader(models.UserIDHeader)
	if err := uuid.Validate(userID); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "invalid user id",
		})
		return
	}

	userWallet, err := h.walletService.GetWallet(c, userID)
	if err != nil {
		if errors.Is(err, domainwallet.ErrWalletNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, models.ErrorResponse{
				Message: domainwallet.ErrWalletNotFound.Error(),
			})
			return
		}

		h.logger.Error("get wallet handler err", slog.Any("error", err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "internal server error",
		})
		return
	}

	c.AbortWithStatusJSON(http.StatusOK, GetWalletResponse{
		ID:        userWallet.ID,
		UserID:    userWallet.UserID,
		Balance:   domainwallet.ConvertFromCentsToDollarsString(userWallet.Balance),
		CreatedAt: userWallet.CreatedAt,
	})
}
