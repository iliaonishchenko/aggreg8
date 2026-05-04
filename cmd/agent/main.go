package main

import (
	"cmp"
	"context"
	"fmt"
	"github.com/iliaonishchenko/aggreg8/internal/agent"
	agentconfig "github.com/iliaonishchenko/aggreg8/internal/config/agent"
	"log"
	"os/signal"
	"syscall"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	fmt.Println("Build version: " + cmp.Or(buildVersion, "N/A"))
	fmt.Println("Build date: " + cmp.Or(buildDate, "N/A"))
	fmt.Println("Build commit: " + cmp.Or(buildCommit, "N/A"))

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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	collector := agent.NewCollector()
	sender := initSender(config)
	app := agent.NewAgent(collector, sender, config.PollInterval, config.ReportInterval, config.RateLimit)
	app.Run(ctx)
}
