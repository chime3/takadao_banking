package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/takadao/banking/internal/handlers"
	"github.com/takadao/banking/internal/middleware"
)

// SetupRouter configures all the routes for the application
func SetupRouter(
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	transactionHandler *handlers.TransactionHandler,
	authMiddleware *middleware.AuthMiddleware,
) *gin.Engine {
	router := gin.Default()

	// API version 1
	api := router.Group("/api/v1")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			// User auth routes
			userAuth := auth.Group("/user")
			{
				userAuth.POST("/register", authHandler.RegisterUser)
				userAuth.POST("/login", authHandler.UserLogin)
			}

			// Admin auth routes
			adminAuth := auth.Group("/admin")
			{
				adminAuth.POST("/login", authHandler.AdminLogin)
				adminAuth.POST("/register", authMiddleware.RequireAuth(), authHandler.RegisterAdmin)
			}
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(authMiddleware.RequireAuth())
		{
			// User routes
			user := protected.Group("/users")
			{
				user.GET("/me", userHandler.GetMe)
				user.PUT("/me", userHandler.UpdateMe)
				user.GET("/balance", userHandler.GetBalances)
			}

			// Admin routes
			admin := protected.Group("/admin")
			admin.Use(authMiddleware.RequireAdmin())
			{
				admin.GET("/users", userHandler.ListUsers)
				admin.GET("/users/:id", userHandler.GetUser)
				admin.PUT("/users/:id", userHandler.UpdateUser)
				admin.DELETE("/users/:id", userHandler.DeleteUser)
				admin.GET("/transactions", transactionHandler.ListTransactions)
				admin.GET("/transactions/:id", transactionHandler.GetTransaction)
			}

			// Transaction routes (for both users and admins)
			transactions := protected.Group("/transactions")
			{
				transactions.GET("/me", transactionHandler.ListMyTransactions)
				transactions.GET("/me/:id", transactionHandler.GetMyTransaction)
				transactions.POST("/deposit", transactionHandler.Deposit)
				transactions.POST("/withdraw", transactionHandler.Withdraw)
				transactions.POST("/transfer", transactionHandler.Transfer)
			}
		}
	}

	return router
}
