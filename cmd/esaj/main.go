// Package main
package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/perebaj/esaj"
	"golang.org/x/net/context"
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

	logger = logger.With("processID", *processID)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	ctx = context.WithValue(ctx, esaj.ProcessIDContextKey, *processID)

	cookieSession, cookiePDFSession, err := esaj.GetCookies(ctx, esajLogin, true, *processID)
	if err != nil {
		logger.Error("error getting cookies: %v", "error", err)
		os.Exit(1)
	}

	client := esaj.New(esaj.Config{
		CookieSession:    cookieSession,
		CookiePDFSession: cookiePDFSession,
	}, &http.Client{
		Timeout: 60 * time.Second,
	})

	err = client.Run(ctx, *processID)
	if err != nil {
		logger.Error("error running the esaj parser", "error", err)
		os.Exit(1)
	}

	logger.Info("all pdfs were downloaded successfully")
}
