package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Note is one row of the notes table.
type Note struct {
	ID        int64     `json:"id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

// NoteStore keeps notes in Postgres. The data lives in the database,
// not in the container, so the API itself stays stateless.
type NoteStore struct {
	db *pgxpool.Pool
}

// NewNoteStore uses an existing pool; it doesn't open or close it (see db.go).
func NewNoteStore(db *pgxpool.Pool) *NoteStore {
	return &NoteStore{db: db}
}

// GET /api/notes: newest first.
func (s *NoteStore) list(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query(r.Context(),
		`SELECT id, text, created_at FROM notes ORDER BY id DESC LIMIT 100`)
	if err != nil {
		serverError(w, err)
		return
	}
	defer rows.Close()

	notes := []Note{}
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Text, &n.CreatedAt); err != nil {
			serverError(w, err)
			return
		}
		notes = append(notes, n)
	}
	if err := rows.Err(); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, notes)
}

// POST /api/notes with {"text": "..."}.
func (s *NoteStore) create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Text string `json:"text"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	text := strings.TrimSpace(input.Text)
	if text == "" || len(text) > 500 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "text must be 1-500 characters"})
		return
	}

	n := Note{Text: text}
	err := s.db.QueryRow(r.Context(),
		`INSERT INTO notes (text) VALUES ($1) RETURNING id, created_at`, text,
	).Scan(&n.ID, &n.CreatedAt)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, n)
}

// serverError logs the real error and returns a generic message to the client.
func serverError(w http.ResponseWriter, err error) {
	slog.Error("database error", "error", err)
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
}
