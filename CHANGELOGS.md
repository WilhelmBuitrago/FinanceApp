# Changelog

Todos los cambios notables de FinanceApp se registran acá. El formato sigue
[Keep a Changelog](https://keepachangelog.com/es/1.1.0/) y el versionado
semántico. La versión **pública** es la del último release de GitHub; `developer`
puede contener trabajo completo aún sin publicar.

## [0.1.1] — completa en `developer`, pendiente de publicación

Ronda de endurecimiento a producción sobre los tres componentes. Sin cambios de
alcance de producto: lo que funcionaba sigue funcionando igual; lo que estaba
roto, duplicado o sin verificar quedó corregido y pineado con tests.

### Backend (Go)

- **Corregido**: la resolución de la cola de revisión aceptaba cualquier tipo de
  ítem para cualquier entrada (`CATEGORY`); ahora valida por tipo
  (`ReviewService.Kind`), con tests de contrato
  (`movement_category_resolve_test.go`, `delete_movement_test.go`).
- **Refactor**: los handlers pasaron de `sql.Open` directo a los repositories
  (`SetCategory` / `CategoryRepository.Delete` + query sqlc `DeleteCategory`);
  `sqlc diff` limpio. `adminPurge` conserva su conexión cruda a propósito.
- Batería en verde: `gofmt`, `goimports`, `go build`, `go vet`, `go test ./...`,
  `golangci-lint run` (0 hallazgos).

### Processor (Python)

- **Corregido (crítico)**: `process_pdf` llamaba `parse_with_retry(dstPath, …)`
  con un nombre que no existe — resto del refactor de pasada única. Toda
  importación de documentos habría explotado con `NameError` en la etapa del
  parser. Invisible para los tests por etapa y para `py_compile` (no resuelve
  nombres); lo encontró `ruff`. Fijado y pineado con
  `tests/test_nu_pipeline_smoke.py`, que corre el pipeline de punta a punta con
  solo la lectura del PDF stubbeada.
- **Herramienta**: `ruff` 0.17.2 integrado a la batería del componente
  (70 hallazgos → 0): modernización de tipos (`List/Dict` → `list/dict`), imports
  ordenados, f-strings sin placeholders, variables y cargas JSON muertas. Los
  tres `except Exception` amplios se conservan con `noqa` y razón documentada:
  son las rutas de estacionamiento del producto.
- Batería en verde: `compileall`, `tabnanny`, `ruff check` y 7 suites.

### Desktop (Tauri + React + Vite)

- **Corregido**: el progreso del updater leía `percent`, que el plugin no emite
  (emite `Started` con el total y `Progress` por chunk); `lib/updateProgress.ts`
  acumula los chunks y decide porcentaje o KB, con tests.
- **Corregido**: la campana del dashboard se suscribe a la colección `review`
  por el bus de invalidación (antes leía solo al montar la página), pineado con
  test.
- **Agregado**: bus de invalidación compartido (`lib/dataBus`) — reemplaza las
  4 recargas completas tras mutaciones y el evento `review-queue-changed`;
  cliente HTTP único con trazadura; hooks de lista y recurso unificados
  (`useListQuery`, `useResource`).
- **Herramienta**: eslint (flat) + prettier + vitest — de 26 advertencias a 0, y
  de cero tests de página a 99 tests en 24 archivos (filtros de Movimientos,
  flujos de Configuración, updater, purga). Regla dura incorporada a `AGENTS.md`:
  el build del front es siempre `npm run build`, nunca `vite build` pelado.
- **Refactor**: eliminación de código muerto (dos configs huérfanas, tres exports
  sin consumo), unificación de duplicaciones (clases de tabla y botón, memo de
  nombres de cuenta, color por signo del monto, primitivas de drawer),
  declaración de tokens de diseño (radios, escala tipográfica completa, una sola
  elevación: cero valores arbitrarios en `src`) y modularización
  (`DashboardPage` 484 → 132 líneas, `DetailDrawer` 321 → 71, máquina del
  updater como hook, geometría del drawer en el layout, tests centralizados en
  `src/test/` con estructura espejada).
- Batería en verde: `npm run build`, `npm run lint`, `npm test`,
  `npx prettier --check .`.

### Repositorio

- `AGENTS.md` §7 documenta la batería obligatoria por componente, con evidencia.
- Todos los componentes declarados en 0.1.1; release aún sin publicar.

## [0.1.0] — primer release público (Windows MSI)

- App de escritorio Tauri + React: el shell lanza el backend Go como proceso
  hijo, espera su health check y recién muestra la ventana. Backend Go con
  SQLite (sqlc). Processor en Python para estados de cuenta de Nu.
- Flujos base: cuentas, importación de documentos (inspección, extracción,
  parseo, validación), movimientos con categorías, cola de revisión humana
  (duplicados, falta de categoría, período no leído, extracción no normalizable),
  dashboard con agregados calculados por el backend y purga administrativa.
- Todos los datos son locales (`%APPDATA%\FinanceApp`); nada sale de la máquina.
