// Package config contains application configs
package config

import (
	"flag"
	"time"

	"github.com/TMWF/url-shortener/internal/config/db"
	"github.com/TMWF/url-shortener/internal/logger"
	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"
)

type Config struct {
	ServerHost           string `env:"SERVER_ADDRESS"`
	BaseURL              string `env:"BASE_URL"`
	EnableHttps          bool   `env:"ENABLE_HTTPS"`
	LogLevel             string `env:"LOG_LEVEL"`
	URLStoragePath       string `env:"FILE_STORAGE_PATH"`
	AuditFileStoragePath string `env:"AUDIT_FILE"`
	AuditURL             string `env:"AUDIT_URL"`
	db.PostgreSQLConfig
	UserJWTConfig
}

func InitialiseConfigs() *Config {
	cfg := &Config{}
	var serverHostFlag string
	var baseURLFlag string
	var enableHtttps bool
	var logLevel string
	var urlStoragePath string
	var databaseDSN string
	var tokenExp int
	var secretKey string
	var auditFilePath string
	var auditURL string

	flag.StringVar(&serverHostFlag, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&baseURLFlag, "b", "http://localhost:8080", "address and port to run server")
	flag.StringVar(&logLevel, "c", "DEBUG", "logging level")
	flag.StringVar(&urlStoragePath, "f", "", "File storage path for urls")
	flag.StringVar(&databaseDSN, "d", "", "PostgreSQL DSN")
	flag.StringVar(&auditFilePath, "audit-file", "", "Audit Event FileStorage Path")
	flag.StringVar(&auditURL, "audit-url", "", "Audit Event Server URL Path")
	flag.IntVar(&tokenExp, "t", 3, "Token expiration in hours")
	flag.StringVar(&secretKey, "s", "supersecretkey", "JWT secret key")
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

	if cfg.EnableHttps || enableHtttps {
		cfg.EnableHttps = true
	}

	if cfg.LogLevel == "" {
		cfg.LogLevel = logLevel
	}

	if cfg.URLStoragePath == "" {
		cfg.URLStoragePath = urlStoragePath
	}

	if cfg.DatabaseDSN == "" {
		logger.GetLogger().Debug("Setting database config")
		cfg.DatabaseDSN = databaseDSN
	}

	if cfg.SecretKey == "" {
		cfg.SecretKey = secretKey
	}

	if cfg.AuditFileStoragePath == "" {
		cfg.AuditFileStoragePath = auditFilePath
	}

	if cfg.AuditURL == "" {
		cfg.AuditURL = auditURL
	}

	cfg.TokenExp = 3 * time.Hour

	return cfg
}
