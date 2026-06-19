// Package db contains database config
package db

type PostgreSQLConfig struct {
	DatabaseDSN string `env:"DATABASE_DSN"`
}
