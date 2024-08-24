package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/perebaj/esaj/clerk"
	"github.com/perebaj/esaj/tracing"
)

// UserStorage is an interface that defines the methods to deal with user in the storage
type UserStorage interface {
	SaveUser(ctx context.Context, user clerk.WebHookEvent) error
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
	default:
		logger.Info(fmt.Sprintf("event %s not supported", clerkWebHook.Type))
		http.Error(w, "event not supported", http.StatusBadRequest)
	}

}

func (h UserHandler) createUser(ctx context.Context, clerkWebHook clerk.WebHookEvent) error {
	err := h.storage.SaveUser(ctx, clerkWebHook)
	return err
}
