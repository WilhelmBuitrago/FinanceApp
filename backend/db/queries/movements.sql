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

-- name: ListMovementsByCategory :many
SELECT m.id, m.financial_event_id, m.account_id, m.source_entry_id, m.movement_date, m.amount_minor, m.description, m.status, m.metadata_json, m.created_at, m.updated_at
FROM movements m
JOIN financial_events fe ON fe.id = m.financial_event_id
WHERE fe.category_id = ?
ORDER BY m.created_at DESC;

-- name: ListMovementsWithEventAndCategory :many
-- Every movement listing needs the category of its financial event: one
-- join replaces the two per-row lookups the handlers used to do.
SELECT m.id, m.financial_event_id, m.account_id, m.source_entry_id, m.movement_date, m.amount_minor, m.description, m.status, m.metadata_json, m.created_at, m.updated_at,
       fe.event_type,
       fe.category_id,
       c.name AS category_name
FROM movements m
JOIN financial_events fe ON fe.id = m.financial_event_id
LEFT JOIN categories c ON c.id = fe.category_id
ORDER BY m.created_at DESC;

-- name: ListMovementsByAccountWithEventAndCategory :many
SELECT m.id, m.financial_event_id, m.account_id, m.source_entry_id, m.movement_date, m.amount_minor, m.description, m.status, m.metadata_json, m.created_at, m.updated_at,
       fe.event_type,
       fe.category_id,
       c.name AS category_name
FROM movements m
JOIN financial_events fe ON fe.id = m.financial_event_id
LEFT JOIN categories c ON c.id = fe.category_id
WHERE m.account_id = ?
ORDER BY m.created_at DESC;

-- name: DeleteMovement :exec
DELETE FROM movements WHERE id = ?;
