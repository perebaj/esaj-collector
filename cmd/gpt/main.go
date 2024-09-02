// Package main ...
package main

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/perebaj/esaj"
	gpt "github.com/perebaj/esaj/openai"
)

func main() {
	cfg := gpt.Config{
		APIToken: esaj.GetEnvWithDefault("OPENAI_API_TOKEN", ""),
	}

	if err := cfg.Validate(); err != nil {
		slog.Error("error validating the configuration", "error", err)
		os.Exit(1)
	}

	c := gpt.New(cfg)

	fileName := "tmp/markdown/1064759-59.2021.8.26.0053_Página 102.pdf.md"

	f, err := os.Open(fileName)
	if err != nil {
		slog.Error("error opening the file", "error", err)
		os.Exit(1)
	}

	defer func() {
		_ = f.Close()
	}()

	b, err := io.ReadAll(f)
	if err != nil {
		slog.Error("error reading the file", "error", err)
		os.Exit(1)
	}

	parsed, err := c.ParsePublication(context.Background(), string(b))
	if err != nil {
		slog.Error("error parsing the publication", "error", err)
		os.Exit(1)
	}

	slog.Info("parsed publication", "parsed", parsed)
}
