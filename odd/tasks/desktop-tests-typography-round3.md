# Desktop tests, tipografía y центральный test — ronda 3

## Objetivo

Tres frentes autorizados sobre `apps/desktop`: (1) tests a nivel de página
(Transacciones con filtros, Settings), (2) la deriva de tipografía fuera de
la escala declarada en `index.css`, (3) mover los `.test` a `src/test/` con
estructura espejada y que sigan pasando.

- **Rama**: `developer` (nunca `main`; work-unit commits; push al cierre).
- **Test runner**: desde `apps/desktop`: `npm run build`, `npm run lint`,
  `npm test`, `npx prettier --check .` tras cada unidad.

## Specs

- S1 — Autorización textual: "Si, continua con los tests de pagina y la deriva
  de tipografia, hay algo mas que quiero que revises si es posible, veo mucho
  codigo .test mezclado con codico de producción, la carpeta
  apps\desktop\src\test esta practicamente sin utilizar, seria posible pasar
  todos los .test a esta carpeta sin que dejen de funcionaro al menos
  adaptarlos para que funcionen y mejorarl a estructura del proyecto, realiza
  estas tres cosas."
- S2 — La mudanza de tests no cambia un solo assertion: mismos mocks, mismos
  fixtures, mismo resultado. Specifiers relativos (incluidos `vi.mock`) se
  reescriben a la estructura espejada `src/test/<sub>/`; la batería verde es
  la prueba.
- S3 — Tests de página nuevos: TransactionsPage (filtros por texto, cuenta y
  estado; chips; estado vacío; apertura del drawer) y SettingsPage (flujo de
  purga con confirmación, alta/borrado de categorías, updater). Mocks a nivel
  de módulos `api/*`; nada habla con el backend real.
- S4 — Tipografía: declarar en `@theme` los tamaños usados y sin declarar
  (`--text-page` 24px, `--text-overline` 18px, `--text-headline` 28px) y
  reemplazar los valores arbitrarios por tokens. Sin cambios de valor: mismo
  tamaño renderizado.
- S5 — Restricción vigente: "todo es optimización manteniendo funcionalidad
  intacta". Un hallazgo de contraste encontrado al pasar
  (`hover:bg-red-50` del borrado en el drawer: rosa sobre casi-blanco en tema
  oscuro) se corrige junto al token que la receta ya prescribe.

## Tasks

| ID | Descripción | Ruta | Estado |
|----|-------------|------|--------|
| A | Mover los `.test` a `src/test/` espejado + reescribir specifiers | inline | done |
| B | Tokens de tipografía + reemplazo de valores arbitrarios | inline | pending |
| C | Tests de página: TransactionsPage y SettingsPage | inline | done |

## Log

- L1 (2026-10-24) — Autorización del mantenedor (ver S1).
- L2 (2026-10-24) — Rondas 1 y 2 cerradas (commits a624049..33d9f41 y
  666f474..e32d1e9). Batería de referencia: build 0, lint 0/0, 88/88 tests
  en 22 archivos, prettier 0. `src/test/` contiene solo `setup.ts`; los 22
  tests viven al lado de su módulo.
- L3 (2026-10-24) — Ronda 3 ejecutada y pusheada:
  - A (6ecabe7): 22 tests mudados a `src/test/` espejado (api, components,
    components/review, hooks, lib); specifiers reescritos por profundidad
    (los de `components/review` necesitan 3 ups, el resto 2). Sin cambios de
    configuración: el include `src/**/*.test.{ts,tsx}` ya los cubría.
    88/88 siguen pasando.
  - B (49db581): `@theme` declara `--text-page` (24px), `--text-headline`
    (28px), `--text-overline` (18px) y `--text-figure` (22px); 13 valores
    arbitrarios + 12 `text-sm`/`text-xs` pasados a tokens (mismo valor
    renderizado, verificado en el CSS built). Hallazgo de contraste corregido:
    `hover:bg-red-50` del borrado → `hover:bg-neg/10`.
  - C (commit final): `test/pages/TransactionsPage.test.tsx` (5 tests: filas
    sin deduplicar, filtro por texto vía URL + chip, filtro por cuenta +
    estado vacío honesto, URL como estado inicial, apertura del drawer) y
    `test/pages/SettingsPage.test.tsx` (6 tests: catálogo por tipo, alta,
    borrado, purga solo tras confirmación destructiva, updater "al día" y
    updater con instalación). 99/99 en 24 archivos.
  - Tres errores autoinfligidos durante la ronda, todos cazados por la
    batería antes de commit: profundidad de `../` en los tests de review
    (3 niveles, sobrecorregidos a 4 y revertidos), profundidad en los tests
    nuevos de `test/pages` (2 niveles, escribí 3) y default export importado
    como named en las dos páginas.
