package config

import "time"

type UserJWTConfig struct {
	TokenExp  time.Duration
	SecretKey string `env:"SECRET_KEY"`
}
