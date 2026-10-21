-- name: CreateEventRelation :one
INSERT INTO event_relations (
    id,
    event_id_a,
    event_id_b,
    relation_type,
    status,
    confidence,
    reason,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING id, event_id_a, event_id_b, relation_type, status, confidence, reason, created_at, updated_at;

-- name: ListRelationsByEvent :many
SELECT id, event_id_a, event_id_b, relation_type, status, confidence, reason, created_at, updated_at FROM event_relations WHERE event_id_a = ? OR event_id_b = ?;
