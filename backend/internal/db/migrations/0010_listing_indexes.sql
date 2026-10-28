-- 0010_listing_indexes.sql
-- The listings the audit found unbacked: every movement query orders by
-- created_at DESC, the account listing filters by account_id, and the
-- hydration joins match financial_event_id. 0009's composite natural key
-- covers account_id only as a prefix (it cannot serve an ORDER BY), and
-- neither financial_event_id nor a date-only lookup has any coverage.
--
-- Additive and idempotent (IF NOT EXISTS): applies cleanly over an
-- existing database. created_at carries DESC to match the ORDER BY.

CREATE INDEX IF NOT EXISTS idx_movements_created_at
    ON movements (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_movements_account_id
    ON movements (account_id);

CREATE INDEX IF NOT EXISTS idx_movements_financial_event_id
    ON movements (financial_event_id);

CREATE INDEX IF NOT EXISTS idx_movements_movement_date
    ON movements (movement_date);
