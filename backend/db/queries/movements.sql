-- name: CreateMovement :one
INSERT INTO movements (
    id,
    financial_event_id,
    account_id,
    source_entry_id,
    movement_date,
    amount_minor,
    description,
    status,
    metadata_json,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING id, financial_event_id, account_id, source_entry_id, movement_date, amount_minor, description, status, metadata_json, created_at, updated_at;

-- name: GetMovement :one
SELECT id, financial_event_id, account_id, source_entry_id, movement_date, amount_minor, description, status, metadata_json, created_at, updated_at FROM movements WHERE id = ?;

-- name: ListMovementsAll :many
SELECT id, financial_event_id, account_id, source_entry_id, movement_date, amount_minor, description, status, metadata_json, created_at, updated_at FROM movements ORDER BY created_at DESC;

-- name: ListMovementsByAccount :many
SELECT id, financial_event_id, account_id, source_entry_id, movement_date, amount_minor, description, status, metadata_json, created_at, updated_at FROM movements WHERE account_id = ? ORDER BY created_at DESC;

-- name: DeleteMovement :exec
DELETE FROM movements WHERE id = ?;
