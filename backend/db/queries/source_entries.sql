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
) RETURNING id, document_id, source_reference, raw_description, raw_amount_minor, raw_date, normalized_description, normalized_amount_minor, normalized_date, page_number, row_number, section, extraction_id, metadata_json, created_at, updated_at, review_status, source_key, corrected_description, corrected_amount_minor, corrected_date;

-- name: GetSourceEntry :one
SELECT id, document_id, source_reference, raw_description, raw_amount_minor, raw_date, normalized_description, normalized_amount_minor, normalized_date, page_number, row_number, section, extraction_id, metadata_json, created_at, updated_at, review_status, source_key, corrected_description, corrected_amount_minor, corrected_date FROM source_entries WHERE id = ?;

-- name: ListSourceEntriesByDocument :many
SELECT id, document_id, source_reference, raw_description, raw_amount_minor, raw_date, normalized_description, normalized_amount_minor, normalized_date, page_number, row_number, section, extraction_id, metadata_json, created_at, updated_at, review_status, source_key, corrected_description, corrected_amount_minor, corrected_date FROM source_entries WHERE document_id = ?;
