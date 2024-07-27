// Package main
package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/perebaj/esaj"
)

func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func numeroDigitoAnoUnificado(processID string) (string, error) {
	regex := regexp.MustCompile(`(\d{7}-\d{2}.\d{4})`)
	matches := regex.FindStringSubmatch(processID)
	if len(matches) == 0 {
		return "", fmt.Errorf("no matches found when searching for numeroDigitoAnoUnificado")
	}
	return matches[1], nil
}

func foroNumeroUnificado(processID string) (string, error) {
	regex := regexp.MustCompile(`(\d{7})-(\d{2}).(\d{4}).(\d{1}).(\d{2}).(\d{4})`)
	matches := regex.FindStringSubmatch(processID)
	if len(matches) == 0 {
		return "", fmt.Errorf("no matches found when searching for foroNumeroUnificado")
	}

	return matches[6], nil
}

func searchDo(jsession string, processID string) (string, error) {
	client := &http.Client{}

	numeroDigitoAnoUnificado, err := numeroDigitoAnoUnificado(processID)
	if err != nil {
		return "", err
	}

	foroNumeroUnificado, err := foroNumeroUnificado(processID)
	if err != nil {
		return "", err
	}

	urlFormated := fmt.Sprintf(`https://esaj.tjsp.jus.br/cpopg/search.do?conversationId=&cbPesquisa=NUMPROC&numeroDigitoAnoUnificado=%s&foroNumeroUnificado=%s&dadosConsulta.valorConsultaNuUnificado=%s&dadosConsulta.valorConsultaNuUnificado=UNIFICADO&dadosConsulta.valorConsulta=&dadosConsulta.tipoNuProcesso=UNIFICADO`, numeroDigitoAnoUnificado, foroNumeroUnificado, processID)

	slog.Info(fmt.Sprintf("urlFormated: %s", urlFormated))

	req, err := http.NewRequest("GET", urlFormated, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Cookie", jsession)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error doing request: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	bodyByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading body: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(bodyByte)))

	if err != nil {
		return "", fmt.Errorf("error initializing goquery new document from reader: %w", err)
	}

	var link string
	doc.Find("tr > td > a.linkMovVincProc").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		href, _ := s.Attr("href")
		if strings.Contains(href, "abrirDocumentoVinculadoMovimentacao.do") {
			link = href
			return false
		}
		return true
	})

	regex := regexp.MustCompile(`processo.codigo=(\w+)`)
	matches := regex.FindStringSubmatch(link)
	if len(matches) == 0 {
		return "", fmt.Errorf("no matches found when searching for processCode")
	}

	processCode := matches[1]
	return processCode, nil
}

func abrirPastaDigital(jsessionid, processCode string) (string, error) {
	formatedURL := fmt.Sprintf("https://esaj.tjsp.jus.br/cpopg/abrirPastaDigital.do?processo.codigo=%s", processCode)
	slog.Info(fmt.Sprintf("formatedURL: %s", formatedURL))

	client := &http.Client{}
	req, err := http.NewRequest("GET", formatedURL, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Cookie", jsessionid)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error doing request: %w", err)
	}

	bodyByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading body: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(bodyByte)))
	if err != nil {
		return "", fmt.Errorf("error initializing goquery new document from reader: %w", err)
	}

	var link string
	doc.Find("body").First().Each(func(_ int, s *goquery.Selection) {
		link = s.Text()
	})

	if link == "" {
		return "", fmt.Errorf("no link found")
	}

	return link, nil
}

func main() {
	logger, err := esaj.NewLoggerSlog(esaj.ConfigLogger{
		Level:  esaj.LevelDebug,
		Format: esaj.FormatJSON,
	})
	if err != nil {
		slog.Info("error initializing logger: %v", "error", err)
		os.Exit(1)
	}

	slog.SetDefault(logger)

	jsession := getEnvWithDefault("JSESSIONID", "")

	if jsession == "" {
		slog.Error("The JSESSIONID environment variable is required")
		os.Exit(1)
	}

	processCode, err := searchDo(jsession, "1029989-06.2022.8.26.0053")
	if err != nil {
		slog.Error("error searching do: %v", "error", err)
		os.Exit(1)
	}

	slog.Info(fmt.Sprintf("processCode: %s", processCode))

	pastaDigitalURL, err := abrirPastaDigital(jsession, processCode)
	if err != nil {
		slog.Error("error opening digital folder: %v", "error", err)
		os.Exit(1)
	}

	slog.Info(fmt.Sprintf("pastaDigitalLink: %s", pastaDigitalURL))
}
