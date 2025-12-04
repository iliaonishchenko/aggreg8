package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/iliaonishchenko/aggreg8/internal/config/server"
	"github.com/iliaonishchenko/aggreg8/internal/handler"
	"github.com/iliaonishchenko/aggreg8/internal/service"
	"log"
	"net/http"
)

func main() {
	cfg, err := server.LoadConfig()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	parseFlags(cfg)

	if err := run(*cfg); err != nil {
		log.Fatalf("error running server: %v", err)
	}
}

func run(cfg server.Config) error {
	fmt.Println("starting server")

	r := chi.NewRouter()

	memStorage := service.NewMemStorage()
	updateHandler := handler.NewUpdateHandler(memStorage)
	allHandler := handler.NewAllMetricsHandler(memStorage)
	getMetricHandler := handler.NewGetMetricHandler(memStorage)

	r.Route("/", func(r chi.Router) {
		r.Get("/", allHandler.HandleAll)
		r.Route("/value/{type}/{name}", func(r chi.Router) {
			r.Get("/", getMetricHandler.HandleGetMetric)
		})
		r.Route("/update/{type}/{name}/{value}", func(r chi.Router) {
			r.Use(middleware.AllowContentType("text/plain"))
			r.Post("/", updateHandler.HandleUpdate)
		})
	})

	return http.ListenAndServe(cfg.ServerAddress, r)
}
