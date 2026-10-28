package domain

import "errors"

// ErrTransferToSelf is the policy a same-account transfer breaks: money
// cannot leave an account to arrive at the same one. It lives in the
// shared vocabulary so both the service and the HTTP layer can name it
// without importing each other.
var ErrTransferToSelf = errors.New("transfer origin and destination must differ")

// ErrTransferAmount is the policy a non-positive transfer amount breaks:
// the amount is the value that moves, and its signs are derived.
var ErrTransferAmount = errors.New("transfer amount must be positive")

// ErrTransferCurrencyMismatch is the policy two accounts with different
// monies break: the amount would be copied as if the currencies matched,
// a silent corruption the day a second one exists.
var ErrTransferCurrencyMismatch = errors.New("transfer needs both accounts in the same currency")

// TransferCommand is a transfer between own accounts: one decision with
// two legs. The amount is submitted positive; the service derives the
// signs (it leaves the origin, it enters the destination).
type TransferCommand struct {
	OriginID      string
	DestinationID string
	AmountMinor   int64
	Date          string
	Description   string
}

// TransferOutcome carries every identity the write produced: the shared
// event, both legs and the relation that binds them.
type TransferOutcome struct {
	EventID               string
	OriginMovementID      string
	DestinationMovementID string
	RelationID            string
}
