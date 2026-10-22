-- 0007_review_items.sql
-- Unified human-review queue. Today three triggers land here:
--   DUPLICATE  - a movement that collided with an existing one
--   CATEGORY   - movements left without category after a category removal
--   EXTRACTION - a movement the parser could not normalize (raw preserved in
--                the referenced source_entry)
-- The payload carries the data the UI shows and the resolve action needs.

CREATE TABLE IF NOT EXISTS review_items (
    id TEXT PRIMARY KEY,
    kind TEXT NOT NULL CHECK (kind IN ('DUPLICATE', 'CATEGORY', 'EXTRACTION')),
    payload_json TEXT NOT NULL,
    source_entry_id TEXT,
    status TEXT NOT NULL CHECK (status IN ('PENDING', 'RESOLVED', 'DISCARDED')),
    created_at TEXT NOT NULL,
    resolved_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_review_items_status
    ON review_items(status, created_at);
