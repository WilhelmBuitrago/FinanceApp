-- 0011_rescale_imported_amounts_to_centavos.sql
-- The ledger's unit is the centavo (amount_minor): the manual path wrote
-- centavos from day one, but the processor's normalize_amount returned
-- whole pesos ('+$700.000,00' -> 700000), so every imported movement was
-- stored 100 times smaller than what it means and the UI divided it again.
-- The processor now returns centavos, decimals preserved, and this migration
-- rescales the rows the old convention already wrote.
--
-- Scope: only imported rows (source_entry_id IS NOT NULL carries the
-- document's own amount). Manual rows were always centavos.
--
-- Idempotent by condition, not by hope: Apply re-runs every migration on
-- each startup and a bare UPDATE would multiply a second time. The marker
-- column (ignored by every query) records that a row was rescaled once.
--
-- Additive for the reader: existing queries select named columns, so the
-- new markers are invisible to them.

ALTER TABLE movements ADD COLUMN amount_scaled_to_centavos INTEGER NOT NULL DEFAULT 0;

UPDATE movements
SET amount_minor = amount_minor * 100,
    amount_scaled_to_centavos = 1
WHERE source_entry_id IS NOT NULL
  AND amount_scaled_to_centavos = 0;

ALTER TABLE source_entries ADD COLUMN normalized_scaled_to_centavos INTEGER NOT NULL DEFAULT 0;

UPDATE source_entries
SET normalized_amount_minor = normalized_amount_minor * 100,
    normalized_scaled_to_centavos = 1
WHERE normalized_amount_minor IS NOT NULL
  AND normalized_scaled_to_centavos = 0;
