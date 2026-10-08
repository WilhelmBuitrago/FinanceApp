package domain

type Document struct {
	ID           string
	AccountID    *string
	DocumentType string
	Institution  *string
	DocumentName *string
	PeriodStart  *string
	PeriodEnd    *string
	DocumentDate *string
	FileHash     *string
	SourceUri    *string
	Status       string
	MetadataJson *string
	CreatedAt    string
	UpdatedAt    string
}
