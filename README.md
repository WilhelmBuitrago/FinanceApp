# FinanceApp

FinanceApp es una aplicación **desktop-first** y **local-first** para la gestión de finanzas personales y el procesamiento de extractos bancarios.

## Resumen ejecutivo

FinanceApp permitirá al usuario:

* administrar múltiples cuentas financieras;
* registrar ingresos, gastos, transferencias y otros eventos financieros;
* administrar efectivo como una cuenta;
* administrar tarjetas de crédito asociadas a cuentas;
* importar extractos bancarios en PDF;
* detectar automáticamente banco, cuenta, periodo y movimientos;
* utilizar extracción de texto nativo cuando sea posible;
* utilizar OCR cuando el PDF no contenga texto utilizable;
* utilizar modelos de IA como mecanismo de fallback cuando los métodos deterministas no sean suficientes;
* detectar posibles transferencias entre cuentas propias;
* relacionar movimientos de distintos extractos;
* detectar documentos y movimientos potencialmente duplicados;
* calcular saldos;
* generar reportes diarios, mensuales, anuales y de periodos personalizados;
* permitir corrección manual de información extraída;
* conservar trazabilidad desde una transacción final hasta el documento y la extracción que la originaron.

### Características del producto

FinanceApp será:

* **desktop-first**;
* **local-first**;
* **offline-capable**;
* orientado inicialmente a un único usuario por instalación;
* diseñado para poder incorporar sincronización/cloud posteriormente.

### Fuera del alcance inicial

La primera versión de FinanceApp no será:

* una aplicación bancaria;
* un sistema de contabilidad empresarial;
* un sistema multiusuario;
* una plataforma de pagos;
* un agregador bancario mediante APIs;
* una aplicación multi-divisa;
* un sistema de inversión;
* un sistema fiscal;
* un sistema de presupuesto avanzado;
* una aplicación móvil nativa.

## Stack tecnológico

| Capa                     | Tecnología   |
| ------------------------ | ------------ |
| Desktop                  | Tauri / Rust |
| Frontend                 | React        |
| Backend                  | Go           |
| Persistencia             | SQLite       |
| Procesamiento documental | Python       |

Las responsabilidades y límites de cada tecnología están definidos en [`ADR-000-project-bootstrap.md`](docs/decisions/ADR-000-project-bootstrap.md).

## Estructura del proyecto

```text
FinanceApp/
├── apps/
│   ├── desktop/          # Aplicación desktop / Tauri
│   └── processor/        # Procesamiento documental / Python
├── backend/              # Backend y dominio / Go
├── data/                 # Datos locales, ignorados por Git
├── docs/                 # Documentación técnica y decisiones
├── frontend/             # Interfaz de usuario / React
├── infrastructure/       # Infraestructura y automatización
├── scripts/              # Scripts auxiliares
└── tests/                # Pruebas de integración y E2E
```

## Documentación

La documentación principal se encuentra en:

* `FinanceApp.md` — definición y planificación general del proyecto.
* `docs/architecture/` — arquitectura.
* `docs/domain/` — dominio.
* `docs/api/` — contratos y API.
* `docs/decisions/` — decisiones arquitectónicas (ADR).
* `docs/experiments/` — experimentos y evaluación.

## Desarrollo

El proyecto se encuentra actualmente en **Fase 0 — Descubrimiento**, **Sprint 0 — Preparación del entorno**.

En esta etapa se establece la infraestructura inicial del repositorio, las convenciones, el tooling y la documentación necesaria para comenzar el desarrollo.

La implementación funcional de los componentes se realizará en fases posteriores.

## Datos locales

Los documentos financieros, bases de datos locales, artefactos de procesamiento y backups no forman parte del repositorio.

```text
data/
├── database/
├── documents/
├── artifacts/
└── backups/
```

Los fixtures de prueba controlados y sintéticos ubicados en `tests/fixtures/` sí pueden formar parte del repositorio.

## Licencia

Consulta [`LICENSE`](LICENSE).
