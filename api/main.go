// A tiny HTTP API for the seminar. Standard library + pgx (Postgres driver).
//
// Configuration comes from environment variables (12-factor):
//
//	PORT         port to listen on (default 8080)
//	APP_MESSAGE  text returned by GET /api/message
//	CORS_ORIGIN  frontend origin allowed to call the API (default http://localhost:5173)
//	DB_URL       Postgres connection string (required), e.g. postgres://user:pass@host:5432/db
//
// Plus the OTEL_* variables in telemetry.go (optional: traces, metrics, logs to Grafana).
package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	// Logs go to stdout as JSON: the platform (Docker, Azure) collects them.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := setupTelemetry(context.Background()); err != nil {
		slog.Error("cannot start telemetry", "error", err)
		os.Exit(1)
	}

	port := getEnv("PORT", "8080")
	message := getEnv("APP_MESSAGE", "Hello from the API")
	corsOrigin := getEnv("CORS_ORIGIN", "http://localhost:5173")

	// The database URL contains a password: required, never logged, never hardcoded.
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		slog.Error("DB_URL is required")
		os.Exit(1)
	}
	notes, err := NewNoteStore(context.Background(), dbURL)
	if err != nil {
		slog.Error("cannot connect to the database", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/message", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"message": message})
	})
	mux.HandleFunc("GET /api/notes", notes.list)
	mux.HandleFunc("POST /api/notes", notes.create)

	// otelhttp goes outermost: it starts the request's span (trace) and records
	// request metrics; the handlers inside see the span in r.Context().
	handler := otelhttp.NewHandler(logRequests(cors(corsOrigin, mux)), "http-server")

	slog.Info("starting API", "port", port, "cors_origin", corsOrigin)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

// cors lets the browser call this API from the frontend's origin.
// Without it, the frontend works locally but fails once API and frontend
// live on different domains in the cloud.
func cors(origin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// logRequests writes one log line per request.
func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		// InfoContext: the log line carries the request's trace ID (links logs to traces in Grafana).
		slog.InfoContext(r.Context(), "request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
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
