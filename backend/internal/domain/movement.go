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
