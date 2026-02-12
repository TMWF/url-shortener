package repository

import (
	"database/sql"

	"github.com/TMWF/url-shortener/internal/config"
	"github.com/TMWF/url-shortener/internal/logger"
	"github.com/TMWF/url-shortener/migrations"
	"go.uber.org/zap"
)

type Storage interface {
	URLStorage
	UserStorage
}

func GetStorage(config *config.Config, db *sql.DB) Storage {
	if db != nil {
		logger.GetLogger().Debug("Setting dbstorage")
		if err := migrations.RunMigrations(db); err != nil {
			logger.GetLogger().Fatal("Failed to run db migrations",
				zap.String("original error message", err.Error()))
		}
		return newDBStorage(db)
	}

	if config.URLStoragePath != "" {
		logger.GetLogger().Debug("Setting filestorage")
		return NewFileStorage(config)
	}

	logger.GetLogger().Debug("Setting memstorage")
	return newMemStorage()
}
