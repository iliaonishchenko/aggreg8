package main

import (
	cryptopkg "github.com/iliaonishchenko/aggreg8/internal/crypto"
	"github.com/iliaonishchenko/aggreg8/internal/agent"
	agentconfig "github.com/iliaonishchenko/aggreg8/internal/config/agent"
	"github.com/iliaonishchenko/aggreg8/internal/logger"
	"github.com/iliaonishchenko/aggreg8/internal/signature"
	"log"
)

func initSender(config *agentconfig.Config) *agent.Sender {
	var sign agent.DataSignature
	var encrypter agent.DataEncrypter
	errorClassifier := agent.NewAgentErrorClassifier()

	if config.Key != "" {
		sign = signature.NewSignature(config.Key)
	}

	if config.CryptoKey != "" {
		enc, err := cryptopkg.LoadPublicKey(config.CryptoKey)
		if err != nil {
			log.Fatalf("error loading public key: %v", err)
		}
		encrypter = enc
	}

	return agent.NewSender(config.ServerAddress, errorClassifier, sign, encrypter)
}

func initLogger(config *agentconfig.Config) {
	if err := logger.Initialize(config.LogLevel); err != nil {
		log.Printf("error initializing logger: %v", err)
	}
}
