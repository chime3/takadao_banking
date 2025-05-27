package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/takadao/banking/internal/service"
)

// AdminHandler handles admin related requests
type AdminHandler struct {
	adminService *service.AdminService
}

// NewAdminHandler creates a new AdminHandler instance
func NewAdminHandler(adminService *service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

// ListAllTransactions godoc
// @Summary      List all transactions
// @Description  Retrieves a paginated list of all transactions in the system
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "Page number (default: 1)"
// @Param        page_size query int false "Items per page (default: 20)"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /admin/transactions [get]
func (h *AdminHandler) ListAllTransactions(c *gin.Context) {
	// Role check is handled by middleware, but we'll double-check here
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		c.Abort()
		return
	}

	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}

	txs, total, err := h.adminService.ListAllTransactions(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"transactions": txs, "total": total})
}

// GetUserBalanceAtTime godoc
// @Summary      Get user balance at specific time
// @Description  Retrieves a user's balance for a specific currency at a given point in time
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        user_id path string true "User ID"
// @Param        currency query string false "Currency code (default: EUR)"
// @Param        at_time query string true "Timestamp in RFC3339 format"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /admin/users/{user_id}/balance [get]
func (h *AdminHandler) GetUserBalanceAtTime(c *gin.Context) {
	// Role check is handled by middleware, but we'll double-check here
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		c.Abort()
		return
	}

	userIDStr := c.Param("user_id")
	currency := c.DefaultQuery("currency", "EUR")
	atTimeStr := c.Query("at_time")
	if userIDStr == "" || atTimeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and at_time are required"})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}
	atTime, err := time.Parse(time.RFC3339, atTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid at_time format, use RFC3339"})
		return
	}
	balance, err := h.adminService.GetUserBalanceAtTime(userID, currency, atTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user_id": userID, "currency": currency, "balance": balance, "at_time": atTime})
}
