// Package gpt from openai.go provides a client to interact with the OpenAI tools.
package gpt

import "fmt"

// Config struct to hold the configuration of the openai client
type Config struct {
	APIToken string
	OrgID    string
}

// Validate will check if the configuration is valid
func (c Config) Validate() error {
	if c.APIToken == "" {
		return fmt.Errorf("API token is required")
	}

	return nil
}
