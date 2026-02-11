package repository

import (
	"database/sql"

	"github.com/TMWF/url-shortener/internal/config"
	"github.com/TMWF/url-shortener/migrations"
)

type Storage interface {
	URLStorage
	UserStorage
}

func GetStorage(config *config.Config, db *sql.DB) Storage {
	if db != nil {
		migrations.RunMigrations(db)
		return newDBStorage(db)
	}

	if config.URLStoragePath != "" {
		return NewFileStorage(config)
	}

	return newMemStorage()
}
