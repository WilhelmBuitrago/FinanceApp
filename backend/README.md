# Backend — FinanceApp

## Resumen

API HTTP en Go (módulo `financeapp/backend`). Sirve todos los datos del
producto al frontend Tauri sobre `http://127.0.0.1:8081`, persiste en SQLite
local y aplica las migraciones al arrancar. También orquesta el procesador de
Python durante la importación de documentos. Es el único componente que toca
la base de datos.

## Estructura

- [`go.mod`](./go.mod) / [`go.sum`](./go.sum) — módulo Go y dependencias fijadas.
- [`sqlc.yaml`](./sqlc.yaml) — configuración de generación sqlc.
- **`cmd/`** — binario ejecutable.
  - **`financeapp/`**
    - [`main.go`](./cmd/financeapp/main.go) — entrypoint: config, migraciones,
      validación del runtime de Python, arranque del servidor.
- **`tests/`** — todas las pruebas del backend (package `tests`, caja negra):
  helpers compartidos, persistencia de prueba y un archivo por área.
- **`internal/`** — código privado del módulo.
  - **`api/`** — capa HTTP.
    - [`router.go`](./internal/api/router.go) — armado de las 33 rutas sobre un
      struct `router` con dependencias inyectadas.
    - [`handlers_accounts.go`](./internal/api/handlers_accounts.go) — instituciones, cuentas y sus subrecursos.
    - [`handlers_movements.go`](./internal/api/handlers_movements.go) — crear/listar/borrar movimientos, asignar categoría.
    - [`handlers_documents.go`](./internal/api/handlers_documents.go) — documentos y cadena de trazabilidad.
    - [`handlers_jobs.go`](./internal/api/handlers_jobs.go) — artefactos y confirm de importaciones.
    - [`handlers_dashboard.go`](./internal/api/handlers_dashboard.go) — flujo de caja y distribuciones.
    - [`handlers_categories.go`](./internal/api/handlers_categories.go) — CRUD de categorías.
    - [`handlers_admin.go`](./internal/api/handlers_admin.go) — purge administrativo.
    - [`handlers_import.go`](./internal/api/handlers_import.go) — subida de documentos y staging del confirm.
    - [`middleware.go`](./internal/api/middleware.go) — CORS, registro de requests y captura de status.
    - [`response.go`](./internal/api/response.go) — helpers `encodeJSON` / `writeRepoError` / `writeLookupError`.
    - [`processor_runner.go`](./internal/api/processor_runner.go) — invocación del procesador Python.
  - **`app/`** — orquestación de casos de uso.
    - [`app.go`](./internal/app/app.go) — wiring de repositorios, router y ciclo de vida.
    - [`import_service.go`](./internal/app/import_service.go) — flujo de
      importación de extractos (transaccional).
    - [`movement_service.go`](./internal/app/movement_service.go) — creación
      transaccional de evento + movimiento + componentes.
  - **`bootstrap/`** — arranque del sistema.
    - [`paths.go`](./internal/bootstrap/paths.go) — rutas de datos por entorno.
    - [`python_runtime.go`](./internal/bootstrap/python_runtime.go) — validación
      del runtime de Python preparado por el instalador.
  - **`config/`** — configuración.
    - [`config.go`](./internal/config/config.go) — lectura de variables de entorno.
  - **`db/`** — persistencia.
    - [`migrations.go`](./internal/db/migrations.go) — lista y aplicación de
      migraciones (embebidas con `go:embed`).
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
  - **`log/`** — [`logger.go`](./internal/log/logger.go) — logging a archivo diario
    con mutex para escrituras concurrentes de los handlers.
  - **`repositories/`** — persistencia SQLite por entidad (100 % sqlc).
    - [`account.go`](./internal/repositories/account.go), [`category.go`](./internal/repositories/category.go),
      [`document.go`](./internal/repositories/document.go), [`financial_event.go`](./internal/repositories/financial_event.go),
      [`installment_plan.go`](./internal/repositories/installment_plan.go), [`movement.go`](./internal/repositories/movement.go),
      [`source_entry.go`](./internal/repositories/source_entry.go), [`relation.go`](./internal/repositories/relation.go),
      [`import_job_repo.go`](./internal/repositories/import_job_repo.go).
    - [`util.go`](./internal/repositories/util.go) — `translateError`, conversores
      `sql.Null*` y taxonomía de errores (`ErrNotFound`, `ErrConflict`,
      `ErrConstraint`, `ErrInternal`, `ErrUnavailable`).
    - [`persistence.go`](./internal/repositories/persistence.go) — wiring de repositorios.
    - [`sqlite.go`](./internal/repositories/sqlite.go) — tipo `DB` (Querier compartido).
- **`db/`** — fuentes SQL de sqlc (entrada del generador).
  - [`accounts.sql`](./db/queries/accounts.sql), [`categories.sql`](./db/queries/categories.sql),
    [`documents.sql`](./db/queries/documents.sql), [`event_relations.sql`](./db/queries/event_relations.sql),
    [`financial_events.sql`](./db/queries/financial_events.sql), [`import_jobs.sql`](./db/queries/import_jobs.sql),
    [`installment_plans.sql`](./db/queries/installment_plans.sql),
    [`installments.sql`](./db/queries/installments.sql), [`movement_components.sql`](./db/queries/movement_components.sql),
    [`movement_relations.sql`](./db/queries/movement_relations.sql), [`movements.sql`](./db/queries/movements.sql),
    [`source_entries.sql`](./db/queries/source_entries.sql).
- `data/`, `logs/`, `out/` — datos, registros y salidas locales (ignorados por git).

## Estado

- **Verde, verificado el 2026-10-21**: `go build ./...` ✅, `go vet ./...` ✅,
  `go test ./...` ✅ (34 pruebas en `tests/`), `sqlc diff` = 0.
- Migraciones **0001–0006** aplicadas en ese orden; catálogo de componentes de
  movimiento en inglés: `CAPITAL`, `INTEREST`, `FEE`, `INSURANCE`, `LATE_FEE`.
- Persistencia: SQLite local en `%APPDATA%\FinanceApp`; nada sale de la máquina.
- El CI (`ci.yml`) **no corre pruebas**: solo valida estructura del repo y
  `sqlc diff`. Las pruebas se corren en local antes de cada commit.
- El instalador MSI prepara el runtime de Python (`prepare-python.ps1` +
  marcador `.python-ready`); el backend solo lo valida al arrancar.
- Pendiente conocido: las tablas de cuotas (`installment_plans`,
  `installments`) y el desglose de componentes existen en el esquema, pero la
  funcionalidad de tarjetas de crédito todavía no se conecta a la API ni a la
  UI; se consolida más adelante.
