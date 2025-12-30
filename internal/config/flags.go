package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerHost string `env:"SERVER_ADDRESS"`
	BaseURL    string `env:"BASE_URL"`
	LogLevel   string `env:"LOG_LEVEL"`
}

func (cfg *Config) ParseFlags() {
	var serverHostFlag string
	var baseURLFlag string
	var logLevel string

	flag.StringVar(&serverHostFlag, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&baseURLFlag, "b", "http://localhost:8080", "address and port to run server")
	flag.StringVar(&logLevel, "c", "INFO", "logging level")
	flag.Parse()

	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
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
}
