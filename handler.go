package grapher

import (
	"encoding/json"
	"net/http"

	"github.com/graph-gophers/graphql-go"
)

type HandlerOption func(h *Handler)

func WithExplorer(name string) HandlerOption {
	return func(h *Handler) {
		h.explorer = name
	}
}

func NewHandler(schema *graphql.Schema, opts ...HandlerOption) *Handler {
	h := &Handler{schema: schema}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

type Handler struct {
	schema   *graphql.Schema
	explorer string
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if e := h.explorer; e == "" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		} else {
			http.ServeFile(w, r, e)
		}
	case http.MethodPost:
		var payload struct {
			Query     string         `json:"query"`
			Operation string         `json:"operationName"`
			Variables map[string]any `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		result := h.schema.Exec(r.Context(), payload.Query, payload.Operation, payload.Variables)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}
