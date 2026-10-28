package tests

import (
	"context"
	"strings"
	"sync"
	"testing"

	"financeapp/backend/internal/domain"
)

/* IDs: the old scheme took a nonce from the wall clock, so two
   movements created in the same clock tick carried the same ID and the
   second insert died on the primary key (a 500 nobody could explain).
   The helper now draws from the OS entropy, and the guarantee is a
   burst of concurrent writers producing distinct identities. */

func TestNewID_IsUniqueAndNamesItsKindUnderABurst(t *testing.T) {
	const n = 5000
	seen := make(map[string]struct{}, n)
	for i := 0; i < n; i++ {
		id := domain.NewID("mv_test")
		if !strings.HasPrefix(id, "mv_test_") {
			t.Fatalf("the id must name its kind, got %s", id)
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("collision after %d ids: %s", i, id)
		}
		seen[id] = struct{}{}
	}
}

func TestTransferService_ConcurrentTransfersNeverCollide(t *testing.T) {
	service, repos, sqlDB := transferEnv(t)
	defer func() { _ = sqlDB.Close() }()
	// Only one account per side: the destination name must differ, so the
	// bursts share accounts and fight over the id space instead.
	seedAccount(t, repos, "acc-x")
	seedAccount(t, repos, "acc-y")

	const burst = 50
	var wg sync.WaitGroup
	ids := make([]string, 0, burst*2)
	var mu sync.Mutex
	for i := 0; i < burst; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out, err := service.Transfer(context.Background(), transferCmd("acc-x", "acc-y", 1000))
			if err != nil {
				t.Errorf("transfer in burst: %v", err)
				return
			}
			mu.Lock()
			ids = append(ids, out.OriginMovementID, out.DestinationMovementID, out.EventID, out.RelationID)
			mu.Unlock()
		}()
	}
	wg.Wait()

	if len(ids) != burst*4 {
		t.Fatalf("every transfer must report its identities, got %d of %d", len(ids), burst*4)
	}
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if _, dup := seen[id]; dup {
			t.Fatalf("concurrent writes collided on %s", id)
		}
		seen[id] = struct{}{}
	}
}
