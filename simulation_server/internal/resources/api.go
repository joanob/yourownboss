package resources

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes attaches resource routes to the router
func RegisterRoutes(r chi.Router, db *sql.DB) {
	repo := NewRepository(db)
	svc := NewService(repo)

	r.Post("/api/resources/upsert", func(w http.ResponseWriter, req *http.Request) {
		// simple auth: Bearer <AUTH_TOKEN>
		r.Post("/api/resources/upsert", func(w http.ResponseWriter, req *http.Request) {
			Resources []ResourceDTO `json:"resources"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err := svc.UpsertResources(context.Background(), db, body.Resources)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	})
}
