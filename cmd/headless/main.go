// Package main from headless/main.go gather all function that support Chrome headless rendering.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/chromedp/cdproto/storage"
	"github.com/chromedp/chromedp"
	"github.com/perebaj/esaj"
)

func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// ESAJLogin is a struct that holds the login information for the ESAJ website.
type ESAJLogin struct {
	username string
	password string
}

func main() {
	logger, err := esaj.NewLoggerSlog(esaj.ConfigLogger{
		Level:  esaj.LevelDebug,
		Format: esaj.FormatJSON,
	})

	if err != nil {
		fmt.Println("Failed to create logger", err)
		os.Exit(1)
	}

	slog.SetDefault(logger)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.Flag("headless", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	esajLogin := ESAJLogin{
		username: getEnvWithDefault("ESAJ_USERNAME", ""),
		password: getEnvWithDefault("ESAJ_PASSWORD", ""),
	}

	if esajLogin.username == "" || esajLogin.password == "" {
		slog.Error("ESAJ_USERNAME and/or ESAJ_PASSWORD not set")
		os.Exit(1)
	}

	err = chromedp.Run(ctx,
		chromedp.Navigate(`https://esaj.tjsp.jus.br/sajcas/login`),
		chromedp.WaitVisible(`#usernameForm`, chromedp.ByID),
		chromedp.SendKeys(`#usernameForm`, esajLogin.username),
		chromedp.SendKeys(`#passwordForm`, esajLogin.password),
		chromedp.WaitVisible(`#pbEntrar`, chromedp.ByID),
		chromedp.Click(`#pbEntrar`, chromedp.ByID),
		chromedp.WaitVisible(`h1.esajTituloPagina`, chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			cookies, err := storage.GetCookies().Do(ctx)
			if err != nil {
				return fmt.Errorf("could not get cookies: %v", err)
			}

			for _, cookie := range cookies {
				cookieByte, err := cookie.MarshalJSON()
				if err != nil {
					return fmt.Errorf("could not marshal cookie: %v", err)
				}

				slog.Info("Cookie", "cookie", string(cookieByte))

			}
			return nil
		}),
	)

	if err != nil {
		slog.Error("Failed to run chromedp", "error", err)
		os.Exit(1)
	}

}
