// Package main
package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/perebaj/esaj"
	"github.com/perebaj/esaj/api"
	"github.com/perebaj/esaj/postgres"
)

func main() {
	logger, err := esaj.NewLoggerSlog(esaj.ConfigLogger{
		Level:  esaj.LevelInfo,
		Format: esaj.FormatLogFmt,
	})
	if err != nil {
		slog.Info("error initializing logger: %v", "error", err)
		os.Exit(1)
	}

	slog.SetDefault(logger)

	postgresCfg := postgres.Config{
		URL:             os.Getenv("POSTGRES_URL"),
		MaxOpenConns:    10,
		MaxIdleConns:    10,
		ConnMaxIdleTime: 10,
	}

	db, err := postgres.OpenDB(postgresCfg)
	if err != nil {
		slog.Error("error opening database", "error", err)
		os.Exit(1)
	}
	storage := postgres.NewStorage(db)

	esaj := esaj.New(esaj.Config{}, &http.Client{
		Timeout: 30 * time.Second,
	})

	mux := api.NewServerMux(storage, esaj)

	slog.Info("server running on port 8080")

	svc := &http.Server{
		Addr:         fmt.Sprintf(":%d", 8080),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	if err := svc.ListenAndServe(); err != nil {
		slog.Error("error starting server", "error", err)
		os.Exit(1)
	}
}
