// Package gpt from chatgpt.go provides the functions to interact with the OpenAI tools.
package gpt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"text/template"

	"github.com/sashabaranov/go-openai"
)

// systemPrompt is the default system prompt for the system
const systemPrompt = `
You will work like a unstructured text parser, I will give you a text and you will parse it according to the specified schema
and fields. You will return a JSON object.
An input example:

TEXT:
# CERTIDÃO DE PUBLICAÇÃO DE RELAÇÃO

Certifico e dou fé que o ato abaixo, constante da relação nº 0262/2021, foi disponibilizado na página 1454/1476 do Diário de Justiça Eletrônico em 27/10/2021. Considera-se a data de publicação em 28/10/2021, primeiro dia útil subsequente à data de disponibilização.

Certifico, ainda, que para efeito de contagem do prazo foram consideradas as seguintes datas.

- 29/10/2021 - Dia do Funcionário Público (Provimento CSM 2631/2021) - Prorrogação
- 01/11/2021 à 01/11/2021 - Suspensão de expediente Prov. CSM n° 2584/2020 - Suspensão
- 02/11/2021 - Finados - Prorrogação

Advogado

- Jonathan Santos Silva (OAB 000000/SP)
- Jonathan Silva Santos (OAB 000000/SP)

Teor do ato: "Vistos. Presentes os requisitos legais, defere-se o pedido de liminar para determinar que a autoridade impetrada aprecie, no prazo de 10 dias, o pedido administrativo formulado pela parte impetrante. Com efeito, o administrado tem direito a obter resposta aos seus pedidos em prazo razoável, em homenagem aos princípios constitucionais da razoável duração do processo, da celeridade e da eficiência. Notifique (m)-se o(s) coator(es), supracitado(s) e no(s) endereço (s) indicado(s), do conteúdo da petição inicial, entregando-lhe(s) a senha de acesso ao processo digital, a fim de que, no prazo de dez dias, preste(m) informações (art. 7º, inciso I da Lei nº 12.016/09). Advirta-se que, nos termos do Comunicado CG nº 879/2016, relativamente aos processos digitais, é obrigatório o uso do formato digital, seja por meio do peticionamento eletrônico pelos órgãos de representação judicial (a ser preferencialmente utilizado), seja por meio do e-mail institucional da Unidade Cartorária onde tramita o feito (sp7faz@tjsp.jus.br). Após, cumpra-se o artigo 7º, II, da Lei n° 12.016/09, intimando-se a Fazenda Pública do Estado de São Paulo pelo portal eletrônico, nos termos do Comunicado Conjunto n° 2536/2017 (Protocolo CPA n° 2016/44379). Findo o prazo, ouça-se o representante do Ministério Público, em dez dias. Oportunamente, tornem conclusos para decisão. Cumpra-se, na forma e sob as penas da Lei, servindo esta decisão como mandado e ofício que poderá, se o caso, ser encaminhado pela parte interessada, nos termos do item 3.b. do Comunicado Conjunto n° 37/2020. Int."

SÃO PAULO, 27 de outubro de 2021.

Marcelo Santos Silva Este documento é cópia do original, assinado digitalmente por Marcelo Santos Silva, liberado nos autos em 27/10/2021 às 14:08 .Escrevente Técnico JudiciárioPara conferir o original, acesse o site https://esaj.tjsp.jus.br/pastadigital/pg/abrirConferenciaDocumento.do, informe o processo 1064759-59.2021.8.26.0053 e código i4kzcPL5.
END TEXT


The output schema must look like this:
{
    "court": "",
    "judge": "",
    "order": "",
    "defendant": [],
    "processID": "1064759-59.2021.8.26.0053",
    "dateIssued": "2021-10-22",
    "notificationDeadlineDays": "10"
}

The order is the summary of the order. Act like a summarizer to synthesize the order. No need to include the full text of the order.
The court is the court that issued the order.
The judge is the judge that issued the order.
The defendant is an array of the defendants in the order. It could be the lawyers or the parties involved.
The processID is the unique identifier of the process. It must be in the format XXXXXXX-XX.XXXX.X.XX.XXXX.
The dateIssued is the date that the order was issued. It must be in the format YYYY-MM-DD.
The notificationDeadlineDays is the number of days that the defendant has to respond to the order. It must be a number.

Do not include any fields that are not specified in the schema.
Do not Generate any fields that are not specified in the schema.
Do not generate any data if you don't have enough confidence in it.
`

// userPrompt is the default user prompt for the system
const userPrompt = `
TEXT
{{.Text}}
END TEXT
`

// Client is the client to interact with the OpenAI API
type Client struct {
	client *openai.Client
	config Config
}

// New initializes a new OpenAI client
func New(cfg Config) *Client {
	c := openai.NewClient(cfg.APIToken)
	return &Client{
		client: c,
		config: cfg,
	}
}

// ParsedPublication is the struct that represents the parsed publication from the raw text input
// this structure is defined in the system prompt. So, to change it. First, you need to change the system prompt.
type ParsedPublication struct {
	// Court is the institution that issued the order
	Court string `json:"court"`
	Judge string `json:"judge"`
	// Order is the summary of the judge's order
	Order string `json:"order"`
	// Defendant is the list of defendants in the order. The guys that will be notified about the new decision
	Defendant []string `json:"defendant"`
	// ProcessID is the unique identifier of the process. In the format XXXXXXX-XX.XXXX.X.XX.XXXX
	ProcessID                string `json:"processID"`
	DateIssued               string `json:"dateIssued"`
	NotificationDeadlineDays string `json:"notificationDeadlineDays"`
}

// ParsePublication receive the raw text of a publication and returns the parsed publication
func (c *Client) ParsePublication(ctx context.Context, text string) (*ParsedPublication, error) {
	prompt, err := createUserPrompt(text)
	if err != nil {
		return nil, fmt.Errorf("error creating user prompt: %v", err)
	}

	request := openai.ChatCompletionRequest{
		Model:       openai.GPT3Dot5Turbo,
		Temperature: math.SmallestNonzeroFloat32,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	resp, err := c.client.CreateChatCompletion(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("error creating chat completion: %v", err)
	}

	respContent := resp.Choices[0].Message.Content

	var parsedPublication ParsedPublication
	err = json.Unmarshal([]byte(respContent), &parsedPublication)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}

	return &parsedPublication, nil
}

func createUserPrompt(text string) (string, error) {
	type userInput struct {
		Text string
	}

	tmpl, err := template.New("parser").Parse(userPrompt)
	if err != nil {
		return "", fmt.Errorf("template parse error: %v", err)
	}

	headerParserInput := userInput{
		Text: text,
	}
	out := bytes.Buffer{}

	err = tmpl.Execute(&out, headerParserInput)
	if err != nil {
		return "", fmt.Errorf("template execute error: %v", err)
	}

	return out.String(), nil
}
