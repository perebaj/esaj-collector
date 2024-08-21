//go:build integration

// Obs: It's recommended to run this tests using the command line located in the Makefile,
// this because all variable injection and important configurations are made there.
// Example: make integration-test testcase=<>
package firestore_test

import (
	"context"
	"testing"

	fs "cloud.google.com/go/firestore"
	"github.com/perebaj/esaj/esaj"
	"github.com/perebaj/esaj/firestore"
	"github.com/perebaj/esaj/tracing"
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
	ctx = tracing.SetTraceIDInContext(ctx, "test-trace-id")

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
	for i, doc := range docs {
		var got map[string]interface{}
		doc.DataTo(&got)

		require.Equal(t, ps[i].ProcessID, got["process_id"])
		require.Equal(t, ps[i].OAB, got["oab"])
		require.Equal(t, ps[i].URL, got["url"])
		require.Equal(t, "test-trace-id", got["trace_id"])
	}

	// update a document that already exists
	ps[0].URL = "http://example.com/updated"
	err = storage.SaveProcessSeeds(ctx, ps)
	require.NoError(t, err)

	iter = collection.Documents(ctx)
	docs, err = iter.GetAll()

	require.Len(t, docs, 3)
	for i, doc := range docs {
		var got map[string]interface{}
		doc.DataTo(&got)

		require.Equal(t, ps[i].URL, got["url"])
	}
}

func TestStorage_GetSeedsByOAB(t *testing.T) {
	ps := []esaj.ProcessSeed{
		{
			ProcessID: "123",
			OAB:       "123",
			URL:       "http://teste.com",
		},
		{
			ProcessID: "456",
			OAB:       "456",
			URL:       "http://example.com",
		},
		{
			ProcessID: "789",
			OAB:       "123",
			URL:       "http://teste1.com",
		},
	}

	ctx := context.TODO()

	c, err := fs.NewClient(ctx, projectID)
	defer cleanup(t, c)

	require.NoError(t, err)
	storage := firestore.NewStorage(c, projectID)
	err = storage.SaveProcessSeeds(ctx, ps)

	require.NoError(t, err)

	got, err := storage.GetSeedsByOAB(ctx, "123")
	require.NoError(t, err)

	require.Len(t, got, 2)

	require.Equal(t, ps[0].ProcessID, got[0].ID)
	require.Equal(t, ps[0].ProcessID, got[0].ProcessID)
	require.Equal(t, ps[0].OAB, got[0].OAB)
	require.Equal(t, ps[0].URL, got[0].URL)
	require.NotNil(t, got[0].CreatedAt)
	require.NotNil(t, got[0].UpdatedAt)

	require.Equal(t, ps[2].ProcessID, got[1].ID)
	require.Equal(t, ps[2].ProcessID, got[1].ProcessID)
	require.Equal(t, ps[2].OAB, got[1].OAB)
	require.Equal(t, ps[2].URL, got[1].URL)
	require.NotNil(t, got[1].CreatedAt)
	require.NotNil(t, got[1].UpdatedAt)
}

func TestStorage_SaveProcessBasicInfo(t *testing.T) {
	pBasicInfo := esaj.ProcessBasicInfo{
		ProcessID:   "123",
		ProcessForo: "123",
		ForoName:    "http://teste.com",
		ProcessCode: "456",
		Judge:       "http://example.com",
		Class:       "123",
		Claimant:    "http://teste1.com",
		Defendant:   "123",
		Vara:        "http://teste.com",
		URL:         "http://example.com",
	}

	ctx := context.TODO()
	ctx = tracing.SetTraceIDInContext(ctx, "test-trace-id")

	c, err := fs.NewClient(ctx, projectID)
	defer cleanup(t, c)

	require.NoError(t, err)
	storage := firestore.NewStorage(c, projectID)

	err = storage.SaveProcessBasicInfo(ctx, pBasicInfo)
	require.NoError(t, err)

	collection := c.Collection("process_basic_info")

	iter := collection.Documents(ctx)
	docs, err := iter.GetAll()
	require.NoError(t, err)

	require.Len(t, docs, 1)

	var got map[string]interface{}
	docs[0].DataTo(&got)

	require.Equal(t, pBasicInfo.ProcessID, got["process_id"])
	require.Equal(t, pBasicInfo.ProcessForo, got["foro_code"])
	require.Equal(t, pBasicInfo.ForoName, got["foro_name"])
	require.Equal(t, pBasicInfo.ProcessCode, got["process_code"])
	require.Equal(t, pBasicInfo.Judge, got["judge"])
	require.Equal(t, pBasicInfo.Class, got["class"])
	require.Equal(t, pBasicInfo.Claimant, got["claimant"])
	require.Equal(t, pBasicInfo.Defendant, got["defendant"])
	require.Equal(t, pBasicInfo.Vara, got["vara"])
	require.Equal(t, "test-trace-id", got["trace_id"])
	require.Equal(t, pBasicInfo.URL, got["url"])

	// update a field to validate if the document is updated
	pBasicInfo.ForoName = "updated value"

	err = storage.SaveProcessBasicInfo(ctx, pBasicInfo)
	require.NoError(t, err)

	iter = collection.Documents(ctx)
	docs, err = iter.GetAll()
	require.NoError(t, err)

	require.Len(t, docs, 1)
	require.Equal(t, "updated value", docs[0].Data()["foro_name"])
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
