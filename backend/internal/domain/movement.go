package domain

type Movement struct {
	ID               string
	FinancialEventID string
	AccountID        string
	SourceEntryID    *string
	MovementDate     string
	AmountMinor      int64
	Description      *string
	Status           string
	MetadataJson     *string
	CreatedAt        string
	UpdatedAt        string
}

type MovementComponent struct {
	ID            string
	MovementID    string
	ComponentType string
	AmountMinor   int64
	Description   *string
	MetadataJson  *string
	CreatedAt     string
	UpdatedAt     string
}

// MovementWithCategory is a movement plus the category of its financial
// event, hydrated by a single join. Every movement listing needs both,
// and fetching them per row cost two queries per movement.
type MovementWithCategory struct {
	Movement
	CategoryID   string
	CategoryName string
	EventType    string
}

// ManualMovementCommand is the manual-creation input the API layer decoded
// and validated, handed to the service that owns the policy (duplicate
// detection, review parking, the atomic write).
type ManualMovementCommand struct {
	AccountID   string
	AmountMinor int64
	Type        string
	ToAccountID string
	Description string
	Date        string
	CategoryID  string
	Force       bool
}

// ManualMovementOutcome is the result: either the created movement's id,
// or the collision parked in the review queue for the human to decide.
type ManualMovementOutcome struct {
	ID             string
	Duplicate      bool
	ParkedReviewID string
}
