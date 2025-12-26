package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerHost string `env:"SERVER_ADDRESS"`
	BaseURL    string `env:"BASE_URL"`
}

func (cfg *Config) ParseFlags() {
	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.ServerHost == "" {
		flag.StringVar(&cfg.ServerHost, "a", "localhost:8080", "address and port to run server")
	}

	if cfg.BaseURL == "" {
		flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "address and port to run server")
	}

	flag.Parse()
}
