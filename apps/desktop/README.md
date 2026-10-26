# Desktop — FinanceApp

## Resumen

Aplicación de escritorio (Tauri 2 + React 18 + Vite + TypeScript + Tailwind 4).
Contiene toda la interfaz de usuario del producto y el shell Rust que empaqueta
la app, lanza el backend Go como proceso hijo y espera su health check antes de
mostrar la ventana. Es el componente que el usuario final ejecuta.

## Estructura

- [`index.html`](./index.html) — entrada HTML de Vite.
- [`package.json`](./package.json) — dependencias y scripts (`dev`, `build`, `preview`, `lint`, `test`).
- [`package-lock.json`](./package-lock.json) — dependencias fijadas.
- [`vite.config.ts`](./vite.config.ts) — build de Vite (dev server, salida `dist`).
- [`tsconfig.json`](./tsconfig.json) — gate de types: cubre `src` y los configs (`tsc --noEmit` typechea también `vite.config.ts` y `vitest.config.ts`).
- [`postcss.config.js`](./postcss.config.js) — pipeline PostCSS con Tailwind 4.
- [`eslint.config.mjs`](./eslint.config.mjs) — configuración de ESLint (flat).
- [`vitest.config.ts`](./vitest.config.ts) — configuración de tests (jsdom).
- [`.prettierrc.json`](./.prettierrc.json) / [`.prettierignore`](./.prettierignore) — estilo de formato.
- **`src/`** — frontend React (TypeScript).
  - [`main.tsx`](./src/main.tsx) — arranque de React y montaje del router.
  - [`index.css`](./src/index.css) — estilos base y directivas de Tailwind.
  - **`api/`** — cliente HTTP, un módulo por recurso del backend.
    - [`client.ts`](./src/api/client.ts) — fetch base, JSON y manejo de errores.
    - [`accounts.ts`](./src/api/accounts.ts) — cuentas bancarias y de tarjeta.
    - [`categories.ts`](./src/api/categories.ts) — categorías del usuario.
    - [`dashboard.ts`](./src/api/dashboard.ts) — flujo de caja y distribuciones.
    - [`documents.ts`](./src/api/documents.ts) — documentos importados.
    - [`financialEvents.ts`](./src/api/financialEvents.ts) — eventos financieros.
    - [`institutions.ts`](./src/api/institutions.ts) — instituciones soportadas.
    - [`jobs.ts`](./src/api/jobs.ts) — trabajos de importación.
    - [`movements.ts`](./src/api/movements.ts) — movimientos.
    - [`review.ts`](./src/api/review.ts) — cola de revisión humana.
    - [`sourceEntries.ts`](./src/api/sourceEntries.ts) — entradas de origen.
    - [`admin.ts`](./src/api/admin.ts) — operaciones administrativas (purga).
  - **`app/`** — armado de la aplicación.
    - [`App.tsx`](./src/app/App.tsx) — componente raíz.
    - [`router.tsx`](./src/app/router.tsx) — rutas de la SPA (compone desde el registro).
    - [`nav.ts`](./src/app/nav.ts) — registro único de rutas: path, etiqueta, icono y grupo (consumen router, Sidebar y Topbar).
    - **`layout/`** — estructura visual.
      - [`AppShell.tsx`](./src/app/layout/AppShell.tsx) — marco general.
      - [`Sidebar.tsx`](./src/app/layout/Sidebar.tsx) — navegación lateral.
      - [`Topbar.tsx`](./src/app/layout/Topbar.tsx) — barra superior.
      - [`PageContainer.tsx`](./src/app/layout/PageContainer.tsx) — contenedor de páginas.
      - [`PageWithDrawer.tsx`](./src/app/layout/PageWithDrawer.tsx) — páginas que abren drawer: reservan su columna (geometría compartida con `DrawerShell`).
  - **`components/`** — componentes compartidos.
    - [`CashFlowChart.tsx`](./src/components/CashFlowChart.tsx) — gráfico de flujo de caja.
    - [`DonutChart.tsx`](./src/components/DonutChart.tsx) — distribuciones.
    - [`ChartCard.tsx`](./src/components/ChartCard.tsx) — tarjeta contenedora de gráficos.
    - [`ConfirmModal.tsx`](./src/components/ConfirmModal.tsx) — confirmación de acciones.
    - [`AddMovementDrawer.tsx`](./src/components/AddMovementDrawer.tsx) — alta manual de movimiento.
    - [`DetailDrawer.tsx`](./src/components/DetailDrawer.tsx) — panel de detalle.
    - [`DocumentDrawer.tsx`](./src/components/DocumentDrawer.tsx) — detalle de documento.
    - [`ImportDocumentDrawer.tsx`](./src/components/ImportDocumentDrawer.tsx) — flujo de importación.
    - [`NewAccountDrawer.tsx`](./src/components/NewAccountDrawer.tsx) — alta de cuenta.
    - [`DrawerShell.tsx`](./src/components/DrawerShell.tsx) — base común de drawers (+ `KV`, estados de sección y la geometría del panel).
    - **`drawer/`** — las secciones del drawer de trazabilidad.
      - [`MovementDetail.tsx`](./src/components/drawer/MovementDetail.tsx) — movimiento: datos locales, asignación de categoría y borrado.
      - [`FinancialEventDetail.tsx`](./src/components/drawer/FinancialEventDetail.tsx) — evento financiero detrás del movimiento.
      - [`SourceEntryDetail.tsx`](./src/components/drawer/SourceEntryDetail.tsx) — entrada documental: la evidencia del PDF.
      - [`DocumentDetail.tsx`](./src/components/drawer/DocumentDetail.tsx) — documento de origen.
    - [`FormField.tsx`](./src/components/FormField.tsx) — campo de formulario.
    - [`SectionHeader.tsx`](./src/components/SectionHeader.tsx) — encabezado de sección.
    - [`SummaryStrip.tsx`](./src/components/SummaryStrip.tsx) — franja de totales.
    - [`CopyId.tsx`](./src/components/CopyId.tsx) — copiar identificador.
    - [`icons.tsx`](./src/components/icons.tsx) — iconografía propia.
    - **`dashboard/`** — las piezas del Overview: la página compone, no renderiza secciones.
      - [`NotificationBell.tsx`](./src/components/dashboard/NotificationBell.tsx) — campana de revisión (popover con Escape/click afuera).
      - [`QuickActions.tsx`](./src/components/dashboard/QuickActions.tsx) — menú "+": importar documento o agregar movimiento.
      - [`NetTotalCard.tsx`](./src/components/dashboard/NetTotalCard.tsx) — neto total y desglose por cuenta.
      - [`FinancialSummarySection.tsx`](./src/components/dashboard/FinancialSummarySection.tsx) — suma real de 30 días.
      - [`CashFlowSection.tsx`](./src/components/dashboard/CashFlowSection.tsx) — flujo de dinero (línea + área).
      - [`DistributionSection.tsx`](./src/components/dashboard/DistributionSection.tsx) — ingresos y egresos clasificados.
      - [`RecentMovementsTable.tsx`](./src/components/dashboard/RecentMovementsTable.tsx) — actividad reciente.
      - [`AttentionSection.tsx`](./src/components/dashboard/AttentionSection.tsx) — documentos que requieren atención.
      - [`RecentDocumentsTable.tsx`](./src/components/dashboard/RecentDocumentsTable.tsx) — documentos recientes.
    - **`review/`** — una card por tipo de ítem de la cola de revisión.
      - [`ReviewCategoryCard.tsx`](./src/components/review/ReviewCategoryCard.tsx) — movimiento sin categoría.
      - [`ReviewPeriodCard.tsx`](./src/components/review/ReviewPeriodCard.tsx) — período de documento no reconocido.
      - [`ReviewMovementCard.tsx`](./src/components/review/ReviewMovementCard.tsx) — duplicado y extracción.
    - **`ui/`** — primitivas de interfaz.
      - [`EmptyState.tsx`](./src/components/ui/EmptyState.tsx) — estado vacío.
      - [`ErrorState.tsx`](./src/components/ui/ErrorState.tsx) — estado de error.
      - [`Notice.tsx`](./src/components/ui/Notice.tsx) — aviso transicional inline (reemplaza a `alert`).
      - [`SkeletonRows.tsx`](./src/components/ui/SkeletonRows.tsx) — filas de carga.
      - [`StatusBadge.tsx`](./src/components/ui/StatusBadge.tsx) — etiqueta de estado.
      - [`buttons.ts`](./src/components/ui/buttons.ts) — clases de botón primario/secundario (una sola definición).
      - [`table.ts`](./src/components/ui/table.ts) — celdas de encabezado de tabla (izquierda y derecha).
  - **`hooks/`** — acceso a datos por recurso.
    - [`useResource.ts`](./src/hooks/useResource.ts) — hook genérico de recursos (por id, con reintento; expone `ResourceHandle`).
    - [`useListQuery.ts`](./src/hooks/useListQuery.ts) — hook genérico de listas (el único; acepta `refreshKey` y colecciones del bus).
    - [`useAccounts.ts`](./src/hooks/useAccounts.ts) — cuentas.
    - [`useDashboard.ts`](./src/hooks/useDashboard.ts) — datos del dashboard.
    - [`useDocuments.ts`](./src/hooks/useDocuments.ts) — documentos.
    - [`useMovements.ts`](./src/hooks/useMovements.ts) — movimientos.
    - [`useAccountNames.ts`](./src/hooks/useAccountNames.ts) — mapa id→nombre de cuenta (las tres pantallas que lo pintaban lo comparten).
    - [`usePopover.ts`](./src/hooks/usePopover.ts) — cierre de popovers por Escape y click afuera.
    - [`useReviewQueue.ts`](./src/hooks/useReviewQueue.ts) — cola de revisión y sus resoluciones.
    - [`useReviewNotifications.ts`](./src/hooks/useReviewNotifications.ts) — campana de notificaciones del dashboard.
    - [`useMovementActions.ts`](./src/hooks/useMovementActions.ts) — categoría y borrado del drawer de trazabilidad.
    - [`useUpdater.ts`](./src/hooks/useUpdater.ts) — máquina del updater (check, ofrecer, descargar, instalar).
  - **`lib/`** — utilidades puras de presentación.
    - [`formatMoney.ts`](./src/lib/formatMoney.ts) — formato de montos.
    - [`formatDate.ts`](./src/lib/formatDate.ts) — formato de fechas.
    - [`formatPeriod.ts`](./src/lib/formatPeriod.ts) — formato de períodos.
    - [`accountType.ts`](./src/lib/accountType.ts) — etiquetas de tipo de cuenta.
    - [`eventType.ts`](./src/lib/eventType.ts) — etiquetas de tipo de evento.
    - [`reviewPayload.ts`](./src/lib/reviewPayload.ts) — parseo defensivo del payload de un ítem en revisión.
    - [`dataBus.ts`](./src/lib/dataBus.ts) — bus de invalidación compartido (colecciones versionadas; `useSyncExternalStore`).
    - [`version.ts`](./src/lib/version.ts) — versión del componente (lee `package.json`).
    - [`updateProgress.ts`](./src/lib/updateProgress.ts) — etiqueta de progreso del updater (porcentaje o KB).
  - **`test/`** — todos los tests del componente, con estructura espejada
    (`test/lib/`, `test/api/`, `test/hooks/`, `test/components/`,
    `test/components/review/`, `test/pages/`). El setup de Vitest vive en
    [`setup.ts`](./src/test/setup.ts); `vitest.config.ts` los descubre con
    `src/**/*.test.{ts,tsx}`.
  - **`types/`** — contratos TypeScript del backend y de la UI, un archivo por
    concepto.
    - [`account.ts`](./src/types/account.ts) — cuenta.
    - [`api.ts`](./src/types/api.ts) — errores de API y estado de consulta.
    - [`category.ts`](./src/types/category.ts) — categoría y `CategoryOption` (la forma que consume la UI).
    - [`document.ts`](./src/types/document.ts) — documento.
    - [`financial-event.ts`](./src/types/financial-event.ts) — evento financiero.
    - [`import-job.ts`](./src/types/import-job.ts) — trabajo de importación.
    - [`institution.ts`](./src/types/institution.ts) — institución.
    - [`movement.ts`](./src/types/movement.ts) — movimiento.
    - [`review-payload.ts`](./src/types/review-payload.ts) — payload de un ítem en revisión (todas las keys opcionales a propósito).
    - [`source-entry.ts`](./src/types/source-entry.ts) — entrada de origen.
  - **`pages/`** — pantallas enrutadas.
    - [`DashboardPage.tsx`](./src/pages/DashboardPage.tsx) — resumen financiero.
    - [`AccountsPage.tsx`](./src/pages/AccountsPage.tsx) — listado de cuentas.
    - [`AccountDetailPage.tsx`](./src/pages/AccountDetailPage.tsx) — detalle de cuenta.
    - [`TransactionsPage.tsx`](./src/pages/TransactionsPage.tsx) — movimientos.
    - [`DocumentsPage.tsx`](./src/pages/DocumentsPage.tsx) — documentos.
    - [`ReviewPage.tsx`](./src/pages/ReviewPage.tsx) — revisión de extracciones.
    - [`ReportsPage.tsx`](./src/pages/ReportsPage.tsx) — reportes.
    - [`SettingsPage.tsx`](./src/pages/SettingsPage.tsx) — configuración.
- **`src-tauri/`** — shell nativo (Rust).
  - [`src/main.rs`](./src-tauri/src/main.rs) — lanza el backend, espera `/health`, lo apaga al cerrar.
  - [`tauri.conf.json`](./src-tauri/tauri.conf.json) — configuración de la app y del bundling MSI.
  - [`Cargo.toml`](./src-tauri/Cargo.toml) — dependencias Rust.
  - [`Cargo.lock`](./src-tauri/Cargo.lock) — dependencias fijadas.
  - [`build.rs`](./src-tauri/build.rs) — script de compilación de Tauri.
  - [`capabilities/default.json`](./src-tauri/capabilities/default.json) — permisos del plugin updater.
  - [`main.wxs`](./src-tauri/main.wxs) — plantilla WiX del instalador: acciones de Python, borrado de AppData, UI de desinstalación.
  - **`icons/`** — iconos de la app (Windows y Android).
- `dist/` — salida de `npm run build` (ignorada por git).
- `node_modules/` — dependencias instaladas (ignorada por git).

## Estado

- Versión del paquete: **0.1.1**. Stack: Tauri 2, React 18, Vite 8, TypeScript 5,
  Tailwind 4, react-router-dom 7.
- `npm run build` **verde** (2026-10-24): `tsc --noEmit && vite build`, sin
  errores ni warnings. Regla dura (AGENTS.md): el build del front es siempre
  `npm run build`, nunca `vite build` pelado. El gate de types cubre también
  los configs (`tsconfig.json` incluye `vite.config.ts` y `vitest.config.ts`).
- `npm run lint` **verde**: 0 errores y 0 warnings (las advertencias de deuda
  documentadas en `eslint.config.mjs` quedaron en cero tras la ronda de
  auditoría).
- `npm test` **verde**: 99 tests en 24 archivos (payload de revisión, contrato
  del cliente HTTP, las formas de request de `api/review`, `api/movements` y
  `api/categories`, `useListQuery` y `useResource` (con el bus), las tres cards
  de revisión, el drawer de detalle (cadena completa, asignación de categoría,
  borrado con confirmación), el drawer de alta, `useReviewQueue`,
  `useReviewNotifications`, las utilidades de formato, la versión, el bus de
  invalidación y las dos pantallas con lógica propia (`TransactionsPage`:
  filtros, chips y drawer; `SettingsPage`: categorías, purga y updater).
  Todos viven en `src/test/` con estructura espejada.
- Refresco de datos: **sin `window.location.reload` en las mutaciones** —
  cada una publica las colecciones que dejó obsoletas (`dataBus`) y las
  pantallas suscritas refetchean solas. La única recarga completa que
  sobrevive es la de la purga (borra todo el sistema: presentar el vacío es
  su semántica, documentada en el código).
- `npx prettier --check .` **verde** (CRLF del repo normalizado a LF).
- La publicación (MSI) la dispara el workflow `windows-tauri.yml` al empujar un
  tag de versión; con el interruptor del usuario, nunca de forma automática
  desde acá.
