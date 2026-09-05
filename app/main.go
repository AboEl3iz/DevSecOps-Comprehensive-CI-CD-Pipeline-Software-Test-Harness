package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

type Application struct {
	isReady *atomic.Bool
	logger  *log.Logger
}

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

type DataResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
	User    string `json:"user,omitempty"`
}

func NewApplication(logger *log.Logger) *Application {
	ready := &atomic.Bool{}
	ready.Store(true)
	return &Application{
		isReady: ready,
		logger:  logger,
	}
}

func (app *Application) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Version:   "1.0.0",
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		app.logger.Printf("HealthHandler encode error: %v", err)
	}
}

func (app *Application) ReadyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !app.isReady.Load() {
		w.WriteHeader(http.StatusServiceUnavailable)
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "not_ready"}); err != nil {
			app.logger.Printf("ReadyHandler encode error: %v", err)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ready"}); err != nil {
		app.logger.Printf("ReadyHandler encode error: %v", err)
	}
}

func (app *Application) DataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"}); err != nil {
			app.logger.Printf("DataHandler encode error: %v", err)
		}
		return
	}

	user := r.Header.Get("X-User-ID")
	if user == "" {
		user = "anonymous"
	}

	resp := DataResponse{
		Message: "DevSecOps secure API data endpoint",
		Status:  "success",
		User:    user,
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		app.logger.Printf("DataHandler encode error: %v", err)
	}
}

func (app *Application) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", app.HealthHandler)
	mux.HandleFunc("/readyz", app.ReadyHandler)
	mux.HandleFunc("/api/v1/data", app.DataHandler)
	return app.SecurityHeadersMiddleware(mux)
}

func (app *Application) SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none';")
		next.ServeHTTP(w, r)
	})
}

func main() {
	logger := log.New(os.Stdout, "[SECURE-APP] ", log.LstdFlags|log.Lshortfile)
	app := NewApplication(logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      app.Routes(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Printf("Starting secure server on port %s...", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server error: %v", err)
		}
	}()

	<-shutdown
	logger.Println("Shutting down server gracefully...")

	app.isReady.Store(false)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Println("Server stopped cleanly.")
}
