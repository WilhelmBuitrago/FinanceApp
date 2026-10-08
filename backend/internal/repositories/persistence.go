package repositories

import (
	"database/sql"

	"financeapp/backend/internal/db"
)

type Persistence struct {
	DB                  *sql.DB
	Querier             db.Querier
	AccountRepo         *AccountRepository
	CategoryRepo        *CategoryRepository
	DocumentRepo        *DocumentRepository
	SourceEntryRepo     *SourceEntryRepository
	FinancialEventRepo  *FinancialEventRepository
	MovementRepo        *MovementRepository
	InstallmentPlanRepo *InstallmentPlanRepository
	RelationRepo        *RelationRepository
}

func NewPersistence(sqlDB *sql.DB) *Persistence {
	querier := db.New(sqlDB)
	repoDB := &DB{
		Querier: querier,
	}
	return &Persistence{
		DB:                  sqlDB,
		Querier:             querier,
		AccountRepo:         NewAccountRepository(repoDB),
		CategoryRepo:        NewCategoryRepository(repoDB),
		DocumentRepo:        NewDocumentRepository(repoDB),
		SourceEntryRepo:     NewSourceEntryRepository(repoDB),
		FinancialEventRepo:  NewFinancialEventRepository(repoDB),
		MovementRepo:        NewMovementRepository(repoDB),
		InstallmentPlanRepo: NewInstallmentPlanRepository(repoDB),
		RelationRepo:        NewRelationRepository(repoDB),
	}
}
