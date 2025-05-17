package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"

	"github.com/ystv/web-auth/infrastructure/db"
	"github.com/ystv/web-auth/infrastructure/db/migrations"
	"github.com/ystv/web-auth/utils"
)

func main() {
	//nolint:reassign
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	logger := utils.NewLogger(zlog.With().
		Str("service", "web-auth").
		Logger(), utils.DefaultSkipper)

	// Load environment
	err := godotenv.Load(".env")
	if err != nil {
		logger.Warn(nil, "failed to load global env file")
	} // Load .env file for production
	err = godotenv.Overload(".env.local") // Load .env.local for developing
	if err != nil {
		logger.Warn(nil, "failed to load env file, using global env")
	}

	downOne := flag.Bool("down_one", false, "undo the last migration instead of upgrading - only use for development!")
	flag.Parse()

	host := os.Getenv("WAUTH_DB_HOST")

	if host == "" {
		logger.Fatal(nil, errors.New("database host not set"))
	}
	dbConnectionString := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		host,
		os.Getenv("WAUTH_DB_PORT"),
		os.Getenv("WAUTH_DB_USER"),
		os.Getenv("WAUTH_DB_NAME"),
		os.Getenv("WAUTH_DB_SSLMODE"),
		os.Getenv("WAUTH_DB_PASS"),
	)
	database := db.NewStore(dbConnectionString, host, logger)

	goose.SetBaseFS(migrations.Migrations)
	if err = goose.SetDialect("postgres"); err != nil {
		logger.Fatal(nil, errors.Errorf("failed to set dialect: %v", err))
	}

	if *downOne {
		if err = goose.Down(database.DB, "."); err != nil {
			logger.Fatal(nil, errors.Errorf("unable to downgrade: %v", err))
		}
		return
	}

	if err = goose.Up(database.DB, "."); err != nil {
		logger.Fatal(nil, errors.Errorf("unable to run migrations: %v", err))
	}

	logger.Info(nil, "migrations ran successfully")
}
