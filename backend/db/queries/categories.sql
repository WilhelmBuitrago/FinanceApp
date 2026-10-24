-- name: CreateCategory :one
INSERT INTO categories (
    id,
    name,
    parent_id,
    category_type,
    status,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?
) RETURNING id, name, parent_id, category_type, status, created_at, updated_at;

-- name: GetCategory :one
SELECT id, name, parent_id, category_type, status, created_at, updated_at FROM categories WHERE id = ?;

-- name: ListCategories :many
SELECT id, name, parent_id, category_type, status, created_at, updated_at FROM categories WHERE status = ? ORDER BY name;

-- name: DeleteCategory :exec
DELETE FROM categories WHERE id = ?;
