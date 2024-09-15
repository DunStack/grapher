package grapher

import (
	"encoding/json"
	"net/http"

	"github.com/graph-gophers/graphql-go"
)

func NewHandler(schema *graphql.Schema) *Handler {
	return &Handler{
		schema: schema,
	}
}

type Handler struct {
	schema *graphql.Schema
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		http.ServeFile(w, r, "static/graphiql.html")
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
