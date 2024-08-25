// Package collector ...
// this function will be triggered by an http request, it will scrape the esaj website and save all processes URLs given an OAB number
package collector

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	fs "cloud.google.com/go/firestore"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/perebaj/esaj/api"
	"github.com/perebaj/esaj/esaj"
	"github.com/perebaj/esaj/firestore"

	"github.com/perebaj/esaj/logger"
)

func init() {
	logger, err := logger.NewLoggerSlog(logger.ConfigLogger{
		Level:  logger.LevelInfo,
		Format: logger.FormatJSON,
	})
	if err != nil {
		slog.Error("error initializing logger", "error", err)
		os.Exit(1)
	}

	slog.SetDefault(logger)

	projectID := "blup-432616"
	databaseName := "blup-db"
	fsClient, err := fs.NewClientWithDatabase(context.Background(), projectID, databaseName)
	if err != nil {
		slog.Error("error initializing firestore client", "error", err)
		os.Exit(1)
	}
	storage := firestore.NewStorage(fsClient, projectID)
	slog.Info("storage initialized")

	esajClient := esaj.New(esaj.Config{
		// CookieSession and CookiePDFSession are not necessary to scrape basic information from the esaj website
		CookieSession:    "",
		CookiePDFSession: "",
	}, &http.Client{
		Timeout: 90 * time.Second,
	})

	handler := api.NewHandler(storage, esajClient)
	// POST /oab-seeder?oab=123456
	// first argument is the entry-point, second is the handler function
	functions.HTTP("fn-process-seeder", handler.OabSeederHandler)
}
