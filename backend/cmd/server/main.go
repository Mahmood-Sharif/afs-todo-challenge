package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"afs-todo-backend/internal/config"
	"afs-todo-backend/internal/database"
	"afs-todo-backend/internal/handlers"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	db, err := database.Connect(context.Background(), cfg.Database)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("database close failed: %v", err)
		}
	}()

	handler := handlers.New(db)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", handler.Health)
	mux.HandleFunc("GET /api/db-health", handler.DatabaseHealth)
	mux.HandleFunc("GET /api/schema-health", handler.SchemaHealth)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("backend connected to postgres at %s:%s", cfg.Database.Host, cfg.Database.Port)
	log.Printf("backend listening on port %s", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
