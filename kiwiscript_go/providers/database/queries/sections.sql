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

-- name: CreateSection :one
INSERT INTO "sections" (
  "title",
  "language_slug",
  "series_slug",
  "description",
  "author_id",
  "position"
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  (
    SELECT COUNT("id") + 1 FROM "sections"
    WHERE "series_slug" = $3::VARCHAR(100)
  )
) RETURNING *;

-- name: UpdateSection :one
UPDATE "sections" SET
  "title" = $1,
  "description" = $2
WHERE "id" = $3
RETURNING *;

-- name: UpdateSectionWithPosition :one
UPDATE "sections" SET
  "title" = $1,
  "description" = $2,
  "position" = $3,
  "updated_at" = now()
WHERE "id" = $4
RETURNING *;

-- name: UpdateSectionIsPublished :one
UPDATE "sections" SET
  "is_published" = $1,
  "updated_at" = now()
WHERE "id" = $2
RETURNING *;

-- name: IncrementSectionPosition :exec
UPDATE "sections" SET
  "position" = "position" + 1
WHERE
  "series_slug" = $1 AND 
  "position" <= $2 AND
  "position" > $3;

-- name: DecrementSectionPosition :exec
UPDATE "sections" SET
  "position" = "position" - 1
WHERE 
  "series_slug" = $1 AND
  "position" > $2 AND 
  "position" <= $3;

-- name: FindSectionById :one
SELECT * FROM "sections"
WHERE "id" = $1 LIMIT 1;

-- name: FindSectionBySlugsAndID :one
SELECT * FROM "sections"
WHERE
    "language_slug" = $1 AND
    "series_slug" = $2 AND
    "id" = $3 
LIMIT 1;

-- name: FindPublishedSectionBySlugsAndIDWithProgress :one
SELECT
    "sections".*,
    "section_progress"."completed_lessons" AS "section_progress_completed_lessons",
    "section_progress"."completed_at" AS "section_progress_completed_at",
    "section_progress"."viewed_at" AS "section_progress_viewed_at"
FROM "sections"
LEFT JOIN "section_progress" ON (
    "sections"."id" = "section_progress"."section_id" AND
    "section_progress"."user_id" = $1
)
WHERE
    "sections"."language_slug" = $2 AND
    "sections"."series_slug" = $3 AND
    "sections"."id" = $4 AND
    "sections"."is_published" = true
LIMIT 1;

-- name: FindPaginatedPublishedSectionsBySlugs :many
SELECT * FROM "sections"
WHERE
    "language_slug" = $1 AND
    "series_slug" = $2 AND
    "is_published" = true
ORDER BY "position" ASC
LIMIT $3 OFFSET $4;

-- name: FindPaginatedSectionsBySlugs :many
SELECT * FROM "sections"
WHERE
    "sections"."language_slug" = $1 AND
    "sections"."series_slug" = $2
ORDER BY "sections"."position" ASC
LIMIT $3 OFFSET $4;

-- name: FindPaginatedPublishedSectionsBySlugsWithProgress :many
SELECT
    "sections".*,
    "section_progress"."completed_lessons" AS "section_progress_completed_lessons",
    "section_progress"."completed_at" AS "section_progress_completed_at",
    "section_progress"."viewed_at" AS "section_progress_viewed_at"
FROM "sections"
LEFT JOIN "section_progress" ON (
    "sections"."id" = "section_progress"."section_id" AND
    "section_progress"."user_id" = $1
)
WHERE
    "sections"."language_slug" = $2 AND
    "sections"."series_slug" = $3 AND
    "sections"."is_published" = true
ORDER BY "sections"."position" ASC
LIMIT $4 OFFSET $5;

-- name: IncrementSectionLessonsCount :exec
UPDATE "sections" SET
  "lessons_count" = "lessons_count" + 1,
  "watch_time_seconds" = "watch_time_seconds" + $2,
  "read_time_seconds" = "read_time_seconds" + $3,
  "updated_at" = now()
WHERE "id" = $1;

-- name: DecrementSectionLessonsCount :exec
UPDATE "sections" SET
  "lessons_count" = "lessons_count" - 1,
  "watch_time_seconds" = "watch_time_seconds" - $2,
  "read_time_seconds" = "read_time_seconds" - $3,
  "updated_at" = now()
WHERE "id" = $1;

-- name: AddSectionWatchTime :exec
UPDATE "sections" SET
  "watch_time_seconds" = "watch_time_seconds" + $1,
  "updated_at" = now()
WHERE "id" = $2;

-- name: AddSectionReadTime :exec
UPDATE "sections" SET
  "read_time_seconds" = "read_time_seconds" + $1,
  "updated_at" = now()
WHERE "id" = $2;

-- name: CountSectionsBySeriesSlug :one
SELECT COUNT("id") AS "count" FROM "sections"
WHERE "series_slug" = $1 LIMIT 1;

-- name: CountPublishedSectionsBySeriesSlug :one
SELECT COUNT("id") AS "count" FROM "sections"
WHERE "series_slug" = $1 AND "is_published" = true LIMIT 1;

-- name: DeleteSectionById :exec
DELETE FROM "sections"
WHERE "id" = $1;
