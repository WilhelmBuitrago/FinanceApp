-- 0005_expand_component_types.sql
-- Widens the component_type catalog to CAPITAL, INTEREST, FEE (cuota de manejo),
-- SEGURO and MORA, for credit card breakdowns coming from issuers such as Nu.
--
-- SQLite cannot ALTER a CHECK constraint, so the only way to change it is to
-- recreate the table. This migration recreates movement_components with the
-- wider catalog and copies the existing rows before dropping the old table.
-- Safe on populated tables too, and trivially cheap while components are empty.

CREATE TABLE movement_components_new (
    id TEXT PRIMARY KEY,
    movement_id TEXT NOT NULL REFERENCES movements(id),
    component_type TEXT NOT NULL CHECK (component_type IN ('CAPITAL', 'INTEREST', 'FEE', 'SEGURO', 'MORA')),
    amount_minor INTEGER NOT NULL,
    description TEXT,
    metadata_json TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

INSERT INTO movement_components_new (
    id, movement_id, component_type, amount_minor, description, metadata_json, created_at, updated_at
)
SELECT
    id, movement_id, component_type, amount_minor, description, metadata_json, created_at, updated_at
FROM movement_components;

DROP TABLE movement_components;

ALTER TABLE movement_components_new RENAME TO movement_components;
