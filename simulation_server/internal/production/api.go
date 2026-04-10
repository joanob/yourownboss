package production

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers production endpoints
func RegisterRoutes(r chi.Router, db *sql.DB) {
	repo := NewRepository(db)
	svc := NewService(repo)

	r.Post("/api/buildings", func(w http.ResponseWriter, req *http.Request) {
		var body BuildingDTO
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := svc.CreateBuilding(req.Context(), db, body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	r.Put("/api/buildings/{id}", func(w http.ResponseWriter, req *http.Request) {
		idStr := chi.URLParam(req, "id")
		var id int64
		_, err := fmt.Sscan(idStr, &id)
		if err != nil || id == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var body BuildingDTO
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := svc.UpdateBuilding(req.Context(), db, id, body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	r.Delete("/api/buildings/{id}", func(w http.ResponseWriter, req *http.Request) {
		idStr := chi.URLParam(req, "id")
		var id int64
		_, err := fmt.Sscan(idStr, &id)
		if err != nil || id == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := svc.DeleteBuilding(req.Context(), db, id); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}
