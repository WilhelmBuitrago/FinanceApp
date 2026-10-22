-- name: ListReviewEntries :many
SELECT id, document_id, raw_description, raw_date, page_number, metadata_json, created_at
FROM source_entries
WHERE review_status = 'REVIEW_REQUIRED'
ORDER BY created_at;

-- name: ResolveSourceEntry :execrows
UPDATE source_entries
SET review_status = 'RESOLVED',
    corrected_description = ?,
    corrected_amount_minor = ?,
    corrected_date = ?,
    updated_at = ?
WHERE id = ? AND review_status = 'REVIEW_REQUIRED';
