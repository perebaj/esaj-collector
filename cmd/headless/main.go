// Package main from headless/main.go gather all function that support Chrome headless rendering.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"os"

	"github.com/chromedp/cdproto/network"
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

// Cookie holds the useful information from the cookies.
type Cookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
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

	var cookies []*network.Cookie
	var pastaVirtualBodyTxt string
	var pastaVirtualHREF string

	err = chromedp.Run(ctx,
		chromedp.Navigate(`https://esaj.tjsp.jus.br/sajcas/login`),
		chromedp.WaitVisible(`#usernameForm`, chromedp.ByID),
		chromedp.SendKeys(`#usernameForm`, esajLogin.username),
		chromedp.SendKeys(`#passwordForm`, esajLogin.password),
		chromedp.WaitVisible(`#pbEntrar`, chromedp.ByID),
		chromedp.Click(`#pbEntrar`, chromedp.ByID),
		chromedp.WaitVisible(`h1.esajTituloPagina`, chromedp.ByQuery),
		chromedp.Navigate("https://esaj.tjsp.jus.br/cpopg/open.do"),
		chromedp.WaitVisible(`a.linkLogo`, chromedp.ByQuery),
		chromedp.Navigate("https://esaj.tjsp.jus.br/cpopg/show.do?processo.codigo=1H000MCVK0000&processo.foro=53&processo.numero=1029989-06.2022.8.26.0053"),
		chromedp.Navigate("https://esaj.tjsp.jus.br/cpopg/abrirPastaDigital.do?processo.codigo=1H000MCVK0000"),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Text(`body`, &pastaVirtualBodyTxt, chromedp.NodeVisible, chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			u, err := url.Parse(pastaVirtualBodyTxt)
			if err != nil {
				return fmt.Errorf("could not parse pastaVirtualHFEF: %v", err)
			}

			pastaVirtualHREF = u.RawQuery

			slog.Info("pastaVirtualURL", "url", pastaVirtualHREF)
			// create a navigation entry
			err = chromedp.Navigate("https://esaj.tjsp.jus.br/pastadigital/abrirPastaProcessoDigital.do?" + pastaVirtualHREF).Do(ctx)
			if err != nil {
				return fmt.Errorf("could not navigate to pastaVirtualURL: %v", err)
			}

			err = chromedp.WaitVisible(`input#salvarButton`, chromedp.ByQuery).Do(ctx)
			if err != nil {
				return fmt.Errorf("could not wait for input#salvarButton: %v", err)
			}

			cookies, err = storage.GetCookies().Do(ctx)
			if err != nil {
				return fmt.Errorf("could not get cookies: %v", err)
			}

			return nil
		}),
	)

	if err != nil {
		slog.Error("Failed to run chromedp", "error", err)
		os.Exit(1)
	}

	err = saveCookies(cookies)
	if err != nil {
		slog.Error("Failed to save cookies", "error", err)
		os.Exit(1)
	}

	slog.Info("Cookies saved!")
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
