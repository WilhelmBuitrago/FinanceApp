# Arquitectura

FinanceApp es una aplicación desktop-first y local-first. Todos los datos
viven en la máquina del usuario; no hay servicios remotos obligatorios.

## Componentes

```
┌──────────────────────── apps/desktop ────────────────────────┐
│ Ventana React/Vite + shell Tauri (Rust)                      │
│ Al arrancar: spawnea el backend, espera /health, le pasa     │
│ variables de entorno. Al cerrar la ventana, mata el backend. │
└───────────────────────────┬─────────────────────────────────┘
                            │ HTTP 127.0.0.1:8081
┌───────────────────────────▼─────────────────────────────────┐
│ backend/ (Go, módulo financeapp/backend)                     │
│ HTTP API, dominio, persistencia SQLite, migraciones.         │
└───────────────────────────┬─────────────────────────────────┘
                            │ subprocess
┌───────────────────────────▼─────────────────────────────────┐
│ apps/processor/ (Python)                                     │
│ Extracción nativa → OCR → IA, normalización, validación.     │
└──────────────────────────────────────────────────────────────┘
                            │
                    data/ y %APPDATA%\FinanceApp (local, nunca en el repo)
```

## Contrato de artefactos Go ↔ Python (v1)

Por cada procesamiento, el processor escribe en
`data/artifacts/<processing_id>/`:

- `manifest.json` — `processing_id`, `processor_version`, `contract_version`,
  `input {sha256, filename, size_bytes}`, `stages`, `status`, `validation`,
  `parsed_metadata`, `movements_count`.
- `inspection.json` — `file {sha256, size_bytes}`, `pages_count`, `warnings`.
- `extraction/pages/<NNN>.txt` — texto nativo por página.
- `parsed.json` — `metadata`, `summary`, `movements[]`.

Go lee `manifest.json` y `parsed.json`. La ruta base se configura con
`ARTIFACTS_PATH`.

## Runtime de Python

- Versión fijada: **3.12.10 x64 embebible**, verificada por SHA256.
- Dependencias fijas en `apps/processor/requirements-lock.txt`.
- Ruedas pre-descargadas (`resources/python-wheels`) para instalación offline.
- Bootstrap idempotente con marcador de estado (`.python-ready`).

## Notas de dominio

- **Catálogo de componentes de movimiento** — regla dura en el plano, en
  inglés: `CAPITAL`, `INTEREST`, `FEE` (cuota de manejo), `INSURANCE`,
  `LATE_FEE`. Vive como CHECK en `movement_components.component_type`
  (migraciones 0005 y 0006). Cada tarjeta de crédito es un instrumento
  separado asociado a una cuenta (una tarjeta Nu no equivale a una de otro
  banco), así que el catálogo crece por emisor. SQLite no permite alterar un
  CHECK: ampliarlo exige recrear la tabla (procedimiento en
  `0005_expand_component_types.sql`), barato mientras la tabla está vacía.
  La migración 0006 normalizó los valores a inglés y traduce los datos
  existentes (`SEGURO`→`INSURANCE`, `MORA`→`LATE_FEE`) antes de copiarlos.
- **Cuotas por plan** — el índice único
  `(installment_plan_id, installment_number)` garantiza un número de cuota por
  plan. Pendiente de decisión: procedimiento si una base preexistente llegara
  a tener duplicados al momento de crear el índice.

## Historial de limpieza

- Migraciones unificadas en `internal/db/migrations/` (fuente de producción,
  embebida con `go:embed`); `sqlc.yaml` y las pruebas apuntan ahí.
- `EnsurePythonEnv` (que referenciaba `apps/processor/requirements.txt`,
  inexistente) fue eliminado: su único consumidor era el bootstrap muerto.
  La instalación usa `requirements-lock.txt` offline con wheels verificados.
