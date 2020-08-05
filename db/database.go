package db

import (
	"context"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/ystv/web-auth/types"
)

var store Store

// Store interface, all functions offered by the db
type Store interface {
	VerifyUser(ctx context.Context, user *types.User) error
	UpdateUser(ctx context.Context, user *types.User) error
	GetPermissions(ctx context.Context, u *types.User) error
}

// DB is the connection pool
type DB struct {
	*sqlx.DB
}

// NewStore initialises the store
func NewStore() (*DB, error) {
	dbpool, err := sqlx.ConnectContext(context.Background(), "postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}
	return &DB{dbpool}, nil
}
