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
| T1 | Migración 0010 con los 4 índices + test de contrato de esquema | inline | pending |
| T2 | Helper de IDs con crypto/rand (manual + transfer) + test de unicidad | inline | pending |
| T3 | LIMIT parametrizado en los listados que crecen + tests HTTP del tope | inline | pending |
| T4 | Borrado sin dejar evento huérfano (movimientos + relaciones + evento) + tests | inline | pending |
| T5 | Timestamps: helper único `domain.NowRFC3339()`, fuera los 18 inline | inline | pending |
| T6 | `parseAmountMinor` (coma decimal colombiana) + drawer + tests | inline | pending |
| T7 | `normalize_amount` whitespace → ValueError + error claro de PDF protegido | inline | pending |

## Log

- L1 (usuario, 2026-10-26): "Crep que todas son de bajo riesgo, ignorando los
  4 metodos muertos. Creo que puedes hacer de pasada todo esto: [...] por
  favor, hazlo" (hallazgos de la auditoría 2026-10-26, obs 96).
