package wallet

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	domainwallet "github.com/jennwah/crypto-assignment/internal/domain/wallet"
	"github.com/jennwah/crypto-assignment/internal/handler/models"
)

type GetWalletTransactionsHistoryResponse struct {
	Transactions []GetWalletTransactionResponse `json:"transactions"`
	Page         int                            `json:"page"`
	PageSize     int                            `json:"page_size"`
	Total        int                            `json:"total"`
	TotalPages   int                            `json:"total_pages"`
}

type GetWalletTransactionResponse struct {
	ID                    string  `json:"id"`
	InitiatorWalletUserID string  `json:"initiator_wallet_user_id"`
	Amount                string  `json:"amount"`
	Type                  string  `json:"type"`
	Status                string  `json:"status"`
	RecipientWalletUserID *string `json:"recipient_wallet_user_id,omitempty"`
	CreatedAt             string  `json:"created_at"`
}

// GetTransactions godoc
// @Summary      Get wallet transactions history
// @Description  Retrieves the wallet transactions history of the user
// @Tags         Wallet
// @Accept       json
// @Produce      json
// @Param        X-USER-ID header string true "User ID (UUID)"
// @Param        page query int false "Page number (default is 1)"
// @Param        pageSize query int false "Number of items per page (default is 10)"
// @Success      200 {object} GetWalletTransactionsHistoryResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      404 {object} models.ErrorResponse
// @Failure      500 {object} models.ErrorResponse
// @Router       /api/v1/wallet/transactions [get]
func (h *Handler) GetTransactions(c *gin.Context) {
	userID := c.GetHeader(models.UserIDHeader)
	if err := uuid.Validate(userID); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "invalid user id",
		})
		return
	}

	pageStr := c.DefaultQuery(models.PageQueryParams, "1")
	pageSizeStr := c.DefaultQuery(models.PageSizeQueryParams, "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "invalid page parameter",
		})
		return
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "invalid pageSize parameter",
		})
		return
	}

	offset := (page - 1) * pageSize
	transactions, total, err := h.walletService.GetWalletTransactionsHistory(
		c,
		userID,
		offset,
		pageSize,
	)
	if err != nil {
		if errors.Is(err, domainwallet.ErrWalletNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, models.ErrorResponse{
				Message: domainwallet.ErrWalletNotFound.Error(),
			})
			return
		}

		h.logger.Error("get wallet transactions history handler err", slog.Any("error", err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "internal server error",
		})
		return
	}

	resp := GetWalletTransactionsHistoryResponse{
		Transactions: make([]GetWalletTransactionResponse, 0, len(transactions)),
		Page:         page,
		PageSize:     pageSize,
		Total:        total,
		TotalPages:   (total + pageSize - 1) / pageSize,
	}

	for _, txn := range transactions {
		resp.Transactions = append(resp.Transactions, GetWalletTransactionResponse{
			ID:                    txn.ID,
			InitiatorWalletUserID: txn.InitiatorWalletUserId,
			Amount:                domainwallet.ConvertFromCentsToDollarsString(txn.Amount),
			Type:                  string(txn.Type),
			Status:                string(txn.Status),
			RecipientWalletUserID: txn.RecipientWalletUserId,
			CreatedAt:             txn.CreatedAt,
		})
	}

	c.AbortWithStatusJSON(http.StatusOK, resp)
}
