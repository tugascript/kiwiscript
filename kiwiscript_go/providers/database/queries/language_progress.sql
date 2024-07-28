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

-- name: CreateLanguageProgress :one
INSERT INTO "language_progress" (
  "language_slug",
  "user_id",
  "is_current"
) VALUES (
  $1,
  $2,
  true
) RETURNING *;

-- name: SetLanguageProgressIsCurrentFalse :exec
UPDATE "language_progress" SET
  "is_current" = false
WHERE "user_id" = $1 AND "language_slug" <> $2;

-- name: FindLanguageProgressBySlugAndUserID :one
SELECT * FROM "language_progress"
WHERE "language_slug" = $1 AND "user_id" = $2 LIMIT 1;

-- name: DeleteLanguageProgressByID :exec
DELETE FROM "language_progress"
WHERE "id" = $1;

-- name: SetLanguageProgressIsCurrentTrue :one
UPDATE "language_progress" SET
  "is_current" = true
WHERE "id" = $1
RETURNING *;

-- name: IncrementLanguageProgressInProgressSeries :exec
UPDATE "language_progress" SET
  "in_progress_series" = "in_progress_series" + 1
WHERE "id" = $1;

-- name: DecrementLanguageProgressInProgressSeries :exec
UPDATE "language_progress" SET
  "in_progress_series" = "in_progress_series" - 1
WHERE "id" = $1;

-- name: AddLanguageProgressCompletedSeries :exec
UPDATE "language_progress" SET
  "completed_series" = "completed_series" + 1,
  "in_progress_series" = "in_progress_series" - 1
WHERE "id" = $1;

-- name: RemoveLanguageProgressCompletedSeries :exec
UPDATE "language_progress" SET
  "completed_series" = "completed_series" - 1,
  "in_progress_series" = "in_progress_series" + 1
WHERE "id" = $1;

-- name: DecrementLanguageProgressCompletedSeries :exec
UPDATE "language_progress" SET
  "completed_series" = "completed_series" - 1
WHERE "id" = $1;
