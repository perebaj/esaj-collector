// Package main from headless/main.go gather all function that support Chrome headless rendering.
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/perebaj/esaj"
)

func main() {
	logger, err := esaj.NewLoggerSlog(esaj.ConfigLogger{
		Level:  esaj.LevelDebug,
		Format: esaj.FormatLogFmt,
	})

	if err != nil {
		fmt.Println("Failed to create logger", err)
		os.Exit(1)
	}

	slog.SetDefault(logger)

	esajLogin := esaj.Login{
		Username: esaj.GetEnvWithDefault("ESAJ_USERNAME", ""),
		Password: esaj.GetEnvWithDefault("ESAJ_PASSWORD", ""),
	}

	if esajLogin.Username == "" || esajLogin.Password == "" {
		slog.Error("ESAJ_USERNAME and/or ESAJ_PASSWORD not set")
		os.Exit(1)
	}

	processID := flag.String("processID", "", "Process ID to search in the format 1016358-63.2020.8.26.0053")
	flag.Parse()

	if *processID == "" {
		slog.Error("processID not set")
		os.Exit(1)
	}

}
