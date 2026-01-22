package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/TMWF/url-shortener/internal/logger"
	"github.com/TMWF/url-shortener/internal/model"
	"github.com/TMWF/url-shortener/internal/util"
	"go.uber.org/zap"
)

var ErrConflict = errors.New("url already exists")

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
	result, err := dbs.db.ExecContext(ctx, "INSERT INTO urls (short_url, original_url) VALUES ($1, $2) ON CONFLICT (original_url) DO NOTHING", id, url)
	if err != nil {
		return "", err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return "", err
	}

	if rowsAffected == 0 {
		row := dbs.db.QueryRowContext(ctx, "SELECT original_url FROM urls WHERE original_url = $1 LIMIT 1", url)
		var shortUrlFromDB string
		if err := row.Scan(&shortUrlFromDB); err != nil {
			logger.GetLogger().Error("Error occured while getting data from database",
				zap.String("original error message", err.Error()),
			)
			return "", err
		}

		return shortUrlFromDB, ErrConflict
	}

	return id, nil
}

func (dbs *dbStorageImpl) SaveBatchURL(ctx context.Context, urlBatch []model.URLBatchRequestDto) ([]model.URLBatchResponseDto, error) {
	tx, err := dbs.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO urls (short_url, original_url) VALUES ($1, $2)")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	result := make([]model.URLBatchResponseDto, 0, len(urlBatch))

	for _, urlModel := range urlBatch {
		id := util.RandomString(8, util.LatinCharSet)

		_, err := stmt.ExecContext(ctx, id, urlModel.OriginalURL)
		if err != nil {
			return nil, err
		}
		responseModel := model.URLBatchResponseDto{CorrelationID: urlModel.CorrelationID, ShortURL: id}
		result = append(result, responseModel)
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return result, nil
}
