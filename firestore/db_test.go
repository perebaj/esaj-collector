//go:build integration

package firestore_test

import (
	"context"
	"testing"

	fs "cloud.google.com/go/firestore"
	"github.com/perebaj/esaj"
	"github.com/perebaj/esaj/firestore"
	"github.com/stretchr/testify/require"
)

// projectID is a mock projectID used for testing
// this const not represent a real projectID, but must be used acc
const projectID = "test-project"

func TestPingFireStore(t *testing.T) {
	c, err := fs.NewClient(context.TODO(), projectID)
	require.NoError(t, err)
	defer cleanup(t, c)

	_, err = c.Collection("test").Doc("test").Set(context.TODO(), map[string]interface{}{"test": "test"})
	require.NoError(t, err)
}

func TestStorage_SaveProcessSeeds(t *testing.T) {
	ps := []esaj.ProcessSeed{
		{
			ProcessID: "123",
			OAB:       "123",
			URL:       "http://example.com",
		},
		{
			ProcessID: "456",
			OAB:       "456",
			URL:       "http://example.com",
		},
		{
			ProcessID: "789",
			OAB:       "789",
			URL:       "http://example.com",
		},
	}

	ctx := context.TODO()

	c, err := fs.NewClient(ctx, projectID)
	defer cleanup(t, c)

	require.NoError(t, err)
	storage := firestore.NewStorage(c, projectID)
	err = storage.SaveProcessSeeds(ctx, ps)
	require.NoError(t, err)

	collection := c.Collection("process_seeds")

	iter := collection.Documents(ctx)
	docs, err := iter.GetAll()
	require.NoError(t, err)

	require.Len(t, docs, 3)
}

// cleanup deletes all collections and documents in the firestore database
// it must be called in all tests that uses the firestore database
func cleanup(t *testing.T, c *fs.Client) {
	ctx := context.TODO()
	iter := c.Collections(ctx)
	for {
		collection, err := iter.Next()
		if err != nil {
			break
		}
		iter := collection.Documents(ctx)
		for {
			doc, err := iter.Next()
			if err != nil {
				break
			}
			_, err = doc.Ref.Delete(ctx)
			require.NoError(t, err)
		}
	}
}
