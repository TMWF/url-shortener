package logger

import (
	"go.uber.org/zap"
)

var logger *zap.Logger

func Initialize(level string) (err error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	logger = zl
	return nil
}

func GetLogger() *zap.Logger {
	if logger == nil {
		Initialize("INFO")
	}
	return logger
}
