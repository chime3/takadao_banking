package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/takadao/banking/internal/middleware"
	"github.com/takadao/banking/internal/service"
)

// AuthHandler handles authentication related requests
type AuthHandler struct {
	userService    *service.UserService
	authMiddleware *middleware.AuthMiddleware
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(userService *service.UserService, authMiddleware *middleware.AuthMiddleware) *AuthHandler {
	return &AuthHandler{
		userService:    userService,
		authMiddleware: authMiddleware,
	}
}

// Common request/response types
type loginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}

type loginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	Role  string `json:"role" example:"user"`
}

// User registration request
type userRegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
}

// Admin registration request (requires admin token)
type adminRegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"admin@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"admin123"`
}

// UserLogin godoc
// @Summary      Login as user
// @Description  Authenticates a regular user and returns a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body loginRequest true "Login credentials"
// @Success      200  {object}  loginResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /auth/user/login [post]
func (h *AuthHandler) UserLogin(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.Authenticate(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Check if user is not an admin
	if user.Role == "admin" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "please use admin login endpoint"})
		return
	}

	token, err := h.authMiddleware.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	c.JSON(http.StatusOK, loginResponse{Token: token, Role: user.Role})
}

// AdminLogin godoc
// @Summary      Login as admin
// @Description  Authenticates an admin user and returns a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body loginRequest true "Admin login credentials"
// @Success      200  {object}  loginResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /auth/admin/login [post]
func (h *AuthHandler) AdminLogin(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.Authenticate(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Check if user is an admin
	if user.Role != "admin" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "admin access required"})
		return
	}

	token, err := h.authMiddleware.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	c.JSON(http.StatusOK, loginResponse{Token: token, Role: user.Role})
}

// RegisterUser godoc
// @Summary      Register new user
// @Description  Creates a new regular user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body userRegisterRequest true "Registration details"
// @Success      201  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Router       /auth/user/register [post]
func (h *AuthHandler) RegisterUser(c *gin.Context) {
	var req userRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.userService.Register(req.Email, req.Password, "user")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "user registered successfully"})
}

// RegisterAdmin godoc
// @Summary      Register new admin
// @Description  Creates a new admin account (requires admin token)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body adminRegisterRequest true "Admin registration details"
// @Success      201  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /auth/admin/register [post]
func (h *AuthHandler) RegisterAdmin(c *gin.Context) {
	// Verify that the requester is an admin
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	var req adminRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.userService.Register(req.Email, req.Password, "admin")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "admin registered successfully"})
}
