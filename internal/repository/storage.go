package repository

import (
	"github.com/TMWF/url-shortener/internal/config"
	"github.com/TMWF/url-shortener/internal/database"
	"github.com/TMWF/url-shortener/internal/logger"
	"github.com/TMWF/url-shortener/migrations"
)

type Storage interface {
	URLStorage
	UserStorage
}

func GetStorage(config *config.Config) Storage {
	if config.DatabaseDSN != "" {
		db, err := database.GetDB(config.DatabaseDSN)
		if err != nil {
			logger.GetLogger().Fatal("Error occured when creating DB connection")
		}
		migrations.RunMigrations(db)
		return newDBStorage(db)
	}

	if config.URLStoragePath != "" {
		return NewFileStorage(config)
	}

	return newMemStorage()
}
