package db

import (
	"flag"

	"github.com/TMWF/url-shortener/internal/logger"
	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"
)

type PostgreSQLConfig struct {
	DatabaseDSN string `env:"DATABASE_DSN"`
}

func (cfg *PostgreSQLConfig) ParseFlags() {
	var databaseDSN string

	flag.StringVar(&databaseDSN, "d", "postgres://user:pass@localhost:5432/db", "PostgreSQL DSN")
	flag.Parse()

	err := env.Parse(cfg)
	if err != nil {
		logger.GetLogger().Fatal("Error occured while trying to parse db configs.",
			zap.String("Original eror message", err.Error()),
		)
	}

	if cfg.DatabaseDSN == "" {
		cfg.DatabaseDSN = databaseDSN
	}
}
