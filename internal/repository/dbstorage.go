package repository

import (
	"context"
	"database/sql"

	"github.com/TMWF/url-shortener/internal/logger"
	"github.com/TMWF/url-shortener/internal/util"
	"go.uber.org/zap"
)

type DBPinger interface {
	PingDB() error
}

type dbStorageImpl struct {
	db *sql.DB
}

func newDBStorage(db *sql.DB) *dbStorageImpl {
	return &dbStorageImpl{db: db}
}

func (dbs *dbStorageImpl) PingDB() error {
	return dbs.db.Ping()
}

// GetURL implements [Storage].
func (dbs *dbStorageImpl) GetURL(ctx context.Context, id string) (string, bool) {
	var originalURL string
	row := dbs.db.QueryRowContext(ctx, "SELECT original_url FROM urls WHERE short_url = $1 LIMIT 1", id)
	if err := row.Scan(&originalURL); err != nil {
		logger.GetLogger().Error("Error occured while getting data from database",
			zap.String("original error message", err.Error()),
		)
		return "", false
	}
	return originalURL, true
}

// SaveURL implements [Storage].
func (dbs *dbStorageImpl) SaveURL(ctx context.Context, url string) (string, error) {
	id := util.RandomString(8, util.LatinCharSet)
	_, err := dbs.db.ExecContext(ctx, "INSERT INTO urls (short_url, original_url) VALUES ($1, $2)", id, url)
	if err != nil {
		return "", err
	}
	return id, nil
}
