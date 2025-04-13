package main

import (
	"database/sql"
	"flag"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/nibroos/s-erp-api/service/internal/config"
	"github.com/nibroos/s-erp-api/service/internal/database"
)

func main() {
	env := os.Getenv("APP_ENV")
	var dbURL string
	if env == "test" {
		dbURL = config.GetTestDatabaseURL()
	} else {
		dbURL = config.GetDatabaseURL()
	}

	migrationPath := flag.String("path", "../../../internal/database/migrations", "Migration files path")
	flag.Parse()

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := database.MigrateUp(db, *migrationPath); err != nil {
		log.Fatal("Migration failed:", err)
	}

	log.Println("Migrations completed successfully")
}
