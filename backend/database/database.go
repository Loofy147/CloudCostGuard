package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "user=postgres password=password dbname=pricing sslmode=disable" // Default for local dev
	}
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	createTable()
	fmt.Println("Database connection successful.")
}

func createTable() {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS aws_prices (
		sku TEXT PRIMARY KEY,
		product_json JSONB,
		terms_json JSONB
	);`

	_, err := DB.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create prices table: %v", err)
	}
}
