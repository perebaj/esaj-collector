// Package gpt from chatgpt.go provides the functions to interact with the OpenAI tools.
package gpt

// SystemPrompt is the default system prompt for the system
const SystemPrompt = `
You will work like a unstructured text parser, I will give you a text and you will parse it according to the specified schema
and fields. You will return a JSON object.
An input example:

TEXT:
Expeça-se mandado de notificação da autoridade
administrativa,  para  cumprir  a  ordem  e  apresentar  as  informações,
no prazo de dez dias.
END TEXT

The fields that must be extracted are:

- deadline - integer - The deadline in days to comply with the order.

The output schema must look like this:
{
	"deadline": 10,

}

Do not include any fields that are not specified in the schema.
Do not Generate any fields that are not specified in the schema.
Do not generate any data if you don't have enough confidence in it.
`

// UserPrompt is the default user prompt for the system
const UserPrompt = `
TEXT
{{.Text}}
END TEXT
`

// type Client struct {
// 	client *openai.Client
// }

// func createUserPrompt() (string, error) {
// 	type userInput struct {
// 		Text string
// 	}

// 	tmpl, err := template.New("systemPrompt").Parse(userPrompt)
// 	if err != nil {
// 		return "", err
// 	}

// 	var input userInput

// }
