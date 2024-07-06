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

-- name: UpdateSeriesPartIsPublished :one
UPDATE "series_parts" SET
  "is_published" = $1
WHERE "id" = $2
RETURNING *;

-- name: UpdateSeriesPartPosition :one
UPDATE "series_parts" SET
  "position" = $1
WHERE "id" = $2
RETURNING *;

-- name: IncrementSeriesPartPosition :exec
UPDATE "series_parts" SET
  "position" = "position" + 1
WHERE "series_id" = $1 AND "position" >= $2;

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

-- name: FindSeriesPartBySeriesId :many
SELECT * FROM "series_parts"
WHERE "series_id" = $1
ORDER BY "position" ASC;


-- name: FindPaginatedSeriesPartsBySeriesId :many
SELECT * FROM "series_parts"
WHERE "series_id" = $1
ORDER BY "position" ASC
LIMIT $2 OFFSET $3;

-- name: CountSeriesPartsBySeriesId :one
SELECT COUNT(*) AS "count" FROM "series_parts"
WHERE "series_id" = $1 LIMIT 1;

-- name: DeleteSeriesPartById :exec
DELETE FROM "series_parts"
WHERE "id" = $1;