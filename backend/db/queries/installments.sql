-- name: CreateInstallment :one
INSERT INTO installments (
    id,
    installment_plan_id,
    installment_number,
    total_installments,
    expected_amount_minor,
    capital_amount_minor,
    interest_amount_minor,
    status,
    expected_date,
    documented_date,
    movement_id,
    metadata_json,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING id, installment_plan_id, installment_number, total_installments, expected_amount_minor, capital_amount_minor, interest_amount_minor, status, expected_date, documented_date, movement_id, metadata_json, created_at, updated_at;

-- name: GetInstallment :one
SELECT id, installment_plan_id, installment_number, total_installments, expected_amount_minor, capital_amount_minor, interest_amount_minor, status, expected_date, documented_date, movement_id, metadata_json, created_at, updated_at FROM installments WHERE id = ?;

-- name: ListInstallmentsByPlan :many
SELECT id, installment_plan_id, installment_number, total_installments, expected_amount_minor, capital_amount_minor, interest_amount_minor, status, expected_date, documented_date, movement_id, metadata_json, created_at, updated_at FROM installments WHERE installment_plan_id = ? ORDER BY installment_number;
