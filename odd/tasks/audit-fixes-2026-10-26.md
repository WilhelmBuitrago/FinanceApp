# Cierres de la auditoría 2026-10-26 (escala, edge cases, estandarización)

- **Objetivo:** ejecutar los hallazgos de la auditoría read-only del 2026-10-26
  que el mantenedor autorizó ("creo que todas son de bajo riesgo... hazlo"):
  índices faltantes, LIMIT en listados, IDs predecibles, evento huérfano al
  borrar, parseo de montos con coma, error de PDF protegido,
  estandarización de timestamps.
- **Rama:** `developer`. **Estrategia:** un commit por unidad, tests primero
  (RED cuando aplica), batería por unidad.
- **Runner backend:** `go test -count=1 ./...` desde `backend/`. **Runner
  desktop:** `npx vitest run` desde `apps/desktop/`. **Runner processor:**
  `<env-python> tests/<file>.py` desde `apps/processor/`.

## Specs

- S1 — Usuario: "Las migraciones tienen 4 índices en total... Faltan los que
  las queries reales usan: movements(account_id), movements(financial_event_id),
  movements(movement_date) y —el que más importa— movements(created_at), que
  es el ORDER BY created_at DESC de todos los listings."
- S2 — Usuario: "Ninguna query tiene LIMIT: /movements, documentos,
  categorías, dashboard y review cargan la tabla completa a memoria."
- S3 — Usuario: "IDs predecibles: manual_movement.go:63,82 y
  transfer_service.go:60 generan IDs con time.Now().UnixNano(). Un doble POST
  en la misma fracción de segundo colisiona (PK duplicada → 500). Consolidar en
  un helper con crypto/rand."
- S4 — Usuario: "Evento huérfano: MovementRepository.Delete (movement.go:215)
  borra el movimiento y no toca su financial_event. Borrar la última pierna de
  un egreso deja un evento fantasma —y no hay ON DELETE CASCADE en ninguna
  migración."
- S5 — Usuario: "Drawer: la coma colombiana se pierde (AddMovementDrawer.tsx:
  64): parseFloat("25,5") = 25. El usuario escribe 25,5 y el sistema guarda
  2500 centavos sin chistar. formatMoney existe; falta un parseAmountMinor
  equivalente."
- S6 — Usuario: "PDF con contraseña (apps/processor/main.py:130): el usuario
  ve traceback crudo en vez de 'el archivo está protegido'."
- S7 — Usuario: "timeNowUTC() existe (handlers_accounts.go:139) pero hay 18
  llamadas inline de time.Now().UTC().Format(time.RFC3339) repartidas en
  api/app."
- S8 — Usuario: "normalize_amount('   ') devuelve 0 en vez de ValueError
  (nu_normalizer.py:86-98). El test documenta el gap pero la función nunca se
  corrigió."
- Fuera de scope (decisión del mantenedor, NO tocar): los 4 métodos de repo
  muertos (ListByType, ListByDocument, ListComponents, ListByUserID).

## Tasks

| ID | Unidad | Ruta | Estado |
|----|--------|------|--------|
| T1 | Migración 0010 con los 4 índices + test de contrato de esquema | inline | done (4adc20b) |
| T2 | Helper de IDs con crypto/rand (manual + transfer) + test de unicidad | inline | done (fb55e08) |
| T3 | LIMIT parametrizado en los listados que crecen + tests HTTP del tope | inline | done (139c453) |
| T4 | Borrado sin dejar evento huérfano (movimientos + relaciones + evento) + tests | inline | done (aa8abef) |
| T5 | Timestamps: helper único `domain.NowRFC3339()`, fuera los 18 inline | inline | done (9b05551) |
| T6 | `parseAmountMinor` (coma decimal colombiana) + drawer + tests | inline | done (91cd809) |
| T7 | `normalize_amount` whitespace → ValueError + error claro de PDF protegido | inline | done (04c8b55) |

## Log

- L1 (usuario, 2026-10-26): "Crep que todas son de bajo riesgo, ignorando los
  4 metodos muertos. Creo que puedes hacer de pasada todo esto: [...] por
  favor, hazlo" (hallazgos de la auditoría 2026-10-26, obs 96).

- L2 (2026-10-26): las 7 unidades quedaron implementadas, verificadas y
  commiteadas. Commits: `4adc20b` (índices 0010 + test de esquema con
  EXPLAIN QUERY PLAN), `fb55e08` (IDs crypto/rand + conexión SQLite con
  busy_timeout/WAL: la ráfaga de 50 transferencias destapó SQLITE_BUSY),
  `139c453` (LIMIT 500/5000 en los 6 listados del histórico),
  `aa8abef` (evento huérfano: se va con su última pierna; el segundo
  delete sigue siendo no-op), `9b05551` (timestamps: 18 inline →
  domain.NowRFC3339), `91cd809` (parseAmountMinor + input decimal: la
  coma colombiana ya no se pierde), `04c8b55` (whitespace ValueError +
  PDF protegido con nombre).
- L3: cada unidad con RED primero donde el comportamiento lo permitía,
  verificación por mutación antes de cerrar (índices: neutralizar el
  join no aplica, se pinza el plan del EXPLAIN; IDs: ráfaga concurrente;
  huérfano: quitar el DeleteFinancialEvent; parser: restaurar
  parseFloat; whitespace: quitar el raise; password: quitar el
  needs_pass) y batería verde por componente al cierre: backend
  build/vet/test/gofmt/lint/sqlc-diff, desktop build/lint/217
  tests/prettier, processor 9 suites + ruff/compileall/tabnanny.
- L4: proceso honesto: un commit (T5) se hizo con el build roto por
  ordenar mal batería y commit; se enmendó antes de pushear. El campo
  de cantidad era type="number": el navegador rechazaba la coma antes
  de que el parser la viera —cambiado a text+inputMode decimal.
- L5: los 4 métodos muertos siguen fuera de scope por decisión del
  mantenedor. Review nativa RDD: transporte caído, sin PASS inventado.
