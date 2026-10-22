package domain

type SourceEntry struct {
	ID                    string
	DocumentID            string
	SourceReference       *string
	RawDescription        *string
	RawAmountMinor        *int64
	RawDate               *string
	NormalizedDescription *string
	NormalizedAmountMinor *int64
	NormalizedDate        *string
	PageNumber            *int64
	RowNumber             *int64
	Section               *string
	ExtractionID          *string
	ReviewStatus          string
	SourceKey             *string
	CorrectedDescription  *string
	CorrectedAmountMinor  *int64
	CorrectedDate         *string
	MetadataJson          *string
	CreatedAt             string
	UpdatedAt             string
}
