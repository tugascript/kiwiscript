-- name: CreateLectureProgress :one
INSERT INTO "lecture_progress" (
  "user_id",
  "lecture_id",
  "language_id",
  "series_progress_id",
  "series_part_progress_id"
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5
) RETURNING *;

-- name: CompleteLectureProgress :exec
UPDATE "lecture_progress" SET
  "is_completed" = true,
  "completed_at" = NOW()
WHERE "id" = $1;

-- name: UncompleteLectureProgress :exec
UPDATE "lecture_progress" SET
  "is_completed" = false,
  "completed_at" = NULL
WHERE "id" = $1;