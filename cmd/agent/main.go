package main

import (
	"context"
	"github.com/iliaonishchenko/aggreg8/internal/agent"
	agentconfig "github.com/iliaonishchenko/aggreg8/internal/config/agent"
	"log"
)

func main() {

	defaultServerAddress := "localhost:8080"
	defaultReportInterval := 10
	defaultPollInterval := 2
	defaultRateLimit := 5

	config, err := agentconfig.LoadConfig()

	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	agent.ParseFlags(config, defaultServerAddress, "", defaultReportInterval, defaultPollInterval, defaultRateLimit)

	initLogger(config)

	context := context.Background()

	collector := agent.NewCollector()
	sender := initSender(config)
	app := agent.NewAgent(collector, sender, config.PollInterval, config.ReportInterval, config.RateLimit)
	app.Run(context)
}
