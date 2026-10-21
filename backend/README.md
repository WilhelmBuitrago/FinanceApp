# Backend — FinanceApp

## Resumen

API HTTP en Go (módulo `financeapp/backend`). Sirve todos los datos del
producto al frontend Tauri sobre `http://127.0.0.1:8081`, persisté en SQLite
local y aplica las migraciones al arrancar. También orquesta el procesador de
Python durante la importación de documentos. Es el único componente que toca
la base de datos.

## Estructura

- [`go.mod`](./go.mod) / [`go.sum`](./go.sum) — módulo Go y dependencias fijadas.
- [`sqlc.yaml`](./sqlc.yaml) — configuración de generación sqlc.
- **`cmd/`** — binario ejecutable.
  - **`financeapp/`**
    - [`main.go`](./cmd/financeapp/main.go) — entrypoint: migraciones, router,
      arranque del servidor. Si las migraciones fallan, termina con error.
- **`internal/`** — código privado del módulo.
  - **`api/`** — capa HTTP.
    - [`router.go`](./internal/api/router.go) — 33 rutas (health, documents,
      accounts, movements, jobs, categories, dashboard, admin/purge, etc.).
    - [`handlers_import.go`](./internal/api/handlers_import.go) — subida y
      confirmación de importaciones.
    - [`processor_runner.go`](./internal/api/processor_runner.go) — invocación
      del procesador Python.
    - [`sprint12_test.go`](./internal/api/sprint12_test.go) — pruebas de la API.
  - **`app/`** — orquestación de casos de uso.
    - [`app.go`](./internal/app/app.go) — wiring de repositorios y router.
    - [`import_service.go`](./internal/app/import_service.go) — flujo de
      importación de extractos.
    - [`import_service_test.go`](./internal/app/import_service_test.go) — pruebas del servicio.
    - [`import_integration_test.go`](./internal/app/import_integration_test.go) — prueba de integración.
    - [`smoke_test.go`](./internal/app/smoke_test.go) — prueba de humo.
  - **`bootstrap/`** — arranque del sistema.
    - [`bootstrap.go`](./internal/bootstrap/bootstrap.go) — base de datos,
      migraciones y rutas de datos.
  - **`config/`** — configuración.
    - [`config.go`](./internal/config/config.go) — lectura de variables de entorno.
  - **`db/`** — persistencia.
    - [`migrations.go`](./internal/db/migrations.go) — lista y aplicación de
      migraciones (embebidas con `go:embed`).
    - [`migrations_test.go`](./internal/db/migrations_test.go) — pruebas de migración.
    - **`migrations/`** — esquema versionado, aplicado en orden.
      - [`0001_initial_schema.sql`](./internal/db/migrations/0001_initial_schema.sql) — esquema inicial.
      - [`0002_import_jobs.sql`](./internal/db/migrations/0002_import_jobs.sql) — trabajos de importación.
      - [`0003_import_review_staging.sql`](./internal/db/migrations/0003_import_review_staging.sql) — revisión de importaciones.
      - [`0004_align_schema_with_generated_queries.sql`](./internal/db/migrations/0004_align_schema_with_generated_queries.sql) — alineación esquema ↔ sqlc.
      - [`0005_expand_component_types.sql`](./internal/db/migrations/0005_expand_component_types.sql) — ampliación del catálogo de componentes.
      - [`0006_normalize_component_types.sql`](./internal/db/migrations/0006_normalize_component_types.sql) — catálogo normalizado a inglés.
    - `models.go`, `querier.go`, `db.go`, `*.sql.go` — **código generado por
      sqlc**: no editar a mano (ver `docs/development.md`).
  - **`domain/`** — entidades puras, sin dependencias de infraestructura.
    - [`account.go`](./internal/domain/account.go), [`category.go`](./internal/domain/category.go),
      [`document.go`](./internal/domain/document.go), [`financial_event.go`](./internal/domain/financial_event.go),
      [`installment_plan.go`](./internal/domain/installment_plan.go), [`movement.go`](./internal/domain/movement.go),
      [`relation.go`](./internal/domain/relation.go), [`source_entry.go`](./internal/domain/source_entry.go).
    - **`entities/`** — [`import_jobs.go`](./internal/domain/entities/import_jobs.go) — entidades de importación.
  - **`log/`** — [`logger.go`](./internal/log/logger.go) — logging estructurado.
  - **`repositories/`** — persistencia SQLite por entidad.
    - [`account.go`](./internal/repositories/account.go), [`category.go`](./internal/repositories/category.go),
      [`document.go`](./internal/repositories/document.go), [`financial_event.go`](./internal/repositories/financial_event.go),
      [`installment_plan.go`](./internal/repositories/installment_plan.go), [`movement.go`](./internal/repositories/movement.go),
      [`source_entry.go`](./internal/repositories/source_entry.go) y sus pruebas.
    - [`util.go`](./internal/repositories/util.go) — `translateError` y taxonomía
      de errores (`ErrNotFound`, `ErrConflict`, `ErrConstraint`, `ErrInternal`,
      `ErrUnavailable`).
- **`db/`** — fuentes SQL de sqlc (entrada del generador).
  - [`accounts.sql`](./db/queries/accounts.sql), [`categories.sql`](./db/queries/categories.sql),
    [`documents.sql`](./db/queries/documents.sql), [`event_relations.sql`](./db/queries/event_relations.sql),
    [`financial_events.sql`](./db/queries/financial_events.sql), [`installment_plans.sql`](./db/queries/installment_plans.sql),
    [`installments.sql`](./db/queries/installments.sql), [`movement_components.sql`](./db/queries/movement_components.sql),
    [`movement_relations.sql`](./db/queries/movement_relations.sql), [`movements.sql`](./db/queries/movements.sql),
    [`source_entries.sql`](./db/queries/source_entries.sql).
- `data/`, `logs/`, `out/` — datos, registros y salidas locales (ignorados por git).

## Estado

- **Verde, verificado el 2026-10-21**: `go build ./...` ✅, `go vet ./...` ✅,
  `go test ./...` ✅ (api, app, db, domain, repositories), `sqlc diff` = 0.
- Migraciones **0001–0006** aplicadas en ese orden; catálogo de componentes de
  movimiento en inglés: `CAPITAL`, `INTEREST`, `FEE`, `INSURANCE`, `LATE_FEE`.
- Persistencia: SQLite local en `%APPDATA%\FinanceApp`; nada sale de la máquina.
- El CI (`ci.yml`) **no corre pruebas**: solo valida estructura del repo y
  `sqlc diff`. Las pruebas se corren en local antes de cada commit.
- Pendiente conocido: las tablas de cuotas (`installment_plans`,
  `installments`) y el desglose de componentes existen en el esquema, pero la
  funcionalidad de tarjetas de crédito todavía no se conecta a la API ni a la
  UI; se consolida más adelante.
