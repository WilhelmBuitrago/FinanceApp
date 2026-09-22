# ADR-000 — Project Bootstrap and Initial Technology Stack

* **Status:** Accepted
* **Date:** 2026-09-22
* **Decision Type:** Architecture / Project Bootstrap

## 1. Context

FinanceApp is a desktop financial-document processing application intended to ingest financial documents, process their contents, normalize extracted information, and expose the resulting data through a local application interface.

The project requires several technically distinct components:

* a desktop application;
* a reactive user interface;
* a local application backend;
* structured local persistence;
* document processing and data extraction;
* automated testing and validation.

The initial architecture therefore needs a technology stack that separates these responsibilities while keeping the application suitable for local execution.

This ADR records the initial technology decisions made during Sprint 0. It establishes the technological boundaries of the main components without defining the detailed implementation of those components.

## 2. Decision

FinanceApp will initially use the following technology stack:

| Component             | Technology         | Primary responsibility                                                          |
| --------------------- | ------------------ | ------------------------------------------------------------------------------- |
| Desktop shell         | Tauri              | Desktop application container and native integration                            |
| Frontend              | React              | User interface and presentation logic                                           |
| Backend               | Go                 | Application/domain services and local API                                       |
| Database              | SQLite             | Local structured persistence                                                    |
| Document processor    | Python             | OCR, parsing, classification, extraction, normalization and document validation |
| Desktop runtime layer | Rust through Tauri | Native desktop capabilities and Tauri integration                               |

The technologies are complementary and are not intended to duplicate each other's responsibilities.

### 2.1 Tauri

Tauri is responsible for providing the desktop application shell and connecting the frontend with native desktop capabilities.

Tauri is **not** responsible for:

* implementing the financial domain;
* processing financial documents;
* performing OCR;
* implementing the primary persistence model;
* replacing the Go backend.

Rust is used where required by the Tauri desktop layer and its native integrations.

### 2.2 React

React is responsible for the graphical user interface.

Its responsibilities include:

* rendering application views;
* handling user interaction;
* managing frontend presentation state;
* displaying document-processing status;
* displaying financial information returned by application services.

React is **not** responsible for:

* direct database access;
* OCR;
* document parsing;
* financial-domain persistence;
* implementing the document-processing pipeline.

### 2.3 Go

Go is responsible for the application backend and domain-facing services.

Its responsibilities include:

* exposing application services;
* coordinating application workflows;
* enforcing domain and application rules;
* interacting with SQLite;
* managing persistence-oriented operations;
* coordinating communication with the document-processing subsystem.

Go is **not** responsible for implementing the Python document-processing algorithms.

### 2.4 SQLite

SQLite is the initial persistence technology.

It is intended for local structured application data, including information generated and managed by FinanceApp.

SQLite is **not** intended to store the original document files themselves.

Original documents and generated processing artifacts belong to the filesystem/data layer, subject to the project's data-management and privacy rules.

### 2.5 Python

Python is responsible for the document-processing subsystem.

Its initial scope includes:

* OCR;
* document parsing;
* document classification;
* information extraction;
* normalization;
* document validation;
* experiments and evaluation related to document processing.

Python is deliberately isolated from the main application/domain backend so that document-processing models, libraries and experimentation can evolve without coupling the entire application to the Python runtime.

## 3. Architectural Boundaries

The initial responsibility boundaries are:

```text
┌──────────────────────────────────────────────┐
│                   Tauri                      │
│          Desktop shell / native layer       │
└──────────────────────┬───────────────────────┘
                       │
                       ▼
┌──────────────────────────────────────────────┐
│                   React                      │
│                User interface                │
└──────────────────────┬───────────────────────┘
                       │
                       ▼
┌──────────────────────────────────────────────┐
│                    Go                        │
│       Application + domain services         │
└───────────────┬──────────────────┬───────────┘
                │                  │
                ▼                  ▼
        ┌──────────────┐   ┌──────────────────┐
        │   SQLite     │   │     Python       │
        │ Persistence  │   │ Document Process │
        └──────────────┘   └──────────────────┘
```

The exact communication mechanism between Go and Python is intentionally **not fixed by this ADR**. That decision will be made when the processing pipeline is designed and will be documented in a separate ADR if it becomes architecturally significant.

Likewise, the detailed API contract between React and Go is outside the scope of this ADR.

## 4. Version Baseline

The following versions describe the development environment used during the initial project bootstrap:

```text
Git       2.51.0.windows.1
Node.js   24.11.1
npm       11.6.2
Go        1.27.0
Python    3.14.0
Rust      1.98.1
Cargo     1.98.1
rustup    1.29.1
```

These versions represent the **bootstrap environment**, not necessarily the permanent compatibility requirements of the application.

The project should define reproducible version constraints through the native configuration mechanisms of each ecosystem once the corresponding component is initialized.

Examples include:

* `package.json` and lockfile for Node.js dependencies;
* `go.mod` for Go;
* `pyproject.toml` and an appropriate lock/dependency mechanism for Python;
* `rust-toolchain.toml` and Cargo configuration for Rust/Tauri.

## 5. Consequences

### Positive consequences

* Clear separation between presentation, application services, persistence and document processing.
* Python can use the machine-learning, OCR and document-processing ecosystem without making it the application's primary runtime.
* Go provides a dedicated application/domain layer independent of the frontend.
* SQLite keeps the initial persistence architecture local and operationally simple.
* Tauri provides a desktop application model while allowing the UI to be developed with React.
* Each major subsystem can be tested independently.

### Negative consequences

* The project introduces multiple programming languages and runtimes.
* Development and CI configuration are more complex than a single-language application.
* Communication between Go and Python introduces an architectural boundary that must be designed carefully.
* Developers must maintain tooling for JavaScript/TypeScript, Rust, Go and Python.
* Cross-language error handling, logging and observability require explicit conventions.

## 6. Scope

This ADR establishes the initial technology stack and broad responsibility boundaries.

It does **not** define:

* the complete domain model;
* database schema;
* API specification;
* document-processing algorithms;
* OCR provider or engine;
* machine-learning models;
* Go/Python communication protocol;
* authentication or authorization;
* deployment architecture;
* production infrastructure;
* backup strategy;
* detailed frontend architecture.

Those decisions belong to subsequent design documents or ADRs when necessary.

## 7. Related Documentation

The detailed project requirements, scope, architecture and development plan are maintained in the project's primary planning and architecture documentation.

Future architectural decisions that materially change the decisions recorded here should be documented through additional ADRs rather than silently modifying this record.
