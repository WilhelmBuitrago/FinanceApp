# Desktop modularización — ronda 2 (lote 4 de la auditoría)

## Objetivo

Modularizar `apps/desktop` sin cambiar funcionalidad: componentes de página
demasiado grandes, padding de drawer duplicado, y la máquina de estados del
updater enterrada en una página. Restricción dura del mantenedor: "todo es
optimización manteniendo funcionalidad intacta".

- **Rama**: `developer` (nunca `main`; work-unit commits; push al cierre).
- **Entrega**: 4 work-unit commits (U1..U4).
- **Test runner**: desde `apps/desktop`: `npm run build` (tsc+vite), `npm run lint`,
  `npm test`, `npx prettier --check .` tras cada unidad.

## Specs

- S1 — Autorización textual: "Vamos con Lote 4, recuerda todo es optimizaición
  maneteniendo funcionalidad intacta, con esto realiza la optimizacion" + pega
  de los cuatro hallazgos de organización del informe.
- S2 — DashboardPage (472 líneas, un componente con 7 secciones + 2 popovers
  inline): extraer campana como componente (consumiendo el hook existente),
  menú "+", NetTotalCard y las secciones como piezas propias. La página queda
  como composición.
- S3 — El padding de drawer (`transition-[padding]` +
  `pr-[calc(min(440px,40vw)_+_16px)]`) que reimplementan TransactionsPage,
  DocumentsPage y AccountDetailPage debe tener una sola definición; el
  orquestador determinó que pertenece al shell, no a las páginas.
- S4 — DetailDrawer (321 líneas, 4 bloques de render): partir en piezas con
  nombres, manteniendo el markup y el orden exactos.
- S5 — SettingsPage (287 líneas): la máquina de estados del updater sale a un
  hook propio (`useUpdater`), la página queda de composición.
- S6 — Cero cambio de comportamiento: misma salida renderizada, mismos eventos,
  mismos textos. La batería completa verde es la prueba; si un cambio pudiera
  alterar funcionalidad, se pregunta antes.

## Tasks

| ID | Descripción | Ruta | Estado |
|----|-------------|------|--------|
| U1 | Dashboard: campana, menú, NetTotalCard y secciones extraídas | inline | done (32f6642) |
| U2 | Padding de drawer centralizado en el shell | inline | done (9b46dae) |
| U3 | DetailDrawer partido en piezas nombradas | inline | done (d8249f4) |
| U4 | useUpdater extraído de SettingsPage | inline | done |

## Log

- L1 (2026-10-24) — Autorización del mantenedor (ver S1). Continúa la ronda 1
  (`odd/tasks/desktop-cleanup-round1.md`, commits a624049/2ea4dcd/33d9f41).
- L2 (2026-10-24) — Ronda 1 cerrada y pusheada; batería de referencia heredada:
  build 0, lint 0/0, 88/88 tests, prettier 0.
- L3 (2026-10-24) — Ronda 2 (lote 4) ejecutada y pusheada a developer:
  - U1 (32f6642): `components/dashboard/` con 9 piezas (NotificationBell,
    QuickActions, NetTotalCard, FinancialSummarySection, CashFlowSection,
    DistributionSection, RecentMovementsTable, AttentionSection,
    RecentDocumentsTable); DashboardPage 484 → 132 líneas.
  - U2 (9b46dae): `drawerReserveClass` en DrawerShell + `PageWithDrawer`
    en app/layout; las 3 páginas con drawer delegan el padding; 0 copias.
  - U3 (d8249f4): `components/drawer/` con 4 secciones + `useMovementActions`;
    DetailDrawer 321 → 71; `KV`/`SectionLoading`/`SectionError` pasaron a
    DrawerShell (la copia de KV en DocumentDrawer, eliminada);
    `useResource` expone `ResourceHandle<T>`.
  - U4 (commit final): `useUpdater` (80 líneas) extraído de SettingsPage
    (287 → 238); la página ya no habla con `@tauri-apps/plugin-updater`.
  - Batería al cierre de cada unidad: build 0, lint 0/0, 88/88, prettier 0.
  - Dos errores autoinfligidos corregidos antes de commit: reemplazar el
    import de StatusBadge en AccountDetailPage (lección repetida de la
    ronda 1) y props tipadas con `ResourceState` en vez del handle real.
