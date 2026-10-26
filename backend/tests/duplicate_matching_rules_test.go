package tests

import (
	"net/http"
	"testing"
)

// TestCreateMovement_DuplicateMatchingRules pins the collision rule that
// used to live inside the handler and now lives in the service and its
// query: same account, date and amount collide, with a description that
// is empty on either side; anything else is a real new movement.
func TestCreateMovement_DuplicateMatchingRules(t *testing.T) {
	router, sqlDB := newTestRouter(t)
	defer func() { _ = sqlDB.Close() }()

	post := func(target, body string) int {
		t.Helper()
		code, _ := hydrationDo(t, router, http.MethodPost, target, body)
		return code
	}
	seedAccount := func(id string) {
		t.Helper()
		if code := post("/accounts", `{"id":"`+id+`","user_id":"u","name":"Nu","account_type":"BANK_ACCOUNT","currency":"COP","institution":"NU"}`); code != http.StatusCreated {
			t.Fatalf("setup account %s: %d", id, code)
		}
	}
	seedMovement := func(account, body string) int {
		t.Helper()
		return post("/movements", body)
	}

	// (a) A request with no description collides with any movement on the
	// natural key: the empty side of the description match.
	seedAccount("acc-dup-empty")
	if code := seedMovement("acc-dup-empty", `{"account_id":"acc-dup-empty","amount_minor":-2500,"type":"EXPENSE","date":"2026-10-19","description":"original"}`); code != http.StatusCreated {
		t.Fatalf("setup: %d", code)
	}
	if code := seedMovement("acc-dup-empty", `{"account_id":"acc-dup-empty","amount_minor":-2500,"type":"EXPENSE","date":"2026-10-19"}`); code != http.StatusConflict {
		t.Errorf("an empty description must still collide on the natural key, got %d", code)
	}

	// (b) A different description is a real movement.
	seedAccount("acc-dup-desc")
	if code := seedMovement("acc-dup-desc", `{"account_id":"acc-dup-desc","amount_minor":-2500,"type":"EXPENSE","date":"2026-10-19","description":"original"}`); code != http.StatusCreated {
		t.Fatalf("setup: %d", code)
	}
	if code := seedMovement("acc-dup-desc", `{"account_id":"acc-dup-desc","amount_minor":-2500,"type":"EXPENSE","date":"2026-10-19","description":"otro"}`); code != http.StatusCreated {
		t.Errorf("a different description must not collide, got %d", code)
	}

	// (c) A different amount is a real movement.
	seedAccount("acc-dup-amount")
	if code := seedMovement("acc-dup-amount", `{"account_id":"acc-dup-amount","amount_minor":-2500,"type":"EXPENSE","date":"2026-10-19","description":"original"}`); code != http.StatusCreated {
		t.Fatalf("setup: %d", code)
	}
	if code := seedMovement("acc-dup-amount", `{"account_id":"acc-dup-amount","amount_minor":-2600,"type":"EXPENSE","date":"2026-10-19","description":"original"}`); code != http.StatusCreated {
		t.Errorf("a different amount must not collide, got %d", code)
	}

	// (d) force=true bypasses the check entirely.
	seedAccount("acc-dup-force")
	if code := seedMovement("acc-dup-force", `{"account_id":"acc-dup-force","amount_minor":-2500,"type":"EXPENSE","date":"2026-10-19","description":"original"}`); code != http.StatusCreated {
		t.Fatalf("setup: %d", code)
	}
	if code := post("/movements?force=true", `{"account_id":"acc-dup-force","amount_minor":-2500,"type":"EXPENSE","date":"2026-10-19","description":"original"}`); code != http.StatusCreated {
		t.Errorf("force=true must bypass the collision check, got %d", code)
	}
}
