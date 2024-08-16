package functions

import (
	"context"
	"log/slog"
	"os"

	fs "cloud.google.com/go/firestore"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/perebaj/esaj"
	"github.com/perebaj/esaj/api"
	"github.com/perebaj/esaj/firestore"
)

func init() {
	logger, err := esaj.NewLoggerSlog(esaj.ConfigLogger{
		Level:  esaj.LevelInfo,
		Format: esaj.FormatLogFmt,
	})
	if err != nil {
		slog.Error("error initializing logger", "error", err)
		os.Exit(1)
	}

	slog.SetDefault(logger)

	projectID := "test"
	database := "test"
	fsClient, err := fs.NewClient(context.TODO(), projectID)
	if err != nil {
		slog.Error("error initializing firestore client", "error", err)
		os.Exit(1)
	}
	storage := firestore.NewStorage(fsClient, projectID, database)
	slog.Info("storage initialized")

	handler := api.NewHandler(storage, nil)

	// POST /oab-seeder?oab=123456
	functions.HTTP("/oab-seeder", handler.OabSeederHandler)
}
