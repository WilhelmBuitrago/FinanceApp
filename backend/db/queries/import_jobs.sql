-- name: CreateImportJob :one
INSERT INTO import_jobs (
    id, document_id, status, created_at, processing_id
) VALUES (
    ?, ?, ?, ?, ?
)
RETURNING id, document_id, status, created_at, started_at, finished_at, error, processing_id;

-- name: GetImportJob :one
SELECT id, document_id, status, created_at, started_at, finished_at, error, processing_id FROM import_jobs WHERE id = ?;

-- name: UpdateImportJob :one
UPDATE import_jobs SET status = ?, started_at = ?, finished_at = ?, error = ?, processing_id = ?
WHERE id = ?
RETURNING id, document_id, status, created_at, started_at, finished_at, error, processing_id;
