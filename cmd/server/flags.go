package main

import (
	"flag"
	"github.com/iliaonishchenko/aggreg8/internal/config/server"
)

func parseFlags(cfg *server.Config, defaultServerAddress string, defaultStoreInterval int, defaultFileStoragePath string, defaultRestore bool, defaultKey string) {
	var serverAddress string
	var storeInterval int
	var fileStoragePath string
	var restore bool
	var databaseDSN string
	var key string
	var auditFile string
	var auditURL string
	var cryptoKey string

	flag.StringVar(&serverAddress, "a", defaultServerAddress, "address and port to run server")
	flag.IntVar(&storeInterval, "i", defaultStoreInterval, "interval to store aggregator data")
	flag.StringVar(&fileStoragePath, "f", defaultFileStoragePath, "file path to store aggregator data")
	flag.BoolVar(&restore, "r", defaultRestore, "restore aggregator data from file on startup")
	flag.StringVar(&databaseDSN, "d", "", "database DSN")
	flag.StringVar(&key, "k", "", "key for signature")
	flag.StringVar(&auditFile, "audit-file", "", "file to store audit data")
	flag.StringVar(&auditURL, "audit-url", "", "url to send audit data")
	flag.StringVar(&cryptoKey, "crypto-key", "", "path to RSA private key PEM file")

	flag.Parse()

	if cfg.ServerAddress == "" {
		cfg.ServerAddress = serverAddress
	}

	if cfg.StoreInterval == nil {
		cfg.StoreInterval = &storeInterval
	}

	if cfg.FileStoragePath == nil {
		cfg.FileStoragePath = &fileStoragePath
	}

	if cfg.Restore == nil {
		cfg.Restore = &restore
	}

	if cfg.DatabaseDSN == "" {
		cfg.DatabaseDSN = databaseDSN
	}
	if cfg.Key == "" {
		cfg.Key = key
	}
	if cfg.AuditFile == "" {
		cfg.AuditFile = auditFile
	}
	if cfg.AuditURL == "" {
		cfg.AuditURL = auditURL
	}
	if cfg.CryptoKey == "" {
		cfg.CryptoKey = cryptoKey
	}
}
