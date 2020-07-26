package db

import (
	"database/sql"
	"log"

	//SQLite3 driver
	_ "github.com/mattn/go-sqlite3"
)

var database Database
var err error

// Database encapsulates database
type Database struct {
	db *sql.DB
}

func (db Database) query(q string, args ...interface{}) (rows *sql.Rows) {
	rows, err := db.db.Query(q, args...)
	if err != nil {
		log.Println(err)
	}
	return rows
}

func init() {
	database.db, err = sql.Open("sqlite3", "./users.db")
	if err != nil {
		log.Fatal(err)
	}
	_, err = database.db.Exec(`CREATE TABLE IF NOT EXISTS "user"
	("user_id" INTEGER PRIMARY KEY AUTOINCREMENT, "username" VARCHAR(50),
	"name" VARCHAR(5), "nickname" VARCHAR(50), "password" VARCHAR(50))`)
	Close()
	database.db, err = sql.Open("sqlite3", "./users.db")
}

// Close function closes this database connection
func Close() {
	database.db.Close()
}

//Query encapsulates running multiple queries which don't do much things
func Query(sql string, args ...interface{}) error {
	SQL, err := database.db.Prepare(sql)
	tx, err := database.db.Begin()
	_, err = tx.Stmt(SQL).Exec(args...)
	if err != nil {
		log.Println("taskQuery: ", err)
		tx.Rollback()
	} else {
		err = tx.Commit()
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return err
}
