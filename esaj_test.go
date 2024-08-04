package esaj

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_numeroDigitoAnoUnificado(t *testing.T) {
	processID := "1029989-06.2022.8.26.0053"
	want := "1029989-06.2022"

	got, err := numeroDigitoAnoUnificado(processID)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func Test_foroNumeroUnificado(t *testing.T) {
	processID := "1029989-06.2022.8.26.0053"
	want := "0053"

	got, err := foroNumeroUnificado(processID)
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

// Passing a invalid body to the request response, this should return an error
// saying that no matches were found.
func Test_ESAJClient_SearchDo_ErrNoMatchesFound(t *testing.T) {
	esajClient := NewESAJClient(Config{}, &http.Client{
		Timeout: time.Second * 2,
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check the request method
		if r.Method != http.MethodGet {
			t.Errorf("expected %s, got %s", http.MethodGet, r.Method)
		}

		// check the request URL
		if r.URL.Path != "/cpopg/search.do" {
			t.Errorf("expected %s, got %s", "/cpopg/search.do", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(``))
	}))

	esajClient.URL = server.URL

	processID := "1029989-06.2022.8.26.0053"
	_, err := esajClient.SearchDo(processID)
	require.Error(t, err)
}

func Test_ESAJClient_SearchDo(t *testing.T) {
	esajClient := NewESAJClient(Config{}, &http.Client{
		Timeout: time.Second * 2,
	})

	bodyHTML := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<title>Test Document</title>
	</head>
	<body>
		<table>
			<tr>
				<td>
					<a class="linkMovVincProc" href="processo.codigo=NOTTHISONE">Document 1</a>
				</td>
			</tr>
			<tr>
				<td>
					<a class="linkMovVincProc" href="abrirDocumentoVinculadoMovimentacao.do?processo.codigo=THISONE">Document 2</a>
				</td>
			</tr>
		</table>
	</body>
	</html>
	`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected %s, got %s", http.MethodGet, r.Method)
		}

		if r.URL.Path != "/cpopg/search.do" {
			t.Errorf("expected %s, got %s", "/cpopg/search.do", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(bodyHTML))
	}))

	esajClient.URL = server.URL

	processID := "1029989-06.2022.8.26.0053"
	got, err := esajClient.SearchDo(processID)
	require.NoError(t, err)

	wantProcessCode := "THISONE"
	assert.Equal(t, wantProcessCode, got)
}

func Test_ESAJClient_pastaDigitalURL_NoLinkFound(t *testing.T) {
	esajClient := NewESAJClient(Config{
		CookieSession: "fake-cookie-session",
	}, &http.Client{
		Timeout: time.Second * 2,
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected %s, got %s", http.MethodGet, r.Method)
		}

		if r.URL.Path != "/cpopg/abrirPastaDigital.do" {
			t.Errorf("expected %s, got %s", "/cpopg/abrirPastaDigital.do", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(``))
	}))

	esajClient.URL = server.URL

	processCode := "PROCESSCODE"
	_, err := esajClient.pastaDigitalURL(processCode)
	require.Error(t, err)

	want := "no link found"
	assert.Equal(t, want, err.Error())
}

func Test_ESAJClient_pastaDigitalURL_invalidAccess(t *testing.T) {
	esajClient := NewESAJClient(Config{
		CookieSession: "fake-cookie-session",
	}, &http.Client{
		Timeout: time.Second * 2,
	})

	bodyHTML := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<title>Test Document</title>
	</head>
	<body>
		Não foi possível validar o seu acesso
	</body>
	</html>
	`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected %s, got %s", http.MethodGet, r.Method)
		}

		if r.URL.Path != "/cpopg/abrirPastaDigital.do" {
			t.Errorf("expected %s, got %s", "/cpopg/abrirPastaDigital.do", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(bodyHTML))
	}))

	esajClient.URL = server.URL

	processCode := "PROCESSCODE"
	_, err := esajClient.pastaDigitalURL(processCode)
	require.Error(t, err)

	want := "access not validated, verify the COOKIESESSION"
	assert.Equal(t, want, err.Error())
}

func Test_ESAJClient_AbrirPastaDigital(t *testing.T) {
	// the server should mock 2 requests:
	// /cpopg/abrirPastaDigital.do
	// /pastadigital/abrirPastaProcessoDigital.do?

	esajClient := NewESAJClient(Config{
		CookieSession: "fake-cookie-session",
	}, &http.Client{
		Timeout: time.Second * 2,
	})

	bodyHTML := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<title>Test Document</title>
	</head>
	<body>https://esaj.tjsp.jus.br/pastadigital/abrirPastaProcessoDigital.do</body>
	</html>
	`

	bodyHTML = strings.ReplaceAll(bodyHTML, "\n", "")
	bodyHTML = strings.ReplaceAll(bodyHTML, "\t", "")

	bodyHTML2 := `
	<html style="overflow: hidden">
	<head>
		<script type="text/javascript">
			var requestScope = [{
				"data": {
					"cdProcessoMaster": null,
					"cdDocumento": "294392168",
					"cdUsuCadastrante": null,
					"dtInclusao": "24\/01\/2024 16:19:49",
					"icon": false,
					"title": "Petição (Outras)",
					"cdTipoDocDigital": "9500",
					"cdProcessoPrinc": null,
					"nuProcessoMaster": null,
					"flPeticaoInicial": true,
					"cdFormatoDoc": 9,
					"deSituacaoProcesso": null,
					"sigiloAbsoluto": false,
					"deSituacaoProcessoMaster": null,
					"flProtocolado": true,
					"cdProcessoOrigem": null
				},
				"children": [{
					"data": {
						"nuPaginas": 16,
						"id_paginacao": 0,
						"icon": false,
						"iconesAss": [{
							"imagem": "logo_cliente.png",
							"alt": "assinado.PNG"
						}],
						"indicePagina": 1,
						"title": "Páginas 1 - 16",
						"parametros": "nuSeqRecurso=00000&nuProcesso=1004257-52.2024.8.26.0053&cdDocumentoOrigem=0&cdDocumento=294392168&conferenciaDocEdigOriginal=false&nmAlias=PG5JM&origemDocumento=P&nuPagina=1&numInicial=1&tpOrigem=2&cdTipoDocDigital=9500&flOrigem=P&deTipoDocDigital=Peti%E7%E3o+%28Outras%29&cdProcesso=1H000QWJM0000&cdFormatoDoc=9&cdForo=53&idDocumento=294392168-1-1&numFinal=16&sigiloExterno=N",
						"contexto": [],
						"tramitacao": null,
						"urlMidiaDigital": null,
						"possuiDocumentoOriginal": false,
						"materializar": false,
						"flProcVirtual": false,
						"documentoSigiloso": false,
						"paginaInicial": false,
						"flAssinado": true
					},
					"attributes": {
						"id": null
					}
				}],
				"id_paginacao": 1,
				"materializar": false,
				"attributes": {
					"ID": "ignorarRaiz"
				}
			}];
		</script>
	</head>
	</html>`

	bodyHTML2 = strings.ReplaceAll(bodyHTML2, "\n", "")
	bodyHTML2 = strings.ReplaceAll(bodyHTML2, "\t", "")

	mux := http.NewServeMux()

	mux.HandleFunc("/cpopg/abrirPastaDigital.do", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(bodyHTML))
	})

	mux.HandleFunc("/pastadigital/abrirPastaProcessoDigital.do", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(bodyHTML2))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	esajClient.URL = server.URL

	processCode := "PROCESSCODE"

	processes, err := esajClient.AbrirPastaProcessoDigital(processCode)
	require.NoError(t, err)

	assert.Len(t, processes, 1)
}

func Test_ESAJClient_pastaDigitalURL(t *testing.T) {
	esajClient := NewESAJClient(Config{
		CookieSession: "fake-cookie-session",
	}, &http.Client{
		Timeout: time.Second * 2,
	})

	bodyHTML := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<title>Test Document</title>
	</head>
	<body>
	<text>https://esaj.tjsp.jus.br/pastadigital/abrirPastaProcessoDigital.do</text>
	</body>
	</html>
	`

	bodyHTML = strings.ReplaceAll(bodyHTML, "\n", "")
	bodyHTML = strings.ReplaceAll(bodyHTML, "\t", "")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected %s, got %s", http.MethodGet, r.Method)
		}

		if r.URL.Path != "/cpopg/abrirPastaDigital.do" {
			t.Errorf("expected %s, got %s", "/cpopg/abrirPastaDigital.do", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(bodyHTML))
	}))

	esajClient.URL = server.URL

	processCode := "PROCESSCODE"
	got, err := esajClient.pastaDigitalURL(processCode)
	require.NoError(t, err)

	assert.Contains(t, got, "/pastadigital/abrirPastaProcessoDigital.do")
}

func Test_getContextWithProcessID(t *testing.T) {
	processID := "1029989-06.2022.8.26.0053"

	ctx := context.Background()
	ctx = context.WithValue(ctx, ProcessIDContextKey, processID)

	got, err := getContextWithProcessID(ctx, "processID")
	require.NoError(t, err)

	assert.Equal(t, processID, got)

	_, err = getContextWithProcessID(ctx, "invalidKey")
	require.Error(t, err)

	// using the typed contextKey
	got, err = getContextWithProcessID(ctx, ProcessIDContextKey)
	require.NoError(t, err)

	assert.Equal(t, processID, got)
}
