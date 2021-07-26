package db

import (
	"context"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // postgres driver
)

// NewStore initialises the store
func NewStore(dataSourceName string) (*sqlx.DB, error) {
	dbpool, err := sqlx.ConnectContext(context.Background(), "postgres", dataSourceName)
	return dbpool, err
}
