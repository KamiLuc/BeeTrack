// Command resetdb drops every table in the database's public schema and
// re-runs migrations, leaving a clean, empty schema. Requires -yes to run,
// since this permanently destroys all data.
//
// Usage (from backend/, with `docker compose up` already running):
//
//	go run ./cmd/resetdb -yes
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/beetrack/backend/internal/config"
	"github.com/beetrack/backend/internal/database"
	"github.com/beetrack/backend/migrations"
	"github.com/joho/godotenv"
)

func main() {
	confirm := flag.Bool("yes", false, "confirm that you want to permanently delete all data")
	flag.Parse()

	if !*confirm {
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/resetdb -yes")
		fmt.Fprintln(os.Stderr, "this permanently deletes all data in the database; pass -yes to confirm")
		os.Exit(1)
	}

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	if _, err := sqlDB.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;"); err != nil {
		log.Fatalf("drop schema: %v", err)
	}
	log.Println("dropped public schema")

	if err := database.Migrate(db, migrations.FS); err != nil {
		log.Fatal(err)
	}
	log.Println("migrations applied — database is clean")
}
