// Package config contains application configs
package config

import (
	"encoding/json"
	"flag"
	"os"
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
	CertFilepath         string `env:"CERT_FILE_PATH"`
	KeyFilePath          string `env:"KEY_FILE_PATH"`
	ConfigFilePath       string `env:"CONFIG"`
	TrustedSubnet        string `env:"TRUSTED_SUBNET"`
	GrpcAddress          string `env:"GRPC_ADDRESS"`
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
	var cConfigPathFlag string
	var configFilePath string
	var certFilePath string
	var keyFilePath string
	var trustedSubnet string
	var grpcAddress string

	flag.StringVar(&serverHostFlag, "a", "", "address and port to run server")
	flag.StringVar(&baseURLFlag, "b", "", "address and port to run server")
	flag.StringVar(&logLevel, "ll", "DEBUG", "logging level")
	flag.BoolVar(&enableHtttps, "s", false, "enable HTTPS flag")
	flag.StringVar(&urlStoragePath, "f", "", "File storage path for urls")
	flag.StringVar(&databaseDSN, "d", "", "PostgreSQL DSN")
	flag.StringVar(&auditFilePath, "audit-file", "", "Audit Event FileStorage Path")
	flag.StringVar(&auditURL, "audit-url", "", "Audit Event Server URL Path")
	flag.StringVar(&certFilePath, "cfp", "cert.pem", "certificate file path")
	flag.StringVar(&keyFilePath, "cfp", "private.pem", "certificate file path")
	flag.IntVar(&tokenExp, "t", 3, "Token expiration in hours")
	flag.StringVar(&secretKey, "sk", "supersecretkey", "JWT secret key")
	flag.StringVar(&cConfigPathFlag, "c", "", "config file path")
	flag.StringVar(&configFilePath, "config", "", "config file path")
	flag.StringVar(&trustedSubnet, "t", "", "можно передать строковое представление бесклассовой адресации (CIDR)")
	flag.StringVar(&grpcAddress, "ga", "localhost:3200", "default grpc server address")
	flag.Parse()

	err := env.Parse(cfg)
	if err != nil {
		logger.GetLogger().Fatal("Error occured while trying to parse configs.",
			zap.String("Original eror message", err.Error()),
		)
	}

	if cfg.ConfigFilePath == "" {
		cfg.ConfigFilePath = cConfigPathFlag

		if cfg.ConfigFilePath == "" {
			cfg.ConfigFilePath = configFilePath
		}
	}

	var jsonConfig jsonConfig
	if cfg.ConfigFilePath != "" {
		temp, err := loadFromJSON(cfg.ConfigFilePath)
		if err != nil {
			logger.GetLogger().Warn("Warning: failed to load config from JSON file", zap.Error(err))
		} else {
			jsonConfig = *temp
		}
	}

	if cfg.TrustedSubnet == "" {
		cfg.TrustedSubnet = trustedSubnet
		if cfg.TrustedSubnet == "" {
			cfg.TrustedSubnet = jsonConfig.TrustedSubnet
		}
	}

	if cfg.CertFilepath == "" {
		cfg.ConfigFilePath = configFilePath
	}

	if cfg.KeyFilePath == "" {
		cfg.KeyFilePath = keyFilePath
	}

	if cfg.ServerHost == "" {
		cfg.ServerHost = serverHostFlag
		if cfg.ServerHost == "" {
			cfg.ServerHost = jsonConfig.ServerAddress
		}
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = baseURLFlag
		if cfg.BaseURL == "" {
			cfg.BaseURL = jsonConfig.BaseURL
		}
	}

	if cfg.EnableHttps || enableHtttps || jsonConfig.EnableHTTPS {
		cfg.EnableHttps = true
	}

	if cfg.LogLevel == "" {
		cfg.LogLevel = logLevel
	}

	if cfg.URLStoragePath == "" {
		cfg.URLStoragePath = urlStoragePath
		if cfg.URLStoragePath == "" {
			cfg.URLStoragePath = jsonConfig.FileStoragePath
		}
	}

	if cfg.DatabaseDSN == "" {
		logger.GetLogger().Debug("Setting database config from flag value")
		cfg.DatabaseDSN = databaseDSN
		if cfg.DatabaseDSN == "" {
			logger.GetLogger().Debug("Setting database config from flag value")
			cfg.DatabaseDSN = jsonConfig.DatabaseDSN
		}
	}

	if cfg.GrpcAddress == "" {
		cfg.GrpcAddress = grpcAddress
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

func loadFromJSON(path string) (*jsonConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var temp jsonConfig
	if err := json.NewDecoder(file).Decode(&temp); err != nil {
		return nil, err
	}

	return &temp, nil
}
