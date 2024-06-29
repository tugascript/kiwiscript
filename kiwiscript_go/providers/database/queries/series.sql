-- name: CreateSeries :one
INSERT INTO "series" (
  "title",
  "slug",
  "description",
  "author_id"
) VALUES (
  $1,
  $2,
  $3,
  $4
) RETURNING *;

-- name: UpdateSeries :one
UPDATE "series" SET
  "title" = $1,
  "slug" = $2,
  "description" = $3
WHERE "id" = $4
RETURNING *;

-- name: UpdateSeriesIsPublished :one
UPDATE "series" SET
  "is_published" = $1
WHERE "id" = $2
RETURNING *;

-- name: UpdateSeriesReviewAvg :exec
UPDATE "series" SET
  "review_avg" = $1
WHERE "id" = $2;

-- name: IncrementSeriesReviewCount :exec
UPDATE "series" SET
  "review_count" = "review_count" + 1
WHERE "id" = $1;

-- name: AddSeriesPartsCount :exec
UPDATE "series" SET
  "series_parts_count" = "series_parts_count" + 1,
  "lectures_count" = "lectures_count" + $2
WHERE "id" = $1;

-- name: DecrementSeriesPartsCount :exec
UPDATE "series" SET
  "series_parts_count" = "series_parts_count" - 1,
  "lectures_count" = "lectures_count" - $2
WHERE "id" = $1;

-- name: IncrementSeriesLecturesCount :exec
UPDATE "series" SET
  "lectures_count" = "lectures_count" + 1
WHERE "id" = $1;

-- name: DecrementSeriesLecturesCount :exec
UPDATE "series" SET
  "lectures_count" = "lectures_count" - 1
WHERE "id" = $1;

-- name: FindSeriesById :one
SELECT * FROM "series"
WHERE "id" = $1 LIMIT 1;

-- name: FindSeriesBySlug :one
SELECT * FROM "series"
WHERE "slug" = $1 LIMIT 1;

-- name: FindAllSeriesOrderBySlug :many
SELECT * FROM "series"
ORDER BY "slug" ASC;

-- name: FindPaginatedSeriesOrderBySlug :many
SELECT * FROM "series"
ORDER BY "slug" ASC
LIMIT $1 OFFSET $2;

-- name: FindAllSeriesOrderById :many
SELECT * FROM "series"
ORDER BY "id" DESC;

-- name: FindPaginatedSeriesOrderById :many
SELECT * FROM "series"
ORDER BY "id" DESC
LIMIT $1 OFFSET $2;

-- name: DeleteSeriesById :exec
DELETE FROM "series"
WHERE "id" = $1;