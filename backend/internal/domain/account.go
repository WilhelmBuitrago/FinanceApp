package domain

type Account struct {
	ID              string
	UserID          string
	Name            string
	AccountType     string
	ParentAccountID *string
	Currency        string
	Status          string
	Institution     *string
	MetadataJson    *string
	CreatedAt       string
	UpdatedAt       string
}
