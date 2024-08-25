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

func TestStorage_GetUser(t *testing.T) {
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

	user, err := storage.GetUser(ctx, "123")
	require.NoError(t, err)

	t.Logf("printing user: %+v", user)
	require.Equal(t, event.Data.ID, user.ID)
	require.Equal(t, event.Data.FirstName, user.FirstName)
	require.Equal(t, event.Data.LastName, user.LastName)
	// TODO(@perebaj) parser the email addresses to validate the fields in the map
	require.NotNil(t, user.EmailAddresses)
	require.Equal(t, event.Data.ImageURL, user.ImageURL)
	require.Equal(t, event.Data.Birthday, user.Birthday)
	require.Equal(t, "2022-05-31T12:56:31-03:00", user.CreatedAt)
	require.Equal(t, "2022-05-31T12:56:31-03:00", user.UpdatedAt)
	require.Equal(t, "test-trace-id", user.TraceID)
	require.Empty(t, user.DeletedAt)

	// Get a user that doesn't exist, it shouldn't return an error
	// just an empty user
	user, err = storage.GetUser(ctx, "non-exitent")
	require.NoError(t, err)

	require.Equal(t, firestore.User{}, user)
}
