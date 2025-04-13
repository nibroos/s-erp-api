package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Migration struct {
	Migration  string
	Batch      int
	ExecutedAt time.Time
}

func MigrateUp(db *sql.DB, migrationPath string) error {
	// Ensure migrations table exists
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS migrations (
            id SERIAL PRIMARY KEY,
            migration VARCHAR(255) NOT NULL UNIQUE,
            batch INTEGER NOT NULL,
            executed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        )
    `)
	if err != nil {
		return err
	}

	// Get executed migrations
	executed := make(map[string]bool)
	rows, err := db.Query("SELECT migration FROM migrations")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var migration string
		if err := rows.Scan(&migration); err != nil {
			return err
		}
		executed[migration] = true
	}

	// Get current batch number
	var batch int
	err = db.QueryRow("SELECT COALESCE(MAX(batch), 0) FROM migrations").Scan(&batch)
	if err != nil {
		return err
	}
	batch++

	// Scan migration files
	files, err := os.ReadDir(migrationPath)
	if err != nil {
		return err
	}

	// Execute pending migrations
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".up.sql") {
			continue
		}

		migrationName := strings.TrimSuffix(file.Name(), ".up.sql")
		if executed[migrationName] {
			continue
		}

		fmt.Printf("Running migration: %s\n", file.Name())

		content, err := os.ReadFile(filepath.Join(migrationPath, file.Name()))
		if err != nil {
			return err
		}

		// Start a new transaction
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %v", err)
		}

		// Ensure transaction is rolled back on error
		defer func() {
			if tx != nil {
				tx.Rollback() // Ignore error from failed rollback
			}
		}()

		// Execute migration
		if _, err := tx.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute %s: %v", file.Name(), err)
		}

		// Record migration in migrations table
		if _, err := tx.Exec(
			"INSERT INTO migrations (migration, batch) VALUES ($1, $2)",
			migrationName, batch,
		); err != nil {
			return fmt.Errorf("failed to record migration %s: %v", migrationName, err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %v", err)
		}

		// Set tx to nil so deferred rollback won't run
		tx = nil
	}

	return nil
}

func MigrateDown(db *sql.DB, migrationPath string) error {
	// Get last batch
	var lastBatch int
	err := db.QueryRow("SELECT COALESCE(MAX(batch), 0) FROM migrations").Scan(&lastBatch)
	if err != nil {
		return err
	}

	// Get migrations from last batch
	rows, err := db.Query("SELECT migration FROM migrations WHERE batch = $1 ORDER BY id DESC", lastBatch)
	if err != nil {
		return err
	}
	defer rows.Close()

	var migrations []string
	for rows.Next() {
		var migration string
		if err := rows.Scan(&migration); err != nil {
			return err
		}
		migrations = append(migrations, migration)
	}

	// Execute down migrations
	for _, migration := range migrations {
		downFile := filepath.Join(migrationPath, migration+".down.sql")
		content, err := os.ReadFile(downFile)
		if err != nil {
			return err
		}

		fmt.Printf("Rolling back: %s\n", migration)

		tx, err := db.Begin()
		if err != nil {
			return err
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute %s: %v", downFile, err)
		}

		if _, err := tx.Exec("DELETE FROM migrations WHERE migration = $1", migration); err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}
