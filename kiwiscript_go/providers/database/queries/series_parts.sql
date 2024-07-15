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
  "series_id",
  "description",
  "author_id",
  "position"
) VALUES (
  $1,
  $2,
  $3,
  $4,
  (
    SELECT COUNT("id") + 1 FROM "series_parts"
    WHERE "series_id" = $2
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
  "position" = $3
WHERE "id" = $4
RETURNING *;

-- name: UpdateSeriesPartIsPublished :one
UPDATE "series_parts" SET
  "is_published" = $1
WHERE "id" = $2
RETURNING *;

-- name: IncrementSeriesPartPosition :exec
UPDATE "series_parts" SET
  "position" = "position" + 1
WHERE
  "series_id" = $1 AND 
  "position" < $2 AND
  "position" >= $3;

-- name: DecrementSeriesPartPosition :exec
UPDATE "series_parts" SET
  "position" = "position" - 1
WHERE 
  "series_id" = $1 AND
  "position" > $2 AND 
  "position" <= $3;

-- name: FindSeriesPartById :one
SELECT * FROM "series_parts"
WHERE "id" = $1 LIMIT 1;

-- name: FindSeriesPartBySeriesIDAndID :one
SELECT * FROM "series_parts"
WHERE "series_id" = $1 AND "id" = $2 LIMIT 1;

-- name: FindPublishedSeriesPartBySeriesIDAndIDWithLectures :many
SELECT 
    "series_parts".*, 
    "lectures"."id" AS "lecture_id", 
    "lectures"."title" AS "lecture_title",
    "lectures"."watch_time_seconds" AS "lecture_watch_time_seconds",
    "lectures"."read_time_seconds" AS "lecture_read_time_seconds",
    "lectures"."is_published" AS "lecture_is_published"
FROM "series_parts"
LEFT JOIN "lectures" ON ("series_parts"."id" = "lectures"."series_part_id" AND "lectures"."is_published" = true)
WHERE 
    "series_parts"."series_id" = $1 AND 
    "series_parts"."id" = $2 AND
    "series_parts"."is_published" = true
ORDER BY "lectures"."position" ASC;

-- name: FindSeriesPartBySeriesIDAndIDWithLectures :many
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
    "series_parts"."series_id" = $1 AND 
    "series_parts"."id" = $2
ORDER BY "lectures"."position" ASC;

-- name: FindPublishedPaginatedSeriesPartsBySeriesIdWithLectures :many
WITH "series_parts" AS (
    SELECT * FROM "series_parts"
    WHERE 
        "series_parts"."series_id" = $1 AND 
        "series_parts"."is_published" = true
    ORDER BY "series_parts"."position" ASC
    LIMIT $2 OFFSET $3
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

-- name: FindPaginatedSeriesPartsBySeriesIdWithLectures :many
WITH "series_parts" AS (
    SELECT * FROM "series_parts"
    WHERE 
        "series_parts"."series_id" = $1
    ORDER BY "series_parts"."position" ASC
    LIMIT $2 OFFSET $3
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


-- name: CountSeriesPartsBySeriesId :one
SELECT COUNT("id") AS "count" FROM "series_parts"
WHERE "series_id" = $1 LIMIT 1;

-- name: DeleteSeriesPartById :exec
DELETE FROM "series_parts"
WHERE "id" = $1;