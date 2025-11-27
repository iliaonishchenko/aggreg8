package main

import (
	"fmt"
	"github.com/iliaonishchenko/aggreg8/internal/agent"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	collector := agent.NewCollector()
	sender := agent.NewSender("http://localhost:8080")
	collectTicker := time.NewTicker(time.Second * 2)
	defer collectTicker.Stop()
	sendTicker := time.NewTicker(time.Second * 10)
	defer sendTicker.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-collectTicker.C:
			collector.Collect()
		case <-sendTicker.C:
			metrics := collector.GetMetrics()
			sender.Send(metrics)
		case <-sigChan:
			fmt.Println("Shutting down...")
			return
		}
	}
}
