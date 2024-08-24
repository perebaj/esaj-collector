//go:build integration

package firestore_test

import (
	"context"
	"testing"

	fs "cloud.google.com/go/firestore"
	"github.com/perebaj/esaj/clerk"
	"github.com/perebaj/esaj/firestore"
	"github.com/perebaj/esaj/tracing"
	"github.com/stretchr/testify/require"
)

func TestStorage_SaveUser(t *testing.T) {
	ctx := context.TODO()

	ctx = tracing.SetTraceIDInContext(ctx, "test-trace-id")
	c, err := fs.NewClient(ctx, projectID)
	require.NoError(t, err)
	defer cleanup(t, c)

	storage := firestore.NewStorage(c, projectID)

	event := clerk.WebHookEvent{
		Data: clerk.Data{
			ID:        "123",
			FirstName: "John",
			LastName:  "Doe",
			EmailAddresses: []clerk.EmailAddress{
				{
					EmailAddress: "teste",
					ID:           "123",
					LinkedTo:     []any{"123"},
					Object:       "email",
				},
			},
			ImageURL:  "image",
			Birthday:  "birday",
			CreatedAt: 1654012591835,
			UpdatedAt: 1654012591835,
		},
	}

	err = storage.SaveUser(ctx, event)
	require.NoError(t, err)

	doc, err := c.Collection("users").Doc("123").Get(ctx)
	require.NoError(t, err)

	var m map[string]interface{}
	doc.DataTo(&m)

	require.Equal(t, event.Data.ID, m["id"])
	require.Equal(t, event.Data.FirstName, m["first_name"])
	require.Equal(t, event.Data.LastName, m["last_name"])
	// TODO(@perebaj) parser the email addresses to validate the fields in the map
	require.NotNil(t, m["email_addresses"])
	require.Equal(t, event.Data.ImageURL, m["image_url"])
	require.Equal(t, event.Data.Birthday, m["birthday"])
	// the date is received in unix timestamp in milliseconds and must be converted to RFC3339
	require.Equal(t, "2022-05-31T12:56:31-03:00", m["created_at"])
	require.Equal(t, "2022-05-31T12:56:31-03:00", m["updated_at"])
	require.Equal(t, "test-trace-id", m["trace_id"])
	require.Nil(t, m["deleted_at"])

	// validate if the update is working
	event.Data.FirstName = "Jane"

	err = storage.SaveUser(ctx, event)
	require.NoError(t, err)

	doc, err = c.Collection("users").Doc("123").Get(ctx)
	require.NoError(t, err)

	doc.DataTo(&m)

	require.Equal(t, "Jane", m["first_name"])
	require.Equal(t, event.Data.LastName, m["last_name"])
	require.NotNil(t, m["email_addresses"])
	require.Equal(t, event.Data.ImageURL, m["image_url"])
	require.Equal(t, event.Data.Birthday, m["birthday"])
	require.Equal(t, "2022-05-31T12:56:31-03:00", m["created_at"])
	require.Equal(t, "2022-05-31T12:56:31-03:00", m["updated_at"])
	require.Equal(t, "test-trace-id", m["trace_id"])
	require.Nil(t, m["deleted_at"])
}

func TestStorage_DeleteUser(t *testing.T) {
	ctx := context.TODO()

	ctx = tracing.SetTraceIDInContext(ctx, "test-trace-id")
	c, err := fs.NewClient(ctx, projectID)
	require.NoError(t, err)
	defer cleanup(t, c)

	storage := firestore.NewStorage(c, projectID)

	event := clerk.WebHookEvent{
		Data: clerk.Data{
			ID:        "123",
			FirstName: "John",
			LastName:  "Doe",
			EmailAddresses: []clerk.EmailAddress{
				{
					EmailAddress: "teste",
					ID:           "123",
					LinkedTo:     []any{"123"},
					Object:       "email",
				},
			},
			ImageURL:  "image",
			Birthday:  "birday",
			CreatedAt: 1654012591835,
			UpdatedAt: 1654012591835,
		},
	}

	err = storage.SaveUser(ctx, event)
	require.NoError(t, err)

	err = storage.DeleteUser(ctx, event)
	require.NoError(t, err)

	doc, err := c.Collection("users").Doc("123").Get(ctx)
	require.NoError(t, err)

	var m map[string]interface{}
	doc.DataTo(&m)

	require.NotNil(t, m["deleted_at"])
}
