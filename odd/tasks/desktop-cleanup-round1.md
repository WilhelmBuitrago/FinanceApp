# Desktop cleanup — ronda 1 (muertos, duplicaciones, prácticas)

## Objetivo

Cerrar los hallazgos de bajo riesgo de la auditoría de `apps/desktop`: archivos
muertos, exports sin consumo, clases Tailwind y memos duplicados, y malas
prácticas concretas. Sin cambios de funcionalidad.

- **Rama**: `developer` (nunca `main`; commits directos en slices revisables).
- **Entrega**: 3 work-unit commits (T1, T2, T3); push a `origin/developer`.
- **Test runner**: desde `apps/desktop`: `npx tsc --noEmit`, `npm run lint`,
  `npm test`, `npx prettier --check .` y `npm run build` al cierre.

## Specs

- S1 — El mantenedor autorizó: "Si, arranca 1+2+3" sobre la auditoría de
  apps/desktop. Los lotes 4 (modularización de Dashboard/DetailDrawer/Settings)
  y 5 (README) quedan FUERA de esta ronda.
- S2 — "no functionality changes; if a change would affect current
  functionality, ask first": cada extracción debe preservar comportamiento;
  la batería completa debe quedar verde como prueba.
- S3 — Muertos a eliminar: `postcss.config.json` (nunca se carga: PostCSS
  prefiere `.js`), `tsconfig.node.json` (huérfano, 0 referencias), tipo
  `ApiErrorResponse` (`src/types/api.ts`), `formatDateTime`
  (`src/lib/formatDate.ts`, solo su test), `formatMoneyFromMinorSafe`
  (`src/lib/formatMoney.ts`, solo su test). Los exports
  `ACCOUNT_TYPE_LABELS`/`EVENT_TYPE_LABELS` pasan a internos (solo sus tests
  los importaban).
- S4 — El hueco que delata `tsconfig.node.json` debe cerrarse: `vite.config.ts`
  y `vitest.config.ts` deben quedar bajo el gate de types.
- S5 — Duplicaciones a unificar en módulos compartidos: `tableHead`/header
  derecho (5 páginas), clases de botón (`primaryButton` ×3,
  `secondaryButton`, `selectClass` de Transactions vs el de `FormField`), memo
  `accountNames` (3 páginas), color por signo del monto (4 sitios). Un módulo
  por concern, archivos chicos, nombres en inglés.
- S6 — Prácticas a corregir: `window.location.href = '/review'` →
  `useNavigate`; campana 🔔 → `icons.tsx`; popovers sin cerrar por
  click-afuera/Escape (z mágico `z-[9999]`); un solo `<h1>` por pantalla
  (ReviewPage tiene dos, más el del Topbar); rutas escritas 3 veces
  (router/Topbar/Sidebar) → una sola config de nav; `rounded-[12px]` ×19 y
  `rounded-[8px]` ×8 → tokens declarados (`index.css`: "If a value is not
  here, it does not belong in a component"); comentarios stale
  (`useListQuery`, `useReviewQueue`).

## Tasks

| ID | Descripción | Ruta | Estado |
|----|-------------|------|--------|
| T1 | Borrar muertos (S3) + gate de types para configs (S4) + tests ajustados | inline | done (a624049) |
| T2 | Módulos compartidos: tabla, botones, controles, `signedAmountClass`, `useAccountNames` (S5) | inline | pending |
| T3 | Prácticas: navigate, icono, popovers, h1, nav config única, tokens de radio, comentarios (S6) | inline | pending |

## Log

- L1 (2026-10-24) — Pedido original del mantenedor, textual: "Bien, ahora viene
  una adutoria general a apps/desktop, carpeta por carpeta, ya sabes buscanod
  codigo muertos, archivos no usados, malas practicas, mala organizacion y
  nomrmas, porisiblidades de optimizacion con refactors + modularización. Veo
  muchos archivos en el root de apps\desktop, son necesarios? Van en ese lugar?
  etc..."
- L2 (2026-10-24) — Autorización de ejecución, textual: "Si, arranca 1+2+3,
  que seria esto." + pega de los lotes 1 (muertos), 2 (duplicaciones) y 3
  (malas prácticas) tal como se reportaron.
- L3 (2026-10-24) — Auditoría de solo lectura completada y guardada en Engram
  (obs 63, topic `audit/desktop-v0.1.1`). Batería desktop de referencia:
  tsc 0, eslint 0/0, 91 tests en 22 archivos, prettier OK. git limpio en
  `developer` (3a96d5c).
