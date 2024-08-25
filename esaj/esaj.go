// Package esaj from esaj.go is a package that provides functions to interact with the TJSP website.
// The function names follow the same naming convention as the original API.
// This package depends of some cookies to work properly, those cookies can only be obtained by using a headless browser.
package esaj

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/perebaj/esaj/tracing"
)

// contextKey is a type used to store the context key.
type contextKey string

var (
	// ProcessIDContextKey is the key used to store the processID in the context.
	ProcessIDContextKey = contextKey("processID")
)

var (
	// ErrSessionExpired is an error that occurs when the access to the TJSP website is expired.
	ErrSessionExpired = errors.New("session expired")
)

// availableProcessStatus is a slice of strings that contains the status of the process that contains information about the deadline.
// Certify that the value is in lower case, because the comparison is case insensitive.
var availableProcessStatus = []string{
	"certidão de publicação",
}

// Config is a struct that contains the configuration of the ESAJClient.
type Config struct {
	// CookieSession is used in the majority of the requests.
	// Example of CookieSession: "JSESSIONID=EACA3333A48456D7953B6331999A4F80.cas11; K-JSESSIONID-nckcjpip=0E4D006FFD78524DBABA78F02E1633FA"
	CookieSession string
	// CookiePDFSession is used for the route the download a PDF.
	// CookiePDFSession example: "JSESSION=8A1F3DCE0D4DC510FFF3305E44ABCC4E.pasta3; K-JSESSIONID-phoaambo=0E4D006FFD78524DBABA78F02E1633FA"
	CookiePDFSession string
}

// Client is a struct that contains the configuration of the client to interact with the TJSP website.
type Client struct {
	Config Config
	Client *http.Client
	// URL is the base URL of the TJSP website.
	URL string
}

// New creates a new esaj Client.
func New(config Config, client *http.Client) *Client {
	return &Client{
		Config: config,
		Client: client,
		URL:    "https://esaj.tjsp.jus.br",
	}
}

// Run is the main function of the Client. It searches for the process in the TJSP website and download the PDF documents.
func (ec Client) Run(ctx context.Context, processID string) error {
	processCode, err := ec.searchDo(processID)
	if err != nil {
		return fmt.Errorf("error searching process: %w", err)
	}

	processes, err := ec.abrirPastaProcessoDigital(processCode)
	if err != nil {
		return fmt.Errorf("error opening digital folder: %w", err)
	}

	for _, p := range processes {
		if slices.Contains(availableProcessStatus, strings.ToLower(p.Data.Title)) {
			err = ec.GetPDF(ctx, processID, p.Children[0].ChildernData)
			if err != nil {
				return fmt.Errorf("error getting pdf: %w", err)
			}
		}
	}
	return nil
}

// searchDo searches for a specific process in the TJSP website and return the processCode. An ID in the format 1H000H91J0000.
// processID: The process ID in the format = 0000001-02.2021.8.26.0000
func (ec Client) searchDo(processID string) (string, error) {
	numeroDigitoAnoUnificado, err := numeroDigitoAnoUnificado(processID)
	if err != nil {
		return "", err
	}

	foroNumeroUnificado, err := foroNumeroUnificado(processID)
	if err != nil {
		return "", err
	}

	urlFormated := ec.URL + fmt.Sprintf(`/cpopg/search.do?conversationId=&cbPesquisa=NUMPROC&numeroDigitoAnoUnificado=%s&foroNumeroUnificado=%s&dadosConsulta.valorConsultaNuUnificado=%s&dadosConsulta.valorConsultaNuUnificado=UNIFICADO&dadosConsulta.valorConsulta=&dadosConsulta.tipoNuProcesso=UNIFICADO`, numeroDigitoAnoUnificado, foroNumeroUnificado, processID)

	req, err := http.NewRequest("GET", urlFormated, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Cookie", ec.Config.CookieSession)

	resp, err := ec.Client.Do(req)
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

// abrirPastaProcessoDigital fetches the digital folder page where all structured data of the process can be found.
// This data is used to download the PDF documents related to the process.
// - processCode: The process code in the format: 1H000H91J0000
func (ec Client) abrirPastaProcessoDigital(processCode string) ([]Process, error) {
	url, err := ec.pastaDigitalURL(processCode)
	if err != nil {
		return nil, fmt.Errorf("error getting pasta digital url: %w", err)
	}

	slog.Debug(fmt.Sprintf("fetching abrir pasta processo digital url: %s", ec.URL+url))
	req, err := http.NewRequest("GET", ec.URL+url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := ec.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error doing request: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	bodyByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %w", err)
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(bodyByte)))
	if err != nil {
		return nil, fmt.Errorf("error initializing goquery new document from reader: %w", err)
	}

	// into the script tag, we can find the requestScope that contains the structured data of the process.
	// this data after parsed, can be used to download the PDF documents.
	var scriptContent string
	doc.Find("script").Each(func(_ int, s *goquery.Selection) {
		scriptContent = s.Text()
	})

	if scriptContent == "" {
		return nil, fmt.Errorf("no script content found")
	}

	regex := regexp.MustCompile(`var requestScope = (.*);`)
	matches := regex.FindStringSubmatch(scriptContent)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no matches found when searching for requestScope")
	}

	var processes []Process
	err = json.Unmarshal([]byte(matches[1]), &processes)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling json: %w", err)
	}

	return processes, nil
}

// GetPDF fetch the pdf document from the TJSP website.
func (ec Client) GetPDF(_ context.Context, processID string, cData ChildrenData) error {
	hrefGetPDF := ec.URL + "/pastadigital/getPDF.do?" + cData.Parametros

	req, err := http.NewRequest("GET", hrefGetPDF, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Cookie", ec.Config.CookiePDFSession)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error doing request %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	bodyByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading body: %w", err)
	}

	if strings.Contains(string(bodyByte), "Sua sessão expirou") {
		return ErrSessionExpired
	}

	fileName := "tmp/" + processID + "_" + cData.Title + ".pdf"
	err = os.WriteFile(fileName, bodyByte, 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}
	slog.Info(fmt.Sprintf("pdf downloaded successfully and saved in: %s", fileName))

	return nil
}

// FetchBasicProcessInfo fetch the html page of the process that contains basic information about legal action.
func (ec Client) FetchBasicProcessInfo(ctx context.Context, u string, processID string) (*ProcessBasicInfo, error) {
	traceID := tracing.GetTraceIDFromContext(ctx)
	logger := slog.With("traceID", traceID, "processID", processID)
	parsedURL, err := url.Parse(u)

	if err != nil {
		logger.Error("error parsing the url", "url", u, "error", err)
		return nil, err
	}

	processCode := parsedURL.Query().Get("processo.codigo")
	processForo := parsedURL.Query().Get("processo.foro")

	if processCode == "" || processForo == "" {
		logger.Error(fmt.Sprintf("error parsing the url: %s. processo.codigo or processo.foro is empty", u))
		return nil, err
	}

	logger.Info("fetching process basic information")

	url := ec.URL + fmt.Sprintf("/cpopg/show.do?processo.codigo=%s&processo.foro=%s&processo.numero=%s",
		processCode,
		processForo,
		processID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error("error creating request", "error", err)
		return nil, err
	}
	req.Header.Set("Cookie", ec.Config.CookieSession)

	resp, err := ec.Client.Do(req)
	if err != nil {
		logger.Error("error doing request", "error", err)
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	bodyByte, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("error reading body", "error", err)
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(bodyByte)))
	if err != nil {
		logger.Error("error initializing goquery new document from reader", "error", err)
		return nil, err
	}

	var processClass string
	doc.Find("#classeProcesso").Each(func(_ int, s *goquery.Selection) {
		processClass = s.Text()
	})

	var foroName string
	doc.Find("#foroProcesso").Each(func(_ int, s *goquery.Selection) {
		foroName = s.Text()
	})

	var vara string
	doc.Find("#varaProcesso").Each(func(_ int, s *goquery.Selection) {
		vara = s.Text()
	})

	var judge string
	doc.Find("#juizProcesso").Each(func(_ int, s *goquery.Selection) {
		judge = s.Text()
	})

	var parties []string
	doc.Find("td.nomeParteEAdvogado").Each(func(_ int, s *goquery.Selection) {
		p := s.Text()
		r := strings.NewReplacer("\n", "", "\t", "")
		p = r.Replace(p)
		parties = append(parties, p)
	})

	if len(parties) < 2 {
		logger.Error("error parsing parties")
		return nil, fmt.Errorf("error parsing parties")
	}

	pBasic := &ProcessBasicInfo{
		ProcessID:   processID,
		ProcessForo: processForo,
		Class:       processClass,
		Vara:        vara,
		Judge:       judge,
		ForoName:    foroName,
		ProcessCode: processCode,
		// TODO(@perebaj) maybe im accessing an index that does not exist. Or maybe the parties are not in the correct order.
		Claimant:  parties[0],
		Defendant: parties[1],
		URL:       u,
	}

	return pBasic, nil
}

// ProcessSeed is the start point to scrape all processes related to a specific OAB number
type ProcessSeed struct {
	ProcessID string `db:"process_id" json:"process_id"`
	OAB       string `db:"oab" json:"oab"`
	URL       string `db:"url" json:"url"`
}

// SearchByOAB is a seeder function that searches for all processes related to a specific OAB number.
// to get all processes hrefs its not necessary to have a valid session.
func (ec Client) SearchByOAB(ctx context.Context, oab string) ([]ProcessSeed, error) {
	traceID := tracing.GetTraceIDFromContext(ctx)
	logger := slog.With("traceID", traceID, "oab", oab)
	// paginaConsulta=1000000000 is a way to find the last page, so we can iterate over all pages.
	// using this output as a range limit.
	fetchURL := ec.URL + fmt.Sprintf("/cpopg/trocarPagina.do?paginaConsulta=1000000000&conversationId=&cbPesquisa=NUMOAB&dadosConsulta.valorConsulta=%s&cdForo=-1", oab)

	req, err := http.NewRequest("GET", fetchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	logger.Info(fmt.Sprintf("searching by all process related to OAB: %s", oab), "url", fetchURL)
	resp, err := ec.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error doing request: %w", err)
	}

	bodyByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(bodyByte)))

	if err != nil {
		return nil, fmt.Errorf("error initializing goquery new document from reader: %w", err)
	}

	// the pagination element in the esaj HTML just contains the penultimate page.
	// so we need to get it and add 1 to get the last page.
	var penultimatePage string
	doc.Find("a.paginacao").Last().Each(func(_ int, s *goquery.Selection) {
		txt := s.Text()
		penultimatePage = txt
	})

	replacer := strings.NewReplacer("\n", "", "\t", "", " ", "")
	penultimatePage = replacer.Replace(penultimatePage)

	// if the penultimatePage is empty, it means that there is only one page.
	var penultimatePageInt int
	// TODO(@perebaj) Maybe if the request failed to retrive the HTML, this value will be empty, so, to avoid
	// miss flow, we need to get different values to have sure that the code its working properly.
	if penultimatePage == "" {
		logger.Info("only one page found, seeting page seek start point to 1")
		penultimatePageInt = 1
	} else {
		logger.Info(fmt.Sprintf("penultimate page found: %s", penultimatePage))
		penultimatePageInt, err = strconv.Atoi(penultimatePage)
		if err != nil {
			return nil, fmt.Errorf("error converting text to number: %w", err)
		}
	}

	lastPage := penultimatePageInt + 1
	logger.Info(fmt.Sprintf("number of pages to iterate: %d", lastPage))
	// iterate over all pages to get all processes hrefs.
	var seeds []ProcessSeed
	// the first page 1 and 0 refers to the same page, so, to avoid duplicate data, we are starting from 1.
	for i := 1; i <= lastPage; i++ {
		fetchURL := ec.URL + fmt.Sprintf("/cpopg/trocarPagina.do?paginaConsulta=%d&cbPesquisa=NUMOAB&dadosConsulta.valorConsulta=%s&cdForo=-1", i, oab)
		logger.Info(fmt.Sprintf("fetching page: %d", i), "url", fetchURL)
		req, err := http.NewRequest("GET", fetchURL, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating request: %w", err)
		}

		resp, err := ec.Client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error doing request: %w", err)
		}

		bodyByte, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading body: %w", err)
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(bodyByte)))
		if err != nil {
			return nil, fmt.Errorf("error initializing goquery new document from reader: %w", err)
		}

		doc.Find("a.linkProcesso").Each(func(_ int, s *goquery.Selection) {
			href, _ := s.Attr("href")
			processID := s.Text()
			// remove all spaces, tabs and new lines.
			processID = replacer.Replace(processID)
			logger.Info(fmt.Sprintf("process found: %s", processID))

			seeds = append(seeds, ProcessSeed{
				ProcessID: processID,
				URL:       ec.URL + href,
				OAB:       oab,
			})
		})
		logger.Info(fmt.Sprintf("number of processes found: %d", len(seeds)))
	}

	return seeds, nil
}

// pastaDigitalURL fetch the html page and return the URL where the pdf documents can be downloaded.
// - processCode: The process code in the format: 1H000H91J0000
func (ec Client) pastaDigitalURL(processCode string) (string, error) {
	formatedURL := ec.URL + fmt.Sprintf("/cpopg/abrirPastaDigital.do?processo.codigo=%s", processCode)

	req, err := http.NewRequest("GET", formatedURL, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Cookie", ec.Config.CookieSession)

	resp, err := ec.Client.Do(req)
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
		return "", ErrSessionExpired
	}

	link = strings.ReplaceAll(link, "\n", "")
	link = strings.ReplaceAll(link, "\t", "")

	linkHREF := strings.Split(link, "https://esaj.tjsp.jus.br")
	if len(linkHREF) < 2 {
		return "", fmt.Errorf("no link found")
	}

	// this linkHREF looks like: /pastadigital/abrirPastaProcessoDigital.do
	return linkHREF[1], nil
}

// More about this way to handle context in Go: https://pkg.go.dev/context#example-WithValue
func getContextWithProcessID(ctx context.Context, k contextKey) (string, error) {
	if v := ctx.Value(k); v != nil {
		return v.(string), nil
	}
	return "", fmt.Errorf("could not get key %s from context", k)
}

// processeID input example: 0000001-02.2021.8.26.0000
// numeroDigitoAnoUnificado output example: 0000001-02.2021.
func numeroDigitoAnoUnificado(processID string) (string, error) {
	regex := regexp.MustCompile(`(\d{7}-\d{2}.\d{4})`)
	matches := regex.FindStringSubmatch(processID)
	if len(matches) == 0 {
		return "", fmt.Errorf("no matches found when searching for numeroDigitoAnoUnificado. processID input: %s", processID)
	}
	return matches[1], nil
}

// processeID input example: 0000001-02.2021.8.26.0054
// foroNumeroUnificado output example: 0054. The last four digits of the processID
func foroNumeroUnificado(processID string) (string, error) {
	regex := regexp.MustCompile(`(\d{7})-(\d{2}).(\d{4}).(\d{1}).(\d{2}).(\d{4})`)
	matches := regex.FindStringSubmatch(processID)
	if len(matches) == 0 {
		return "", fmt.Errorf("no matches found when searching for foroNumeroUnificado. processID input: %s", processID)
	}

	return matches[6], nil
}
