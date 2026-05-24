// Package agent содержит конфигурацию агента сбора метрик.
package agent

import (
	"github.com/caarlos0/env/v11"
)

// Config описывает параметры конфигурации агента, загружаемые из переменных окружения.
type Config struct {
	ServerAddress  string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	Key            string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
	CryptoKey      string `env:"CRYPTO_KEY"`
	LogLevel       string `env:"LOG_LEVEL" envDefault:"info"`
	ConfigFile     string `env:"CONFIG"`
	GRPCAddress    string `env:"GRPC_ADDRESS"`
}

// LoadConfig загружает конфигурацию агента из переменных окружения.
func LoadConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
