package domain

type InstallmentPlan struct {
	ID                    string
	AccountID             string
	OriginEventID         *string
	TotalAmountMinor      *int64
	TotalInstallments     int64
	PaidInstallments      int64
	RemainingInstallments *int64
	Status                string
	StartDate             *string
	ExpectedEndDate       *string
	MetadataJson          *string
	CreatedAt             string
	UpdatedAt             string
}

type Installment struct {
	ID                  string
	InstallmentPlanID   string
	InstallmentNumber   int64
	TotalInstallments   int64
	ExpectedAmountMinor *int64
	CapitalAmountMinor  *int64
	InterestAmountMinor *int64
	Status              string
	ExpectedDate        *string
	DocumentedDate      *string
	MovementID          *string
	MetadataJson        *string
	CreatedAt           string
	UpdatedAt           string
}
