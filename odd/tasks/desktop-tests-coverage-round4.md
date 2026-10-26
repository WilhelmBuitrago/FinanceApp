# Cobertura de tests del desktop — ronda 4 (hallazgo D2 de la auditoría)

- **Objetivo:** cubrir con tests los módulos de `apps/desktop/src` que quedaron
  sin suite tras las rondas 1-3: los 9 componentes de `components/dashboard/`,
  las 4 secciones de `components/drawer/` y los hooks sin test
  (`useAccountNames`, `useAccounts`, `useMovements`, `useDocuments`,
  `useDashboard`, `useMovementActions`, `useUpdater`).
- **Rama:** `developer`. **Estrategia:** unidades de commit por tema, batería
  desktop por unidad.
- **Runner:** `npx vitest run` desde `apps/desktop/`; gate de types
  `npm run build`; `npx eslint .`; `npx prettier --check .`.

## Specs

- S1 — El usuario: "Ahora realiza D2." (hallazgo D2 de la auditoría:
  "componentes agregados después de la ronda 3 sin tests: `dashboard/*`
  (9), `drawer/*` (4 secciones) y los hooks `useMovementActions`,
  `useUpdater`, `useAccounts`, `useDashboard`, `useDocuments`,
  `useMovements`, `useAccountNames`, `usePopover`"). `usePopover` quedó
  cubierto en la ronda 1 (commit 71f7352).
- S2 — Tests de comportamiento, no de implementación: estado cargando,
  error, vacío y con datos por cada componente; las acciones del usuario
  (asignar categoría, borrar, reintentar, abrir/cerrar popovers) a través
  de la interfaz pública.
- S3 — Ubicación y nomenclatura: espejo de `src/` bajo `src/test/`
  (`test/components/dashboard/`, `test/components/drawer/`,
  `test/hooks/`), un archivo por módulo, specifiers con la profundidad
  correcta (3 ups desde `test/components/<subdir>/`, 2 desde `test/hooks/`).
- S4 — Mocks: los módulos de `api/` vía `vi.mock`; `dataBus` vía `vi.mock`
  para asertar invalidaciones; `@tauri-apps/plugin-updater` vía `vi.mock`
  para el updater. Nada de bases de datos ni red real.
- S5 — Toda prueba de test nuevo se verifica contra una mutación del código
  productivo (evidencia RED) antes de cerrar la unidad.

## Tasks

| ID | Unidad | Ruta | Estado |
|----|--------|------|--------|
| T1 | Hooks de datos: `useAccountNames`, `useAccounts`, `useMovements`, `useDocuments`, `useDashboard` | inline | done |
| T2 | Máquinas de estado: `useMovementActions`, `useUpdater` | inline | pending |
| T3 | Dashboard I: `NetTotalCard`, `FinancialSummarySection`, `AttentionSection`, `RecentMovementsTable`, `RecentDocumentsTable` | inline | pending |
| T4 | Dashboard II: `CashFlowSection`, `DistributionSection`, `NotificationBell`, `QuickActions` | inline | pending |
| T5 | Drawer: `MovementDetail`, `FinancialEventDetail`, `SourceEntryDetail`, `DocumentDetail` | inline | pending |

## Log

- L1 (usuario, 2026-10-25): "Ahora realiza D2. y para P1-bis explicalo bien
  detalladamente."
