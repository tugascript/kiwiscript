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

-- name: CreateSeriesPartProgress :one
INSERT INTO "series_part_progress" (
  "language_slug",
  "series_slug",
  "series_part_id",
  "language_progress_id",
  "series_progress_id",
  "user_id",
  "is_current"
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  true
) RETURNING *;

-- name: SetSeriesPartProgressIsCurrentFalse :exec
UPDATE "series_part_progress" SET
    "is_current" = false
WHERE
    "user_id" = $1 AND
    "language_slug" = $2 AND
    "series_slug" = $3 AND
    "series_part_id" <> $4;

-- name: SetSeriesPartProgressIsCurrentTrue :one
UPDATE "series_part_progress" SET
  "is_current" = true
WHERE "id" = $1
RETURNING *;

-- name: FindSeriesPartProgressBySlugsAndUserID :one
SELECT * FROM "series_part_progress"
WHERE
    "language_slug" = $1 AND
    "series_slug" = $2 AND
    "series_part_id" = $3 AND
    "user_id" = $4
LIMIT 1;

-- name: FindSeriesPartProgressBySeriesProgressID :many
SELECT * FROM "series_part_progress"
WHERE "series_progress_id" = $1
ORDER BY "id" DESC;

-- name: DeleteSeriesPartProgress :exec
DELETE FROM "series_part_progress"
WHERE "id" = $1;

