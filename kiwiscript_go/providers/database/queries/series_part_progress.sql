-- name: CreateSeriesPartProgress :one
INSERT INTO "series_part_progress" (
  "user_id",
  "series_part_id",
  "language_id",
  "series_progress_id"
) VALUES (
  $1,
  $2,
  $3,
  $4
) RETURNING *;

-- name: IncrementSeriesPartProgressLecturesCount :exec
UPDATE "series_part_progress" SET
  "lectures_count" = "lectures_count" + 1
WHERE "id" = $1;

-- name: DecrementSeriesPartProgressLecturesCount :exec
UPDATE "series_part_progress" SET
  "lectures_count" = "lectures_count" - 1
WHERE "id" = $1;

-- name: CompleteAndIncrementSeriesPartProgressLecturesCount :exec
UPDATE "series_part_progress" SET
  "is_completed" = true,
  "lectures_count" = "lectures_count" + 1,
  "completed_at" = NOW()
WHERE "id" = $1;

-- name: UncompleteAndDecrementSeriesPartProgressLecturesCount :exec
UPDATE "series_part_progress" SET
  "is_completed" = false,
  "lectures_count" = "lectures_count" - 1,
  "completed_at" = NULL
WHERE "id" = $1;