package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/takadao/banking/internal/auth"
	"github.com/takadao/banking/internal/models"
	"github.com/takadao/banking/internal/repository"
	"github.com/takadao/banking/internal/service"
)

// UserHandler handles user-related requests
type UserHandler struct {
	userService     *service.UserService
	transactionRepo *repository.TransactionRepository
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(userService *service.UserService, transactionRepo *repository.TransactionRepository) *UserHandler {
	return &UserHandler{
		userService:     userService,
		transactionRepo: transactionRepo,
	}
}

type balanceResponse struct {
	Currency string  `json:"currency" example:"EUR"`
	Amount   float64 `json:"amount" example:"1000.50"`
}

type balancesResponse struct {
	Balances []balanceResponse `json:"balances"`
}

// GetBalances godoc
// @Summary      Get user balances
// @Description  Retrieves all balances for the authenticated user
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  balancesResponse
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /users/balance [get]
func (h *UserHandler) GetBalances(c *gin.Context) {
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// For simplicity, get all balances for the user
	var balances []balanceResponse
	if err := h.transactionRepo.GetDB().Raw("SELECT currency, amount FROM balances WHERE user_id = ?", userID).Scan(&balances).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get balances"})
		return
	}

	c.JSON(http.StatusOK, balancesResponse{Balances: balances})
}

// GetMe godoc
// @Summary      Get current user profile
// @Description  Returns the profile of the currently authenticated user
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  models.User
// @Failure      401  {object}  map[string]string
// @Router       /users/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	userID, _ := c.Get("user_id")
	user, err := h.userService.GetByID(uuid.MustParse(userID.(string)))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// UpdateMe godoc
// @Summary      Update current user profile
// @Description  Updates the profile of the currently authenticated user
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body models.User true "User update details"
// @Success      200  {object}  models.User
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /users/me [put]
func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure user can only update their own profile
	user.ID = uuid.MustParse(userID.(string))
	// Prevent role modification
	user.Role = "" // This will be ignored in the update

	updatedUser, err := h.userService.Update(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

// ListUsers godoc
// @Summary      List all users
// @Description  Returns a list of all users (admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   models.User
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /admin/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Role check is handled by middleware, but we'll double-check here
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		c.Abort()
		return
	}

	users, err := h.userService.ListAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

// GetUser godoc
// @Summary      Get user by ID
// @Description  Returns a specific user by ID (admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  models.User
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /admin/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	// Role check is handled by middleware, but we'll double-check here
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		c.Abort()
		return
	}

	id := c.Param("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	user, err := h.userService.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUser godoc
// @Summary      Update user
// @Description  Updates a specific user by ID (admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID"
// @Param        request body models.User true "User update details"
// @Success      200  {object}  models.User
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /admin/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	// Role check is handled by middleware, but we'll double-check here
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		c.Abort()
		return
	}

	id := c.Param("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user.ID = userID
	updatedUser, err := h.userService.Update(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

// DeleteUser godoc
// @Summary      Delete user
// @Description  Deletes a specific user by ID (admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /admin/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// Role check is handled by middleware, but we'll double-check here
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		c.Abort()
		return
	}

	id := c.Param("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	if err := h.userService.Delete(userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}
