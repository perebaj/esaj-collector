// Package main from headless/main.go gather all function that support Chrome headless rendering.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/chromedp/cdproto/network"
	"github.com/perebaj/esaj"
)

func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Cookie holds the useful information from the cookies.
type Cookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

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

	cookies, err := esaj.GetCookies(esajLogin, true, *processID)
	if err != nil {
		slog.Error("Failed to get cookies", "error", err)
		os.Exit(1)
	}

	err = saveCookies(cookies)
	if err != nil {
		slog.Error("Failed to save cookies", "error", err)
		os.Exit(1)
	}

	slog.Info("Cookies saved successfully")
}

func saveCookies(cookies []*network.Cookie) error {
	cookiesJSON := make([]Cookie, 0, len(cookies))
	for _, cookie := range cookies {
		cookiesJSON = append(cookiesJSON, Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,
		})
	}

	cookiesBytes, err := json.Marshal(cookiesJSON)
	if err != nil {
		return fmt.Errorf("failed to marshal cookies to json: %v", err)
	}

	err = os.WriteFile("cookies.json", cookiesBytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to write cookies to file: %v", err)
	}

	return nil
}
