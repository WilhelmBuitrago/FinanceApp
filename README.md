# FinanceApp

[![CI](https://github.com/WilhelmBuitrago/FinanceApp/actions/workflows/ci.yml/badge.svg)](https://github.com/WilhelmBuitrago/FinanceApp/actions/workflows/ci.yml)
[![Build Windows MSI](https://github.com/WilhelmBuitrago/FinanceApp/actions/workflows/windows-tauri.yml/badge.svg)](https://github.com/WilhelmBuitrago/FinanceApp/actions/workflows/windows-tauri.yml)
[![Version](https://img.shields.io/badge/version-0.1.0-blue)](https://github.com/WilhelmBuitrago/FinanceApp/releases)
[![Download](https://img.shields.io/badge/download-Windows%20MSI-0078D6?logo=windows&logoColor=white)](https://github.com/WilhelmBuitrago/FinanceApp/releases/latest)
[![License](https://img.shields.io/badge/license-proprietary-critical)](LICENSE)

![Tauri](https://img.shields.io/badge/Tauri-2.12.3-FFC131?logo=tauri&logoColor=black)
![React](https://img.shields.io/badge/React-18.3.1-61DAFB?logo=react&logoColor=black)
![Vite](https://img.shields.io/badge/Vite-8.3.4-646CFF?logo=vite&logoColor=white)
![Go](https://img.shields.io/badge/Go-1.27.3-00ADD8?logo=go&logoColor=white)
![Python](https://img.shields.io/badge/Python-3.13.17-3776AB?logo=python&logoColor=white)
![SQLite](https://img.shields.io/badge/SQLite-003B57?logo=sqlite&logoColor=white)
![Platform](https://img.shields.io/badge/Platform-Windows%20x64-0078D6?logo=windows)

**Gestión de finanzas personales local-first y desktop-first para Windows.**
Administrá cuentas, importá extractos bancarios en PDF y obtené movimientos,
saldos y reportes con trazabilidad completa — sin depender de la nube.

## Características

- Múltiples cuentas financieras, efectivo y tarjetas de crédito asociadas.
- Registro de ingresos, gastos, transferencias y otros eventos financieros.
- Importación de extractos en PDF con detección automática de banco, cuenta,
  período y movimientos.
- Extracción de texto nativa, con OCR e IA como mecanismos de fallback cuando
  los métodos deterministas no alcanzan.
- Detección de transferencias entre cuentas propias y de movimientos
  potencialmente duplicados.
- Cálculo de saldos y reportes diarios, mensuales, anuales y de períodos
  personalizados.
- Corrección manual de la información extraída y trazabilidad desde cada
  transacción hasta el documento que la originó.
- Actualizaciones automáticas integradas (Tauri updater).

## Stack

| Capa | Tecnología | Versión |
| --- | --- | --- |
| Desktop | Tauri / Rust | 2.12.3 (CLI) · updater 2.14.0 |
| Frontend | React / Vite | React 18.3.1 · Vite 8.3.4 |
| Backend | Go | 1.27.3 |
| Persistencia | SQLite | driver modernc.org/sqlite 1.59.0 (pure Go) |
| Procesamiento documental | Python | 3.13.17 (embebible x64) |
| Toolchain de build | Node / npm | 24 LTS (CI fija major 24) |

## Estructura del proyecto

```text
FinanceApp/
├── apps/
│   ├── desktop/          # App de escritorio (Tauri + React/Vite)
│   └── processor/        # Procesamiento documental (Python)
├── backend/              # API, dominio y persistencia (Go, módulo financeapp/backend)
├── cmd/                  # Bootstrap legacy (no referenciado; pendiente de decisión)
├── data/                 # Datos locales (SQLite, artefactos, logs) — ignorado por Git
├── docs/                 # Documentación para desarrolladores — local, no versionada
├── .github/              # CI y build de releases
├── README.md
└── LICENSE
```

## Desarrollo

Cada componente corre por separado. Requisitos: Go, Node y Python; las
versiones exactas quedan fijadas en `backend/go.mod` y
`apps/processor/requirements-lock.txt`.

```bash
# Backend — http://127.0.0.1:8081/health
cd backend && go run ./cmd/financeapp

# Interfaz — http://localhost:5174
cd apps/desktop && npm ci && npm run dev

# App de escritorio completa (compila el backend a resources y levanta Tauri)
cd apps/desktop && npx tauri dev

# Procesador — pruebas
cd apps/processor && python -m pytest tests
```

## Datos locales

La base de datos, los artefactos de procesamiento y los logs viven en
`%APPDATA%\FinanceApp` y nunca forman parte del repositorio.

## Releases

Las versiones se publican como releases de GitHub con instalador MSI para
Windows x64. La última versión está en
[Releases](https://github.com/WilhelmBuitrago/FinanceApp/releases/latest).

- **v0.1.0** — versión actual publicada.
- **v0.1.1** — en desarrollo: optimización y buenas prácticas de proyecto.

## Licencia

Software propietario distribuido como source-visible.
© 2026 Wilhelm David Buitrago Garcia. Todos los derechos reservados.
Ver [LICENSE](LICENSE).
