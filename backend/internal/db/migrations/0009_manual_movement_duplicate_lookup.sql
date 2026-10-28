-- 0009_manual_movement_duplicate_lookup.sql
-- Backs the manual-entry duplicate check: the handler used to load the
-- whole movements table into memory on every POST /movements to hunt a
-- collision. The lookup is now one indexed query.
--
-- Additive and idempotent (IF NOT EXISTS): applies cleanly over an
-- existing database.

CREATE INDEX IF NOT EXISTS idx_movements_natural_key
    ON movements (account_id, movement_date, amount_minor);
