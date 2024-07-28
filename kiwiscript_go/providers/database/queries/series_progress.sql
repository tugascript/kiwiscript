-- Copyright (C) 2024 Afonso Barracha
-- 
-- This file is part of KiwiScript.
-- 
-- KiwiScript is free software: you can redistribute it and/or modify
-- it under the terms of the GNU General Public License as published by
-- the Free Software Foundation, either version 3 of the License, or
-- (at your option) any later version.
-- 
-- KiwiScript is distributed in the hope that it will be useful,
-- but WITHOUT ANY WARRANTY; without even the implied warranty of
-- MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
-- GNU General Public License for more details.
-- 
-- You should have received a copy of the GNU General Public License
-- along with KiwiScript.  If not, see <https://www.gnu.org/licenses/>.

-- name: CreateSeriesProgress :one
INSERT INTO "series_progress" (
  "language_slug",
  "series_slug",
  "language_progress_id",
  "user_id",
  "is_current"
) VALUES (
  $1,
  $2,
  $3,
  $4,
  true
) RETURNING *;

-- name: SetSeriesProgressIsCurrentFalse :exec
UPDATE "series_progress" SET
  "is_current" = false
WHERE "user_id" = $1 AND "language_slug" = $2 AND "series_slug" <> $3;

-- name: SetSeriesProgressIsCurrentTrue :one
UPDATE "series_progress" SET
  "is_current" = true
WHERE "id" = $1
RETURNING *;

-- name: FindSeriesProgressBySlugAndUserID :one
SELECT * FROM "series_progress"
WHERE
    "language_slug" = $1 AND
    "series_slug" = $2 AND
    "user_id" = $3
LIMIT 1;

-- name: FindSeriesProgressByLanguageProgressID :many
SELECT * FROM "series_progress"
WHERE "language_progress_id" = $1
ORDER BY "id" DESC;

-- name: DeleteSeriesProgress :exec
DELETE FROM "series_progress"
WHERE "id" = $1;

-- name: IncrementSeriesProgressInProgressParts :exec
UPDATE "series_progress" SET
  "in_progress_parts" = "in_progress_parts" + 1
WHERE "id" = $1;

-- name: DecrementSeriesProgressInProgressParts :exec
UPDATE "series_progress" SET
  "in_progress_parts" = "in_progress_parts" - 1
WHERE "id" = $1;

-- name: AddSeriesProgressCompletedParts :exec
UPDATE "series_progress" SET
  "completed_parts" = "completed_parts" + 1,
  "in_progress_parts" = "in_progress_parts" - 1
WHERE "id" = $1;

-- name: RemoveSeriesProgressCompletedParts :exec
UPDATE "series_progress" SET
  "completed_parts" = "completed_parts" - 1,
  "in_progress_parts" = "in_progress_parts" + 1
WHERE "id" = $1;

-- name: DecrementSeriesProgressCompletedParts :exec
UPDATE "series_progress" SET
  "completed_parts" = "completed_parts" - 1
WHERE "id" = $1;