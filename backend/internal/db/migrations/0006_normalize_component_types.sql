-- 0006_normalize_component_types.sql
-- Normalizes the component_type catalog to English only:
-- CAPITAL, INTEREST, FEE, INSURANCE (was SEGURO), LATE_FEE (was MORA).
--
-- Same procedure as 0005: SQLite cannot ALTER a CHECK, so the table is
-- recreated. Existing rows written with the Spanish values are translated
-- before the copy, so databases with data do not fail the CHECK on insert.

CREATE TABLE movement_components_new (
    id TEXT PRIMARY KEY,
    movement_id TEXT NOT NULL REFERENCES movements(id),
    component_type TEXT NOT NULL CHECK (component_type IN ('CAPITAL', 'INTEREST', 'FEE', 'INSURANCE', 'LATE_FEE')),
    amount_minor INTEGER NOT NULL,
    description TEXT,
    metadata_json TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

UPDATE movement_components SET component_type = 'INSURANCE' WHERE component_type = 'SEGURO';

UPDATE movement_components SET component_type = 'LATE_FEE' WHERE component_type = 'MORA';

INSERT INTO movement_components_new (
    id, movement_id, component_type, amount_minor, description, metadata_json, created_at, updated_at
)
SELECT
    id, movement_id, component_type, amount_minor, description, metadata_json, created_at, updated_at
FROM movement_components;

DROP TABLE movement_components;

ALTER TABLE movement_components_new RENAME TO movement_components;
