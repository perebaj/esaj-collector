// Package main
package main

import (
	"encoding/json"
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

func searchDo(cookieSession string, processID string) (string, error) {
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
	req.Header.Set("Cookie", cookieSession)

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

func abrirPastaDigital(cookieSession, processCode string) (string, error) {
	formatedURL := fmt.Sprintf("https://esaj.tjsp.jus.br/cpopg/abrirPastaDigital.do?processo.codigo=%s", processCode)
	slog.Info(fmt.Sprintf("formatedURL: %s", formatedURL))

	client := &http.Client{}
	req, err := http.NewRequest("GET", formatedURL, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Cookie", cookieSession)

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

	if strings.Contains(link, "Não foi possível validar o seu acesso") {
		return "", fmt.Errorf("access not validated, verify the COOKIESESSION")
	}

	return link, nil
}

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

	cookieSession, err := formatCookies()
	if err != nil {
		slog.Error("error formatting cookies: %v", "error", err)
		os.Exit(1)
	}

	processCode, err := searchDo(cookieSession, "1029989-06.2022.8.26.0053")
	if err != nil {
		slog.Error("error searching do: %v", "error", err)
		os.Exit(1)
	}

	slog.Info(fmt.Sprintf("processCode: %s", processCode))

	pastaDigitalURL, err := abrirPastaDigital(cookieSession, processCode)
	if err != nil {
		slog.Error("error opening digital folder", "error", err)
		os.Exit(1)
	}

	slog.Info(fmt.Sprintf("pastaDigitalLink: %s", pastaDigitalURL))

	client := &http.Client{}
	req, err := http.NewRequest("GET", pastaDigitalURL, nil)
	if err != nil {
		slog.Error("error creating request: %v", "error", err)
		os.Exit(1)
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("error doing request: %v", "error", err)
		os.Exit(1)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	bodyByte, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("error reading body: %v", "error", err)
		os.Exit(1)
	}

	err = os.WriteFile("pasta_digital.html", bodyByte, 0644)
	if err != nil {
		slog.Error("error writing file: %v", "error", err)
		os.Exit(1)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(bodyByte)))
	if err != nil {
		slog.Error("error initializing goquery new document from reader: %v", "error", err)
		os.Exit(1)
	}

	var scriptContent string
	doc.Find("script").Each(func(_ int, s *goquery.Selection) {
		scriptContent = s.Text()
	})

	if scriptContent == "" {
		slog.Error("no script content found")
		os.Exit(1)
	}

	err = os.WriteFile("script_content.html", []byte(scriptContent), 0644)
	if err != nil {
		slog.Error("error writing file: %v", "error", err)
		os.Exit(1)
	}

	regex := regexp.MustCompile(`var requestScope = (.*);`)
	matches := regex.FindStringSubmatch(scriptContent)
	if len(matches) == 0 {
		slog.Error("no matches found when searching for requestScope")
		os.Exit(1)
	}

	err = os.WriteFile("request_scope.json", []byte(matches[1]), 0644)
	if err != nil {
		slog.Error("error writing file: %v", "error", err)
		os.Exit(1)
	}

	var processes []esaj.Process
	err = json.Unmarshal([]byte(matches[1]), &processes)
	if err != nil {
		slog.Error("error unmarshalling json: %v", "error", err)
		os.Exit(1)
	}

	slog.Info(fmt.Sprintf("parametros get pdf: %s", processes[0].Children[0].ChildernData.Parametros))

	hrefGetPDF := "https://esaj.tjsp.jus.br/pastadigital/getPDF.do?" + processes[0].Children[0].ChildernData.Parametros

	slog.Info(fmt.Sprintf("hrefGetPDF: %s", hrefGetPDF))

	cookiePDFSession, err := formatCookieGetPDF()
	if err != nil {
		slog.Error("error formatting cookies: %v", "error", err)
		os.Exit(1)
	}

	req, err = http.NewRequest("GET", hrefGetPDF, nil)
	if err != nil {
		slog.Error("error creating request: %v", "error", err)
		os.Exit(1)
	}

	req.Header.Set("Cookie", cookiePDFSession)

	resp, err = client.Do(req)
	if err != nil {
		slog.Error("error doing request: %v", "error", err)
		os.Exit(1)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	bodyByte, err = io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("error reading body: %v", "error", err)
		os.Exit(1)
	}

	if strings.Contains(string(bodyByte), "Sua sessão expirou") {
		slog.Error("access not authorized to get pdf")
		os.Exit(1)
	}

	err = os.WriteFile("documento.pdf", bodyByte, 0644)
	if err != nil {
		slog.Error("error writing file: %v", "error", err)
		os.Exit(1)
	}
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
		if cookie.Name == "JSESSIONID" && strings.Contains(cookie.Value, "pasta6") {
			cookieHeader = fmt.Sprintf("%s=%s;", cookie.Name, cookie.Value)
		}

		if strings.Contains(cookie.Name, "K-JSESSIONID-phoaambo") {
			cookieHeader = fmt.Sprintf("%s %s=%s;", cookieHeader, cookie.Name, cookie.Value)
		}
	}

	// remove the last character, a additional semicolon
	cookieHeader = cookieHeader[:len(cookieHeader)-1]
	slog.Info(fmt.Sprintf("cookieHeader: %s", cookieHeader))

	return cookieHeader, nil
}
