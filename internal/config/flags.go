package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerHost string
	BaseURL    string
}

func (cfg *Config) ParseFlags() {
	cfg.ServerHost = os.Getenv("SERVER_ADDRESS")

	if cfg.ServerHost == "" {
		flag.StringVar(&cfg.ServerHost, "a", "localhost:8080", "address and port to run server")
	}

	cfg.BaseURL = os.Getenv("BASE_URL")

	if cfg.BaseURL == "" {
		flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "address and port to run server")
	}

	flag.Parse()
}
