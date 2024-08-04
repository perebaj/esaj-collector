// Package main
package main

import (
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/perebaj/esaj"
	"golang.org/x/net/context"
)

// AvailableProcessStatus is a slice of strings that contains the status of the process that contains information about the deadline.
var AvailableProcessStatus = []string{
	"Certidão",
	"Decisão",
	"Certidão de publicação",
}

func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
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

	ctx = context.WithValue(ctx, esaj.ProcessIDContextKey, *processID)

	_, err = esaj.GetCookies(ctx, esajLogin, true, *processID)
	if err != nil {
		logger.Error("error getting cookies: %v", "error", err)
		os.Exit(1)
	}

	// cookieSession, cookiePDFSession := parseCookies(cookies)

	// processCode, err := esaj.SearchDo(cookieSession, *processID)
	// if err != nil {
	// 	logger.Error("error searching process", "error", err)
	// 	os.Exit(1)
	// }

	// logger.Info(fmt.Sprintf("processCode was found: %s for the processID: %s", processCode, *processID))

	// processes, err := esaj.AbrirPastaProcessoDigital(cookieSession, processCode)
	// if err != nil {
	// 	logger.Error("error opening digital folder", "error", err)
	// 	os.Exit(1)
	// }

	// for _, processo := range processes {
	// 	if slices.Contains(availableProcessStatus, processo.Data.Title) {

	// 		err = esaj.GetPDF(ctx, cookiePDFSession, processo.Children[0].ChildernData)
	// 		if err != nil {
	// 			logger.Error("error getting pdf: %v", "error", err)
	// 		}
	// 	}
	// }

	// logger.Info("pdf downloaded successfully")
}

// parseCookies receives a slice of cookies and returns two strings that contains the cookieSession and cookiePDFSession.
// each one is used in different types of http requests.
// the first string return is the cookieSession and the second is the cookiePDFSession
// cookiesSession example: "JSESSIONID=EACA3333A48456D7953B6331999A4F80.cas11; K-JSESSIONID-nckcjpip=0E4D006FFD78524DBABA78F02E1633FA"
// cookiesPDFSession example: "JSESSION=8A1F3DCE0D4DC510FFF3305E44ABCC4E.pasta3; K-JSESSIONID-phoaambo=0E4D006FFD78524DBABA78F02E1633FA"
// func parseCookies(cookies []*network.Cookie) (string, string) {
// 	var cookieSession string
// 	var cookiePDFSession string
// 	for _, cookie := range cookies {
// 		if cookie.Name == "JSESSIONID" && strings.Contains(cookie.Value, "cpopg") {
// 			cookieSession = fmt.Sprintf("%s=%s;", cookie.Name, cookie.Value)
// 		}

// 		if strings.Contains(cookie.Name, "K-JSESSIONID-knbbofpc") {
// 			cookieSession = fmt.Sprintf("%s %s=%s;", cookieSession, cookie.Name, cookie.Value)
// 		}

// 		if cookie.Name == "JSESSIONID" && strings.Contains(cookie.Value, "pasta") {
// 			cookiePDFSession = fmt.Sprintf("%s=%s;", cookie.Name, cookie.Value)
// 		}

// 		if strings.Contains(cookie.Name, "K-JSESSIONID-phoaambo") {
// 			cookiePDFSession = fmt.Sprintf("%s %s=%s;", cookiePDFSession, cookie.Name, cookie.Value)
// 		}
// 	}
// 	return cookieSession, cookiePDFSession
// }
