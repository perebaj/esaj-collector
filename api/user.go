// Package api user.go has the handlers to deal with the http requests related to user management.
//
//go:generate mockgen -source user.go -destination ../mock/user_mock.go -package mock
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/perebaj/esaj/clerk"
	"github.com/perebaj/esaj/firestore"
	"github.com/perebaj/esaj/tracing"
)

// UserStorage is an interface that defines the methods to deal with user in the storage
type UserStorage interface {
	SaveUser(ctx context.Context, user clerk.WebHookEvent) error
	DeleteUser(ctx context.Context, user clerk.WebHookEvent) error
	GetUser(ctx context.Context, userID string) (firestore.User, error)
}

// UserHandler gather third party services to create an user
type UserHandler struct {
	storage UserStorage
}

// NewUserHandler creates a new UserHandler to deal with user management
func NewUserHandler(storage UserStorage) UserHandler {
	return UserHandler{
		storage: storage,
	}
}

// ClerkWebHookHandler is a handler that receives a clerk webhook and create a user with the data
func (h UserHandler) ClerkWebHookHandler(w http.ResponseWriter, r *http.Request) {
	traceID := r.Header.Get(GCPTraceHeader)
	ctx := r.Context()

	ctx = tracing.SetTraceIDInContext(ctx, traceID)

	logger := slog.With("traceID", traceID)

	var clerkWebHook clerk.WebHookEvent
	err := json.NewDecoder(r.Body).Decode(&clerkWebHook)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		logger.Error("error decoding clerk webhook", "error", err)
		return
	}

	switch clerkWebHook.Type {
	case "user.created":
		logger.Info("creating user", "clerkWebHook", clerkWebHook)
		err := h.createUser(ctx, clerkWebHook)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("error creating user", "error", err, "request_body", clerkWebHook)
			return
		}
	case "user.updated":
		logger.Info("updating user", "clerkWebHook", clerkWebHook)
		err := h.createUser(ctx, clerkWebHook)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("error updating user", "error", err, "request_body", clerkWebHook)
			return
		}
	case "user.deleted":
		logger.Info("deleting user", "clerkWebHook", clerkWebHook)
		err := h.storage.DeleteUser(ctx, clerkWebHook)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("error deleting user", "error", err, "request_body", clerkWebHook)
			return
		}

	default:
		logger.Error(fmt.Sprintf("event %s not supported", clerkWebHook.Type))
		http.Error(w, "event not supported", http.StatusBadRequest)
	}

}

func (h UserHandler) createUser(ctx context.Context, clerkWebHook clerk.WebHookEvent) error {
	err := h.storage.SaveUser(ctx, clerkWebHook)
	return err
}

// GetUserHandler is a handler that receives a user_id and return the user data
// If the user does not exist, it will return a 404 status code
func (h UserHandler) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	traceID := r.Header.Get(GCPTraceHeader)
	ctx := r.Context()

	ctx = tracing.SetTraceIDInContext(ctx, traceID)

	logger := slog.With("traceID", traceID)

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		logger.Error("user_id is required")
		return
	}

	user, err := h.storage.GetUser(ctx, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("error getting user", "error", err)
		return
	}

	if user.ID == "" || user.DeletedAt != "" {
		http.Error(w, "user not found", http.StatusNotFound)
		logger.Info("user not found", "user_id", userID)
		return
	}

	logger.Info(fmt.Sprintf("user %s found", userID), "user", user)
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("error encoding user", "error", err)
		return
	}
}
