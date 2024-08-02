package esaj

import (
	"testing"
)

func Test_showDoURL(t *testing.T) {
	got := showDoURL("123456", "São Paulo", "7890")
	want := "https://esaj.tjsp.jus.br/cpopg/show.do?processo.codigo=123456&processo.foro=S%C3%A3o+Paulo&processo.numero=7890"
	if want != got {
		t.Errorf("showDoURL was incorrect, got: %s, want: %s.", got, want)
	}

	got = showDoURL("987654", "Campinas", "54321")
	want = "https://esaj.tjsp.jus.br/cpopg/show.do?processo.codigo=987654&processo.foro=Campinas&processo.numero=54321"
	if want != got {
		t.Errorf("showDoURL was incorrect, got: %s, want: %s.", got, want)
	}
}

func Test_abrirPastaDigitalDoURL(t *testing.T) {
	got := abrirPastaDigitalDoURL("123456")

	want := "https://esaj.tjsp.jus.br/cpopg/abrirPastaDigital.do?processo.codigo=123456"
	if want != got {
		t.Errorf("abrirPastaDigitalDoURL was incorrect, got: %s, want: %s.", got, want)
	}

	got = abrirPastaDigitalDoURL("987 654")
	want = "https://esaj.tjsp.jus.br/cpopg/abrirPastaDigital.do?processo.codigo=987+654"

	if want != got {
		t.Errorf("abrirPastaDigitalDoURL was incorrect, got: %s, want: %s.", got, want)
	}
}

func Test_searchDoURL(t *testing.T) {
	processID := "1029989-06.2022.8.26.0053"
	got, err := searchDoURL(processID)
	if err != nil {
		t.Errorf("searchDoURL was incorrect, got: %s, want: nil.", err)
	}

	want := "https://esaj.tjsp.jus.br/cpopg/search.do?conversationId=&cbPesquisa=NUMPROC&numeroDigitoAnoUnificado=1029989-06.2022&foroNumeroUnificado=0053&dadosConsulta.valorConsultaNuUnificado=1029989-06.2022.8.26.0053&dadosConsulta.valorConsultaNuUnificado=UNIFICADO&dadosConsulta.valorConsulta=&dadosConsulta.tipoNuProcesso=UNIFICADO"

	if want != got {
		t.Errorf("searchDoURL was incorrect, got: %s, want: %s.", got, want)
	}
}
