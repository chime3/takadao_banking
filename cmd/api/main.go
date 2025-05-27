package main

import (
	"log"
	"os"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/takadao/banking/docs"
	"github.com/takadao/banking/internal/config"
	"github.com/takadao/banking/internal/handlers"
	"github.com/takadao/banking/internal/middleware"
	"github.com/takadao/banking/internal/repository"
	"github.com/takadao/banking/internal/routes"
	"github.com/takadao/banking/internal/service"
)

// @title           Banking API
// @version         1.0
// @description     A banking service API in Go using Gin framework.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database connection
	db, err := config.NewDatabaseConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo)
	transactionService := service.NewTransactionService(transactionRepo)

	// Initialize JWT middleware
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}
	authMiddleware := middleware.NewAuthMiddleware(jwtSecret)

	// Setup routes
	router := routes.SetupRouter(
		handlers.NewAuthHandler(userService, authMiddleware),
		handlers.NewUserHandler(userService, transactionRepo),
		handlers.NewTransactionHandler(transactionService),
		authMiddleware,
	)

	// Add Swagger documentation
	docs.SwaggerInfo.BasePath = "/api/v1"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
