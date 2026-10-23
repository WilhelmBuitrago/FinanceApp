-- 0008_document_period_review.sql
-- Adds the DOCUMENT_PERIOD kind to the unified human-review queue: the
-- statement period could not be read from the extraction, and the processor
-- confirmed it on a silent retry. Resolution asks the human for month+year
-- and derives the parked movement dates against it.
--
-- SQLite cannot ALTER a CHECK constraint, so the table is recreated the
-- same way migration 0005 widened movement_components: new table, copy the
-- existing rows, drop the old one. Safe on populated queues. The index is
-- rebuilt at the end (DROP TABLE drops it with the old table).

CREATE TABLE review_items_new (
    id TEXT PRIMARY KEY,
    kind TEXT NOT NULL CHECK (kind IN ('DUPLICATE', 'CATEGORY', 'EXTRACTION', 'DOCUMENT_PERIOD')),
    payload_json TEXT NOT NULL,
    source_entry_id TEXT,
    status TEXT NOT NULL CHECK (status IN ('PENDING', 'RESOLVED', 'DISCARDED')),
    created_at TEXT NOT NULL,
    resolved_at TEXT
);

INSERT INTO review_items_new (
    id, kind, payload_json, source_entry_id, status, created_at, resolved_at
)
SELECT
    id, kind, payload_json, source_entry_id, status, created_at, resolved_at
FROM review_items;

DROP TABLE review_items;

ALTER TABLE review_items_new RENAME TO review_items;

CREATE INDEX IF NOT EXISTS idx_review_items_status
    ON review_items(status, created_at);
