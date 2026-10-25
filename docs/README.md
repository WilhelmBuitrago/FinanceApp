# Documentación para desarrolladores — FinanceApp

> Esta documentación es para desarrolladores y **sí** se versiona en el
> repositorio. El `README.md` de la raíz es la cara pública del producto.

## Resumen

Documentación técnica del proyecto: arquitectura de los tres componentes
(desktop, processor, backend), el flujo de desarrollo local (requisitos,
migraciones, regeneración de sqlc, variables de entorno) y las convenciones de
ramas, commits, pruebas y releases. Es la referencia que cualquier agente o
desarrollador lee antes de tocar código.

## Estructura

- [`architecture.md`](./architecture.md) — componentes por runtime, contrato de
  artefactos Go ↔ Python, runtime de Python, catálogo de componentes y notas de
  dominio.
- [`development.md`](./development.md) — requisitos, cómo correr cada
  componente, migraciones, regeneración de sqlc, verificación de duplicados de
  cuotas y variables de entorno.
- [`conventions.md`](./conventions.md) — ramas, commits, pruebas, CI y
  releases.

## Estado

- Vigente a la fecha de la revisión 2026-10-21 (ronda v0.1.1):
  - Backend documentado contra la realidad: 33 rutas HTTP, migraciones
    0001–0006, catálogo de componentes en inglés.
  - `README.md` de cada carpeta principal (`apps/desktop`,
    `apps/processor`, `backend`) estandarizados con el formato definido en
    `AGENTS.md` sección 6.
  - Los tests del processor quedan marcados como **no verificados**: sin
    intérprete de Python en la máquina de mantenimiento no se pueden observar.
- Versión pública del producto: **v0.1.0** (último release publicado, Windows
  MSI). **v0.1.1** — completa en `developer`, pendiente de publicación
  (optimización y buenas prácticas sobre los tres componentes). Detalle en
  [CHANGELOG.md](../CHANGELOG.md).
