package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/takadao/banking/internal/auth"
	"github.com/takadao/banking/internal/service"
)

// TransactionHandler handles transaction-related requests
type TransactionHandler struct {
	transactionService *service.TransactionService
}

// NewTransactionHandler creates a new TransactionHandler instance
func NewTransactionHandler(transactionService *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
	}
}

type depositWithdrawRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0" example:"100.50"`
	Currency    string  `json:"currency" binding:"required,len=3" example:"EUR"`
	Description string  `json:"description" example:"Initial deposit"`
}

type transferRequest struct {
	RecipientID string  `json:"recipient_id" binding:"required,uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	Amount      float64 `json:"amount" binding:"required,gt=0" example:"50.25"`
	Currency    string  `json:"currency" binding:"required,len=3" example:"EUR"`
	Description string  `json:"description" example:"Payment for services"`
}

// ListMyTransactions godoc
// @Summary      List user's transactions
// @Description  Returns a list of transactions for the authenticated user
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   models.Transaction
// @Failure      401  {object}  map[string]string
// @Router       /transactions/me [get]
func (h *TransactionHandler) ListMyTransactions(c *gin.Context) {
	userID, _ := c.Get("user_id")
	transactions, err := h.transactionService.ListByUserID(uuid.MustParse(userID.(string)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list transactions"})
		return
	}
	c.JSON(http.StatusOK, transactions)
}

// GetMyTransaction godoc
// @Summary      Get user's transaction
// @Description  Returns a specific transaction for the authenticated user
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Transaction ID"
// @Success      200  {object}  models.Transaction
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /transactions/me/{id} [get]
func (h *TransactionHandler) GetMyTransaction(c *gin.Context) {
	id := c.Param("id")
	transactionID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction ID"})
		return
	}

	userID, _ := c.Get("user_id")
	transaction, err := h.transactionService.GetByIDAndUserID(transactionID, uuid.MustParse(userID.(string)))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

// ListTransactions godoc
// @Summary      List all transactions
// @Description  Returns a list of all transactions (admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   models.Transaction
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /admin/transactions [get]
func (h *TransactionHandler) ListTransactions(c *gin.Context) {
	// Role check is handled by middleware, but we'll double-check here
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		c.Abort()
		return
	}

	transactions, err := h.transactionService.ListAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list transactions"})
		return
	}
	c.JSON(http.StatusOK, transactions)
}

// GetTransaction godoc
// @Summary      Get transaction
// @Description  Returns a specific transaction by ID (admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Transaction ID"
// @Success      200  {object}  models.Transaction
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /admin/transactions/{id} [get]
func (h *TransactionHandler) GetTransaction(c *gin.Context) {
	// Role check is handled by middleware, but we'll double-check here
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		c.Abort()
		return
	}

	id := c.Param("id")
	transactionID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction ID"})
		return
	}

	transaction, err := h.transactionService.GetByID(transactionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

// Deposit godoc
// @Summary      Make a deposit
// @Description  Deposits money into the user's account
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body depositWithdrawRequest true "Deposit details"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /transactions/deposit [post]
func (h *TransactionHandler) Deposit(c *gin.Context) {
	var req depositWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if err := h.transactionService.Deposit(userID, req.Amount, req.Currency, req.Description); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deposit successful"})
}

// Withdraw godoc
// @Summary      Make a withdrawal
// @Description  Withdraws money from the user's account
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body depositWithdrawRequest true "Withdrawal details"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /transactions/withdraw [post]
func (h *TransactionHandler) Withdraw(c *gin.Context) {
	var req depositWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if err := h.transactionService.Withdraw(userID, req.Amount, req.Currency, req.Description); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "withdrawal successful"})
}

// Transfer godoc
// @Summary      Transfer money
// @Description  Transfers money to another user
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body transferRequest true "Transfer details"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /transactions/transfer [post]
func (h *TransactionHandler) Transfer(c *gin.Context) {
	var req transferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	recipientID, err := uuid.Parse(req.RecipientID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid recipient_id"})
		return
	}
	if err := h.transactionService.Transfer(userID, recipientID, req.Amount, req.Currency, req.Description); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "transfer successful"})
}
