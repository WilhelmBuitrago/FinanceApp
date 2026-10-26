package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"financeapp/backend/internal/log"
	"financeapp/backend/internal/repositories"
)

// encodeJSON writes payload as JSON and logs an encode failure instead of
// silently truncating the response: once the status line was written, a
// truncated body reaches the client as invalid JSON with no trace server-side.
func encodeJSON(w http.ResponseWriter, payload any) {
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Error(fmt.Sprintf("json encode failed: %v", err))
	}
}

// writeJSONObject writes payload as the response with the given status.
// The content type is set first: every response body in this API is JSON,
// and a wrong declaration breaks any client that trusts the header.
func writeJSONObject(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	encodeJSON(w, payload)
}

// writeJSONError is the single writer for API errors: {"error": msg} with
// an honest content type. It replaces http.Error, which declares
// text/plain and left every error body mislabeled.
func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSONObject(w, status, map[string]string{"error": msg})
}

// writeRepoError maps the repository error taxonomy to HTTP status codes so a
// failed write is reported as a real error instead of a fabricated success.
func writeRepoError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repositories.ErrNotFound):
		writeJSONError(w, http.StatusNotFound, "not found")
	case errors.Is(err, repositories.ErrConflict):
		writeJSONError(w, http.StatusConflict, "conflict")
	case errors.Is(err, repositories.ErrConstraint):
		writeJSONError(w, http.StatusBadRequest, "constraint violation")
	default:
		writeJSONError(w, http.StatusInternalServerError, "internal error")
	}
}
