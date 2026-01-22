package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/iliaonishchenko/aggreg8"
	"github.com/iliaonishchenko/aggreg8/internal/config/server"
	"github.com/iliaonishchenko/aggreg8/internal/handler"
	"github.com/iliaonishchenko/aggreg8/internal/logger"
	"github.com/iliaonishchenko/aggreg8/internal/repository"
	"github.com/iliaonishchenko/aggreg8/internal/router"
	"github.com/iliaonishchenko/aggreg8/internal/service"
	"github.com/iliaonishchenko/aggreg8/internal/service/sync"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"log"
	"net/http"
	"time"
)

func main() {
	defaultServerAddress := "localhost:8080"
	defaultStoreInterval := 300
	defaultFileStoragePath := "./snapshot.json"
	defaultRestore := false
	cfg, err := server.LoadConfig()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	parseFlags(cfg, defaultServerAddress, defaultStoreInterval, defaultFileStoragePath, defaultRestore)

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatalf("error initializing logger: %v", err)
	}

	if cfg.DatabaseDSN != "" {
		db, err := sql.Open("pgx", cfg.DatabaseDSN)
		if err != nil {
			logger.Log.Fatal("error connecting to database", zap.Error(err))
		}
		defer db.Close()
		aggreg8.RunMigrations(db)
	}

	memStorage := service.NewMemStorage()
	fileStorage := repository.NewFileStorage()

	if *cfg.Restore {
		metrics, err := fileStorage.ReadFromFile(*cfg.FileStoragePath)
		if err != nil {
			logger.Log.Fatal("error restoring metrics from file", zap.Error(err))
		}
		for _, metric := range metrics {
			memStorage.UpdateMetric(metric)
		}
	}

	var storage service.MetricStorage
	var cancelFunc context.CancelFunc

	if *cfg.StoreInterval == 0 {
		storage = sync.NewSyncStorage(memStorage, fileStorage, *cfg.FileStoragePath)
	} else {
		storage = memStorage
		interval := time.Duration(*cfg.StoreInterval) * time.Second
		persister := service.NewPersister(memStorage, fileStorage, *cfg.FileStoragePath, interval)
		ctx, cancel := context.WithCancel(context.Background())
		cancelFunc = cancel
		go persister.Start(ctx)
	}

	if err := run(*cfg, storage); err != nil {
		if cancelFunc != nil {
			cancelFunc()
		}
		logger.Log.Fatal("error starting server", zap.Error(err))
	}
}

func run(cfg server.Config, memStorage service.MetricStorage) error {

	r := chi.NewRouter()

	db, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}
	database := repository.NewDatabase(db)

	updateHandler := handler.NewUpdateHandler(memStorage)
	allHandler := handler.NewAllMetricsHandler(memStorage)
	getMetricHandler := handler.NewGetMetricHandler(memStorage)
	pingHandler := handler.NewPingHandler(database)

	r.Use(logger.WithLogger)
	r.Use(router.WithCompression)

	r.Route("/", func(r chi.Router) {
		r.Get("/ping", pingHandler.HandlePing)
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
