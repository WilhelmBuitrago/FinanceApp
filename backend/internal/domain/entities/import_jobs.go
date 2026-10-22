package entities

// ImportJob represents an import pipeline job for a document.
// TODO: Move to internal/domain if entities package is consolidated.
type ImportJob struct {
	ID           string
	DocumentID   string
	Status       string
	CreatedAt    string
	StartedAt    *string
	FinishedAt   *string
	Error        *string
	ProcessingID *string
}

// TODO: Add validation methods and constants for statuses.
