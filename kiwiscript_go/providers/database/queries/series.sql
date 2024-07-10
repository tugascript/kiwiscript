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

-- name: CreateSeries :one
INSERT INTO "series" (
  "title",
  "slug",
  "description",
  "author_id"
) VALUES (
  $1,
  $2,
  $3,
  $4
) RETURNING *;

-- name: UpdateSeries :one
UPDATE "series" SET
  "title" = $1,
  "slug" = $2,
  "description" = $3
WHERE "id" = $4
RETURNING *;

-- name: UpdateSeriesIsPublished :one
UPDATE "series" SET
  "is_published" = $1
WHERE "id" = $2
RETURNING *;

-- name: UpdateSeriesReviewAvg :exec
UPDATE "series" SET
  "review_avg" = $1
WHERE "id" = $2;

-- name: IncrementSeriesReviewCount :exec
UPDATE "series" SET
  "review_count" = "review_count" + 1
WHERE "id" = $1;

-- name: AddSeriesPartsCount :exec
UPDATE "series" SET
  "series_parts_count" = "series_parts_count" + 1,
  "lectures_count" = "lectures_count" + $2
WHERE "id" = $1;

-- name: DecrementSeriesPartsCount :exec
UPDATE "series" SET
  "series_parts_count" = "series_parts_count" - 1,
  "lectures_count" = "lectures_count" - $2
WHERE "id" = $1;

-- name: IncrementSeriesLecturesCount :exec
UPDATE "series" SET
  "lectures_count" = "lectures_count" + 1
WHERE "id" = $1;

-- name: DecrementSeriesLecturesCount :exec
UPDATE "series" SET
  "lectures_count" = "lectures_count" - 1
WHERE "id" = $1;

-- name: FindSeriesById :one
SELECT * FROM "series"
WHERE "id" = $1 LIMIT 1;

-- name: FindSeriesBySlug :one
SELECT * FROM "series"
WHERE "slug" = $1 LIMIT 1;

-- name: DeleteSeriesById :exec
DELETE FROM "series"
WHERE "id" = $1;

-- name: FindSeriesBySlugWithJoins :many
SELECT 
    "series".*,
    "tags"."name" AS "tag_name",
    "users"."first_name" AS "author_first_name",
    "users"."last_name" AS "author_last_name"
FROM "series"
LEFT JOIN "users" ON "series"."author_id" = "users"."id"
LEFT JOIN "series_tags" ON "series"."id" = "series_tags"."series_id"
    LEFT JOIN "tags" ON "series_tags"."tag_id" = "tags"."id"
WHERE 
    "series"."is_published" = true AND 
    "series"."slug" = $1 
LIMIT 1;
