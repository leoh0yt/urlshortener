package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eqweqr/urlshortener/config"
	"github.com/eqweqr/urlshortener/handler"
	"github.com/eqweqr/urlshortener/middleware"
	"github.com/eqweqr/urlshortener/service"
	"github.com/eqweqr/urlshortener/storage"
	"github.com/eqweqr/urlshortener/storage/memory"
	"github.com/eqweqr/urlshortener/storage/postgres"
)

func main() {
	var (
		storageFlag = flag.String("storage", "memory", "Storage type: memory or postgres")
		portFlag    = flag.String("port", "8080", "Storage type: memory or postgresql")
	)
	flag.Parse()

	cfg := config.LoadConfig()

	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.Logs.Level)); err != nil {
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	var store storage.Storage
	var err error

	switch *storageFlag {
	case "postgres":
		store, err = postgres.NewPostgresStorage(cfg.DB)
		if err != nil {
			logger.Error("Failed to initialize PostgreStorage", "error", err)
			os.Exit(1)
		}
		defer store.Close()
		logger.Info("Using PostgreStorage")
	case "memory":
		store = memory.NewStorage()
		logger.Info("Using MemoryStorage")
	default:
		logger.Error("Undefined storage type", "type", fmt.Sprintf("'%s'", *storageFlag))
		os.Exit(1)
	}

	urlService := service.NewUrlService(store)
	urlHandler := handler.NewHandler(urlService, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/shorten", urlHandler.Shorten)
	mux.HandleFunc("/", urlHandler.Resolve)

	handlerWithLog := middleware.Logger(logger)(mux)
	handlerFinal := middleware.Recover(logger)(handlerWithLog)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", *portFlag),
		Handler: handlerFinal,
	}

	go func() {
		logger.Info("Starting server", "port", portFlag)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	cxt, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(cxt); err != nil {
		logger.Error("Error during shutdown", "error", err)
		_ = server.Close()
	}
	logger.Info("Server stopped")

}
