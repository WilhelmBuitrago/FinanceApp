-- name: CreateMovementRelation :one
INSERT INTO movement_relations (
    id,
    movement_id_a,
    movement_id_b,
    relation_type,
    status,
    confidence,
    reason,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING id, movement_id_a, movement_id_b, relation_type, status, confidence, reason, created_at, updated_at;

-- name: GetMovementRelation :one
SELECT id, movement_id_a, movement_id_b, relation_type, status, confidence, reason, created_at, updated_at FROM movement_relations WHERE id = ?;

-- name: ListRelationsByMovement :many
SELECT id, movement_id_a, movement_id_b, relation_type, status, confidence, reason, created_at, updated_at FROM movement_relations WHERE movement_id_a = ? OR movement_id_b = ?;
