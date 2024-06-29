-- name: CreateSeriesPart :one
INSERT INTO "series_parts" (
  "title",
  "series_id",
  "description",
  "author_id",
  "position"
) VALUES (
  $1,
  $2,
  $3,
  $4,
  (
    SELECT COUNT("id") + 1 FROM "series_parts"
    WHERE "series_id" = $2
  )
) RETURNING *;

-- name: UpdateSeriesPart :one
UPDATE "series_parts" SET
  "title" = $1,
  "description" = $2
WHERE "id" = $3
RETURNING *;

-- name: UpdateSeriesPartIsPublished :one
UPDATE "series_parts" SET
  "is_published" = $1
WHERE "id" = $2
RETURNING *;

-- name: UpdateSeriesPartPosition :one
UPDATE "series_parts" SET
  "position" = $1
WHERE "id" = $2
RETURNING *;

-- name: IncrementSeriesPartPosition :exec
UPDATE "series_parts" SET
  "position" = "position" + 1
WHERE "series_id" = $1 AND "position" >= $2;

-- name: DecrementSeriesPartPosition :exec
UPDATE "series_parts" SET
  "position" = "position" - 1
WHERE 
    "series_id" = $1 AND
    "position" > $2 AND 
    "position" <= $3;

-- name: FindSeriesPartById :one
SELECT * FROM "series_parts"
WHERE "id" = $1 LIMIT 1;

-- name: FindSeriesPartBySeriesId :many
SELECT * FROM "series_parts"
WHERE "series_id" = $1
ORDER BY "position" ASC;


-- name: FindPaginatedSeriesPartsBySeriesId :many
SELECT * FROM "series_parts"
WHERE "series_id" = $1
ORDER BY "position" ASC
LIMIT $2 OFFSET $3;

-- name: CountSeriesPartsBySeriesId :one
SELECT COUNT(*) AS "count" FROM "series_parts"
WHERE "series_id" = $1 LIMIT 1;

-- name: DeleteSeriesPartById :exec
DELETE FROM "series_parts"
WHERE "id" = $1;