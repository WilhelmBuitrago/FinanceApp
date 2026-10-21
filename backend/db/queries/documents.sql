-- name: CreateDocument :one
INSERT INTO documents (
    id,
    account_id,
    document_type,
    institution,
    document_name,
    period_start,
    period_end,
    document_date,
    file_hash,
    source_uri,
    status,
    metadata_json,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING id, account_id, document_type, institution, document_name, period_start, period_end, document_date, file_hash, source_uri, status, metadata_json, created_at, updated_at;

-- name: GetDocument :one
SELECT id, account_id, document_type, institution, document_name, period_start, period_end, document_date, file_hash, source_uri, status, metadata_json, created_at, updated_at FROM documents WHERE id = ?;

-- name: ListDocumentsAll :many
SELECT id, account_id, document_type, institution, document_name, period_start, period_end, document_date, file_hash, source_uri, status, metadata_json, created_at, updated_at FROM documents ORDER BY document_date DESC;

-- name: ListDocumentsByAccount :many
SELECT id, account_id, document_type, institution, document_name, period_start, period_end, document_date, file_hash, source_uri, status, metadata_json, created_at, updated_at FROM documents WHERE account_id = ? ORDER BY document_date DESC;
