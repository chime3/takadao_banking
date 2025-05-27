package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"

	_ "github.com/lib/pq"
)

func main() {
	dbURL := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSL_MODE"),
	)
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	files, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		log.Fatalf("failed to list migration files: %v", err)
	}
	sort.Strings(files)

	for _, file := range files {
		log.Printf("Applying migration: %s", file)
		content, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatalf("failed to read migration %s: %v", file, err)
		}
		if _, err := db.Exec(string(content)); err != nil {
			log.Fatalf("failed to execute migration %s: %v", file, err)
		}
	}
	log.Println("Migrations applied successfully.")
} 