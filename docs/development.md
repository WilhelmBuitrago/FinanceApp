# Guía de desarrollo

## Requisitos

| Componente | Versión | Notas |
| --- | --- | --- |
| Go | la declarada en `backend/go.mod` | módulo único `financeapp/backend`, vive en `backend/` |
| Node + npm | 20 | frontend de escritorio |
| Python | 3.12 | el runtime productivo es el embebible 3.12.10 x64 |

## Backend (Go)

```bash
cd backend
go run ./cmd/financeapp
```

- Expone HTTP en `127.0.0.1:8081`. Health check: `GET /health`.
- Pruebas: `go test ./...` desde `backend/`.

## Desktop (Tauri + React)

Opción A — solo la interfaz:

```bash
cd apps/desktop
npm ci
npm run dev        # Vite en http://localhost:5174
```

Opción B — aplicación de escritorio completa:

```bash
cd backend
go build -o ../apps/desktop/src-tauri/resources/financeapp-backend.exe ./cmd/financeapp
cd ../apps/desktop
npx tauri dev
```

- El shell de Tauri spawnea el backend desde
  `src-tauri/resources/financeapp-backend.exe`, espera `/health` y le pasa las
  variables de entorno. Si el ejecutable no está en `resources/`, la app
  falla al arrancar: por eso la opción B lo compila primero.
- Build de producción: `npm run build` (genera `dist/`) + `cargo tauri build`
  (bundle MSI). Es exactamente lo que hace el workflow de releases.

## Procesador (Python)

```bash
cd apps/processor
python -m pytest tests
```

- Punto de entrada: `src/main.py` (lo invoca el backend como subprocess).
- Dependencias fijas en `requirements-lock.txt` (el build de release falla si
  no existe). Instalación local: `pip install -r requirements-lock.txt`.

## Variables de entorno del backend

Tauri las setea al spawnear el backend (`apps/desktop/src-tauri/src/main.rs`):

| Variable | Valor en la app empaquetada |
| --- | --- |
| `PORT` | `8081` |
| `DATABASE_PATH` | `%APPDATA%\FinanceApp\data\db\finance.db` |
| `ARTIFACTS_PATH` | `%APPDATA%\FinanceApp\data\artifacts` |
| `LOGS_PATH` | `%APPDATA%\FinanceApp\logs` |
| `PROCESSOR_SCRIPT` | `resources/processor/src/main.py` |
| `PROCESSOR_REQUIREMENTS` | `resources/processor/requirements-lock.txt` |
| `PROCESSOR_PYTHON` | venv indicado por el marcador `.python-ready` |
| `PYTHON_BUNDLE_PATH` / `PYTHON_WHEELS_PATH` | recursos opcionales para modo offline |

- El instalador deja el marcador `%PROGRAMDATA%\FinanceApp\.python-ready`
  con el `venv_path` creado por `prepare-python.ps1`.
