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
	// POST /clerk-webhook
	// first argument is the entry-point, second is the handler function
	functions.HTTP("fn-clerk-webhook", handler.ClerkWebHookHandler)
}
