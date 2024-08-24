package firestore

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/perebaj/esaj/clerk"
	"github.com/perebaj/esaj/tracing"
)

// SaveUser receive a generic clerk webhook event and save the user in the firestore database
func (s *Storage) SaveUser(ctx context.Context, event clerk.WebHookEvent) error {
	traceID := tracing.GetTraceIDFromContext(ctx)
	collection := s.client.Collection("users")
	docRef := collection.Doc(event.Data.ID)
	m := make(map[string]interface{})

	m["id"] = event.Data.ID
	m["first_name"] = event.Data.FirstName
	m["last_name"] = event.Data.LastName
	m["email_addresses"] = event.Data.EmailAddresses
	m["image_url"] = event.Data.ImageURL
	m["birthday"] = event.Data.Birthday
	// Date is received in unix timestamp in milliseconds and must be converted to RFC3339
	m["created_at"] = time.Unix(0, event.Data.CreatedAt*int64(time.Millisecond)).Format(time.RFC3339)
	m["updated_at"] = time.Unix(0, event.Data.UpdatedAt*int64(time.Millisecond)).Format(time.RFC3339)
	m["trace_id"] = traceID

	_, err := docRef.Set(ctx, m)

	return err
}

// DeleteUser receive a generic clerk webhook event and delete the user in the firestore database
// This delete is a soft delete, the user is not removed from the database, but a deleted_at field is added
// with the current time
func (s *Storage) DeleteUser(ctx context.Context, event clerk.WebHookEvent) error {
	traceID := tracing.GetTraceIDFromContext(ctx)
	slog.Info(fmt.Sprintf("deleting user %s", event.Data.ID), "traceID", traceID, "user_id", event.Data.ID)
	collection := s.client.Collection("users")
	docRef := collection.Doc(event.Data.ID)

	_, err := docRef.Update(ctx, []firestore.Update{
		{
			Path:  "deleted_at",
			Value: time.Now().Format(time.RFC3339),
		},
	})

	return err
}
