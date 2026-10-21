# Processor — FinanceApp

## Resumen

Procesador de documentos en Python. Toma un PDF de extracto bancario, lo
inspecciona (hash, tamaño), extrae el texto nativo, escribe artefactos en
disco, parsea el estado de cuenta de Nu y valida el resultado, dejando un
`manifest.json` por ejecución. Lo invoca el backend Go (`processor_runner`)
durante el flujo de importación; no tiene servidor propio.

## Estructura

- **`src/`** — punto de entrada y utilidades.
  - [`main.py`](./src/main.py) — pipeline completo: inspección → extracción →
    parseo → validación → manifest. Versión de processor 0.1.0, contrato 1.
  - [`benchmark.py`](./src/benchmark.py) — medición de rendimiento del pipeline.
- **`extraction/`** — lectura del PDF y escritura de artefactos.
  - [`__init__.py`](./extraction/__init__.py) — marcador de paquete.
  - [`inspection.py`](./extraction/inspection.py) — inspección del archivo (sha256, tamaño).
  - [`extractor.py`](./extraction/extractor.py) — extracción de texto nativo.
  - [`artifacts.py`](./extraction/artifacts.py) — escritura de artefactos en disco.
- **`parser/`** — interpretación del extracto.
  - [`__init__.py`](./parser/__init__.py) — marcador de paquete.
  - [`nu_account.py`](./parser/nu_account.py) — parser de estados de cuenta Nu.
- **`normalization/`** — normalización de datos extraídos.
  - [`__init__.py`](./normalization/__init__.py) — marcador de paquete.
  - [`nu_normalizer.py`](./normalization/nu_normalizer.py) — normalizador de
    descripciones y montos de Nu.
- **`validation/`** — reglas de calidad sobre el documento parseado.
  - [`__init__.py`](./validation/__init__.py) — marcador de paquete.
  - [`validator.py`](./validation/validator.py) — validador de estados de cuenta Nu.
  - [`benchmark.py`](./validation/benchmark.py) — benchmark de validación.
- **`tests/`** — pruebas con pytest.
  - [`test_nu_parser.py`](./tests/test_nu_parser.py) — reglas del parser.
  - [`test_nu_account_rules.py`](./tests/test_nu_account_rules.py) — reglas de negocio de Nu.
- [`requirements-lock.txt`](./requirements-lock.txt) — dependencias fijadas;
  es la fuente de verdad de versiones.

## Estado

- Versión: **0.1.0**; versión de contrato de artefactos: **1** (registrada en
  el manifest de cada ejecución).
- Alcance actual: **solo extractos de cuenta Nu**. Otros emisores quedan fuera
  hasta que se consoliden (las tarjetas de crédito, por ejemplo, se integran
  más adelante).
- Las pruebas existentes **no pudieron ejecutarse en la máquina de
  mantenimiento** (2026-10-21): no hay intérprete de Python disponible en el
  entorno (`python` no resuelve en PATH). El CI tampoco las corre: `ci.yml`
  solo verifica estructura del repo y `sqlc diff`. Estado real de los tests:
  **no verificado**.
- La carpeta quedó fuera del ciclo RED→GREEN de esta revisión por esa razón:
  hay tests, pero sin intérprete no se puede observar nada. Pendiente: correr
  `pytest` en una máquina con el entorno de `requirements-lock.txt`.
