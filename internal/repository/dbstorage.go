package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/TMWF/url-shortener/internal/logger"
	"github.com/TMWF/url-shortener/internal/model"
	"github.com/TMWF/url-shortener/internal/util"
	"go.uber.org/zap"
)

var ErrConflict = errors.New("url already exists")
var ErrURLDeleted = errors.New("url was deleted")
var ErrURLNotFound = errors.New("original url not found in storage")
var ErrUserIDAbsent = errors.New("UserID unexpectedly not found in context")

type DBStorage interface {
	Storage
	PingDB() error
	DeleteUserURLs([]model.DeleteUserURLsJobModel) error
}

type dbStorageImpl struct {
	db            *sql.DB
	urlDeleteJobs chan model.DeleteUserURLsJobModel
}

func newDBStorage(db *sql.DB) *dbStorageImpl {
	return &dbStorageImpl{db: db, urlDeleteJobs: make(chan model.DeleteUserURLsJobModel, 1024)}
}

func (dbs *dbStorageImpl) PingDB() error {
	return dbs.db.Ping()
}

// GetURL implements [Storage].
func (dbs *dbStorageImpl) GetURL(ctx context.Context, id string) (string, error) {
	var originalURL string
	var isDeleted bool
	row := dbs.db.QueryRowContext(ctx, "SELECT original_url, is_deleted FROM urls WHERE short_url = $1 LIMIT 1", id)
	if err := row.Scan(&originalURL, &isDeleted); err != nil {
		logger.GetLogger().Error("Error occured while getting data from database",
			zap.String("original error message", err.Error()),
		)
		return "", err
	}

	if isDeleted {
		return "", ErrURLDeleted
	}
	return originalURL, nil
}

// SaveURL implements [Storage].
func (dbs *dbStorageImpl) SaveURL(ctx context.Context, url string) (string, error) {
	userID, ok := ctx.Value(util.UserID).(int)

	if !ok || userID < 1 {
		logger.GetLogger().Error("UserID unexpectedly not found in context")

		return "", ErrUserIDAbsent
	}

	id := util.RandomString(8, util.LatinCharSet)

	tx, err := dbs.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, "INSERT INTO urls (short_url, original_url) VALUES ($1, $2) ON CONFLICT (original_url) DO NOTHING", id, url)
	if err != nil {
		return "", err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return "", err
	}

	var urlID int
	if rowsAffected == 0 {
		logger.GetLogger().Debug("Database conflict detected")
		row := tx.QueryRowContext(ctx, "SELECT id, short_url FROM urls WHERE original_url = $1 LIMIT 1", url)
		var shortURLFromDB string

		if err := row.Scan(&urlID, &shortURLFromDB); err != nil {
			logger.GetLogger().Error("Error occured while getting data from database",
				zap.String("original error message", err.Error()),
			)
			return "", err
		}

		// _, err := tx.ExecContext(ctx, "INSERT INTO urls_users (url_id, user_id) VALUES ($1, $2)", urlID, userID)
		// if err != nil {
		// 	logger.GetLogger().Debug("Error while inserting into urls_users")
		// 	return "", err
		// }

		if err = tx.Commit(); err != nil {
			logger.GetLogger().Debug("Error while commiting transaction")
			return "", err
		}
		logger.GetLogger().Debug("Returning ErrConflict from dbStorage")
		return shortURLFromDB, ErrConflict
	}

	row := tx.QueryRowContext(ctx, "SELECT id FROM urls WHERE original_url = $1 LIMIT 1", url)
	if err := row.Scan(&urlID); err != nil {
		logger.GetLogger().Error("Error occured while getting data from database",
			zap.String("original error message", err.Error()),
		)
		return "", err
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO urls_users (url_id, user_id) VALUES ($1, $2)", urlID, userID)
	if err != nil {
		return "", nil
	}

	if err = tx.Commit(); err != nil {
		return "", err
	}
	logger.GetLogger().Debug("Successfully saved url in database without error", zap.String("url", url))
	return id, nil
}

func (dbs *dbStorageImpl) SaveBatchURL(ctx context.Context, urlBatch []model.URLBatchRequestDto) ([]string, error) {
	userID, ok := ctx.Value(util.UserID).(int)

	if !ok || userID < 1 {
		logger.GetLogger().Error("UserID unexpectedly not found in context")

		return nil, ErrUserIDAbsent
	}

	tx, err := dbs.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	insertUrlsStatement, err := tx.PrepareContext(ctx, "INSERT INTO urls (short_url, original_url) VALUES ($1, $2) RETURNING ID")
	if err != nil {
		return nil, err
	}
	defer insertUrlsStatement.Close()

	insertUsersUrlsStatement, err := tx.PrepareContext(ctx, "INSERT INTO urls_users (url_id, user_id) VALUES ($1, $2)")
	if err != nil {
		return nil, err
	}
	defer insertUrlsStatement.Close()

	result := make([]string, 0, len(urlBatch))

	for _, urlModel := range urlBatch {
		id := util.RandomString(8, util.LatinCharSet)
		var urlDatabaseID int
		row := insertUrlsStatement.QueryRowContext(ctx, id, urlModel.OriginalURL)
		if err := row.Scan(&urlDatabaseID); err != nil {
			logger.GetLogger().Error("Error occured while getting data from database",
				zap.String("original error message", err.Error()),
			)
			return nil, err
		}

		_, err := insertUsersUrlsStatement.ExecContext(ctx, urlDatabaseID, userID)
		if err != nil {
			return nil, err
		}

		result = append(result, id)
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return result, nil
}

func (dbs *dbStorageImpl) GetUsersURLs(ctx context.Context) ([]model.GetUserURLsResponseModel, error) {
	userID, ok := ctx.Value(util.UserID).(int)
	logger.GetLogger().Debug("User ID is", zap.Int("userID", userID))

	if !ok || userID < 0 {
		logger.GetLogger().Error("UserID unexpectedly not found in context")
		return nil, ErrUserIDAbsent
	}

	rows, err := dbs.db.QueryContext(
		ctx,
		"SELECT short_url, original_url FROM urls "+
			"JOIN urls_users as uu ON urls.id = uu.url_id "+
			"WHERE uu.user_id = $1",
		userID,
	)

	if err != nil {
		logger.GetLogger().Error("Error occured while getting data from database",
			zap.String("original error message", err.Error()),
		)
		return nil, err
	}
	defer rows.Close()

	result := make([]model.GetUserURLsResponseModel, 0)

	for rows.Next() {
		var model model.GetUserURLsResponseModel
		err = rows.Scan(&model.ShortURL, &model.OriginalURL)
		if err != nil {
			return nil, err
		}

		result = append(result, model)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (dbs *dbStorageImpl) DeleteUserURLs(jobModels []model.DeleteUserURLsJobModel) error {
	var values []string
	var args []any
	var base int
	for _, jobModel := range jobModels {
		userID := jobModel.UserID
		for _, urlID := range jobModel.URLIDs {
			// PostgreSQL требует шаблоны в формате ($1, $2) для каждой вставки
			params := fmt.Sprintf("($%d::text, $%d::int)", base+1, base+2)
			values = append(values, params)
			args = append(args, urlID, userID)
			base += 2
		}
	}

	query := `UPDATE urls AS u1
		SET u1.is_deleted = TRUE
		FROM (VALUES 
    	` + strings.Join(values, ",") + `
		) AS v(urlId, userId)
		WHERE u1.id = (SELECT id FROM urls u 
		JOIN urls_users uu 
		ON u.id = uu.url_id 
		WHERE u.short_url = v.urlId AND uu.user_id = v.userId);`

	logger.GetLogger().Debug("DELETE Query is: " + query)
	logger.GetLogger().Debug("Arguments for Query are: ", zap.Any("args", args))
	logger.GetLogger().Debug("Args slice length is", zap.Int("argsLength", len(args)))
	_, err := dbs.db.Exec(query, args...)

	return err
}

func (dbs *dbStorageImpl) SaveUser(ctx context.Context) (int, error) {
	var userID int
	err := dbs.db.QueryRowContext(ctx, "INSERT INTO users DEFAULT VALUES RETURNING ID").Scan(&userID)
	if err != nil {
		logger.GetLogger().Error(
			"Error while getting user ID from database",
			zap.String("original error message", err.Error()),
		)

		return -1, err
	}

	return userID, nil
}
