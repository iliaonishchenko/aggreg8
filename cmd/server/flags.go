package main

import (
	"flag"
	"github.com/iliaonishchenko/aggreg8/internal/config/server"
)

func parseFlags(cfg *server.Config, defaultServerAddress string) {

	if cfg.ServerAddress == "" {
		flag.StringVar(&cfg.ServerAddress, "a", defaultServerAddress, "address and port to run server")
	}

	flag.Parse()
}
