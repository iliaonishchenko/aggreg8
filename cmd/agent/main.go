package main

import (
	"flag"
	"github.com/iliaonishchenko/aggreg8/internal/agent"
	agentconfig "github.com/iliaonishchenko/aggreg8/internal/config/agent"
	"github.com/iliaonishchenko/aggreg8/internal/signature"
	"log"
)

func parseFlags(cfg *agentconfig.Config, defaultServerAddress, defaultKey string, defaultReportInterval, defaultPollInterval, defaultRateLimit int) {
	var serverAddress string
	var key string
	var reportInterval int
	var pollInterval int
	var rateLimit int

	flag.StringVar(&serverAddress, "a", defaultServerAddress, "address and port to run server")
	flag.IntVar(&reportInterval, "r", defaultReportInterval, "report interval in seconds")
	flag.IntVar(&pollInterval, "p", defaultPollInterval, "poll interval in seconds")
	flag.StringVar(&key, "k", defaultKey, "key")
	flag.IntVar(&rateLimit, "l", defaultRateLimit, "rate limit")

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
}

func main() {

	defaultServerAddress := "localhost:8080"
	defaultReportInterval := 10
	defaultPollInterval := 2
	defaultRateLimit := 5

	config, err := agentconfig.LoadConfig()

	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	parseFlags(config, defaultServerAddress, "", defaultReportInterval, defaultPollInterval, defaultRateLimit)

	collector := agent.NewCollector()
	sender := initSender(config)
	app := agent.NewAgent(collector, sender, config.PollInterval, config.ReportInterval, config.RateLimit)
	app.Run()
}

func initSender(config *agentconfig.Config) *agent.Sender {
	var sender *agent.Sender
	var sign *signature.Signature
	errorClassifier := agent.NewAgentErrorClassifier()

	if config.Key != "" {
		sign = signature.NewSignature(config.Key)
		sender = agent.NewSender(config.ServerAddress, errorClassifier, sign)
	} else {
		sender = agent.NewSender(config.ServerAddress, errorClassifier, nil)
	}

	return sender
}
