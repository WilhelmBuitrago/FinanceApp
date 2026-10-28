# Backend — FinanceApp

## Resumen

API HTTP en Go (módulo `financeapp/backend`). Sirve todos los datos del
producto al frontend Tauri sobre `http://127.0.0.1:8081`, persiste en SQLite
local y aplica las migraciones al arrancar. También orquesta el procesador de
Python durante la importación de documentos. Es el único componente que toca
la base de datos.

Versión del componente: **0.1.1**.

## Estructura

- [`go.mod`](./go.mod) / [`go.sum`](./go.sum) — módulo Go y dependencias fijadas.
- [`sqlc.yaml`](./sqlc.yaml) — configuración de generación sqlc.
- **`cmd/`** — binario ejecutable.
  - **`financeapp/`**
    - [`main.go`](./cmd/financeapp/main.go) — entrypoint: config, migraciones,
      validación del runtime de Python, arranque del servidor.
- **`tests/`** — todas las pruebas del backend (package `tests`, caja negra):
  helpers compartidos (`helpers_test.go`), persistencia de prueba y un archivo
  por área.
- **`internal/`** — código privado del módulo.
  - **`api/`** — capa HTTP: 34 rutas sobre un struct `router` con dependencias
    inyectadas.
    - [`router.go`](./internal/api/router.go) — armado de rutas.
    - [`handlers_accounts.go`](./internal/api/handlers_accounts.go) — instituciones, cuentas y sus subrecursos.
    - [`handlers_movements.go`](./internal/api/handlers_movements.go) — crear/listar/borrar movimientos, asignar categoría.
    - [`handlers_documents.go`](./internal/api/handlers_documents.go) — documentos y cadena de trazabilidad.
    - [`handlers_jobs.go`](./internal/api/handlers_jobs.go) — estado y artefactos de las importaciones.
    - [`handlers_dashboard.go`](./internal/api/handlers_dashboard.go) — flujo de caja y distribuciones.
    - [`handlers_categories.go`](./internal/api/handlers_categories.go) — CRUD de categorías.
    - [`handlers_admin.go`](./internal/api/handlers_admin.go) — purge administrativo.
    - [`handlers_import.go`](./internal/api/handlers_import.go) — subida de documentos; el staging delega en `ImportService`.
    - [`handlers_review.go`](./internal/api/handlers_review.go) — cola de revisión humana: resolver, descartar y resolver período de documento.
    - [`middleware.go`](./internal/api/middleware.go) — CORS, registro de requests y captura de status.
    - [`response.go`](./internal/api/response.go) — helpers `encodeJSON` / `writeRepoError` / `writeLookupError`.
    - [`processor_runner.go`](./internal/api/processor_runner.go) — invocación del procesador Python (stdout = contrato JSON, stderr = log).
  - **`app/`** — orquestación de casos de uso.
    - [`app.go`](./internal/app/app.go) — wiring de repositorios, router y ciclo de vida.
    - [`import_service.go`](./internal/app/import_service.go) — flujo de
      importación de extractos (transaccional, único dueño del staging).
    - [`movement_service.go`](./internal/app/movement_service.go) — creación
      transaccional de evento + movimiento + componentes.
    - [`movement_write.go`](./internal/app/movement_write.go) — escritura
      transaccional compartida por movimientos y correcciones de revisión.
    - [`review_service.go`](./internal/app/review_service.go) — cola unificada
      de revisión (duplicados, categorías, extracción, período de documento).
    - [`period_dates.go`](./internal/app/period_dates.go) — derivación de
      fechas contra el período confirmado (meses ES/EN, años bisiestos).
  - **`bootstrap/`** — arranque del sistema.
    - [`paths.go`](./internal/bootstrap/paths.go) — rutas de datos por entorno.
    - [`python_runtime.go`](./internal/bootstrap/python_runtime.go) — validación
      del runtime de Python preparado por el instalador.
  - **`config/`** — configuración.
    - [`config.go`](./internal/config/config.go) — lectura de variables de entorno.
  - **`db/`** — persistencia.
    - [`migrations.go`](./internal/db/migrations.go) — lista y aplicación de
      migraciones (embebidas con `go:embed`; cada archivo se parte por `;`).
    - **`migrations/`** — esquema versionado, aplicado en orden.
      - [`0001_initial_schema.sql`](./internal/db/migrations/0001_initial_schema.sql) — esquema inicial.
      - [`0002_import_jobs.sql`](./internal/db/migrations/0002_import_jobs.sql) — trabajos de importación.
      - [`0003_import_review_staging.sql`](./internal/db/migrations/0003_import_review_staging.sql) — revisión de importaciones.
      - [`0004_align_schema_with_generated_queries.sql`](./internal/db/migrations/0004_align_schema_with_generated_queries.sql) — alineación esquema ↔ sqlc.
      - [`0005_expand_component_types.sql`](./internal/db/migrations/0005_expand_component_types.sql) — ampliación del catálogo de componentes.
      - [`0006_normalize_component_types.sql`](./internal/db/migrations/0006_normalize_component_types.sql) — catálogo normalizado a inglés.
      - [`0007_review_items.sql`](./internal/db/migrations/0007_review_items.sql) — cola de revisión humana unificada.
      - [`0008_document_period_review.sql`](./internal/db/migrations/0008_document_period_review.sql) — kind `DOCUMENT_PERIOD` en la cola.
    - `models.go`, `querier.go`, `db.go`, `*.sql.go` — **código generado por
      sqlc**: no editar a mano (ver `docs/development.md`).
  - **`domain/`** — entidades puras, sin dependencias de infraestructura: un
    archivo por tipo ([account](./internal/domain/account.go),
    [category](./internal/domain/category.go),
    [document](./internal/domain/document.go),
    [financial_event](./internal/domain/financial_event.go),
    [import_job](./internal/domain/import_job.go),
    [installment_plan](./internal/domain/installment_plan.go),
    [movement](./internal/domain/movement.go),
    [relation](./internal/domain/relation.go),
    [source_entry](./internal/domain/source_entry.go)).
  - **`log/`** — [`logger.go`](./internal/log/logger.go) — logging a archivo diario
    con mutex para escrituras concurrentes de los handlers.
  - **`repositories/`** — persistencia SQLite por entidad (100 % sqlc).
    - [`account.go`](./internal/repositories/account.go), [`category.go`](./internal/repositories/category.go),
      [`document.go`](./internal/repositories/document.go), [`financial_event.go`](./internal/repositories/financial_event.go),
      [`installment_plan.go`](./internal/repositories/installment_plan.go), [`movement.go`](./internal/repositories/movement.go),
      [`source_entry.go`](./internal/repositories/source_entry.go), [`relation.go`](./internal/repositories/relation.go),
      [`import_job_repo.go`](./internal/repositories/import_job_repo.go),
      [`review_items_repository.go`](./internal/repositories/review_items_repository.go).
    - [`util.go`](./internal/repositories/util.go) — `translateError`, conversores
      `sql.Null*` y taxonomía de errores (`ErrNotFound`, `ErrConflict`,
      `ErrConstraint`, `ErrInternal`, `ErrUnavailable`).
    - [`persistence.go`](./internal/repositories/persistence.go) — wiring de repositorios.
    - [`sqlite.go`](./internal/repositories/sqlite.go) — tipo `DB` (Querier compartido).
    - Nota: cuotas (`installment_plan.go`) y relaciones (`relation.go`) son
      scaffolding para features planeadas: completos y testeados, sin rutas
      que los consuman todavía (ver comentario en cada archivo).
- **`db/`** — fuentes SQL de sqlc (entrada del generador).
  - [`accounts.sql`](./db/queries/accounts.sql), [`categories.sql`](./db/queries/categories.sql),
    [`documents.sql`](./db/queries/documents.sql), [`event_relations.sql`](./db/queries/event_relations.sql),
    [`financial_events.sql`](./db/queries/financial_events.sql), [`import_jobs.sql`](./db/queries/import_jobs.sql),
    [`installment_plans.sql`](./db/queries/installment_plans.sql),
    [`installments.sql`](./db/queries/installments.sql), [`movement_components.sql`](./db/queries/movement_components.sql),
    [`movement_relations.sql`](./db/queries/movement_relations.sql), [`movements.sql`](./db/queries/movements.sql),
    [`review_items.sql`](./db/queries/review_items.sql), [`source_entries.sql`](./db/queries/source_entries.sql).
- `data/`, `logs/`, `out/` — datos, registros y salidas locales (ignorados por git).

## Estado

- **Verde, verificado el 2026-10-28**: `go build ./...`, `go vet ./...`,
  `go test ./...` ✅ (suite completa en `tests/`), `gofmt`/`goimports` sin
  diferencias, `golangci-lint` 0 findings, `sqlc diff` = 0.
- Migraciones **0001–0008** aplicadas en ese orden; cola de revisión con
  kinds `DUPLICATE`, `CATEGORY`, `EXTRACTION` y `DOCUMENT_PERIOD`.
- Persistencia: SQLite local en `%APPDATA%\FinanceApp`; nada sale de la máquina.
- El CI (`ci.yml`) **no corre pruebas**: solo valida `sqlc diff`. Las pruebas
  se corren en local antes de cada commit.
- El instalador MSI prepara el runtime de Python (`prepare-python.ps1` +
  marcador `.python-ready`); el backend solo lo valida al arrancar.
- Pendiente conocido: las tablas de cuotas (`installment_plans`,
  `installments`) y el desglose de componentes existen en el esquema, pero la
  funcionalidad de tarjetas de crédito todavía no se conecta a la API ni a la
  UI; se consolida más adelante.
