package functions

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/perebaj/esaj"
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

	functions.HTTP("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})
}
