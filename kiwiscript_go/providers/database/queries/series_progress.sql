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
  "user_id",
  "series_id",
  "language_id"
) VALUES (
  $1,
  $2,
  $3
) RETURNING *;

-- name: IncrementSeriesProgressLecturesCount :exec
UPDATE "series_progress" SET
  "lectures_count" = "lectures_count" + 1
WHERE "id" = $1;

-- name: DecrementSeriesProgressLecturesCount :exec
UPDATE "series_progress" SET
  "lectures_count" = "lectures_count" - 1
WHERE "id" = $1;

-- name: IncrementSeriesProgressPartAndLecturesCount :exec
UPDATE "series_progress" SET
  "part" = "part" + 1,
  "lectures_count" = "lectures_count" + 1
WHERE "id" = $1;

-- name: DecrementSeriesProgressPartAndLecturesCount :exec
UPDATE "series_progress" SET
  "part" = "part" - 1,
  "lectures_count" = "lectures_count" - 1
WHERE "id" = $1;

-- name: CompleteAndIncrementSeriesProgressPartAndLecturesSeriesProgress :exec
UPDATE "series_progress" SET
  "is_completed" = true,
  "part" = "part" + 1,
  "lectures_count" = "lectures_count" + 1,
  "completed_at" = NOW()
WHERE "id" = $1;