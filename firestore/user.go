package firestore

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/perebaj/esaj/clerk"
	"github.com/perebaj/esaj/tracing"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

// User is the struct that represents the user in the firestore database
type User struct {
	ID             string `firestore:"id"`
	FirstName      string `firestore:"first_name"`
	LastName       string `firestore:"last_name"`
	EmailAddresses []any  `firestore:"email_addresses"`
	ImageURL       string `firestore:"image_url"`
	Birthday       string `firestore:"birthday"`
	CreatedAt      string `firestore:"created_at"`
	UpdatedAt      string `firestore:"updated_at"`
	DeletedAt      string `firestore:"deleted_at"`
	TraceID        string `firestore:"trace_id"`
}

// GetUser get a user from the firestore database
// If the user does not exist, it will return an empty user and a nil error.
func (s *Storage) GetUser(ctx context.Context, userID string) (User, error) {
	traceID := tracing.GetTraceIDFromContext(ctx)
	slog.Info(fmt.Sprintf("getting user %s", userID), "traceID", traceID, "user_id", userID)
	collection := s.client.Collection("users")
	docRef := collection.Doc(userID)

	doc, err := docRef.Get(ctx)
	if status.Code(err) == codes.NotFound {
		return User{}, nil
	}

	if err != nil {
		return User{}, fmt.Errorf("error getting user %s: %w", userID, err)
	}

	var user User
	err = doc.DataTo(&user)
	if err != nil {
		return User{}, fmt.Errorf("error parsing user data: %w", err)
	}

	return user, nil
}
