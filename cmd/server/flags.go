package main

import (
	"flag"
	"github.com/iliaonishchenko/aggreg8/internal/config/server"
)

func parseFlags(cfg *server.Config) {
	// регистрируем переменную flagRunAddr
	// как аргумент -a со значением :8080 по умолчанию
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "address and port to run server")
	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()
}
