package main

import (
	"flag"
	"log"
	"time"

	"github.com/iliaonishchenko/aggreg8/internal/config/external"
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
	var configFile string
	var trustedSubnet string
	var grpcAddress string

	flag.StringVar(&serverAddress, "a", "", "address and port to run server")
	flag.IntVar(&storeInterval, "i", 0, "interval to store aggregator data (seconds)")
	flag.StringVar(&fileStoragePath, "f", "", "file path to store aggregator data")
	flag.BoolVar(&restore, "r", false, "restore aggregator data from file on startup")
	flag.StringVar(&databaseDSN, "d", "", "database DSN")
	flag.StringVar(&key, "k", "", "key for signature")
	flag.StringVar(&auditFile, "audit-file", "", "file to store audit data")
	flag.StringVar(&auditURL, "audit-url", "", "url to send audit data")
	flag.StringVar(&cryptoKey, "crypto-key", "", "path to RSA private key PEM file")
	flag.StringVar(&configFile, "c", "", "path to config file")
	flag.StringVar(&configFile, "config", "", "path to config file")
	flag.StringVar(&trustedSubnet, "t", "", "trusted subnet in CIDR notation")
	flag.StringVar(&grpcAddress, "g", "", "address and port to run gRPC server")
	flag.StringVar(&grpcAddress, "grpc-address", "", "address and port to run gRPC server")

	flag.Parse()

	if cfg.ConfigFile == "" {
		cfg.ConfigFile = configFile
	}

	var fileCfg *external.ExternalServerConfiguration
	if cfg.ConfigFile != "" {
		var err error
		fileCfg, err = external.ConfigurationFromFile[external.ExternalServerConfiguration](cfg.ConfigFile)
		if err != nil {
			log.Fatalf("error reading config file: %v", err)
		}
	}

	if cfg.ServerAddress == "" {
		cfg.ServerAddress = serverAddress
	}
	if cfg.ServerAddress == "" && fileCfg != nil {
		cfg.ServerAddress = fileCfg.Address
	}
	if cfg.ServerAddress == "" {
		cfg.ServerAddress = defaultServerAddress
	}

	if cfg.DatabaseDSN == "" {
		cfg.DatabaseDSN = databaseDSN
	}
	if cfg.DatabaseDSN == "" && fileCfg != nil {
		cfg.DatabaseDSN = fileCfg.DatabaseDSN
	}

	if cfg.CryptoKey == "" {
		cfg.CryptoKey = cryptoKey
	}
	if cfg.CryptoKey == "" && fileCfg != nil {
		cfg.CryptoKey = fileCfg.CryptoKey
	}

	if cfg.Key == "" {
		cfg.Key = key
	}
	if cfg.Key == "" {
		cfg.Key = defaultKey
	}
	if cfg.AuditFile == "" {
		cfg.AuditFile = auditFile
	}
	if cfg.AuditURL == "" {
		cfg.AuditURL = auditURL
	}

	if cfg.TrustedSubnet == "" {
		cfg.TrustedSubnet = trustedSubnet
	}
	if cfg.TrustedSubnet == "" && fileCfg != nil {
		cfg.TrustedSubnet = fileCfg.TrustedSubnet
	}

	if cfg.GRPCAddress == "" {
		cfg.GRPCAddress = grpcAddress
	}
	if cfg.GRPCAddress == "" && fileCfg != nil {
		cfg.GRPCAddress = fileCfg.GRPCAddress
	}

	if cfg.StoreInterval == nil && storeInterval != 0 {
		cfg.StoreInterval = &storeInterval
	}
	if cfg.StoreInterval == nil && fileCfg != nil && fileCfg.StoreInterval != "" {
		d, err := time.ParseDuration(fileCfg.StoreInterval)
		if err != nil {
			log.Fatalf("invalid store_interval in config file: %v", err)
		}
		sec := int(d.Seconds())
		cfg.StoreInterval = &sec
	}
	if cfg.StoreInterval == nil {
		cfg.StoreInterval = &defaultStoreInterval
	}

	if cfg.FileStoragePath == nil && fileStoragePath != "" {
		cfg.FileStoragePath = &fileStoragePath
	}
	if cfg.FileStoragePath == nil && fileCfg != nil && fileCfg.StoreFileName != "" {
		cfg.FileStoragePath = &fileCfg.StoreFileName
	}
	if cfg.FileStoragePath == nil {
		cfg.FileStoragePath = &defaultFileStoragePath
	}

	explicit := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { explicit[f.Name] = true })
	if cfg.Restore == nil && explicit["r"] {
		cfg.Restore = &restore
	}
	if cfg.Restore == nil && fileCfg != nil && fileCfg.Restore != nil {
		cfg.Restore = fileCfg.Restore
	}
	if cfg.Restore == nil {
		cfg.Restore = &defaultRestore
	}
}
