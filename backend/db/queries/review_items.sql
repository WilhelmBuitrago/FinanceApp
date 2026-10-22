-- name: CreateReviewItem :exec
INSERT INTO review_items (
    id, kind, payload_json, source_entry_id, status, created_at
) VALUES (
    ?, ?, ?, ?, 'PENDING', ?
);

-- name: ListPendingReviewItems :many
SELECT id, kind, payload_json, source_entry_id, status, created_at, resolved_at
FROM review_items
WHERE status = 'PENDING'
ORDER BY created_at;

-- name: GetReviewItem :one
SELECT id, kind, payload_json, source_entry_id, status, created_at, resolved_at
FROM review_items WHERE id = ?;

-- name: ResolveReviewItem :execrows
UPDATE review_items SET status = 'RESOLVED', resolved_at = ?
WHERE id = ? AND status = 'PENDING';

-- name: DiscardReviewItem :execrows
UPDATE review_items SET status = 'DISCARDED', resolved_at = ?
WHERE id = ? AND status = 'PENDING';
