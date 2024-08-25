// Package collector ...
// This function will be triggered by an http request, to create, update or delete a user in the database.
package collector

import (
	"context"
	"log/slog"
	"os"

	fs "cloud.google.com/go/firestore"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/perebaj/esaj/api"
	"github.com/perebaj/esaj/firestore"
	"github.com/perebaj/esaj/logger"
)

func init() {
	logger, err := logger.NewLoggerSlog(logger.ConfigLogger{
		Level:  logger.LevelInfo,
		Format: logger.FormatGCP,
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

	handler := api.NewUserHandler(storage)
	functions.HTTP("fn-clerk-webhook", handler.ClerkWebHookHandler)
}
