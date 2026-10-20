# Convenciones

## Ramas

- `main` — estable. La CI corre sobre `main` y los tags `v*` publican release.
- `developer` — rama de trabajo desde la v0.1.1 en adelante.

## Commits

- Conventional Commits: `feat`, `fix`, `chore`, `docs`, `refactor`, `test`, `ci`.
- Sin `Co-Authored-By` ni atribución a IA.
- Unidades de commit revisables, con tests y documentación junto al cambio.

## Pruebas

- Backend: `go test ./...` (en `backend/`).
- Procesador: `python -m pytest` (en `apps/processor`).
- Desktop: suite de UI a definir en v0.1.1.

## CI y releases

- **CI** (`.github/workflows/ci.yml`): verifica archivos y directorios
  requeridos del repositorio en `main`.
- **Build Windows MSI** (`.github/workflows/windows-tauri.yml`): corre en push
  de tags `v*`. Compila backend, frontend y MSI, y publica el release con el
  manifiesto `latest.json` que consume el actualizador de Tauri.

## Documentación

- `docs/` se versiona: contiene la documentación para desarrolladores
  (arquitectura, guía de desarrollo y convenciones).
- El `README.md` es la cara pública; debe describir la estructura real del
  repositorio y no enlazar archivos ausentes.
