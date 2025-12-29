package main

import (
	"flag"
	"github.com/iliaonishchenko/aggreg8/internal/config/server"
)

func parseFlags(cfg *server.Config, defaultServerAddress string, defaultStoreInterval int, defaultFileStoragePath string, defaultRestore bool) {

	if cfg.ServerAddress == "" {
		flag.StringVar(&cfg.ServerAddress, "a", defaultServerAddress, "address and port to run server")
	}

	if cfg.StoreInterval == nil {
		var storeInterval int
		flag.IntVar(&storeInterval, "i", defaultStoreInterval, "interval to store aggregator data")
		cfg.StoreInterval = &storeInterval
	}

	if cfg.FileStoragePath == nil {
		var fileStoragePath string
		flag.StringVar(&fileStoragePath, "f", defaultFileStoragePath, "file path to store aggregator data")
		cfg.FileStoragePath = &fileStoragePath
	}

	if cfg.Restore == nil {
		var restore bool
		flag.BoolVar(&restore, "r", defaultRestore, "restore aggregator data from file on startup")
		cfg.Restore = &restore
	}

	flag.Parse()
}
