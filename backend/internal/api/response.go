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

// writeRepoError maps the repository error taxonomy to HTTP status codes so a
// failed write is reported as a real error instead of a fabricated success.
func writeRepoError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repositories.ErrNotFound):
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
	case errors.Is(err, repositories.ErrConflict):
		http.Error(w, `{"error":"conflict"}`, http.StatusConflict)
	case errors.Is(err, repositories.ErrConstraint):
		http.Error(w, `{"error":"constraint violation"}`, http.StatusBadRequest)
	default:
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
	}
}
