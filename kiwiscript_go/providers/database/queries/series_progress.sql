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
  "user_id"
) VALUES (
  $1,
  $2,
  $3,
  $4
) RETURNING *;

-- name: UpdateSeriesProgressViewedAt :exec
UPDATE "series_progress" SET
  "viewed_at" = NOW()
WHERE "id" = $1;

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
ORDER BY "viewed_at" DESC;

-- name: DeleteSeriesProgress :exec
DELETE FROM "series_progress"
WHERE "id" = $1;

-- name: IncrementSeriesProgressCompletedSections :one
UPDATE "series_progress" SET
    "completed_sections" = "series_progress"."completed_sections" + 1,
    "completed_lessons" = "series_progress"."completed_lessons" + 1,
    "completed_at" = CASE
        WHEN (
            "series"."sections_count" + 1 > "series_progress"."completed_sections" AND
            "series_progress"."completed_at" IS NULL
        ) THEN (NOW())
        ELSE "series_progress"."completed_at"
    END
FROM "series"
WHERE "series_progress"."id" = $1
RETURNING "series_progress".*;

-- name: DecrementSeriesProgressCompletedSections :exec
UPDATE "series_progress" SET
  "completed_sections" = "completed_sections" - 1,
  "completed_lessons" = "completed_lessons" - $1
WHERE "id" = $2;

-- name: RemoveSeriesProgressCompletedLessons :exec
UPDATE "series_progress" SET
  "completed_lessons" = "completed_lessons" - $1
WHERE "id" = $2;

-- name: IncrementSeriesProgressCompletedLessons :exec
UPDATE "series_progress" SET
  "completed_lessons" = "completed_lessons" + 1
WHERE "id" = $1;

-- name: DecrementSeriesProgressCompletedLessons :exec
UPDATE "series_progress" SET
  "completed_lessons" = "completed_lessons" - 1
WHERE "id" = $1;