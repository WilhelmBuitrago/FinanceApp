# Diseño de Runtime de Python reproducible para FinanceApp Windows

## Versión exacta de Python

**Python 3.12.10 x86_64 Windows embeddable**

- Motivo: última versión de 3.12 con instalador oficial, compatible con:
  - `pdfplumber >=0.11.0` → Python >=3.8
  - `pymupdf >=1.26` → Python >=3.10
  - `pydantic >=2.11` → Python >=3.10
- URL oficial: `https://www.python.org/ftp/python/3.12.10/python-3.12.10-embed-amd64.zip`
- SHA256 a centralizar en configuración: `...` *a rellenar con hash oficial*
- `get-pip.py`: `https://bootstrap.pypa.io/get-pip.py`
- Licencia Python: PSF 2.0, redistribuible.

## Dependencias bloqueadas

`apps/processor/requirements-lock.txt`:
```
pdfplumber==0.11.10
pydantic==2.11.7
pymupdf==1.26.3
```
Sin versiones sueltas. `requirements.txt` se reemplaza por lock en producción.

## Estructura de rutas

- Runtime base: `%APPDATA%\FinanceApp\python-runtime\3.12.10`
- Venv: `%APPDATA%\FinanceApp\venv\Scripts\python.exe`
- Scripts procesador: `resources/processor` empaquetados en MSI
- Config: `backend/internal/config/config.go` expone constantes

## Flujo de bootstrap idempotente

1. **Chequeo rápido**
   - Existe `%APPDATA%\FinanceApp\venv\Scripts\python.exe` ?
   - `python --version` == 3.12.10 ?
   - Archivo marcador `.python-ready` con hash de lock coincide.
   - Si sí → salir.

2. **Descarga y verificación**
   - Descargar zip embeddable y get-pip.py.
   - Verificar SHA256 contra valores centralizados.
   - Extraer a `python-runtime/3.12.10`.

3. **Instalación pip**
   - `python.exe get-pip.py --no-warn-script-location`
   - `python.exe -m pip install --upgrade pip`

4. **Entorno virtual**
   - `python.exe -m venv %APPDATA%\FinanceApp\venv`
   - Activar venv, instalar `requirements-lock.txt` con `--no-index` y wheels pre-cached? 
   - Para offline: descargar wheels en CI y empaquetar en `resources/python-wheels`.

5. **Verificación**
   - `venv\Scripts\python.exe -c "import pdfplumber,pydantic,fitz; print('ok')"`
   - Si falla → error con código y log.

6. **Resiliencia**
   - Proceso reanudable: descargas a temp, luego move atómico.
   - No tocar DB/arts/docs.
   - Logs en `logs/python-bootstrap-YYYYMMDD.log`.

## Integración con Tauri MSI

- No cambiar MSI para descargar en instalación. Bootstrapper corre **en primer arranque** del backend Go.
- `apps/desktop/src-tauri/src/main.rs` ya spawnea backend con env vars. Añadir llamada a `pythonbootstrap.Ensure()` antes de `http.ListenAndServe`.
- El workflow `windows-tauri.yml` copia `requirements-lock.txt` y `python-runtime.json` a `resources`.
- Sin Go/Node en cliente.

## Errores y UX

- Si falla descarga/verificación/venv/install: backend retorna HTTP 503 con mensaje `PROCESSOR_NOT_READY`, UI muestra error comprensible y log detallado en `logs`.
- No marcar instalación como exitosa si el procesador no está listo.

## Licencias

- Python PSF 2.0 → redistribución permitida.
- PyMuPDF 1.26.3 → AGPL-3.0. *Nota*: si FinanceApp es propietario, se requiere licencia comercial de Artifex. Evaluar sustitución por `pdfminer.six` solo para inspección.
- pdfplumber MIT, pydantic MIT.

## Próximos pasos de implementación

1. Crear `backend/internal/python/bootstrap.go` con Ensure().
2. Añadir constantes de versión y hashes en `config.go`.
3. Generar `requirements-lock.txt` y copiar a resources en CI.
4. Modificar `backend/cmd/financeapp/main.go` para llamar Ensure antes de router.
5. Pruebas de idempotencia y recuperación.
