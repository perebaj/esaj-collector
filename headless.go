package esaj

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/storage"
	"github.com/chromedp/chromedp"
)

// Login is a struct that holds the login information for the ESAJ website.
type Login struct {
	Username string
	Password string
}

// GetCookies use a headless browser to simulate the login and all the steps to retrive the cookies from the ESAJ website.
// - headless is a boolean that defines if the browser should be headless or not. For production, it must be true.
// - processoID example: 1016358-63.2020.8.26.0053
func GetCookies(ctx context.Context, esajLogin Login, headless bool, processoID string) (string, string, error) {
	logger := slog.With("processID", processoID)

	logger.Debug(fmt.Sprintf("GetCookies headless initialized with the headless option: %v", headless))

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.Flag("headless", headless),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	var cookies []*network.Cookie

	searchDo, err := searchDoURL(processoID)
	if err != nil {
		return "", "", fmt.Errorf("bulding the searchDoURL: %v", err)
	}
	logger.Debug("searchDoURL", "url", searchDo)

	err = chromedp.Run(ctx,
		chromedp.Navigate(`https://esaj.tjsp.jus.br/sajcas/login`),
		chromedp.WaitVisible(`#usernameForm`, chromedp.ByID),
		chromedp.SendKeys(`#usernameForm`, esajLogin.Username),
		chromedp.SendKeys(`#passwordForm`, esajLogin.Password),
		chromedp.WaitVisible(`#pbEntrar`, chromedp.ByID),
		chromedp.Click(`#pbEntrar`, chromedp.ByID),
		chromedp.WaitVisible(`h1.esajTituloPagina`, chromedp.ByQuery),
		chromedp.Navigate("https://esaj.tjsp.jus.br/cpopg/open.do"),
		chromedp.WaitVisible(`a.linkLogo`, chromedp.ByQuery),
		// navigate through the searchDo page to extract the process.codigo, key to follow the next steps.
		chromedp.Navigate(searchDo),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Extracting the current URL, thi one contains the process.codigo, the necessary key to follow the next steps.
			var searchDoURLWithProcessCode string
			err = chromedp.Location(&searchDoURLWithProcessCode).Do(ctx)
			if err != nil {
				return fmt.Errorf("could not get the url: %v", err)
			}

			logger.Debug(fmt.Sprintf("searchDoURLWithProcessCode: %s", searchDoURLWithProcessCode))

			err = chromedp.Navigate(searchDoURLWithProcessCode).Do(ctx)
			if err != nil {
				return fmt.Errorf(fmt.Sprintf("could not navigate to search.do URL %s: %v", searchDoURLWithProcessCode, err))
			}

			u, err := url.Parse(searchDoURLWithProcessCode)
			if err != nil {
				return fmt.Errorf("could not parse searchDoURLWithProcessCode %s: %v", searchDoURLWithProcessCode, err)
			}

			processoCodigo := u.Query().Get("processo.codigo")
			if processoCodigo == "" {
				return fmt.Errorf("could not get process.codigo from searchDoURLWithProcessCode %s", searchDoURLWithProcessCode)
			}

			abrirPastaDigitalDoURL := abrirPastaDigitalDoURL(processoCodigo)

			// abrirPastaDigital.do is the page that retrieves the page where we can find all the pdfs of the process.
			// we need to get the HREF of this page to navigate to it. This because, each time that we access this page,
			// ESAJ generates a new "ticket"to access the page.
			err = chromedp.Navigate(abrirPastaDigitalDoURL).Do(ctx)
			if err != nil {
				return fmt.Errorf("could not navigate to abrirPastaDigitalDoURL: %v", err)
			}

			err = chromedp.WaitVisible(`body`, chromedp.ByQuery).Do(ctx)
			if err != nil {
				return fmt.Errorf("could not wait for body: %v", err)
			}

			// pastaVirtualBodyTxt saves the body text to parse the href to navigate to the pastaDigital
			var pastaVirtualBodyTxt string
			err = chromedp.Text(`body`, &pastaVirtualBodyTxt, chromedp.NodeVisible, chromedp.ByQuery).Do(ctx)
			if err != nil {
				return fmt.Errorf("could not get text from body: %v", err)
			}

			u, err = url.Parse(pastaVirtualBodyTxt)
			if err != nil {
				return fmt.Errorf("could not parse pastaVirtualBodyTxt: %v", err)
			}

			// pastaDigitalHREF parse the BodyText to a valid href to navigate to the pastadigital
			pastaDigitalHREF := u.RawQuery

			logger.Debug("parsed pasta digital href", "href", pastaDigitalHREF)

			cookies, err = navigatePastaVirtualURL(ctx, "https://esaj.tjsp.jus.br/pastadigital/abrirPastaProcessoDigital.do?"+pastaDigitalHREF)
			if err != nil {
				return fmt.Errorf("could not navigate to pastaVirtualURL: %v", err)
			}

			return nil
		}),
	)

	if err != nil {
		return "", "", fmt.Errorf("could not get cookies: %v", err)
	}

	cookieSession, cookiePDFSession := parseCookies(cookies)

	return cookieSession, cookiePDFSession, nil
}

// parseCookies receives a slice of cookies and returns two strings that contains the cookieSession and cookiePDFSession.
// each one is used in different types of http requests.
// the first string return is the cookieSession and the second is the cookiePDFSession
// cookiesSession example: "JSESSIONID=EACA3333A48456D7953B6331999A4F80.cas11; K-JSESSIONID-nckcjpip=0E4D006FFD78524DBABA78F02E1633FA"
// cookiesPDFSession example: "JSESSION=8A1F3DCE0D4DC510FFF3305E44ABCC4E.pasta3; K-JSESSIONID-phoaambo=0E4D006FFD78524DBABA78F02E1633FA"
func parseCookies(cookies []*network.Cookie) (string, string) {
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
	return cookieSession, cookiePDFSession
}

// showDoURL is the page that retreive the specific information about a process.
// - processoCodigo example: 1H000H91J0000. Important to mentioned that this ID does not have a defined pattern, it's a internal ID from the ESAJ
// the only thing that we can assume is that it is a string with 13 characters.
// - processoForo example: 53 or 0053
// - processID example: 1016358-63.2020.8.26.0053
func showDoURL(processoCodigo, processoForo, processID string) string {
	// The url.QueryEscape is used to escape the special characters to avoid errors.
	processoForo = url.QueryEscape(processoForo)
	processoCodigo = url.QueryEscape(processoCodigo)
	processID = url.QueryEscape(processID)

	return fmt.Sprintf("https://esaj.tjsp.jus.br/cpopg/show.do?processo.codigo=%s&processo.foro=%s&processo.numero=%s", processoCodigo, processoForo, processID)
}

// searchDoURL retrive the page that we need to access to get the processoCodigo.
// - processID example: 1016358-63.2020.8.26.0053
func searchDoURL(processID string) (string, error) {
	foro, err := foroNumeroUnificado(processID)
	if err != nil {
		return "", err
	}

	numDigAno, err := numeroDigitoAnoUnificado(processID)
	if err != nil {
		return "", err
	}

	//TODO(@perebaj): reduce this string to a more readable format.
	return fmt.Sprintf(`https://esaj.tjsp.jus.br/cpopg/search.do?conversationId=&cbPesquisa=NUMPROC&numeroDigitoAnoUnificado=%s&foroNumeroUnificado=%s&dadosConsulta.valorConsultaNuUnificado=%s&dadosConsulta.valorConsultaNuUnificado=UNIFICADO&dadosConsulta.valorConsulta=&dadosConsulta.tipoNuProcesso=UNIFICADO`, numDigAno, foro, processID),
		nil
}

// abrirPastaDigitalDoURL is the page that retreive all the pdfs documents of the process.
// - processoCodigo example: 1H000H91J0000. Important to mentioned that this ID does not have a defined pattern, it's a internal ID from the ESAJ
func abrirPastaDigitalDoURL(processoCodigo string) string {
	// The url.QueryEscape is used to escape the special characters to avoid errors.
	processoCodigo = url.QueryEscape(processoCodigo)

	return fmt.Sprintf("https://esaj.tjsp.jus.br/cpopg/abrirPastaDigital.do?processo.codigo=%s", processoCodigo)
}

func navigatePastaVirtualURL(ctx context.Context, pastaVirtualURL string) ([]*network.Cookie, error) {
	processID, err := getContextWithProcessID(ctx, ProcessIDContextKey)
	if err != nil {
		return nil, fmt.Errorf("error getting processID from context: %w", err)
	}

	logger := slog.With("processID", processID)
	logger.Debug("navigating to pastaVirtualURL", "url", pastaVirtualURL)
	err = chromedp.Navigate(pastaVirtualURL).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not navigate to pastaVirtualURL: %v", err)
	}

	err = chromedp.WaitVisible(`input#salvarButton`, chromedp.ByQuery).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not wait for input#salvarButton: %v", err)
	}

	cookies, err := storage.GetCookies().Do(ctx)

	return cookies, err
}
