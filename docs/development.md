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
- Intérprete de pruebas: el env de conda `agents`
  (`C:\Users\wilhe\miniconda3\envs\agent\python.exe`, Python 3.15); el `python`
  de PATH es el alias de Microsoft Store y falla. Los unitarios corren con
  `<python.exe> tests/test_nu_account_rules.py`.
- Dependencias fijas en `requirements-lock.txt` (el build de release falla si
  no existe). Instalación local: `pip install -r requirements-lock.txt`.
  Solo `pymupdf` se usa hoy en el código; `pdfplumber` y `pydantic` están
  reservadas para funciones planificadas y el backend verifica las tres
  versiones al arrancar: quitarlas del lock sin quitar la verificación
  rompería el arranque (detalle en el README de `apps/processor`).

## Esquema y sqlc (código generado)

El código de `backend/internal/db/` es **generado**: no se edita a mano. La
cadena es: migraciones (`backend/internal/db/migrations/`) + consultas
(`backend/db/queries/`) → sqlc genera el Go.

Requisito local: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1`
(misma versión que verifica la CI).

Flujo para actualizar el esquema:

1. Crear o editar una migración en `backend/internal/db/migrations/`.
2. Actualizar las consultas en `backend/db/queries/` si hace falta.
3. Regenerar: `cd backend && sqlc generate`.
4. Verificar: `sqlc diff` (debe terminar sin diferencias).
5. `go build ./... && go test ./...`.
6. Commitear migración, consultas **y** código regenerado juntos.

Reglas que evitan el desfase que rompió las pruebas de `repositories`:

- El `SELECT`/`RETURNING` lista las columnas en el orden del esquema:
  columnas de la tabla base primero, las agregadas por `ALTER` después. Con
  otro orden sqlc genera tipos `...Row` en lugar del modelo.
- Toda consulta termina en `;`.
- La CI corre `sqlc diff` en cada push a `main`: si cambiás migraciones o
  consultas sin regenerar, el pipeline falla y el commit ni llega a revisión.

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

## Pruebas

Todas las pruebas del backend viven en `backend/tests/` (una sola carpeta, un
solo paquete `tests`). Se corren con `go test ./...` desde `backend/` como
siempre. Los helpers compartidos están en `tests/helpers_test.go`; la
persistencia de pruebas se construye con `repositories.NewTestPersistence`,
que aplica las migraciones reales resolviendo la ruta por ubicación del
código (no por directorio de trabajo), así los tests corren desde cualquier
lugar.

## Duplicados de cuotas (operación)

La unicidad del número de cuota por plan está garantizada por el índice
`uq_installments_plan_number` y validada antes de insertar en
`InstallmentPlanRepository.Create` (`ErrConflict` si el número ya existe en
el plan). Para bases manipuladas a mano o restauradas desde un backup, esta
query detecta duplicados antes de que una migración falle al crear el índice:

```sql
SELECT installment_plan_id, installment_number, COUNT(*) AS veces, GROUP_CONCAT(id) AS ids
FROM installments
GROUP BY installment_plan_id, installment_number
HAVING COUNT(*) > 1;
```

Si devuelve filas, decidir qué cuota sobrevive y borrar el resto conservando
la más antigua (`MIN(rowid)`). Nunca automático: borrar registros
financieros es una decisión del producto, no de la infraestructura.
