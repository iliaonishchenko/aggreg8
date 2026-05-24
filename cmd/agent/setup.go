package main

import (
	"log"

	"github.com/iliaonishchenko/aggreg8/internal/agent"
	agentconfig "github.com/iliaonishchenko/aggreg8/internal/config/agent"
	"github.com/iliaonishchenko/aggreg8/internal/crypto"
	"github.com/iliaonishchenko/aggreg8/internal/logger"
	"github.com/iliaonishchenko/aggreg8/internal/signature"
)

func initSender(config *agentconfig.Config) agent.SenderService {
	if config.GRPCAddress != "" {
		return initGRPCSender(config)
	}
	return initHTTPSender(config)
}

func initHTTPSender(config *agentconfig.Config) *agent.Sender {
	var sign agent.DataSignature
	var encrypter agent.DataEncrypter
	errorClassifier := agent.NewAgentErrorClassifier()

	if config.Key != "" {
		sign = signature.NewSignature(config.Key)
	}

	if config.CryptoKey != "" {
		enc, err := crypto.LoadPublicKey(config.CryptoKey)
		if err != nil {
			log.Fatalf("error loading public key: %v", err)
		}
		encrypter = enc
	}

	localIP := agent.LocalOutboundIP(config.ServerAddress)
	if localIP == "" {
		log.Printf("warning: could not determine local outbound IP for X-Real-IP header")
	}

	return agent.NewSender(config.ServerAddress, errorClassifier, sign, encrypter, localIP)
}

func initGRPCSender(config *agentconfig.Config) *agent.GRPCSender {
	localIP := agent.LocalOutboundIP(config.GRPCAddress)
	if localIP == "" {
		log.Printf("warning: не удалось определить локальный IP для метаданных x-real-ip")
	}

	sender, err := agent.NewGRPCSender(config.GRPCAddress, localIP)
	if err != nil {
		log.Fatalf("ошибка создания gRPC-клиента: %v", err)
	}
	return sender
}

func initLogger(config *agentconfig.Config) {
	if err := logger.Initialize(config.LogLevel); err != nil {
		log.Printf("error initializing logger: %v", err)
	}
}
