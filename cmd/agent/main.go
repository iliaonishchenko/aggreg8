package main

import (
	"flag"
	"fmt"
	"github.com/iliaonishchenko/aggreg8/internal/agent"
	agentconfig "github.com/iliaonishchenko/aggreg8/internal/config/agent"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	defaultServerAddress := "localhost:8080"
	defaultReportInterval := 10
	defaultPollInterval := 2

	config, err := agentconfig.LoadConfig()

	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	if config.ServerAddress == "" {
		flag.StringVar(&config.ServerAddress, "a", defaultServerAddress, "address and port to run server")
	}
	if config.ReportInterval == 0 {
		flag.IntVar(&config.ReportInterval, "r", defaultReportInterval, "report interval in seconds")
	}
	if config.PollInterval == 0 {
		flag.IntVar(&config.PollInterval, "p", defaultPollInterval, "poll interval in seconds")
	}

	flag.Parse()

	collector := agent.NewCollector()
	sender := agent.NewSender(config.ServerAddress)
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
	for _, metric := range metrics {
		err := sender.SendJSON(metric)
		if err != nil {
			return err
		}
	}
	return nil
}
