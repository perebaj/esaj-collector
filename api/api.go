// Package api from api/api.go has the handlers to deal with the http requests.
package api

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/perebaj/esaj"
)

type storage interface {
	SaveProcessSeeds(ctx context.Context, ps []esaj.ProcessSeed) (int64, error)
}

type esajClient interface {
	SearchByOAB(oab string) ([]esaj.ProcessSeed, error)
}

type handler struct {
	storage storage
	esaj    esajClient
}

func newHandler(storage storage, esaj esajClient) handler {
	return handler{
		storage: storage,
		esaj:    esaj,
	}
}

func (h handler) oabSeederHandler(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.storage.SaveProcessSeeds(ctx, seed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"found":` + strconv.FormatInt(result, 10) + `}`))

}

// NewServerMux creates a new http.ServeMux with the routes for the api
func NewServerMux(storage storage, esajClient esajClient) *http.ServeMux {
	h := newHandler(storage, esajClient)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /processes", h.oabSeederHandler)

	return mux
}
