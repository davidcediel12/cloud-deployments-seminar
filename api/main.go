// A tiny HTTP API for the seminar. Standard library + pgx (Postgres driver).
//
// Configuration comes from environment variables (12-factor):
//
//	PORT         port to listen on (default 8080)
//	APP_MESSAGE  text returned by GET /api/message
//	APP_VERSION  version returned by GET /api/message (set in the image by CI: the commit SHA)
//	CORS_ORIGIN  frontend origin allowed to call the API (default http://localhost:5173)
//	DB_URL       Postgres connection string (required), e.g. postgres://user:pass@host:5432/db
package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	// Logs go to stdout as JSON: the platform (Docker, Azure) collects them.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	port := getEnv("PORT", "8080")
	message := getEnv("APP_MESSAGE", "Hello from the API")
	version := getEnv("APP_VERSION", "dev")
	corsOrigin := getEnv("CORS_ORIGIN", "http://localhost:5173")

	// The database URL contains a password: required, never logged, never hardcoded.
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		slog.Error("DB_URL is required")
		os.Exit(1)
	}
	// Fail fast with a clear log line instead of hanging when the DB is unreachable
	// (e.g. a firewall that drops packets): the pod then shows CrashLoopBackOff + this error.
	slog.Info("connecting to the database")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	db, err := connectDB(ctx, dbURL)
	cancel()
	if err != nil {
		slog.Error("cannot connect to the database", "error", err)
		os.Exit(1)
	}
	notes := NewNoteStore(db)
	h := &health{}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /livez", h.livez)
	mux.HandleFunc("GET /readyz", h.readyz)
	mux.HandleFunc("GET /api/message", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"message": message, "version": version})
	})
	mux.HandleFunc("GET /api/notes", notes.list)
	mux.HandleFunc("POST /api/notes", notes.create)

	server := &http.Server{Addr: ":" + port, Handler: logRequests(cors(corsOrigin, mux))}

	slog.Info("starting API", "port", port, "cors_origin", corsOrigin)
	serve(server, h) // blocks until shutdown
	db.Close()
	slog.Info("stopped")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
