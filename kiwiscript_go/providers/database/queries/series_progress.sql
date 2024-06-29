-- name: CreateSeriesProgress :one
INSERT INTO "series_progress" (
  "user_id",
  "series_id",
  "language_id"
) VALUES (
  $1,
  $2,
  $3
) RETURNING *;

-- name: IncrementSeriesProgressLecturesCount :exec
UPDATE "series_progress" SET
  "lectures_count" = "lectures_count" + 1
WHERE "id" = $1;

-- name: DecrementSeriesProgressLecturesCount :exec
UPDATE "series_progress" SET
  "lectures_count" = "lectures_count" - 1
WHERE "id" = $1;

-- name: IncrementSeriesProgressPartAndLecturesCount :exec
UPDATE "series_progress" SET
  "part" = "part" + 1,
  "lectures_count" = "lectures_count" + 1
WHERE "id" = $1;

-- name: DecrementSeriesProgressPartAndLecturesCount :exec
UPDATE "series_progress" SET
  "part" = "part" - 1,
  "lectures_count" = "lectures_count" - 1
WHERE "id" = $1;

-- name: CompleteAndIncrementSeriesProgressPartAndLecturesSeriesProgress :exec
UPDATE "series_progress" SET
  "is_completed" = true,
  "part" = "part" + 1,
  "lectures_count" = "lectures_count" + 1,
  "completed_at" = NOW()
WHERE "id" = $1;