//go:build integration

// Obs: It's recommended to run this tests using the command line located in the Makefile,
// this because all variable injection and important configurations are made there.
// Example: make integration-test testcase=<>
package firestore_test

import (
	"context"
	"testing"
	"time"

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
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	ctx = tracing.SetTraceIDInContext(ctx, "test-trace-id")

	c, err := fs.NewClient(ctx, projectID)
	defer cleanup(t, c)

	require.NoError(t, err)
	storage := firestore.NewStorage(c, projectID)

	pBasicInfo := esaj.ProcessBasicInfo{
		ProcessID:   "123",
		ProcessForo: "123",
		ForoName:    "Foro Test",
		ProcessCode: "456",
		Judge:       "Judge Test",
		Class:       "Class Test",
		Claimant:    "Claimant Test",
		Defendant:   "Defendant Test",
		Vara:        "Vara Test",
		URL:         "http://example.com",
		OAB:         "OAB123",
	}

	// Test initial save
	err = storage.SaveProcessBasicInfo(ctx, pBasicInfo)
	require.NoError(t, err)

	collection := c.Collection("process_basic_info")
	doc, err := collection.Doc(pBasicInfo.ProcessID).Get(ctx)
	require.NoError(t, err)
	var got map[string]interface{}
	err = doc.DataTo(&got)
	require.NoError(t, err)

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
	require.Len(t, got["oabs"], 1)

	// Test update with same OAB
	pBasicInfo.ForoName = "Updated Foro"
	err = storage.SaveProcessBasicInfo(ctx, pBasicInfo)
	require.NoError(t, err)

	doc, err = collection.Doc(pBasicInfo.ProcessID).Get(ctx)
	require.NoError(t, err)
	err = doc.DataTo(&got)
	require.NoError(t, err)

	require.Equal(t, "Updated Foro", got["foro_name"])
	require.Len(t, got["oabs"], 1)

	// Test update with new OAB
	pBasicInfo.OAB = "OAB456"
	err = storage.SaveProcessBasicInfo(ctx, pBasicInfo)
	require.NoError(t, err)

	doc, err = collection.Doc(pBasicInfo.ProcessID).Get(ctx)
	require.NoError(t, err)
	err = doc.DataTo(&got)
	require.NoError(t, err)

	require.Len(t, got["oabs"], 2)

	// Test update with existing OAB (should not add duplicate)
	pBasicInfo.OAB = "OAB123"
	err = storage.SaveProcessBasicInfo(ctx, pBasicInfo)
	require.NoError(t, err)

	doc, err = collection.Doc(pBasicInfo.ProcessID).Get(ctx)
	require.NoError(t, err)
	err = doc.DataTo(&got)
	require.NoError(t, err)

	require.Len(t, got["oabs"], 2)
}

func TestStorage_ProcessBasicInfoByOAB(t *testing.T) {
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
		OAB:         "123",
	}

	err = storage.SaveProcessBasicInfo(ctx, pBasicInfo)
	require.NoError(t, err)

	pBasicInfo2 := esaj.ProcessBasicInfo{
		ProcessID:   "789",
		ProcessForo: "123",
		ForoName:    "http://teste.com",
		ProcessCode: "456",
		Judge:       "http://example.com",
		Class:       "123",
		Claimant:    "http://teste1.com",
		Defendant:   "123",
		Vara:        "http://teste.com",
		URL:         "http://example.com",
		OAB:         "123",
	}

	err = storage.SaveProcessBasicInfo(ctx, pBasicInfo2)
	require.NoError(t, err)

	got, err := storage.ProcessBasicInfoByOAB(ctx, "123")
	require.NoError(t, err)

	require.Len(t, got, 2)

	require.Equal(t, pBasicInfo.ProcessID, got[0].ProcessID)
	require.Equal(t, pBasicInfo.ProcessForo, got[0].ProcessForo)
	require.Equal(t, pBasicInfo.ForoName, got[0].ForoName)
	require.Equal(t, pBasicInfo.ProcessCode, got[0].ProcessCode)
	require.Equal(t, pBasicInfo.Judge, got[0].Judge)
	require.Equal(t, pBasicInfo.Class, got[0].Class)
	require.Equal(t, pBasicInfo.Claimant, got[0].Claimant)
	require.Equal(t, pBasicInfo.Defendant, got[0].Defendant)
	require.Equal(t, pBasicInfo.Vara, got[0].Vara)
	require.Equal(t, pBasicInfo.URL, got[0].URL)

	require.Equal(t, pBasicInfo2.ProcessID, got[1].ProcessID)
	require.Equal(t, pBasicInfo2.ProcessForo, got[1].ProcessForo)
	require.Equal(t, pBasicInfo2.ForoName, got[1].ForoName)
	require.Equal(t, pBasicInfo2.ProcessCode, got[1].ProcessCode)
	require.Equal(t, pBasicInfo2.Judge, got[1].Judge)
	require.Equal(t, pBasicInfo2.Class, got[1].Class)
	require.Equal(t, pBasicInfo2.Claimant, got[1].Claimant)
	require.Equal(t, pBasicInfo2.Defendant, got[1].Defendant)
	require.Equal(t, pBasicInfo2.Vara, got[1].Vara)
	require.Equal(t, pBasicInfo2.URL, got[1].URL)
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
