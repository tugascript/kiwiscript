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

-- name: CreateSectionProgress :one
INSERT INTO "section_progress" (
  "language_slug",
  "series_slug",
  "section_id",
  "language_progress_id",
  "series_progress_id",
  "user_id"
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6
) RETURNING *;

-- name: UpdateSectionProgressViewedAt :exec
UPDATE "section_progress"
SET "viewed_at" = NOW()
WHERE "id" = $1;

-- name: FindSectionProgressBySlugsAndUserID :one
SELECT * FROM "section_progress"
WHERE
    "language_slug" = $1 AND
    "series_slug" = $2 AND
    "section_id" = $3 AND
    "user_id" = $4
LIMIT 1;

-- name: FindSectionProgressByID :one
SELECT * FROM "section_progress"
WHERE "id" = $1
LIMIT 1;

-- name: DeleteSectionProgress :exec
DELETE FROM "section_progress"
WHERE "id" = $1;

-- name: IncrementSectionProgressCompletedLessons :one
UPDATE "section_progress"
SET
    "completed_lessons" = "section_progress"."completed_lessons" + 1,
    "completed_at" = CASE
        WHEN "sections"."lessons_count" + 1 > "section_progress"."completed_lessons" THEN (NOW())
        ELSE "section_progress"."completed_at"
    END
FROM "sections"
WHERE
    "section_progress"."section_id" = "sections"."id" AND
    "section_progress"."id" = $1
RETURNING "section_progress".*;

-- name: DecrementSectionProgressCompletedLessons :exec
UPDATE "section_progress"
SET
    "completed_lessons" = "section_progress"."completed_lessons" -1,
    "completed_at" = CASE
        WHEN "section_progress"."completed_lessons" IS NOT NULL THEN (NULL)
        ELSE "section_progress"."completed_at"
    END
FROM "sections"
WHERE
    "section_progress"."section_id" = "sections"."id" AND
    "section_progress"."id" = $1;