package db

import (
	"context"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // postgres driver
	"github.com/pkg/errors"

	"github.com/ystv/web-auth/utils"
)

// NewStore initialises the store
func NewStore(dataSourceName string, host string, logger *utils.Logger) *sqlx.DB {
	db, err := sqlx.ConnectContext(context.Background(), "postgres", dataSourceName)
	if err != nil {
		logger.Fatal(nil, errors.Errorf("db failed: %+v", err))
	}

	logger.Debug(nil, "connected to db: %s", host)

	return db
}
