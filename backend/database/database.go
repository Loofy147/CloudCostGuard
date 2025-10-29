package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// DB is the global database connection pool.
var DB *sql.DB

// InitDB initializes the connection to the PostgreSQL database.
// It uses the DATABASE_URL environment variable for the connection string.
// If DATABASE_URL is not set, it falls back to a default local development connection string.
// It also creates the aws_prices table if it doesn't already exist.
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
		terms_json JSONB,
		last_updated TIMESTAMPTZ NOT NULL
	);`

	_, err := DB.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create prices table: %v", err)
	}
}
