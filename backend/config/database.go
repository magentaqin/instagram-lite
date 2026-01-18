package config

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "instagram.db")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Test connection
	if err = DB.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Database connected successfully")

	// Run migrations
	if err = runMigrations(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	log.Println("Database migrated successfully")
}

func runMigrations() error {
	migrationFile := filepath.Join("migrations", "001_init.sql")
	sqlBytes, err := os.ReadFile(migrationFile)
	if err != nil {
		return err
	}

	_, err = DB.Exec(string(sqlBytes))
	return err
}
