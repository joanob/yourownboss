package simulation

import (
	"database/sql"
	"encoding/json"
	"fmt"
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

	// GET process resources with names
	r.Get("/api/processes/{id}/resources", func(w http.ResponseWriter, req *http.Request) {
		idStr := chi.URLParam(req, "id")
		var id int64
		_, err := fmt.Sscan(idStr, &id)
		if err != nil || id == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		pres, err := repo.GetProcessResourcesDetailed(req.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(pres)
	})

	// GET simulations by process
	r.Get("/api/simulations", func(w http.ResponseWriter, req *http.Request) {
		q := req.URL.Query().Get("process_id")
		if q == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var pid int64
		_, err := fmt.Sscan(q, &pid)
		if err != nil || pid == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		sims, err := repo.ListSimulationsByProcess(req.Context(), pid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sims)
	})
}
