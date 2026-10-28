-- name: CreateSourceEntry :one
INSERT INTO source_entries (
    id,
    document_id,
    source_reference,
    raw_description,
    raw_amount_minor,
    raw_date,
    normalized_description,
    normalized_amount_minor,
    normalized_date,
    page_number,
    row_number,
    section,
    extraction_id,
    review_status,
    source_key,
    metadata_json,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING id, document_id, source_reference, raw_description, raw_amount_minor, raw_date, normalized_description, normalized_amount_minor, normalized_date, page_number, row_number, section, extraction_id, metadata_json, created_at, updated_at, review_status, source_key, corrected_description, corrected_amount_minor, corrected_date, normalized_scaled_to_centavos;

-- name: GetSourceEntry :one
SELECT id, document_id, source_reference, raw_description, raw_amount_minor, raw_date, normalized_description, normalized_amount_minor, normalized_date, page_number, row_number, section, extraction_id, metadata_json, created_at, updated_at, review_status, source_key, corrected_description, corrected_amount_minor, corrected_date, normalized_scaled_to_centavos FROM source_entries WHERE id = ?;

-- name: ListSourceEntriesByDocument :many
SELECT id, document_id, source_reference, raw_description, raw_amount_minor, raw_date, normalized_description, normalized_amount_minor, normalized_date, page_number, row_number, section, extraction_id, metadata_json, created_at, updated_at, review_status, source_key, corrected_description, corrected_amount_minor, corrected_date, normalized_scaled_to_centavos FROM source_entries WHERE document_id = ?;

-- name: ListReviewRequiredByDocument :many
SELECT id, document_id, source_reference, raw_description, raw_amount_minor, raw_date, normalized_description, normalized_amount_minor, normalized_date, page_number, row_number, section, extraction_id, metadata_json, created_at, updated_at, review_status, source_key, corrected_description, corrected_amount_minor, corrected_date, normalized_scaled_to_centavos FROM source_entries WHERE document_id = ? AND review_status = 'REVIEW_REQUIRED' ORDER BY row_number;

-- name: ResolveSourceEntryDerivedDate :exec
UPDATE source_entries
SET normalized_date = ?, normalized_description = ?, normalized_amount_minor = ?, review_status = 'RESOLVED', updated_at = ?
WHERE id = ?;
