// Package main
package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/perebaj/esaj"
)

// Cookie holds the useful information from the cookies.
type Cookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// formatCookies reads the cookies.json file and formats the cookies to be used in the requests.
func formatCookies() (string, error) {
	cookies, err := os.ReadFile("cookies.json")
	if err != nil {
		return "", fmt.Errorf("error reading cookies: %w", err)
	}

	var cookiesJSON []Cookie

	err = json.Unmarshal(cookies, &cookiesJSON)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling cookies: %w", err)
	}

	var cookieHeader string
	for _, cookie := range cookiesJSON {
		if cookie.Name == "JSESSIONID" && strings.Contains(cookie.Value, "cpopg") {
			cookieHeader = fmt.Sprintf("%s=%s;", cookie.Name, cookie.Value)
		}

		if strings.Contains(cookie.Name, "K-JSESSIONID-knbbofpc") {
			cookieHeader = fmt.Sprintf("%s %s=%s;", cookieHeader, cookie.Name, cookie.Value)
		}
	}

	// remove the last character, a additional semicolon
	cookieHeader = cookieHeader[:len(cookieHeader)-1]
	slog.Info(fmt.Sprintf("cookieHeader: %s", cookieHeader))

	return cookieHeader, nil
}

func formatCookieGetPDF() (string, error) {
	cookies, err := os.ReadFile("cookies.json")
	if err != nil {
		return "", fmt.Errorf("error reading cookies: %w", err)
	}

	var cookiesJSON []Cookie

	err = json.Unmarshal(cookies, &cookiesJSON)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling cookies: %w", err)
	}

	var cookieHeader string
	for _, cookie := range cookiesJSON {
		if cookie.Name == "JSESSIONID" && strings.Contains(cookie.Value, "pasta") {
			cookieHeader = fmt.Sprintf("%s=%s;", cookie.Name, cookie.Value)
		}

		if strings.Contains(cookie.Name, "K-JSESSIONID-phoaambo") {
			cookieHeader = fmt.Sprintf("%s %s=%s;", cookieHeader, cookie.Name, cookie.Value)
		}
	}

	// remove the last character, a additional semicolon
	cookieHeader = cookieHeader[:len(cookieHeader)-1]
	slog.Info(fmt.Sprintf("cookieHeader get pdf: %s", cookieHeader))

	return cookieHeader, nil
}

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

	cookieSession, err := formatCookies()
	if err != nil {
		slog.Error("error formatting cookies: %v", "error", err)
		os.Exit(1)
	}

	processCode := "1H000H91J0000"

	processes, err := esaj.AbrirPastaProcessoDigital(cookieSession, processCode)
	if err != nil {
		slog.Error("error opening digital folder", "error", err)
		os.Exit(1)
	}

	slog.Info("processes: %v", "processes [0] param", processes[0].Children[0].ChildernData.Parametros)

	cookiePDFSession, err := formatCookieGetPDF()
	if err != nil {
		slog.Error("error formatting cookies: %v", "error", err)
		os.Exit(1)
	}

	err = esaj.GetPDF(cookiePDFSession, processes[0].Children[0].ChildernData.Parametros)
	if err != nil {
		slog.Error("error getting pdf: %v", "error", err)
		os.Exit(1)
	}
}
