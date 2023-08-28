package db

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // postgres driver
)

// NewStore initialises the store
func NewStore(dataSourceName string, host string, debug bool) *sqlx.DB {
	db, err := sqlx.ConnectContext(context.Background(), "postgres", dataSourceName)
	if err != nil {
		if debug {
			log.Printf("db failed: %+v", err)
		} else {
			log.Fatalf("db failed: %+v", err)
		}
	} else {
		log.Printf("connected to db: %s", host)
	}
	return db
}
