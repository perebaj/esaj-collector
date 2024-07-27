package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Config struct {
	JSESSIONID string
}

func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func searchDo(jsession string) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://esaj.tjsp.jus.br/cpopg/search.do?conversationId=&cbPesquisa=NUMPROC&numeroDigitoAnoUnificado=1029989-06.2022&foroNumeroUnificado=0053&dadosConsulta.valorConsultaNuUnificado=1029989-06.2022.8.26.0053&dadosConsulta.valorConsultaNuUnificado=UNIFICADO&dadosConsulta.valorConsulta=&dadosConsulta.tipoNuProcesso=UNIFICADO", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Cookie", jsession)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(bodyText)))

	if err != nil {
		log.Fatal(err)
	}

	var link string
	doc.Find("tr > td > a.linkMovVincProc").EachWithBreak(func(i int, s *goquery.Selection) bool {
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
		log.Fatal("No matches found")
	}

	processCode := matches[1]
	log.Println(processCode)
}

func main() {
	jsession := getEnvWithDefault("JSESSIONID", "")

	if jsession == "" {
		log.Fatal("The JSESSIONID environment variable is required")
	}

	searchDo(jsession)
}
