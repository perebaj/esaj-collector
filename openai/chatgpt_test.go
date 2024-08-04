package gpt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUserPrompt(t *testing.T) {
	actual, err := createUserPrompt("hello world")
	require.NoError(t, err)

	actual = removeSpaces(t, actual)

	expected := `
	TEXT
	hello world
	END TEXT`

	expected = removeSpaces(t, expected)

	assert.Equal(t, expected, actual)
}

func removeSpaces(t *testing.T, s string) string {
	t.Helper()
	s = strings.ReplaceAll(s, "\t", "")
	s = strings.ReplaceAll(s, "\n", "")

	return s
}
