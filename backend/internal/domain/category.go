package domain

type Category struct {
	ID           string
	Name         string
	ParentID     *string
	CategoryType *string
	Status       string
	CreatedAt    string
	UpdatedAt    string
}
