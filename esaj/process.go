// Package esaj from process.go follow the same naming convention as the original API.
package esaj

// Process ...
type Process struct {
	Children []Children `json:"children"`
	Data     Data       `json:"data"`
}

// Data ...
type Data struct {
	CdProcessoMaster         any    `json:"cdProcessoMaster"`
	CdDocumento              string `json:"cdDocumento"`
	CdUsuCadastrante         any    `json:"cdUsuCadastrante"`
	DtInclusao               string `json:"dtInclusao"`
	Icon                     bool   `json:"icon"`
	Title                    string `json:"title"`
	CdTipoDocDigital         string `json:"cdTipoDocDigital"`
	CdProcessoPrinc          any    `json:"cdProcessoPrinc"`
	NuProcessoMaster         any    `json:"nuProcessoMaster"`
	FlPeticaoInicial         bool   `json:"flPeticaoInicial"`
	CdFormatoDoc             int    `json:"cdFormatoDoc"`
	DeSituacaoProcesso       any    `json:"deSituacaoProcesso"`
	SigiloAbsoluto           bool   `json:"sigiloAbsoluto"`
	DeSituacaoProcessoMaster any    `json:"deSituacaoProcessoMaster"`
	FlProtocolado            bool   `json:"flProtocolado"`
	CdProcessoOrigem         any    `json:"cdProcessoOrigem"`
}

// Children ...
type Children struct {
	ChildernData ChildrenData       `json:"data"`
	IDPaginacao  int                `json:"id_paginacao"`
	Materializar bool               `json:"materializar"`
	Attributes   ChildrenAttributes `json:"attributes"`
}

// ChildrenAttributes ...
type ChildrenAttributes struct {
	ID any `json:"id"`
}

// ChildrenData ...
type ChildrenData struct {
	NuPaginas   int  `json:"nuPaginas"`
	IDPaginacao int  `json:"id_paginacao"`
	Icon        bool `json:"icon"`
	IconesAss   []struct {
		Imagem string `json:"imagem"`
		Alt    string `json:"alt"`
	} `json:"iconesAss"`
	IndicePagina            int    `json:"indicePagina"`
	Title                   string `json:"title"`
	Parametros              string `json:"parametros"`
	Contexto                []any  `json:"contexto"`
	Tramitacao              any    `json:"tramitacao"`
	URLMidiaDigital         any    `json:"urlMidiaDigital"`
	PossuiDocumentoOriginal bool   `json:"possuiDocumentoOriginal"`
	Materializar            bool   `json:"materializar"`
	FlProcVirtual           bool   `json:"flProcVirtual"`
	DocumentoSigiloso       bool   `json:"documentoSigiloso"`
	PaginaInicial           bool   `json:"paginaInicial"`
	FlAssinado              bool   `json:"flAssinado"`
}

// ProcessBasicInfo as the name says, is the basic information of a process.
type ProcessBasicInfo struct {
	// OAB is the OAB of the process.
	OAB string
	// ProcessID example: "1007573-30.2024.8.26.0229"
	ProcessID string
	// ProcessForo example: "0053"
	ProcessForo string
	// ForoName example: "Tribunal de Justiça do Estado de São Paulo"/"Foro de Hortolândia"
	ForoName string
	// ProcessCode. Example: 6D0008MAZ0000
	ProcessCode string
	// Judge...
	Judge string
	// Class is the class of the process. Example: "Habilitação de Crédito"
	Class string
	// Claimant is who is claiming for something in the process.
	Claimant string
	// Defendant is who is being claimed in the process.
	Defendant string
	// Vara is the court where the process is being processed.
	Vara string
	// URL is the URL of the process in the TJSP website.
	// Example: https://esaj.tjsp.jus.br/cpopg/show.do?processo.codigo=1HZX5Q48A0000&processo.foro=53&paginaConsulta=17&cbPesquisa=NUMOAB&dadosConsulta.valorConsulta=103289&cdForo=-1
	URL string
}
