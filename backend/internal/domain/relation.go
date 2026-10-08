package domain

type MovementRelation struct {
	ID           string
	MovementIDA  string
	MovementIDB  string
	RelationType string
	Status       string
	Confidence   *float64
	Reason       *string
	CreatedAt    string
	UpdatedAt    string
}

type EventRelation struct {
	ID           string
	EventIDA     string
	EventIDB     string
	RelationType string
	Status       string
	Confidence   *float64
	Reason       *string
	CreatedAt    string
	UpdatedAt    string
}
