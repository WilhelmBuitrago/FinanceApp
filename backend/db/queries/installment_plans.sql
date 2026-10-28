-- name: CreateInstallmentPlan :one
INSERT INTO installment_plans (
    id,
    account_id,
    origin_event_id,
    total_amount_minor,
    total_installments,
    paid_installments,
    status,
    start_date,
    created_at,
    updated_at,
    remaining_installments,
    expected_end_date,
    metadata_json
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING id, account_id, origin_event_id, total_amount_minor, total_installments, paid_installments, status, start_date, created_at, updated_at, remaining_installments, expected_end_date, metadata_json;

-- name: GetInstallmentPlan :one
SELECT id, account_id, origin_event_id, total_amount_minor, total_installments, paid_installments, status, start_date, created_at, updated_at, remaining_installments, expected_end_date, metadata_json FROM installment_plans WHERE id = ?;

-- name: ListInstallmentPlansByAccount :many
SELECT id, account_id, origin_event_id, total_amount_minor, total_installments, paid_installments, status, start_date, created_at, updated_at, remaining_installments, expected_end_date, metadata_json FROM installment_plans WHERE account_id = ? ORDER BY created_at DESC;
