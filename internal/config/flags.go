package config

import (
	"flag"

	"github.com/TMWF/url-shortener/internal/config/db"
	"github.com/TMWF/url-shortener/internal/logger"
	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"
)

type Config struct {
	ServerHost     string `env:"SERVER_ADDRESS"`
	BaseURL        string `env:"BASE_URL"`
	LogLevel       string `env:"LOG_LEVEL"`
	URLStoragePath string `env:"FILE_STORAGE_PATH"`
	db.PostgreSQLConfig
}

func InitialiseConfigs() *Config {
	cfg := &Config{}
	var serverHostFlag string
	var baseURLFlag string
	var logLevel string
	var urlStoragePath string
	var databaseDSN string

	flag.StringVar(&serverHostFlag, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&baseURLFlag, "b", "http://localhost:8080", "address and port to run server")
	flag.StringVar(&logLevel, "c", "INFO", "logging level")
	flag.StringVar(&urlStoragePath, "f", "", "File storage path for urls")
	flag.StringVar(&databaseDSN, "d", "", "PostgreSQL DSN")
	flag.Parse()

	err := env.Parse(cfg)
	if err != nil {
		logger.GetLogger().Fatal("Error occured while trying to parse configs.",
			zap.String("Original eror message", err.Error()),
		)
	}

	if cfg.ServerHost == "" {
		cfg.ServerHost = serverHostFlag
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = baseURLFlag
	}

	if cfg.LogLevel == "" {
		cfg.LogLevel = logLevel
	}

	if cfg.URLStoragePath == "" {
		cfg.URLStoragePath = urlStoragePath
	}

	if cfg.DatabaseDSN == "" {
		cfg.DatabaseDSN = databaseDSN
	}

	return cfg
}
