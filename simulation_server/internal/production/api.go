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

	// GET all buildings
	r.Get("/api/buildings", func(w http.ResponseWriter, req *http.Request) {
		buildings, err := repo.GetAllBuildingsDetailed(req.Context())
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(buildings)
	})

	// GET single building
	r.Get("/api/buildings/{id}", func(w http.ResponseWriter, req *http.Request) {
		idStr := chi.URLParam(req, "id")
		var id int64
		_, err := fmt.Sscan(idStr, &id)
		if err != nil || id == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		b, err := repo.GetBuildingDetailed(req.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if b == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(b)
	})

	// POST bulk upsert production data (array of buildings with related data)
	r.Post("/api/production", func(w http.ResponseWriter, req *http.Request) {
		var body []BuildingDTO
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		for _, b := range body {
			if err := svc.UpdateBuilding(req.Context(), db, b.ID, b); err != nil {
				if err.Error() == "building not found" {
					if err2 := svc.CreateBuilding(req.Context(), db, b); err2 != nil {
						http.Error(w, err2.Error(), http.StatusBadRequest)
						return
					}
				} else {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			}
		}
		w.WriteHeader(http.StatusOK)
	})

	// POST single upsert
	r.Post("/api/production/{id}", func(w http.ResponseWriter, req *http.Request) {
		var body BuildingDTO
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// try update first
		if err := svc.UpdateBuilding(req.Context(), db, body.ID, body); err != nil {
			if err.Error() == "building not found" {
				if err2 := svc.CreateBuilding(req.Context(), db, body); err2 != nil {
					http.Error(w, err2.Error(), http.StatusBadRequest)
					return
				}
			} else {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	})
}
