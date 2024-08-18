// Package tracing turn easy to save and retrieve the traceID from the context.
package tracing

import "context"

type contextKey string

var (
	// TraceIDContextKey is a context key to store the traceID
	TraceIDContextKey = contextKey("traceID")
)

// GetTraceIDFromContext retrieves the traceID from the context
func GetTraceIDFromContext(ctx context.Context) string {
	traceID, ok := ctx.Value(TraceIDContextKey).(string)
	if !ok {
		return ""
	}
	return traceID
}

// SetTraceIDInContext sets the traceID in the context
func SetTraceIDInContext(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDContextKey, traceID)
}
