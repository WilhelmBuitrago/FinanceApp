PRAGMA foreign_keys = ON;

ALTER TABLE source_entries ADD COLUMN review_status TEXT NOT NULL DEFAULT 'VALID';
ALTER TABLE source_entries ADD COLUMN source_key TEXT;
ALTER TABLE source_entries ADD COLUMN corrected_description TEXT;
ALTER TABLE source_entries ADD COLUMN corrected_amount_minor INTEGER;
ALTER TABLE source_entries ADD COLUMN corrected_date TEXT;

CREATE INDEX IF NOT EXISTS idx_source_entries_document_extraction ON source_entries(document_id, extraction_id);

CREATE UNIQUE INDEX IF NOT EXISTS uq_source_entries_key ON source_entries(document_id, extraction_id, source_key) WHERE source_key IS NOT NULL;
