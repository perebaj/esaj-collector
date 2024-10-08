// Package api from api/api.go has the handlers to deal with the http requests.
//
//go:generate mockgen -source api.go -destination ../mock/api_mock.go -package mock
package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/perebaj/esaj/esaj"
	"github.com/perebaj/esaj/tracing"
)

// GCPTraceHeader is a default header send through http requests that can be used in the tracking
const GCPTraceHeader = "X-Cloud-Trace-Context"

// Storage is an interface that defines the methods to deal with storage
type Storage interface {
	SaveProcessSeeds(ctx context.Context, ps []esaj.ProcessSeed) error
	ProcessBasicInfoByOAB(ctx context.Context, oab string) ([]esaj.ProcessBasicInfo, error)
}

type esajClient interface {
	SearchByOAB(ctx context.Context, oab string) ([]esaj.ProcessSeed, error)
}

// Handler is a struct that holds the storage and esaj client
type Handler struct {
	storage Storage
	esaj    esajClient
}

// NewHandler creates a new Handler struct
func NewHandler(storage Storage, esaj esajClient) Handler {
	return Handler{
		storage: storage,
		esaj:    esaj,
	}
}

// OabSeederHandler is a handler that receives a oab query parameter and search for the process seeds in the esaj website
func (h Handler) OabSeederHandler(w http.ResponseWriter, r *http.Request) {
	traceID := r.Header.Get(GCPTraceHeader)
	ctx := r.Context()

	ctx = tracing.SetTraceIDInContext(ctx, traceID)

	logger := slog.With("traceID", traceID)
	oab := r.URL.Query().Get("oab")
	if oab == "" {
		http.Error(w, "oab is required", http.StatusBadRequest)
		return
	}

	seed, err := h.esaj.SearchByOAB(ctx, oab)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("error searching by oab", "error", err)
		return
	}

	logger.Info("saving process seeds", "seeds", seed)
	err = h.storage.SaveProcessSeeds(ctx, seed)
	if err != nil {
		logger.Error("error saving process seeds", "error", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
}

// ProcessesByOABHandler is a handler that receives a oab query parameter and search for the processes in the firestore database
func (h Handler) ProcessesByOABHandler(w http.ResponseWriter, r *http.Request) {
	traceID := r.Header.Get(GCPTraceHeader)
	ctx := r.Context()

	ctx = tracing.SetTraceIDInContext(ctx, traceID)

	logger := slog.With("traceID", traceID)
	oab := r.URL.Query().Get("oab")
	if oab == "" {
		http.Error(w, "oab is required", http.StatusBadRequest)
		return
	}

	processes, err := h.storage.ProcessBasicInfoByOAB(ctx, oab)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("error searching by oab", "error", err)
		return
	}

	logger.Info("processes found", "processes", processes)
	w.WriteHeader(http.StatusOK)
}
