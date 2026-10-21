package api

import (
	"errors"
	"net/http"

	"financeapp/backend/internal/repositories"
)

// responseWriter captures the status code written by a handler so the
// logging middleware can report it.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// writeLookupError maps a repository lookup failure to its HTTP status:
// not found is a client error, everything else is a server error.
func writeLookupError(w http.ResponseWriter, err error) {
	if errors.Is(err, repositories.ErrNotFound) {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	http.Error(w, `{"error":"lookup failed"}`, http.StatusInternalServerError)
}

// corsMiddleware allows the Vite dev server and the Tauri desktop shell, and
// answers preflight requests without hitting the mux.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		// Allow Vite dev server and Tauri desktop origins
		allowedOrigins := map[string]bool{
			"http://localhost:5173": true,
			"http://localhost:5174": true,
			"http://127.0.0.1:5173": true,
			"http://127.0.0.1:5174": true,
		}
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
