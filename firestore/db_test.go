//go:build integration

package firestore_test

import (
	"context"
	"testing"

	fs "cloud.google.com/go/firestore"
	"github.com/stretchr/testify/require"
)

func TestPingFireStore(t *testing.T) {
	c, err := fs.NewClient(context.TODO(), "test-project")
	require.NoError(t, err)

	//create a simple document to test the connection
	_, err = c.Collection("test").Doc("test").Set(context.TODO(), map[string]interface{}{"test": "test"})
	require.NoError(t, err)
}
