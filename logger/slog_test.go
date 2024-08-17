package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewLoggerSlog(t *testing.T) {
	l, err := NewLoggerSlog(ConfigLogger{
		Level:  LevelInfo,
		Format: FormatGCP,
	})

	require.NoError(t, err)

	require.NotNil(t, l)
	l.Info("test")
}
