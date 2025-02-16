package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"

	"github.com/ystv/web-auth/infrastructure/db"
	"github.com/ystv/web-auth/infrastructure/db/migrations"
)

func main() {
	// Load environment
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("failed to load global env file")
	} // Load .env file for production
	err = godotenv.Overload(".env.local") // Load .env.local for developing
	if err != nil {
		log.Println("failed to load env file, using global env")
	}

	downOne := flag.Bool("down_one", false, "undo the last migration instead of upgrading - only use for development!")
	flag.Parse()

	if os.Getenv("WAUTH_DB_HOST") == "" {
		log.Fatalf("database host not set")
	}
	dbConnectionString := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		os.Getenv("WAUTH_DB_HOST"),
		os.Getenv("WAUTH_DB_PORT"),
		os.Getenv("WAUTH_DB_USER"),
		os.Getenv("WAUTH_DB_NAME"),
		os.Getenv("WAUTH_DB_SSLMODE"),
		os.Getenv("WAUTH_DB_PASS"),
	)
	database, err := db.NewStore(dbConnectionString)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

	goose.SetBaseFS(migrations.Migrations)
	if err = goose.SetDialect("postgres"); err != nil {
		log.Fatalf("failed to set dialect: %v", err)
	}

	if *downOne {
		if err = goose.Down(database.DB, "."); err != nil {
			log.Fatalf("unable to downgrade: %v", err)
		}
		return
	}

	if err = goose.Up(database.DB, "."); err != nil {
		log.Fatalf("unable to run migrations: %v", err)
	}

	log.Println("migrations ran successfully")
}
