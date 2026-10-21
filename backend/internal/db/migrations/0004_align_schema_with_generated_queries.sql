-- 0004_align_schema_with_generated_queries.sql
-- Aligns the schema with the tables and columns used by the generated
-- sqlc queries (backend/internal/db). Additive only: no data migration.

CREATE TABLE IF NOT EXISTS movement_components (
    id TEXT PRIMARY KEY,
    movement_id TEXT NOT NULL REFERENCES movements(id),
    component_type TEXT NOT NULL CHECK (component_type IN ('CAPITAL', 'INTEREST')),
    amount_minor INTEGER NOT NULL,
    description TEXT,
    metadata_json TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS movement_relations (
    id TEXT PRIMARY KEY,
    movement_id_a TEXT NOT NULL REFERENCES movements(id),
    movement_id_b TEXT NOT NULL REFERENCES movements(id),
    relation_type TEXT NOT NULL,
    status TEXT NOT NULL,
    confidence REAL,
    reason TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    CHECK (movement_id_a <> movement_id_b)
);

ALTER TABLE installment_plans ADD COLUMN remaining_installments INTEGER;
ALTER TABLE installment_plans ADD COLUMN expected_end_date TEXT;
ALTER TABLE installment_plans ADD COLUMN metadata_json TEXT;

-- One installment number per plan (unique index keeps 0001 additive).
CREATE UNIQUE INDEX IF NOT EXISTS uq_installments_plan_number
    ON installments(installment_plan_id, installment_number);
