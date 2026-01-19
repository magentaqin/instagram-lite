package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

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
	// 0) Create migrations tracking table
	if _, err := DB.Exec(`
CREATE TABLE IF NOT EXISTS schema_migrations (
  filename TEXT PRIMARY KEY,
  applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	dir := "migrations"

	// 1) Read all *.sql under migrations/
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".sql") {
			files = append(files, name)
		}
	}

	// Ensure stable execution order (001_..., 002_..., ...)
	sort.Strings(files)

	// 2) Apply migrations
	for _, name := range files {
		applied, err := isMigrationApplied(DB, name)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		path := filepath.Join(dir, name)
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		tx, err := DB.Begin()
		if err != nil {
			return fmt.Errorf("begin tx for %s: %w", name, err)
		}

		if _, err := tx.Exec(string(sqlBytes)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("exec migration %s: %w", name, err)
		}

		if _, err := tx.Exec(
			`INSERT INTO schema_migrations(filename, applied_at) VALUES(?, ?)`,
			name,
			time.Now().UTC().Format(time.RFC3339Nano),
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", name, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}
	}

	return nil
}

func isMigrationApplied(db *sql.DB, filename string) (bool, error) {
	var one int
	err := db.QueryRow(`SELECT 1 FROM schema_migrations WHERE filename = ? LIMIT 1`, filename).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check migration %s: %w", filename, err)
	}
	return true, nil
}