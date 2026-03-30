// Package server содержит конфигурацию сервера метрик.
package server

import "github.com/caarlos0/env/v11"

// Config описывает параметры конфигурации сервера, загружаемые из переменных окружения.
type Config struct {
	ServerAddress   string  `env:"ADDRESS"`
	LogLevel        string  `env:"LOG_LEVEL" envDefault:"info"`
	StoreInterval   *int    `env:"STORE_INTERVAL"`
	FileStoragePath *string `env:"FILE_STORAGE_PATH"`
	Restore         *bool   `env:"RESTORE"`
	DatabaseDSN     string  `env:"DATABASE_DSN"`
	Key             string  `env:"KEY"`
	AuditFile       string  `env:"AUDIT_FILE"`
	AuditURL        string  `env:"AUDIT_URL"`
}

// LoadConfig загружает конфигурацию сервера из переменных окружения.
func LoadConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
