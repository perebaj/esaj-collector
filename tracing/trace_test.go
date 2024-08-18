package tracing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTraceContext(t *testing.T) {
	ctx := context.Background()
	traceID := "123456"
	ctx = SetTraceIDInContext(ctx, traceID)

	got := GetTraceIDFromContext(ctx)
	require.Equal(t, traceID, got)
}
