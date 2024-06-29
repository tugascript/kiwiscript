-- name: CreateLecture :one
INSERT INTO "lectures" (
  "title",
  "video",
  "duration_seconds",
  "description",
  "author_id",
  "series_id",
  "series_part_id",
  "language_id",
  "position"
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  (
    SELECT COUNT("id") + 1 FROM "lectures"
    WHERE "series_part_id" = $6
  )
) RETURNING *;

-- name: UpdateLecture :one
UPDATE "lectures" SET
  "title" = $1,
  "video" = $2,
  "duration_seconds" = $3,
  "description" = $4
WHERE "id" = $5
RETURNING *;

-- name: UpdateLectureIsPublished :one
UPDATE "lectures" SET
  "is_published" = $1
WHERE "id" = $2
RETURNING *;

-- name: UpdateLecturePosition :one
UPDATE "lectures" SET
  "position" = $1
WHERE "id" = $2
RETURNING *;

-- name: IncrementLecturePosition :exec
UPDATE "lectures" SET
  "position" = "position" + 1
WHERE "series_part_id" = $1 AND "position" >= $2;

-- name: DecrementLecturePosition :exec
UPDATE "lectures" SET
  "position" = "position" - 1
WHERE 
    "series_part_id" = $1 AND 
    "position" > $2 AND 
    "position" <= $3;

-- name: IncrementLectureCommentsCount :exec
UPDATE "lectures" SET
  "comments_count" = "comments_count" + 1
WHERE "id" = $1;

-- name: DecrementLectureCommentsCount :exec
UPDATE "lectures" SET
  "comments_count" = "comments_count" - 1
WHERE "id" = $1;

-- name: FindLectureById :one
SELECT * FROM "lectures"
WHERE "id" = $1 LIMIT 1;

-- name: FindLecturesBySeriesPartId :many
SELECT * FROM "lectures"
WHERE "series_part_id" = $1
ORDER BY "position" ASC;

-- name: FindPaginatedLecturesBySeriesPartId :many
SELECT * FROM "lectures"
WHERE "series_part_id" = $1
ORDER BY "position" ASC
LIMIT $2 OFFSET $3;