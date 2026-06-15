package main

import (
	"context"
	"log"
	"net/http"
	"time"

	authpkg "afs-todo-backend/internal/auth"
	"afs-todo-backend/internal/config"
	"afs-todo-backend/internal/database"
	"afs-todo-backend/internal/handlers"
	"afs-todo-backend/internal/middleware"
	"afs-todo-backend/internal/users"
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
	userRepository := users.NewRepository(db.SQL())
	tokenManager := authpkg.NewTokenManager(cfg.Auth.JWTSecret, cfg.Auth.JWTExpiryHours)
	authHandler := authpkg.NewHandler(userRepository, tokenManager)
	authMiddleware := middleware.NewAuth(authpkg.CookieName, tokenManager)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", handler.Health)
	mux.HandleFunc("GET /api/db-health", handler.DatabaseHealth)
	mux.HandleFunc("GET /api/schema-health", handler.SchemaHealth)
	mux.HandleFunc("POST /api/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/auth/logout", authHandler.Logout)
	mux.Handle("GET /api/auth/me", authMiddleware.RequireAuth(http.HandlerFunc(authHandler.Me)))

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
