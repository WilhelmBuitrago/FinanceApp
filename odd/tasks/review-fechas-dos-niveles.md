# Two-level date review: document period vs movement dates

Objective: when the statement PERIOD is unreadable, the pipeline parks one
document-level review item and re-derives the movement dates once the
human fixes the period. When a MOVEMENT date is unreadable while the
period is fine, only that movement is parked (kind EXTRACTION, same as
amounts today). The import never blocks on dates. Branch: developer.
Runners: go test ./... from backend/;
C:\Users\wilhe\miniconda3\envs\agent\python.exe tests/test_nu_account_rules.py
from apps/processor.

## Current mechanism (verified)

- normalization.nu_normalizer.normalize_date(period_str, date_original)
  extracts year+month from the PERIOD FIRST (regex ([A-Za-z]{3})\s+(\d{4})
  over period_str) and only then parses the movement date. An unreadable
  period raises before touching any date: EVERY movement date then fails
  and the parser sets date_normalized = '' silently (nu_account.py, the
  except around normalize_date).
- validation.validator marks the whole document invalid when any movement
  has an empty normalized date. One header defect blocks the entire import.
- review_items.kind has a SQL CHECK ('DUPLICATE','CATEGORY','EXTRACTION')
  in migration 0007: a new kind requires a migration.
- app.ReviewService.Resolve dispatches by kind (switch) inside one
  transaction; EXTRACTION marks the source entry resolved and creates the
  movement from the correction.

## Specs

- S1: An unreadable period produces exactly ONE parked review item of a new
  kind (DOCUMENT_PERIOD) carrying the raw period, the document id and the
  account id. No per-movement date items are created for this cause.
- S2: The import of that document is NOT blocked: movements import with
  their own dates whenever derivable (movement year first, period year as
  fallback); the validator does not invalidate the document for parked
  dates.
- S3: Resolving the DOCUMENT_PERIOD item records the corrected period on
  the document (period_start/period_end) and marks the item RESOLVED. It
  never rewrites movements. Discard keeps the movements already imported.
- S4: With a readable period, an unreadable MOVEMENT date parks only that
  movement (kind EXTRACTION, reason "unparseable date"), symmetric with
  the existing amount behavior. The rest of the document imports.
- S5: Migration 0008 widens the review_items kind CHECK additively (create
  new table + copy + swap), per the repo migration rules; sqlc diff stays
  clean.

## Tasks

| ID | Specs | Route | Status |
| --- | --- | --- | --- |
| T1 | S1,S2,S4 | feature doc + memory mirror (this document) | done |
| T2 | S1,S4 | parser: period check before the loop; per-movement date park (RED tests) | pending |
| T3 | S1,S2 | processor contract: document_review in parsed/manifest; validator accepts derived dates | pending |
| T4 | S1,S3,S5 | backend: migration 0008, kind DOCUMENT_PERIOD, resolve re-derives dates | pending |
| T5 | S1,S5 | integration: ImportService parks document review item | pending |
| T6 | all | GREEN + full verification | pending |
| T7 | S1,S3 | frontend document-review UI | DEFERRED (frontend pass comes later) |

## Decisions (resolved with the maintainer)

- D1 (REVISED): NO date re-derivation on resolve. The period always comes
  from the extraction for Nu ("Período" / "01 - 30 ABR 2026"): if it did
  not parse, that is a parser defect, not an ambiguous document. Resolving
  the document item only records the corrected period on the document;
  movements are never rewritten. (Maintainer: "Las fechas de movimientos
  pueden ayudar como comprobación mas no como verdad.")
- D2: with the period parked, movements still import (ACTIVE) as long as
  their own date is derivable. A movement without a derivable date is
  parked individually (S4), never silently emptied.
- D3: migration 0008 widens the review_items kind CHECK (create+copy+swap).
- D4: normalize_date prefers the year on the movement's own date, falling
  back to the period's year. A movement carrying "10 abr 2026" normalizes
  even when the period is unreadable.

## Reprogrammed mechanism (parser first)

1. _parse_metadata reads the real period format robustly (same line or the
   next one, with/without accent, "01 - 30 ABR 2026" style ranges).
2. If no period parses: the parser emits ONE document-level review item
   (kind DOCUMENT_PERIOD downstream) with the raw text, and keeps parsing
   movements on their own.
3. normalize_date: year from the movement date if present, else from the
   period; neither -> the movement parks as EXTRACTION ("unparseable
   date"), the rest of the document imports.
4. Backend: resolve of DOCUMENT_PERIOD updates the document period fields
   and marks the item RESOLVED; no movement rewrite.

## Known risks

- The period feeds the YEAR of every movement. A corrected period whose
  month/year shifts the movement across a year boundary (e.g. 31 dic vs
  01 ene) re-derives each date with the corrected year, but a movement
  dated in a month different from the period month keeps its own month
  (current normalize_date behavior, pinned).
- Discard of a document-period item leaves movements permanently without
  normalized dates: the UI must show them as pending, not as clean data.

## Log

- L1 (user, verbatim): "Sigue con el documento feature para hacer
  'período del documento ilegible → review de documento (¿el período es X?
  Sí / corregir), fecha de movimiento ilegible → review de movimiento'."
- L2 (user, earlier, verbatim): "aunque la fecha del documento falle puede
  que no falle la fecha de los movimientos de ese documento... una cosa es
  fallar con la fecha del documento y otra con la de los movimientos."
