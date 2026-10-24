-- name: ResolveSourceEntry :execrows
UPDATE source_entries
SET review_status = 'RESOLVED',
    corrected_description = ?,
    corrected_amount_minor = ?,
    corrected_date = ?,
    updated_at = ?
WHERE id = ? AND review_status = 'REVIEW_REQUIRED';
