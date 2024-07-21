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

-- name: CreateSeriesPart :one
INSERT INTO "series_parts" (
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
    SELECT COUNT("id") + 1 FROM "series_parts"
    WHERE "series_slug" = $3
  )
) RETURNING *;

-- name: UpdateSeriesPart :one
UPDATE "series_parts" SET
  "title" = $1,
  "description" = $2
WHERE "id" = $3
RETURNING *;

-- name: UpdateSeriesPartWithPosition :one
UPDATE "series_parts" SET
  "title" = $1,
  "description" = $2,
  "position" = $3,
  "updated_at" = now()
WHERE "id" = $4
RETURNING *;

-- name: UpdateSeriesPartIsPublished :one
UPDATE "series_parts" SET
  "is_published" = $1,
  "updated_at" = now()
WHERE "id" = $2
RETURNING *;

-- name: IncrementSeriesPartPosition :exec
UPDATE "series_parts" SET
  "position" = "position" + 1
WHERE
  "series_slug" = $1 AND 
  "position" < $2 AND
  "position" >= $3;

-- name: DecrementSeriesPartPosition :exec
UPDATE "series_parts" SET
  "position" = "position" - 1
WHERE 
  "series_slug" = $1 AND
  "position" > $2 AND 
  "position" <= $3;

-- name: FindSeriesPartById :one
SELECT * FROM "series_parts"
WHERE "id" = $1 LIMIT 1;

-- name: FindSeriesPartByLanguageSlugSeriesSlugAndID :one
SELECT * FROM "series_parts"
WHERE
    "language_slug" = $1 AND
    "series_slug" = $2 AND
    "id" = $3 
LIMIT 1;

-- name: FindPublishedSeriesPartByLanguageSlugSeriesSlugAndIDWithLectures :many
SELECT 
    "series_parts".*, 
    "lectures"."id" AS "lecture_id", 
    "lectures"."title" AS "lecture_title",
    "lectures"."watch_time_seconds" AS "lecture_watch_time_seconds",
    "lectures"."read_time_seconds" AS "lecture_read_time_seconds",
    "lectures"."is_published" AS "lecture_is_published"
FROM "series_parts"
LEFT JOIN "lectures" ON (
    "series_parts"."id" = "lectures"."series_part_id" AND 
    "lectures"."is_published" = true
)
WHERE
    "series_parts"."language_slug" = $1 AND
    "series_parts"."series_slug" = $2 AND 
    "series_parts"."id" = $3 AND
    "series_parts"."is_published" = true
ORDER BY "lectures"."position" ASC;

-- name: FindSeriesPartByLanguageSlugSeriesSlugAndIDWithLectures :many
SELECT 
    "series_parts".*, 
    "lectures"."id" AS "lecture_id", 
    "lectures"."title" AS "lecture_title",
    "lectures"."watch_time_seconds" AS "lecture_watch_time_seconds",
    "lectures"."read_time_seconds" AS "lecture_read_time_seconds",
    "lectures"."is_published" AS "lecture_is_published"
FROM "series_parts"
LEFT JOIN "lectures" ON ("series_parts"."id" = "lectures"."series_part_id")
WHERE 
    "series_parts"."language_slug" = $1 AND
    "series_parts"."series_slug" = $2 AND 
    "series_parts"."id" = $3
ORDER BY "lectures"."position" ASC;

-- name: FindPaginatedPublishedSeriesPartsByLanguageSlugAndSeriesSlugWithLectures :many
WITH "series_parts" AS (
    SELECT * FROM "series_parts"
    WHERE 
        "series_parts"."language_slug" = $1 AND
        "series_parts"."series_slug" = $2 AND 
        "series_parts"."is_published" = true
    ORDER BY "series_parts"."position" ASC
    LIMIT $3 OFFSET $4
)
SELECT 
    "series_parts".*, 
    "lectures"."id" AS "lecture_id", 
    "lectures"."title" AS "lecture_title",
    "lectures"."watch_time_seconds" AS "lecture_watch_time_seconds",
    "lectures"."read_time_seconds" AS "lecture_read_time_seconds",
    "lectures"."is_published" AS "lecture_is_published"
FROM "series_parts"
LEFT JOIN "lectures" ON (
    "series_parts"."id" = "lectures"."series_part_id" AND 
    "lectures"."is_published" = true
)
ORDER BY 
    "series_parts"."position" ASC,
    "lectures"."position" ASC;

-- name: FindPaginatedSeriesPartsByLanguageSlugAndSeriesSlugWithLectures :many
WITH "series_parts" AS (
    SELECT * FROM "series_parts"
    WHERE 
        "series_parts"."language_slug" = $1 AND
        "series_parts"."series_slug" = $2
    ORDER BY "series_parts"."position" ASC
    LIMIT $3 OFFSET $4
)
SELECT 
    "series_parts".*, 
    "lectures"."id" AS "lecture_id", 
    "lectures"."title" AS "lecture_title",
    "lectures"."watch_time_seconds" AS "lecture_watch_time_seconds",
    "lectures"."read_time_seconds" AS "lecture_read_time_seconds",
    "lectures"."is_published" AS "lecture_is_published"
FROM "series_parts"
LEFT JOIN "lectures" ON ("series_parts"."id" = "lectures"."series_part_id")
ORDER BY 
    "series_parts"."position" ASC,
    "lectures"."position" ASC;

-- name: FindPaginatedSeriesPartsByLanguageSlugAndSeriesSlugWithPublishedLectures :many
WITH "series_parts" AS (
    SELECT * FROM "series_parts"
    WHERE 
        "series_parts"."language_slug" = $1 AND
        "series_parts"."series_slug" = $2
    ORDER BY "series_parts"."position" ASC
    LIMIT $3 OFFSET $4
)
SELECT 
    "series_parts".*, 
    "lectures"."id" AS "lecture_id", 
    "lectures"."title" AS "lecture_title",
    "lectures"."watch_time_seconds" AS "lecture_watch_time_seconds",
    "lectures"."read_time_seconds" AS "lecture_read_time_seconds",
    "lectures"."is_published" AS "lecture_is_published"
FROM "series_parts"
LEFT JOIN "lectures" ON (
    "series_parts"."id" = "lectures"."series_part_id" AND 
    "lectures"."is_published" = true
)
ORDER BY 
    "series_parts"."position" ASC,
    "lectures"."position" ASC;

-- name: IncrementSeriesPartLecturesCount :exec
UPDATE "series_parts" SET
  "lectures_count" = "lectures_count" + 1,
  "watch_time_seconds" = "watch_time_seconds" + $2,
  "read_time_seconds" = "read_time_seconds" + $3,
  "updated_at" = now()
WHERE "id" = $1;

-- name: DecrementSeriesPartLecturesCount :exec
UPDATE "series_parts" SET
  "lectures_count" = "lectures_count" - 1,
  "watch_time_seconds" = "watch_time_seconds" - $2,
  "read_time_seconds" = "read_time_seconds" - $3,
  "updated_at" = now()
WHERE "id" = $1;

-- name: AddSeriesPartWatchTime :exec
UPDATE "series_parts" SET
  "watch_time_seconds" = "watch_time_seconds" + $1,
  "updated_at" = now()
WHERE "id" = $2;

-- name: AddSeriesPartReadTime :exec
UPDATE "series_parts" SET
  "read_time_seconds" = "read_time_seconds" + $1,
  "updated_at" = now()
WHERE "id" = $2;

-- name: CountSeriesPartsBySeriesSlug :one
SELECT COUNT("id") AS "count" FROM "series_parts"
WHERE "series_slug" = $1 LIMIT 1;

-- name: DeleteSeriesPartById :exec
DELETE FROM "series_parts"
WHERE "id" = $1;