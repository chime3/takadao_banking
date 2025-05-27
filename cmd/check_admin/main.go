package main

import (
	"fmt"
	"log"
	"os"

	"github.com/takadao/banking/internal/config"
	"github.com/takadao/banking/internal/models"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := config.NewDatabaseConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Check for admin user
	var adminUser models.User
	result := db.Where("role = ?", "admin").First(&adminUser)
	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			fmt.Println("No admin user found. Creating one...")

			// Create admin user
			adminUser = models.User{
				Email:    os.Getenv("ADMIN_EMAIL"),
				Password: os.Getenv("ADMIN_PASSWORD"),
				Role:     "admin",
			}

			if err := adminUser.HashPassword(); err != nil {
				log.Fatalf("Failed to hash password: %v", err)
			}

			if err := db.Create(&adminUser).Error; err != nil {
				log.Fatalf("Failed to create admin user: %v", err)
			}
			fmt.Println("Admin user created successfully!")
		} else {
			log.Fatalf("Error checking for admin user: %v", result.Error)
		}
	} else {
		fmt.Printf("Admin user found:\nEmail: %s\nRole: %s\nCreated at: %s\n",
			adminUser.Email,
			adminUser.Role,
			adminUser.CreatedAt.Format("2006-01-02 15:04:05"))
	}
}
