package hello

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/perebaj/esaj/api"
	"github.com/perebaj/esaj/esaj"
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

	clerkKey := esaj.GetEnvWithDefault("CLERK_SECRET_KEY", "")
	if clerkKey == "" {
		slog.Error("CLERK_SECRET_KEY is required")
		os.Exit(1)
	}

	clerk.SetKey(clerkKey)
	handler := http.HandlerFunc(hello)
	functions.HTTP("jojo-auth", api.ProtectRouteMiddleware(handler).ServeHTTP)
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"message": "hello world"}`))
}
