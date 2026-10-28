-- name: CreateMovementComponent :one
INSERT INTO movement_components (
    id,
    movement_id,
    component_type,
    amount_minor,
    description,
    metadata_json,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING id, movement_id, component_type, amount_minor, description, metadata_json, created_at, updated_at;

-- name: ListComponentsByMovement :many
SELECT id, movement_id, component_type, amount_minor, description, metadata_json, created_at, updated_at FROM movement_components WHERE movement_id = ?;
