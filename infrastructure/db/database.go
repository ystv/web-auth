package db

import (
	"context"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // postgres driver
)

// DB is the connection pool
type DB struct {
	*sqlx.DB
}

// NewStore initialises the store
func NewStore(dataSourceName string) (*DB, error) {
	dbpool, err := sqlx.ConnectContext(context.Background(), "postgres", dataSourceName)
	return &DB{dbpool}, err
}
