-- name: CreateFinancialEvent :one
INSERT INTO financial_events (
    id,
    event_type,
    event_date,
    status,
    verification_status,
    description,
    category_id,
    confidence,
    metadata_json,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING id, event_type, event_date, status, verification_status, description, category_id, confidence, metadata_json, created_at, updated_at;

-- name: GetFinancialEvent :one
SELECT id, event_type, event_date, status, verification_status, description, category_id, confidence, metadata_json, created_at, updated_at FROM financial_events WHERE id = ?;

-- name: UpdateFinancialEventCategory :execrows
UPDATE financial_events SET category_id = ?, updated_at = ? WHERE id = ?;

-- name: ListFinancialEventsByType :many
SELECT id, event_type, event_date, status, verification_status, description, category_id, confidence, metadata_json, created_at, updated_at FROM financial_events WHERE event_type = ? ORDER BY event_date DESC;
