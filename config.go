// Package esaj config.go gather functions to deal with configurations and environment variables.
// Don't know if is the best name, but it's a good start.
package esaj

import "os"

// GetEnvWithDefault returns the value of the environment variable key if it exists, otherwise it returns defaultValue.
// The validation if the env key must be filled or not is up to the caller.
func GetEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
