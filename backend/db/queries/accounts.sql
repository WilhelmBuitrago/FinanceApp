-- name: CreateAccount :one
INSERT INTO accounts (
    id,
    user_id,
    name,
    account_type,
    parent_account_id,
    currency,
    status,
    institution,
    metadata_json,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING id, user_id, name, account_type, parent_account_id, currency, status, institution, metadata_json, created_at, updated_at;

-- name: GetAccount :one
SELECT id, user_id, name, account_type, parent_account_id, currency, status, institution, metadata_json, created_at, updated_at FROM accounts WHERE id = ?;

-- name: ListAccounts :many
SELECT id, user_id, name, account_type, parent_account_id, currency, status, institution, metadata_json, created_at, updated_at FROM accounts WHERE user_id = ? ORDER BY created_at DESC;

-- name: ListAccountsAll :many
SELECT id, user_id, name, account_type, parent_account_id, currency, status, institution, metadata_json, created_at, updated_at FROM accounts ORDER BY created_at DESC;
