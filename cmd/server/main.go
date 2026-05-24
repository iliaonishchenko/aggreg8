package main

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/iliaonishchenko/aggreg8"
	"github.com/iliaonishchenko/aggreg8/internal/audit"
	"github.com/iliaonishchenko/aggreg8/internal/config/server"
	cryptopkg "github.com/iliaonishchenko/aggreg8/internal/crypto"
	"github.com/iliaonishchenko/aggreg8/internal/grpcserver"
	"github.com/iliaonishchenko/aggreg8/internal/handler"
	"github.com/iliaonishchenko/aggreg8/internal/logger"
	pb "github.com/iliaonishchenko/aggreg8/internal/proto"
	"github.com/iliaonishchenko/aggreg8/internal/repository"
	"github.com/iliaonishchenko/aggreg8/internal/repository/file"
	"github.com/iliaonishchenko/aggreg8/internal/router"
	"github.com/iliaonishchenko/aggreg8/internal/service"
	"github.com/iliaonishchenko/aggreg8/internal/service/memory"
	"github.com/iliaonishchenko/aggreg8/internal/service/pg"
	"github.com/iliaonishchenko/aggreg8/internal/service/sync"
	"github.com/iliaonishchenko/aggreg8/internal/signature"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func initStorage(ctx context.Context, cfg *server.Config) (service.MetricStorage, *repository.MetricsRepository) {
	var storage service.MetricStorage
	var repo *repository.MetricsRepository

	if cfg.DatabaseDSN != "" {
		db, err := sql.Open("pgx", cfg.DatabaseDSN)
		if err != nil {
			logger.Log.Fatal("error connecting to database", zap.Error(err))
		}
		aggreg8.RunMigrations(db)

		classifier := repository.NewPostgresErrorClassifier()
		repo = repository.NewMetricsRepository(db, classifier)
		storage = pg.NewPostgresStorage(repo)
	} else {
		memStorage := memory.NewMemStorage()
		fileStorage := file.NewFileStorage()

		if *cfg.Restore {
			metrics, err := fileStorage.ReadFromFile(*cfg.FileStoragePath)
			if err != nil {
				logger.Log.Fatal("error restoring metrics from file", zap.Error(err))
			}
			for _, metric := range metrics {
				memStorage.UpdateMetric(metric)
			}
		}

		if *cfg.StoreInterval == 0 {
			storage = sync.NewSyncStorage(memStorage, fileStorage, *cfg.FileStoragePath)
		} else {
			storage = memStorage
			interval := time.Duration(*cfg.StoreInterval) * time.Second
			persister := service.NewPersister(memStorage, fileStorage, *cfg.FileStoragePath, interval)
			go persister.Start(ctx)
		}
	}

	return storage, repo
}

func initNotifier(cfg *server.Config) (audit.Notifier, []func() error) {
	notifier := audit.NewNotifier(10)
	var closers []func() error

	if cfg.AuditFile != "" {
		fileObserver, err := audit.NewFileObserver(cfg.AuditFile)
		if err != nil {
			log.Fatalf("error creating file observer: %v", err)
		}
		notifier.Register(fileObserver)
		closers = append(closers, fileObserver.Close)
	}
	if cfg.AuditURL != "" {
		retryClient := retryablehttp.NewClient()
		notifier.Register(audit.NewHTTPObserver(cfg.AuditURL, retryClient.StandardClient()))
	}
	return notifier, closers
}

func main() {
	fmt.Println("Build version: " + cmp.Or(buildVersion, "N/A"))
	fmt.Println("Build date: " + cmp.Or(buildDate, "N/A"))
	fmt.Println("Build commit: " + cmp.Or(buildCommit, "N/A"))

	defaultServerAddress := "localhost:8080"
	defaultStoreInterval := 300
	defaultFileStoragePath := "./snapshot.json"
	defaultRestore := false
	defaultKey := ""
	defaultGRPCAddress := "localhost:3200"

	cfg, err := server.LoadConfig()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	parseFlags(cfg, defaultServerAddress, defaultStoreInterval, defaultFileStoragePath, defaultRestore, defaultKey)

	if cfg.GRPCAddress == "" {
		cfg.GRPCAddress = defaultGRPCAddress
	}

	if cfg.TrustedSubnet != "" {
		if _, _, err := net.ParseCIDR(cfg.TrustedSubnet); err != nil {
			log.Fatalf("invalid trusted_subnet %q: %v", cfg.TrustedSubnet, err)
		}
	}

	var decrypter router.Decrypter
	if cfg.CryptoKey != "" {
		dec, err := cryptopkg.LoadPrivateKey(cfg.CryptoKey)
		if err != nil {
			log.Fatalf("error loading private key: %v", err)
		}
		decrypter = dec
	}

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatalf("error initializing logger: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	storage, repo := initStorage(ctx, cfg)
	notifier, closers := initNotifier(cfg)
	defer func() {
		for _, closer := range closers {
			closer()
		}
	}()

	if err := run(ctx, *cfg, storage, repo, notifier, decrypter); err != nil {
		logger.Log.Fatal("error starting server", zap.Error(err))
	}
}

func run(ctx context.Context, cfg server.Config, storage service.MetricStorage, repo *repository.MetricsRepository, notifier audit.Notifier, decrypter router.Decrypter) error {

	r := chi.NewRouter()

	recorder := service.NewRecorder(storage, notifier)
	updateHandler := handler.NewUpdateHandler(recorder)
	allHandler := handler.NewAllMetricsHandler(storage)
	getMetricHandler := handler.NewGetMetricHandler(storage)
	pingHandler := handler.NewPingHandler(repo)
	sign := signature.NewSignature(cfg.Key)

	r.Use(logger.WithLogger)
	if subnet := router.WithTrustedSubnet(cfg.TrustedSubnet); subnet != nil {
		r.Use(subnet)
	}
	if decrypter != nil {
		r.Use(router.WithDecryption(decrypter))
	}
	r.Use(router.WithCompression)
	r.Use(router.WithHash(sign))

	r.Mount("/debug", middleware.Profiler())

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
		r.Route("/updates", func(r chi.Router) {
			r.Use(middleware.AllowContentType("application/json"))
			r.Post("/", updateHandler.HandleBatchUpdateJSON)
		})
		r.Route("/update/{type}/{name}/{value}", func(r chi.Router) {
			r.Use(middleware.AllowContentType("text/plain"))
			r.Post("/", updateHandler.HandleUpdate)
		})
	})

	srv := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: r,
	}

	errCh := make(chan error, 2)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	grpcSrv, grpcErr := startGRPCServer(cfg, recorder, errCh)
	if grpcErr != nil {
		return grpcErr
	}

	select {
	case <-ctx.Done():
		logger.Log.Info("Shutting down server...")
		if grpcSrv != nil {
			grpcSrv.GracefulStop()
		}
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		if grpcSrv != nil {
			grpcSrv.GracefulStop()
		}
		return err
	}
}

func startGRPCServer(cfg server.Config, recorder *service.Recorder, errCh chan<- error) (*grpc.Server, error) {
	if cfg.GRPCAddress == "" {
		return nil, nil
	}

	lis, err := net.Listen("tcp", cfg.GRPCAddress)
	if err != nil {
		return nil, fmt.Errorf("ошибка прослушивания gRPC-адреса %q: %w", cfg.GRPCAddress, err)
	}

	var grpcOpts []grpc.ServerOption
	if interceptor := grpcserver.TrustedSubnetInterceptor(cfg.TrustedSubnet); interceptor != nil {
		grpcOpts = append(grpcOpts, grpc.UnaryInterceptor(interceptor))
	}
	grpcSrv := grpc.NewServer(grpcOpts...)
	pb.RegisterMetricsServer(grpcSrv, grpcserver.NewMetricsServer(recorder))

	logger.Log.Info("Запуск gRPC-сервера", zap.String("address", cfg.GRPCAddress))
	go func() {
		defer lis.Close()
		if err := grpcSrv.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	return grpcSrv, nil
}
