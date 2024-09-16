package grapher

import (
	"context"
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

type WithContextFunc func(r *http.Request) context.Context

func WithContext(withContext WithContextFunc) HandlerOption {
	return func(h *Handler) {
		h.withContext = withContext
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
	schema      *graphql.Schema
	explorer    string
	withContext WithContextFunc
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
		ctx := r.Context()
		if h.withContext != nil {
			ctx = h.withContext(r)
		}
		result := h.schema.Exec(ctx, payload.Query, payload.Operation, payload.Variables)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}
