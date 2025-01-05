package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type HealthResponse struct {
	Status      string
	Version     string
	Description string
	Time        string
}

// https://datatracker.ietf.org/doc/html/draft-inadarei-api-health-check-06#name-api-health-response
func (appHandler *AppHandler) Health(mux chi.Router) {
	mux.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response := HealthResponse{
			Status:      "pass",
			Version:     "0.0.1",
			Description: "API Build for `Forum Regional de Cybesecurite de la CEMAC`",
			Time:        time.Now().Format(time.RFC3339),
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "error encoding the result", http.StatusBadRequest)
			return
		}
	})
}
