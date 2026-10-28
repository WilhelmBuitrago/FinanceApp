package tests

import (
	"testing"
	"time"

	"financeapp/backend/internal/domain"
)

// The timestamp contract every writer relies on: UTC, RFC3339, the
// shape the schema stores and the front formats.
func TestNowRFC3339_StampsUTCInRFC3339(t *testing.T) {
	stamp := domain.NowRFC3339()
	parsed, err := time.Parse(time.RFC3339, stamp)
	if err != nil {
		t.Fatalf("the stamp must be RFC3339: %v (%s)", err, stamp)
	}
	if parsed.Location() != time.UTC {
		t.Errorf("the stamp must be UTC, got %s", parsed.Location())
	}
}
