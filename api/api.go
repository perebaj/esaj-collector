// Package api from api/api.go has the handlers to deal with the http requests.
package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/perebaj/esaj/esaj"
)

type storage interface {
	SaveProcessSeeds(ctx context.Context, ps []esaj.ProcessSeed) error
}

type esajClient interface {
	SearchByOAB(oab string) ([]esaj.ProcessSeed, error)
}

// Handler is a struct that holds the storage and esaj client
type Handler struct {
	storage storage
	esaj    esajClient
}

// NewHandler creates a new Handler struct
func NewHandler(storage storage, esaj esajClient) Handler {
	return Handler{
		storage: storage,
		esaj:    esaj,
	}
}

// OabSeederHandler is a handler that receives a oab query parameter and search for the process seeds in the esaj website
func (h Handler) OabSeederHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	oab := r.URL.Query().Get("oab")
	if oab == "" {
		http.Error(w, "oab is required", http.StatusBadRequest)
		return
	}

	seed, err := h.esaj.SearchByOAB(oab)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		slog.Error("error searching by oab", "error", err)
		return
	}

	slog.Info("saving process seeds", "seeds", seed)
	err = h.storage.SaveProcessSeeds(ctx, seed)
	if err != nil {
		slog.Error("error saving process seeds", "error", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
}
