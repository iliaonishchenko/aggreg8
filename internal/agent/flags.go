package agent

import (
	"flag"
	agentconfig "github.com/iliaonishchenko/aggreg8/internal/config/agent"
)

func ParseFlags(cfg *agentconfig.Config, defaultServerAddress, defaultKey string, defaultReportInterval, defaultPollInterval, defaultRateLimit int) {
	var serverAddress string
	var key string
	var reportInterval int
	var pollInterval int
	var rateLimit int
	var cryptoKey string

	flag.StringVar(&serverAddress, "a", defaultServerAddress, "address and port to run server")
	flag.IntVar(&reportInterval, "r", defaultReportInterval, "report interval in seconds")
	flag.IntVar(&pollInterval, "p", defaultPollInterval, "poll interval in seconds")
	flag.StringVar(&key, "k", defaultKey, "key")
	flag.IntVar(&rateLimit, "l", defaultRateLimit, "rate limit")
	flag.StringVar(&cryptoKey, "crypto-key", "", "path to RSA public key PEM file")

	flag.Parse()

	if cfg.ServerAddress == "" {
		cfg.ServerAddress = serverAddress
	}
	if cfg.ReportInterval == 0 {
		cfg.ReportInterval = reportInterval
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = pollInterval
	}
	if cfg.Key == "" {
		cfg.Key = key
	}
	if cfg.RateLimit == 0 {
		cfg.RateLimit = rateLimit
	}
	if cfg.CryptoKey == "" {
		cfg.CryptoKey = cryptoKey
	}
}
