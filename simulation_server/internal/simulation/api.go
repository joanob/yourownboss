package simulation

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers simulation endpoints
func RegisterRoutes(r chi.Router, db *sql.DB) {
	repo := NewRepository(db)
	svc := NewService(repo)

	// POST /api/simulations -> start background simulation
	r.Post("/api/simulations", func(w http.ResponseWriter, req *http.Request) {
		var body SimulationRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := svc.StartSimulation(req.Context(), db, body); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
}
