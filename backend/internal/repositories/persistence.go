package repositories

import (
	"database/sql"

	"financeapp/backend/internal/db"
)

type Persistence struct {
	DB                 *sql.DB
	Querier            db.Querier
	AccountRepo        *AccountRepository
	CategoryRepo       *CategoryRepository
	DocumentRepo       *DocumentRepository
	SourceEntryRepo    *SourceEntryRepository
	FinancialEventRepo *FinancialEventRepository
	MovementRepo       *MovementRepository
	// The next two are scaffolding for features that are designed and
	// tested but not wired to the API yet (credit-card installments,
	// transfers between own accounts). See each repository's comment.
	InstallmentPlanRepo *InstallmentPlanRepository
	RelationRepo        *RelationRepository
	ImportJobRepo       *ImportJobRepo
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
		ImportJobRepo:       NewImportJobRepo(repoDB),
	}
}
