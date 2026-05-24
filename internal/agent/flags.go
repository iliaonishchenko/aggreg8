package agent

import (
	"flag"
	"log"
	"time"

	agentconfig "github.com/iliaonishchenko/aggreg8/internal/config/agent"
	"github.com/iliaonishchenko/aggreg8/internal/config/external"
)

func ParseFlags(cfg *agentconfig.Config, defaultServerAddress, defaultKey string, defaultReportInterval, defaultPollInterval, defaultRateLimit int) {
	var serverAddress string
	var key string
	var reportInterval int
	var pollInterval int
	var rateLimit int
	var cryptoKey string
	var configFile string
	var grpcAddress string

	flag.StringVar(&serverAddress, "a", "", "address and port to run server")
	flag.IntVar(&reportInterval, "r", 0, "report interval in seconds")
	flag.IntVar(&pollInterval, "p", 0, "poll interval in seconds")
	flag.StringVar(&key, "k", "", "key")
	flag.IntVar(&rateLimit, "l", 0, "rate limit")
	flag.StringVar(&cryptoKey, "crypto-key", "", "path to RSA public key PEM file")
	flag.StringVar(&configFile, "c", "", "path to config file")
	flag.StringVar(&configFile, "config", "", "path to config file")
	flag.StringVar(&grpcAddress, "g", "", "gRPC server address (если задан, агент отправляет метрики по gRPC)")
	flag.StringVar(&grpcAddress, "grpc-address", "", "gRPC server address (если задан, агент отправляет метрики по gRPC)")

	flag.Parse()

	if cfg.ConfigFile == "" {
		cfg.ConfigFile = configFile
	}

	var fileCfg *external.ExternalAgentConfiguration
	if cfg.ConfigFile != "" {
		var err error
		fileCfg, err = external.ConfigurationFromFile[external.ExternalAgentConfiguration](cfg.ConfigFile)
		if err != nil {
			log.Fatalf("error reading config file: %v", err)
		}
	}

	if cfg.ServerAddress == "" {
		cfg.ServerAddress = serverAddress
	}
	if cfg.ServerAddress == "" && fileCfg != nil {
		cfg.ServerAddress = fileCfg.Address
	}
	if cfg.ServerAddress == "" {
		cfg.ServerAddress = defaultServerAddress
	}

	if cfg.CryptoKey == "" {
		cfg.CryptoKey = cryptoKey
	}
	if cfg.CryptoKey == "" && fileCfg != nil {
		cfg.CryptoKey = fileCfg.CryptoKey
	}

	if cfg.Key == "" {
		cfg.Key = key
	}
	if cfg.Key == "" && fileCfg != nil {
		cfg.Key = fileCfg.Key
	}
	if cfg.Key == "" {
		cfg.Key = defaultKey
	}

	if cfg.ReportInterval == 0 && reportInterval != 0 {
		cfg.ReportInterval = reportInterval
	}
	if cfg.ReportInterval == 0 && fileCfg != nil && fileCfg.ReportInterval != "" {
		d, err := time.ParseDuration(fileCfg.ReportInterval)
		if err != nil {
			log.Fatalf("invalid report_interval in config file: %v", err)
		}
		cfg.ReportInterval = int(d.Seconds())
	}
	if cfg.ReportInterval == 0 {
		cfg.ReportInterval = defaultReportInterval
	}

	if cfg.PollInterval == 0 && pollInterval != 0 {
		cfg.PollInterval = pollInterval
	}
	if cfg.PollInterval == 0 && fileCfg != nil && fileCfg.PollInterval != "" {
		d, err := time.ParseDuration(fileCfg.PollInterval)
		if err != nil {
			log.Fatalf("invalid poll_interval in config file: %v", err)
		}
		cfg.PollInterval = int(d.Seconds())
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = defaultPollInterval
	}

	if cfg.RateLimit == 0 && rateLimit != 0 {
		cfg.RateLimit = rateLimit
	}
	if cfg.RateLimit == 0 && fileCfg != nil && fileCfg.RateLimit != 0 {
		cfg.RateLimit = fileCfg.RateLimit
	}
	if cfg.RateLimit == 0 {
		cfg.RateLimit = defaultRateLimit
	}

	if cfg.GRPCAddress == "" {
		cfg.GRPCAddress = grpcAddress
	}
	if cfg.GRPCAddress == "" && fileCfg != nil {
		cfg.GRPCAddress = fileCfg.GRPCAddress
	}
}
