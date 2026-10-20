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

## Notas conocidas (a resolver en v0.1.1)

- Migraciones duplicadas: `backend/db/migrations/` y
  `backend/internal/db/migrations/` contienen el mismo juego de archivos.
  Falta unificar la fuente de verdad.
- `backend/internal/bootstrap/bootstrap.go` referencia
  `apps/processor/requirements.txt`, pero solo existe `requirements-lock.txt`.
- `cmd/bootstrap/main.go` no pertenece al módulo Go (no compila) y no tiene
  referencias en el repo. Pendiente de decisión: mover a `backend/cmd/` o
  eliminar.
