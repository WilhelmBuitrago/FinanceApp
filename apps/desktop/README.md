# Desktop — FinanceApp

## Resumen

Aplicación de escritorio (Tauri 2 + React 18 + Vite + TypeScript + Tailwind 4).
Contiene toda la interfaz de usuario del producto y el shell Rust que empaqueta
la app, lanza el backend Go como proceso hijo y espera su health check antes de
mostrar la ventana. Es el componente que el usuario final ejecuta.

## Estructura

- [`index.html`](./index.html) — entrada HTML de Vite.
- [`package.json`](./package.json) — dependencias y scripts (`dev`, `build`, `preview`).
- [`package-lock.json`](./package-lock.json) — dependencias fijadas.
- [`postcss.config.js`](./postcss.config.js) — pipeline PostCSS con Tailwind 4.
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
    - [`sourceEntries.ts`](./src/api/sourceEntries.ts) — entradas de origen.
    - [`admin.ts`](./src/api/admin.ts) — operaciones administrativas (purga).
  - **`app/`** — armado de la aplicación.
    - [`App.tsx`](./src/app/App.tsx) — componente raíz.
    - [`router.tsx`](./src/app/router.tsx) — rutas de la SPA.
    - **`layout/`** — estructura visual.
      - [`AppShell.tsx`](./src/app/layout/AppShell.tsx) — marco general.
      - [`Sidebar.tsx`](./src/app/layout/Sidebar.tsx) — navegación lateral.
      - [`Topbar.tsx`](./src/app/layout/Topbar.tsx) — barra superior.
      - [`PageContainer.tsx`](./src/app/layout/PageContainer.tsx) — contenedor de páginas.
  - **`components/`** — componentes compartidos.
    - [`CashFlowChart.tsx`](./src/components/CashFlowChart.tsx) — gráfico de flujo de caja.
    - [`DonutChart.tsx`](./src/components/DonutChart.tsx) — distribuciones.
    - [`ChartCard.tsx`](./src/components/ChartCard.tsx) — tarjeta contenedora de gráficos.
    - [`ConfirmModal.tsx`](./src/components/ConfirmModal.tsx) — confirmación de acciones.
    - [`DetailDrawer.tsx`](./src/components/DetailDrawer.tsx) — panel de detalle.
    - [`DocumentDrawer.tsx`](./src/components/DocumentDrawer.tsx) — detalle de documento.
    - [`ImportDocumentDrawer.tsx`](./src/components/ImportDocumentDrawer.tsx) — flujo de importación.
    - [`ImportStatus.tsx`](./src/components/ImportStatus.tsx) — estado del job de importación.
    - [`NewAccountDrawer.tsx`](./src/components/NewAccountDrawer.tsx) — alta de cuenta.
    - [`DrawerShell.tsx`](./src/components/DrawerShell.tsx) — base común de drawers.
    - [`FormField.tsx`](./src/components/FormField.tsx) — campo de formulario.
    - [`SectionHeader.tsx`](./src/components/SectionHeader.tsx) — encabezado de sección.
    - [`SummaryStrip.tsx`](./src/components/SummaryStrip.tsx) — franja de totales.
    - [`CopyId.tsx`](./src/components/CopyId.tsx) — copiar identificador.
    - [`icons.tsx`](./src/components/icons.tsx) — iconografía propia.
    - **`ui/`** — primitivas de interfaz.
      - [`EmptyState.tsx`](./src/components/ui/EmptyState.tsx) — estado vacío.
      - [`ErrorState.tsx`](./src/components/ui/ErrorState.tsx) — estado de error.
      - [`SkeletonRows.tsx`](./src/components/ui/SkeletonRows.tsx) — filas de carga.
      - [`StatusBadge.tsx`](./src/components/ui/StatusBadge.tsx) — etiqueta de estado.
  - **`hooks/`** — acceso a datos por recurso.
    - [`useResource.ts`](./src/hooks/useResource.ts) — hook genérico de recursos.
    - [`useAccounts.ts`](./src/hooks/useAccounts.ts) — cuentas.
    - [`useDashboard.ts`](./src/hooks/useDashboard.ts) — datos del dashboard.
    - [`useDocuments.ts`](./src/hooks/useDocuments.ts) — documentos.
    - [`useMovements.ts`](./src/hooks/useMovements.ts) — movimientos.
  - **`lib/`** — utilidades puras de presentación.
    - [`formatMoney.ts`](./src/lib/formatMoney.ts) — formato de montos.
    - [`formatDate.ts`](./src/lib/formatDate.ts) — formato de fechas.
    - [`formatPeriod.ts`](./src/lib/formatPeriod.ts) — formato de períodos.
    - [`accountType.ts`](./src/lib/accountType.ts) — etiquetas de tipo de cuenta.
    - [`eventType.ts`](./src/lib/eventType.ts) — etiquetas de tipo de evento.
  - **`types/`** — contratos TypeScript del backend, un archivo por entidad.
    - [`account.ts`](./src/types/account.ts) — cuenta.
    - [`api.ts`](./src/types/api.ts) — errores de API y estado de consulta.
    - [`category.ts`](./src/types/category.ts) — categoría.
    - [`document.ts`](./src/types/document.ts) — documento.
    - [`financial-event.ts`](./src/types/financial-event.ts) — evento financiero.
    - [`import-job.ts`](./src/types/import-job.ts) — trabajo de importación.
    - [`institution.ts`](./src/types/institution.ts) — institución.
    - [`movement.ts`](./src/types/movement.ts) — movimiento.
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
  - [`custom.wxs`](./src-tauri/custom.wxs) — personalización del instalador WiX.
  - **`icons/`** — iconos de la app (Windows y Android).
- `dist/` — salida de `npm run build` (ignorada por git).
- `node_modules/` — dependencias instaladas (ignorada por git).

## Estado

- Versión del paquete: **0.1.0**. Stack: Tauri 2, React 18, Vite 8, TypeScript 5,
  Tailwind 4, react-router-dom 7.
- `npm run build` **verde** (2026-10-24): `tsc --noEmit && vite build` en ~11 s,
  sin errores ni warnings.
- La publicación (MSI) la dispara el workflow `windows-tauri.yml` al empujar un
  tag de versión; con el interruptor del usuario, nunca de forma automática
  desde acá.
