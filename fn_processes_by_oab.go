// An API endpoint that returns a list of processes by OAB

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

	// This endpoint is not using the esaj client, so we don't need to load it here
	// POST /processes-by-oab?oab=123456
	handler := api.NewHandler(storage, nil)
	functions.HTTP("fn-processes-by-oab", handler.ProcessesByOABHandler)
}
