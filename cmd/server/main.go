package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/iliaonishchenko/aggreg8/internal/handler"
	"github.com/iliaonishchenko/aggreg8/internal/service"
	"net/http"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
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

	return http.ListenAndServe(`:8080`, r)
}
