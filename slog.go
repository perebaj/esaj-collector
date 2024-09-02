// Package esaj from slog.go aims to gather all the functions that initialize the structure logs.
package esaj

import (
	"fmt"
	"log/slog"
	"math"
	"os"
	"strings"
)

const (
	// LevelDebug is the debug level
	LevelDebug = "debug"
	// LevelInfo is the info level
	LevelInfo = "info"
	// LevelWarn is the warn level
	LevelWarn = "warn"
	// LevelError is the error level
	LevelError = "error"
)

const (
	// FormatLogFmt is the logfmt format
	FormatLogFmt = "logfmt"
	// FormatJSON is the JSON format
	FormatJSON = "json"
	// FormatGCP is the GCP format
	FormatGCP = "gcp"
)

// AvailableLogLevels is a list of supported logging levels
var AvailableLogLevels = []string{
	// LevelDebug is the debug level
	LevelDebug,
	// LevelInfo is the info level
	LevelInfo,
	// LevelWarn is the warn level
	LevelWarn,
	// LevelError is the error level
	LevelError,
}

// AvailableLogFormats is a list of supported log formats
var AvailableLogFormats = []string{
	// FormatLogFmt is the logfmt format
	FormatLogFmt,
	// FormatJSON is the JSON format
	FormatJSON,
	// FormatGCP is the GCP format
	FormatGCP,
}

// ConfigLogger is a struct that holds the configuration for the logger.
type ConfigLogger struct {
	Level  string
	Format string
}

// NewLoggerSlog returns a *slog.Logger that prints in the provided format at the
// provided level with a UTC timestamp and the caller of the log entry.
func NewLoggerSlog(c ConfigLogger) (*slog.Logger, error) {
	lvlOption, err := parseLevel(c.Level)
	if err != nil {
		return nil, err
	}

	handler, err := getHandlerFromFormat(c.Format, slog.HandlerOptions{
		Level:       lvlOption,
		AddSource:   true,
		ReplaceAttr: replaceSlogAttributes,
	})
	if err != nil {
		return nil, err
	}

	return slog.New(handler), nil
}

// replaceSlogAttributes replaces fields that were added by default by slog, but had different
// formats or key names in github.com/go-kit/log. The operator was originally implemented with go-kit/log,
// so we use these replacements to make the migration smoother.
func replaceSlogAttributes(_ []string, a slog.Attr) slog.Attr {
	if a.Key == "level" {
		return slog.Attr{
			Key:   "severity",
			Value: a.Value,
		}
	}

	if a.Key == "msg" {
		return slog.Attr{
			Key:   "message",
			Value: a.Value,
		}
	}

	return a
}

// getHandlerFromFormat returns a slog.Handler based on the provided format and slog options.
func getHandlerFromFormat(format string, opts slog.HandlerOptions) (slog.Handler, error) {
	var handler slog.Handler
	switch strings.ToLower(format) {
	case FormatLogFmt:
		handler = slog.NewTextHandler(os.Stdout, &opts)
		return handler, nil
	case FormatJSON:
		handler = slog.NewJSONHandler(os.Stdout, &opts)
		return handler, nil
	case FormatGCP:
		handler = slog.NewJSONHandler(os.Stdout, &opts)
		return handler, nil
	default:
		return nil, fmt.Errorf("log format %s unknown, %v are possible values", format, AvailableLogFormats)
	}
}

// parseLevel returns the slog.Level based on the provided string.
func parseLevel(lvl string) (slog.Level, error) {
	switch strings.ToLower(lvl) {
	case LevelDebug:
		return slog.LevelDebug, nil
	case LevelInfo:
		return slog.LevelInfo, nil
	case LevelWarn:
		return slog.LevelWarn, nil
	case LevelError:
		return slog.LevelError, nil
	default:
		return math.MaxInt, fmt.Errorf("log log_level %s unknown, %v are possible values", lvl, AvailableLogLevels)
	}
}
