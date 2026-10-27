package domain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

/* NewID mints an identity for a row the app creates by hand. The first
   scheme took a nonce from the wall clock: two rows written in the same
   clock tick collided on the primary key, and the failed insert surfaced
   as a 500 nobody could explain. OS entropy cannot collide in a burst. */

func NewID(prefix string) string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		// The OS entropy pool is a hard dependency of every write path:
		// falling back to a predictable scheme would resurrect the
		// collision this replacement exists to remove.
		panic(fmt.Sprintf("crypto/rand unavailable: %v", err))
	}
	return prefix + "_" + hex.EncodeToString(buf)
}
