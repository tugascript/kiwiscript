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

-- name: CreateLecture :one
INSERT INTO "lectures" (
  "title",
  "description",
  "author_id",
  "series_part_id",
  "position"
) VALUES (
  $1,
  $2,
  $3,
  $4,
  (
    SELECT COUNT("id") + 1 FROM "lectures"
    WHERE "series_part_id" = $4
  )
) RETURNING *;

-- name: UpdateLecture :one
UPDATE "lectures" SET
  "title" = $1,
  "description" = $2
WHERE "id" = $3
RETURNING *;

-- name: UpdateLectureWithPosition :one
UPDATE "lectures" SET
  "title" = $1,
  "description" = $2,
  "position" = $3
WHERE "id" = $4
RETURNING *;

-- name: UpdateLectureIsPublished :one
UPDATE "lectures" SET
  "is_published" = $1
WHERE "id" = $2
RETURNING *;

-- name: UpdateLecturePosition :one
UPDATE "lectures" SET
  "position" = $1
WHERE "id" = $2
RETURNING *;

-- name: IncrementLecturePosition :exec
UPDATE "lectures" SET
  "position" = "position" + 1
WHERE
  "series_part_id" = $1 AND
  "position" < $2 AND
  "position" >= $3;

-- name: DecrementLecturePosition :exec
UPDATE "lectures" SET
  "position" = "position" - 1
WHERE 
    "series_part_id" = $1 AND 
    "position" > $2 AND 
    "position" <= $3;

-- name: IncrementLectureCommentsCount :exec
UPDATE "lectures" SET
  "comments_count" = "comments_count" + 1
WHERE "id" = $1;

-- name: DecrementLectureCommentsCount :exec
UPDATE "lectures" SET
  "comments_count" = "comments_count" - 1
WHERE "id" = $1;

-- name: FindPublishedLectureByIDs :one
SELECT * FROM "lectures"
WHERE 
  "is_published" = true AND 
  "series_part_id" = $1 AND
  "id" = $2
LIMIT 1;

-- name: FindLectureByIDs :one
SELECT * FROM "lectures"
WHERE
  "series_part_id" = $1 AND
  "id" = $2
LIMIT 1;

-- name: FindLecturesBySeriesPartID :many
SELECT * FROM "lectures"
WHERE "series_part_id" = $1 AND "is_published" = $2
ORDER BY "position" ASC;

-- name: FindPaginatedLecturesBySeriesPartID :many
SELECT * FROM "lectures"
WHERE "series_part_id" = $1
ORDER BY "position" ASC
LIMIT $2 OFFSET $3;

-- name: CountLecturesBySeriesPartID :one
SELECT COUNT("id") FROM "lectures"
WHERE "series_part_id" = $1
LIMIT 1;

-- name: DeleteLectureByID :exec
DELETE FROM "lectures"
WHERE "id" = $1;