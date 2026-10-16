# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Stack
Tauri / Rust desktop shell, React frontend, Go backend, SQLite persistence, Python document processor. Current stack confirmed per README and ADR-000. Future phone app mentioned but not scoped.

## Users
Primary users: individuals managing personal finances on desktop. Situation: importing bank statements in PDF, needing accurate records without cloud dependency. Job: administer multiple financial accounts, record income/expenses/transfers, reconcile statements, generate reports. Single user per installation initially.

## Product Purpose
FinanceApp makes personal finance management and bank statement processing possible locally and offline. It lets users import bank PDFs, automatically detect bank/account/period and movements via native text extraction, OCR fallback, and AI fallback when deterministic methods are insufficient, detect inter-account transfers, relate movements across statements, flag duplicates, calculate balances, and generate daily/monthly/annual/custom reports with manual correction and full traceability from transaction to source document and extraction.

## Positioning
Local-first, desktop-first personal finance with automatic PDF extraction pipeline. Meaningfully different mechanism: local execution with native text → OCR → AI fallback chain, transfer detection between own accounts, duplicate detection, and full provenance from document to transaction, without requiring bank APIs or cloud aggregation.

## Operating Context
Desktop-first, local-first, offline-capable. Workflows: import PDFs, review automated extraction, correct manually, categorize movements, review balances and reports. Initial focus on single user per install. Planned evolution may include synchronization/cloud later, and a phone app in the future per user note.

## Capabilities and Constraints
Confirmed capabilities: multiple accounts, income/expense/transfer events, cash as account, credit cards linked to accounts, PDF import with bank/account/period/movement auto-detection, text extraction → OCR → AI fallback, transfer detection, movement relation, duplicate detection, balance calculation, daily/monthly/annual/custom reports, manual correction, traceability to source document and extraction.

Confirmed constraints: not a banking app, not enterprise accounting, not multi-user initially, not payments platform, not bank API aggregator, not multi-currency initially, not investment system, not fiscal system, not advanced budgeting, not native mobile app initially. Undecided: communication protocol between Go and Python, detailed API contracts, authentication/authorization, deployment architecture, backup strategy.

## Brand Commitments
Name: FinanceApp. Commitments: desktop-first, local-first, offline-capable. Single user per install for v1. Technology stack preserved: Tauri/Rust desktop, React frontend, Go backend, SQLite persistence, Python document processor.

## Evidence on Hand
README.md — executive summary, scope, stack, project structure.
docs/decisions/ADR-000-project-bootstrap.md — technology stack and responsibility boundaries.
User confirmation of primary user, differentiation, and constraints.

## Product Principles
- Local-first privacy: all data stays on the user's machine by default.
- Automatic with human oversight: extraction is automated with transparent manual correction.
- Provenance and traceability: every transaction links to source document and extraction.
- Desktop-first clarity: optimize for focused desktop workflows.
- Extensible boundaries: keep processing, domain, and UI loosely coupled.

## Accessibility & Inclusion
Not established yet.
