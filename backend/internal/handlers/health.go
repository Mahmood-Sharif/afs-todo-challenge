package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
)

type DatabaseHealthChecker interface {
	Ping(ctx context.Context) error
}

type Handler struct {
	database DatabaseHealthChecker
}

type healthResponse struct {
	Status   string `json:"status"`
	Service  string `json:"service"`
	Database string `json:"database,omitempty"`
	Message  string `json:"message,omitempty"`
}

func New(database DatabaseHealthChecker) *Handler {
	return &Handler{database: database}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{
		Status:  "ok",
		Service: "backend",
	})
}

func (h *Handler) DatabaseHealth(w http.ResponseWriter, r *http.Request) {
	if err := h.database.Ping(r.Context()); err != nil {
		log.Printf("database health check failed: %v", err)
		writeJSON(w, http.StatusServiceUnavailable, healthResponse{
			Status:  "error",
			Service: "database",
			Message: "database unavailable",
		})
		return
	}

	writeJSON(w, http.StatusOK, healthResponse{
		Status:   "ok",
		Service:  "database",
		Database: "postgres",
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload healthResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}
