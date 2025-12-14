package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/iliaonishchenko/aggreg8/internal/config/server"
	"github.com/iliaonishchenko/aggreg8/internal/handler"
	"github.com/iliaonishchenko/aggreg8/internal/logger"
	"github.com/iliaonishchenko/aggreg8/internal/service"
	"go.uber.org/zap"
	"log"
	"net/http"
)

func main() {
	defaultServerAddress := "localhost:8080"
	cfg, err := server.LoadConfig()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	parseFlags(cfg, defaultServerAddress)

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatalf("error initializing logger: %v", err)
	}

	if err := run(*cfg); err != nil {
		logger.Log.Fatal("error starting server", zap.Error(err))
	}
}

func run(cfg server.Config) error {

	r := chi.NewRouter()

	memStorage := service.NewMemStorage()
	updateHandler := handler.NewUpdateHandler(memStorage)
	allHandler := handler.NewAllMetricsHandler(memStorage)
	getMetricHandler := handler.NewGetMetricHandler(memStorage)

	r.Use(logger.WithLogger)

	r.Route("/", func(r chi.Router) {
		r.Get("/", allHandler.HandleAll)
		r.Route("/value/{type}/{name}", func(r chi.Router) {
			r.Get("/", getMetricHandler.HandleGetMetric)
		})
		r.Route("/value", func(r chi.Router) {
			r.Use(middleware.AllowContentType("application/json"))
			r.Post("/", getMetricHandler.HandleGetMetricJSON)
		})
		r.Route("/update", func(r chi.Router) {
			r.Use(middleware.AllowContentType("application/json"))
			r.Post("/", updateHandler.HandleUpdateJSON)
		})
		r.Route("/update/{type}/{name}/{value}", func(r chi.Router) {
			r.Use(middleware.AllowContentType("text/plain"))
			r.Post("/", updateHandler.HandleUpdate)
		})
	})

	return http.ListenAndServe(cfg.ServerAddress, r)
}
