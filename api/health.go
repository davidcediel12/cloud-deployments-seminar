package main

import (
	"net/http"
	"sync/atomic"
)

// health serves the probes for the platform (Kubernetes). Neither probe checks
// the database: a DB outage hits every pod at once, so restarting pods
// (liveness) or removing all of them from the Service (readiness) would only
// make things worse.
type health struct {
	ready atomic.Bool // true while serving, false once shutdown starts
}

// GET /livez. Liveness: "is the process alive?"
// Failing it makes Kubernetes restart the pod.
func (h *health) livez(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GET /readyz. Readiness: "should this pod get traffic?"
// Failing it removes the pod from the Service but doesn't restart it.
// Here it fails while shutting down (see serve in server.go).
func (h *health) readyz(w http.ResponseWriter, r *http.Request) {
	if !h.ready.Load() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "shutting down"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
