package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
)

type DatabaseHealthChecker interface {
	Ping(ctx context.Context) error
	SchemaTables(ctx context.Context) ([]string, error)
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

type schemaHealthResponse struct {
	Status  string   `json:"status"`
	Service string   `json:"service"`
	Tables  []string `json:"tables,omitempty"`
	Message string   `json:"message,omitempty"`
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

func (h *Handler) SchemaHealth(w http.ResponseWriter, r *http.Request) {
	tables, err := h.database.SchemaTables(r.Context())
	if err != nil {
		log.Printf("schema health check failed: %v", err)
		writeJSON(w, http.StatusServiceUnavailable, schemaHealthResponse{
			Status:  "error",
			Service: "schema",
			Message: "schema unavailable",
		})
		return
	}

	if len(tables) != 2 {
		writeJSON(w, http.StatusServiceUnavailable, schemaHealthResponse{
			Status:  "error",
			Service: "schema",
			Tables:  tables,
			Message: "required tables missing",
		})
		return
	}

	writeJSON(w, http.StatusOK, schemaHealthResponse{
		Status:  "ok",
		Service: "schema",
		Tables:  tables,
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}
