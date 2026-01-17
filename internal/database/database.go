package database

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var dbConn *sql.DB

func GetDB(dsn string) (*sql.DB, error) {
	if dbConn != nil {
		return dbConn, nil
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	dbConn = db
	return db, nil
}
