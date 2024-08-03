// Package main
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/perebaj/esaj"
	"golang.org/x/net/context"
)

func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

type contextKey string

var (
	processIDContextKey = contextKey("processID")
)

func main() {
	logger, err := esaj.NewLoggerSlog(esaj.ConfigLogger{
		Level:  esaj.LevelDebug,
		Format: esaj.FormatLogFmt,
	})
	if err != nil {
		slog.Info("error initializing logger: %v", "error", err)
		os.Exit(1)
	}

	slog.SetDefault(logger)

	esajLogin := esaj.Login{
		Username: getEnvWithDefault("ESAJ_USERNAME", ""),
		Password: getEnvWithDefault("ESAJ_PASSWORD", ""),
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

	ctx = context.WithValue(ctx, processIDContextKey, *processID)

	cookies, err := esaj.GetCookies(ctx, esajLogin, true, *processID)
	if err != nil {
		logger.Error("error getting cookies: %v", "error", err)
		os.Exit(1)
	}

	var cookieSession string
	var cookiePDFSession string
	for _, cookie := range cookies {
		if cookie.Name == "JSESSIONID" && strings.Contains(cookie.Value, "cpopg") {
			cookieSession = fmt.Sprintf("%s=%s;", cookie.Name, cookie.Value)
		}

		if strings.Contains(cookie.Name, "K-JSESSIONID-knbbofpc") {
			cookieSession = fmt.Sprintf("%s %s=%s;", cookieSession, cookie.Name, cookie.Value)
		}

		if cookie.Name == "JSESSIONID" && strings.Contains(cookie.Value, "pasta") {
			cookiePDFSession = fmt.Sprintf("%s=%s;", cookie.Name, cookie.Value)
		}

		if strings.Contains(cookie.Name, "K-JSESSIONID-phoaambo") {
			cookiePDFSession = fmt.Sprintf("%s %s=%s;", cookiePDFSession, cookie.Name, cookie.Value)
		}
	}

	processCode, err := esaj.SearchDo(cookieSession, *processID)
	if err != nil {
		logger.Error("error searching process", "error", err)
		os.Exit(1)
	}

	logger.Info(fmt.Sprintf("processCode was found: %s for the processID: %s", processCode, *processID))

	processes, err := esaj.AbrirPastaProcessoDigital(cookieSession, processCode)
	if err != nil {
		logger.Error("error opening digital folder", "error", err)
		os.Exit(1)
	}

	logger.Info("processes: %v", "processes [0] param", processes[0].Children[0].ChildernData.Parametros)

	err = esaj.GetPDF(cookiePDFSession, processes[0].Children[0].ChildernData.Parametros)
	if err != nil {
		logger.Error("error getting pdf: %v", "error", err)
		os.Exit(1)
	}

	logger.Info("pdf downloaded successfully")
}
