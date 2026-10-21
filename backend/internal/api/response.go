package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	logpkg "financeapp/backend/internal/log"
)

// encodeJSON writes payload as JSON and logs an encode failure instead of
// silently truncating the response: once the status line was written, a
// truncated body reaches the client as invalid JSON with no trace server-side.
func encodeJSON(w http.ResponseWriter, payload any) {
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logpkg.Error(fmt.Sprintf("json encode failed: %v", err))
	}
}
