package domain

type FinancialEvent struct {
	ID                 string
	EventType          string
	EventDate          *string
	Status             string
	VerificationStatus string
	Description        *string
	CategoryID         *string
	Confidence         *float64
	MetadataJson       *string
	CreatedAt          string
	UpdatedAt          string
}
