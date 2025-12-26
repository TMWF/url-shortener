package config

import "flag"

type Config struct {
	ServerHost string
	BaseURL    string
}

func (cfg *Config) ParseFlags() {
	flag.StringVar(&cfg.ServerHost, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "address and port to run server")
	flag.Parse()
}
