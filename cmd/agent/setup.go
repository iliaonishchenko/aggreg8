package main

import (
	"github.com/iliaonishchenko/aggreg8/internal/agent"
	agentconfig "github.com/iliaonishchenko/aggreg8/internal/config/agent"
	"github.com/iliaonishchenko/aggreg8/internal/logger"
	"github.com/iliaonishchenko/aggreg8/internal/signature"
	"log"
)

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

func initLogger(config *agentconfig.Config) {
	if err := logger.Initialize(config.LogLevel); err != nil {
		log.Printf("error initializing logger: %v", err)
	}
}
