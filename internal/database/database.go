// Package database contains method responsible for setting up a connection to database
package database

import (
	"database/sql"
	"errors"
	"time"

	"github.com/TMWF/url-shortener/internal/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var dbConn *sql.DB
var ErrEmptyDSN = errors.New("DSN string is empty")

func GetDB(dsn string) (*sql.DB, error) {
	if dbConn != nil {
		logger.GetLogger().Debug("Returning existing db")
		return dbConn, nil
	}

	if dsn == "" {
		logger.GetLogger().Debug("DSN string is empty")
		return nil, ErrEmptyDSN
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.GetLogger().Debug("Error occured while opening sql/db")
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	dbConn = db
	logger.GetLogger().Debug("Successfuly created sql/db")
	return db, nil
}
