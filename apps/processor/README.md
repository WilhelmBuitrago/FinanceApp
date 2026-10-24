# Processor — FinanceApp

## Resumen

Procesador de documentos en Python. Toma un PDF de extracto bancario, lo
inspecciona (hash, tamaño), extrae el texto nativo, escribe artefactos en
disco, parsea el estado de cuenta de Nu y valida el resultado, dejando un
`manifest.json` por ejecución. Lo invoca el backend Go (`processor_runner`)
durante el flujo de importación; no tiene servidor propio.

Versión del componente: **0.1.1**; versión de contrato de artefactos: **1**
(registrada en el manifest de cada ejecución).

## Estructura

- **`src/`** — puntos de entrada (CLIs).
  - [`main.py`](./src/main.py) — pipeline completo: inspección → extracción →
    parseo → validación → manifest.
  - [`benchmark.py`](./src/benchmark.py) — CLI de medición del parser.
- **`extraction/`** — lectura del PDF y escritura de artefactos.
  - [`__init__.py`](./extraction/__init__.py) — marcador de paquete.
  - [`_fitz.py`](./extraction/_fitz.py) — importación de PyMuPDF que protege
    el contrato de stdout (el import de fitz escribe a stderr).
  - [`inspection.py`](./extraction/inspection.py) — inspección del archivo (sha256, tamaño).
  - [`extractor.py`](./extraction/extractor.py) — extracción de texto nativo.
  - [`capture.py`](./extraction/capture.py) — pasada única: inspección y extracción sobre el mismo documento abierto.
  - [`artifacts.py`](./extraction/artifacts.py) — escritura de artefactos en disco.
- **`parser/`** — interpretación del extracto.
  - [`__init__.py`](./parser/__init__.py) — marcador de paquete.
  - [`nu_account.py`](./parser/nu_account.py) — parser de estados de cuenta Nu
    (escaneo, tolerancias, estacionado para revisión).
- **`normalization/`** — normalización de datos extraídos.
  - [`__init__.py`](./normalization/__init__.py) — marcador de paquete.
  - [`nu_normalizer.py`](./normalization/nu_normalizer.py) — normalizador de
    descripciones y montos de Nu.
- **`validation/`** — reglas de calidad sobre el documento parseado.
  - [`__init__.py`](./validation/__init__.py) — marcador de paquete.
  - [`validator.py`](./validation/validator.py) — validador de estados de cuenta Nu.
- **`benchmarks/`** — medición de calidad del parser contra ground truth.
  - [`__init__.py`](./benchmarks/__init__.py) — marcador de paquete.
  - [`benchmarks.py`](./benchmarks/benchmarks.py) — librería de benchmark;
    su CLI vive en `src/benchmark.py`.
- **`tests/`** — pruebas (todas corren con el intérprete del env `agents`,
  sin pytest: cada archivo tiene bloque `__main__`).
  - [`test_nu_parser.py`](./tests/test_nu_parser.py) — integración del parser
    contra un artefacto real (hace skip si no está presente).
  - [`test_nu_account_rules.py`](./tests/test_nu_account_rules.py) — unitarios
    del normalizador, marcadores del parser e independencia de movimientos.
  - [`test_nu_parser_robustness.py`](./tests/test_nu_parser_robustness.py) —
    estacionado ante desvíos de formato, absorción, ausencia de manifest,
    montos en cero, nombres de página inválidos.
  - [`test_nu_period_dates.py`](./tests/test_nu_period_dates.py) — revisión de
    fechas a dos niveles (período del documento vs. fecha de movimiento).
  - [`test_nu_extraction_retry.py`](./tests/test_nu_extraction_retry.py) —
    reintento silencioso cuando el período no se lee.
  - [`test_extraction_capture.py`](./tests/test_extraction_capture.py) — pasada
    única de inspección + extracción.
  - [`test_nu_pipeline_smoke.py`](./tests/test_nu_pipeline_smoke.py) — smoke de
    `process_pdf` de punta a punta (solo se stubea la lectura del PDF): existe
    porque un `NameError` en la orquestación sobrevivió a los tests por etapa y
    a `py_compile`.
- [`requirements-lock.txt`](./requirements-lock.txt) — dependencias fijadas;
  es la fuente de verdad de versiones.

## Dependencias

Fijadas en [`requirements-lock.txt`](./requirements-lock.txt) (fuente de verdad
de versiones; el instalador las deja en el venv versionado):

| Paquete | Versión | Uso actual |
|---|---|---|
| `pymupdf` (fitz) | 1.26.3 | **Sí** — extracción de texto nativo e inspección del PDF (`extraction/`) |
| `pdfplumber` | 0.11.10 | **Reservado, sin uso actual** — ninguna función del processor lo importa |
| `pydantic` | 2.11.7 | **Reservado, sin uso actual** — ninguna función del processor lo importa |

Las dos reservadas se mantienen a propósito: son dependencias previstas para
funciones planificadas del processor, no basura. El backend **verifica las tres
versiones exactas al arrancar** (`internal/bootstrap/python_runtime.go`), así
que un venv que las omita hace fallar el arranque del backend aunque hoy no se
usen. Si algún día se decide prescindir de ellas, hay que quitarlas del lock
**y** de esa verificación juntas.

## Estado

- Alcance actual: **solo extractos de cuenta Nu**. Otros emisores quedan fuera
  hasta que se consoliden (las tarjetas de crédito, por ejemplo, se integran
  más adelante).
- El intérprete de Python vive en el env de conda `agents`
  (`C:\Users\wilhe\miniconda3\envs\agent\python.exe`, Python 3.15). El `python`
  a secas en PATH es el alias de Microsoft Store y falla: usar siempre la ruta
  completa.
- Las suites de `tests/` **corren y pasan** con ese intérprete
  (`<python.exe> tests/<archivo>.py`; no hace falta pytest porque los archivos
  tienen bloque `__main__`): **7 suites**.
- Batería de verificación del componente (AGENTS.md §7), toda verde al
  2026-10-24: `compileall` (compila), `tabnanny` (indentación consistente),
  `ruff check` (0 hallazgos; ruff instalado en el env) y las 7 suites.
- Los tests de parser (`test_nu_parser.py`) **hacen skip** si no encuentran el
  artefacto local en `data/artifacts/`: son integración contra datos reales
  que solo existen en la máquina de mantenimiento.
- Nota de versiones: el env trae PyMuPDF 1.28.2 y el lock fija 1.26.3 (la que
  verifica el backend al arrancar). Los tests de parseo/normalización no usan
  fitz; los de extracción correrían contra una versión distinta a producción.
