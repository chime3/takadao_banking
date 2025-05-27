package main

import (
	"fmt"
	"log"

	"github.com/takadao/banking/internal/config"
	"github.com/takadao/banking/internal/models"
)

func main() {
	fmt.Println("Starting admin password reset...")

	// Load configuration
	fmt.Println("Loading configuration...")
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Println("Configuration loaded successfully")

	// Connect to database
	fmt.Println("Connecting to database...")
	db, err := config.NewDatabaseConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	fmt.Println("Connected to database successfully")

	// Find admin user
	fmt.Println("Looking for admin user...")
	var adminUser models.User
	result := db.Where("role = ?", "admin").First(&adminUser)
	if result.Error != nil {
		log.Fatalf("Error finding admin user: %v", result.Error)
	}
	fmt.Printf("Found admin user: %s\n", adminUser.Email)

	// Set new password
	newPassword := "admin123" // You can change this to any password you want
	fmt.Printf("Setting new password: %s\n", newPassword)
	adminUser.Password = newPassword

	// Hash the new password
	fmt.Println("Hashing new password...")
	if err := adminUser.HashPassword(); err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}
	fmt.Println("Password hashed successfully")

	// Update the user in database
	fmt.Println("Updating admin user in database...")
	if err := db.Save(&adminUser).Error; err != nil {
		log.Fatalf("Failed to update admin password: %v", err)
	}

	fmt.Printf("\nAdmin password reset successfully!\n")
	fmt.Printf("Email: %s\n", adminUser.Email)
	fmt.Printf("New password: %s\n", newPassword)
}
