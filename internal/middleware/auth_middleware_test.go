package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/takadao/banking/internal/models"
)

func setupTestRouter(middleware *AuthMiddleware) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Test endpoint that requires auth
	router.GET("/protected", middleware.RequireAuth(), func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"role":    role,
		})
	})

	// Test endpoint that requires admin
	router.GET("/admin", middleware.RequireAdmin(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})

	return router
}

func TestRequireAuth(t *testing.T) {
	middleware := NewAuthMiddleware("test-secret")
	router := setupTestRouter(middleware)

	// Create a test user
	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  "user",
	}

	// Generate a valid token
	token, err := middleware.GenerateToken(user)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "No Authorization Header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   map[string]interface{}{"error": "authorization header is required"},
		},
		{
			name:           "Invalid Authorization Format",
			authHeader:     "InvalidFormat",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   map[string]interface{}{"error": "invalid authorization header format"},
		},
		{
			name:           "Invalid Token",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   map[string]interface{}{"error": "invalid token"},
		},
		{
			name:           "Valid Token",
			authHeader:     "Bearer " + token,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"user_id": user.ID.String(),
				"role":    user.Role,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Perform the request
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse response body
			var response map[string]interface{}
			if tt.expectedStatus == http.StatusOK {
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody["user_id"], response["user_id"])
				assert.Equal(t, tt.expectedBody["role"], response["role"])
			} else {
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody["error"], response["error"])
			}
		})
	}
}

func TestRequireAdmin(t *testing.T) {
	middleware := NewAuthMiddleware("test-secret")
	router := setupTestRouter(middleware)

	// Create test users
	adminUser := &models.User{
		ID:    uuid.New(),
		Email: "admin@example.com",
		Role:  "admin",
	}
	regularUser := &models.User{
		ID:    uuid.New(),
		Email: "user@example.com",
		Role:  "user",
	}

	// Generate tokens
	adminToken, err := middleware.GenerateToken(adminUser)
	assert.NoError(t, err)
	userToken, err := middleware.GenerateToken(regularUser)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "No Authorization Header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   map[string]interface{}{"error": "authorization header is required"},
		},
		{
			name:           "Regular User Token",
			authHeader:     "Bearer " + userToken,
			expectedStatus: http.StatusForbidden,
			expectedBody:   map[string]interface{}{"error": "admin access required"},
		},
		{
			name:           "Admin User Token",
			authHeader:     "Bearer " + adminToken,
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]interface{}{"message": "admin access granted"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/admin", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Perform the request
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse response body
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody["error"], response["error"])
			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, tt.expectedBody["message"], response["message"])
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	middleware := NewAuthMiddleware("test-secret")

	tests := []struct {
		name    string
		user    *models.User
		wantErr bool
	}{
		{
			name: "Valid User",
			user: &models.User{
				ID:    uuid.New(),
				Email: "test@example.com",
				Role:  "user",
			},
			wantErr: false,
		},
		{
			name: "Admin User",
			user: &models.User{
				ID:    uuid.New(),
				Email: "admin@example.com",
				Role:  "admin",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := middleware.GenerateToken(tt.user)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, token)

			// Verify token can be parsed and contains correct claims
			claims := jwt.MapClaims{}
			parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte("test-secret"), nil
			})
			assert.NoError(t, err)
			assert.True(t, parsedToken.Valid)
			assert.Equal(t, tt.user.ID.String(), claims["user_id"])
			assert.Equal(t, tt.user.Email, claims["email"])
			assert.Equal(t, tt.user.Role, claims["role"])
		})
	}
}
