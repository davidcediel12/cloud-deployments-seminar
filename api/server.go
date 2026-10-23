package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Together these must stay below the pod's terminationGracePeriodSeconds
// (default 30s), or Kubernetes kills the pod before it has finished.
const (
	drainDelay      = 5 * time.Second  // keep serving while the pod is removed from the Service
	shutdownTimeout = 10 * time.Second // time for in-flight requests to finish
)

// serve runs the server until SIGTERM (Kubernetes, docker stop) or Ctrl+C,
// then shuts down gracefully. It returns once the server has stopped.
func serve(server *http.Server, h *health) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, os.Interrupt)

	go func() {
		h.ready.Store(true)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server stopped", "error", err)
			os.Exit(1)
		}
	}()

	sig := <-stop
	slog.Info("shutting down", "signal", sig.String())

	// 1. Fail readiness and keep serving for a moment, so the platform stops
	//    sending new requests here first. Skipped for Ctrl+C: no load balancer locally.
	h.ready.Store(false)
	if sig == syscall.SIGTERM {
		time.Sleep(drainDelay)
	}

	// 2. Stop accepting connections and let in-flight requests finish.
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("shutdown did not finish cleanly", "error", err)
	}
}
