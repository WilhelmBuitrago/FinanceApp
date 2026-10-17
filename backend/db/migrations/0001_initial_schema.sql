PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS accounts (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    name TEXT NOT NULL,
    account_type TEXT NOT NULL,
    parent_account_id TEXT,
    currency TEXT NOT NULL,
    status TEXT NOT NULL,
    institution TEXT,
    metadata_json TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS categories (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    parent_id TEXT,
    category_type TEXT,
    status TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS documents (
    id TEXT PRIMARY KEY,
    account_id TEXT,
    document_type TEXT NOT NULL,
    institution TEXT,
    document_name TEXT,
    period_start TEXT,
    period_end TEXT,
    document_date TEXT,
    file_hash TEXT,
    source_uri TEXT,
    status TEXT NOT NULL,
    metadata_json TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS source_entries (
    id TEXT PRIMARY KEY,
    document_id TEXT NOT NULL,
    source_reference TEXT,
    raw_description TEXT,
    raw_amount_minor INTEGER,
    raw_date TEXT,
    normalized_description TEXT,
    normalized_amount_minor INTEGER,
    normalized_date TEXT,
    page_number INTEGER,
    row_number INTEGER,
    section TEXT,
    extraction_id TEXT,
    metadata_json TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS financial_events (
    id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    event_date TEXT,
    status TEXT NOT NULL,
    verification_status TEXT NOT NULL,
    description TEXT,
    category_id TEXT,
    confidence REAL,
    metadata_json TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS movements (
    id TEXT PRIMARY KEY,
    financial_event_id TEXT NOT NULL,
    account_id TEXT NOT NULL,
    source_entry_id TEXT,
    movement_date TEXT NOT NULL,
    amount_minor INTEGER NOT NULL,
    description TEXT,
    status TEXT NOT NULL,
    metadata_json TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS event_relations (
    id TEXT PRIMARY KEY,
    event_id_a TEXT NOT NULL,
    event_id_b TEXT NOT NULL,
    relation_type TEXT NOT NULL,
    status TEXT NOT NULL,
    confidence REAL,
    reason TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS installment_plans (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    origin_event_id TEXT,
    total_amount_minor INTEGER,
    total_installments INTEGER NOT NULL,
    paid_installments INTEGER NOT NULL,
    status TEXT NOT NULL,
    start_date TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS installments (
    id TEXT PRIMARY KEY,
    installment_plan_id TEXT NOT NULL,
    installment_number INTEGER NOT NULL,
    total_installments INTEGER NOT NULL,
    expected_amount_minor INTEGER,
    capital_amount_minor INTEGER,
    interest_amount_minor INTEGER,
    status TEXT NOT NULL,
    expected_date TEXT,
    documented_date TEXT,
    movement_id TEXT,
    metadata_json TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
