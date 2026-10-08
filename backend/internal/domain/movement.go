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
