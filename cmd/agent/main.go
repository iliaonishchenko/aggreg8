package main

import (
	"flag"
	"fmt"
	"github.com/iliaonishchenko/aggreg8/internal/agent"
	agentconfig "github.com/iliaonishchenko/aggreg8/internal/config/agent"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/iliaonishchenko/aggreg8/internal/signature"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func parseFlags(cfg *agentconfig.Config, defaultServerAddress, defaultKey string, defaultReportInterval, defaultPollInterval int) {
	var serverAddress string
	var key string
	var reportInterval int
	var pollInterval int

	flag.StringVar(&serverAddress, "a", defaultServerAddress, "address and port to run server")
	flag.IntVar(&reportInterval, "r", defaultReportInterval, "report interval in seconds")
	flag.IntVar(&pollInterval, "p", defaultPollInterval, "poll interval in seconds")
	flag.StringVar(&key, "k", defaultKey, "key")

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
}

func main() {

	defaultServerAddress := "localhost:8080"
	defaultReportInterval := 10
	defaultPollInterval := 2

	config, err := agentconfig.LoadConfig()

	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	parseFlags(config, defaultServerAddress, "", defaultReportInterval, defaultPollInterval)

	collector := agent.NewCollector()
	sender := initSender(config)
	collectTicker := time.NewTicker(time.Duration(config.PollInterval) * time.Second)
	defer collectTicker.Stop()
	sendTicker := time.NewTicker(time.Duration(config.ReportInterval) * time.Second)
	defer sendTicker.Stop()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-collectTicker.C:
			collector.Collect()
		case <-sendTicker.C:
			metrics := collector.GetMetrics()
			if err = sendMetrics(metrics, sender); err != nil {
				log.Printf("error sending metrics: %v", err)
			}
		case <-sigChan:
			fmt.Println("Shutting down...")
			return
		}
	}
}

func sendMetrics(metrics []*models.Metrics, sender *agent.Sender) error {
	return sender.SendJSONWithRetries(metrics...)
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
