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
  "user_id"
) VALUES (
  $1,
  $2
) RETURNING *;

-- name: UpdateLanguageProgressViewedAt :exec
UPDATE "language_progress" SET
  "viewed_at" = NOW()
WHERE "id" = $1;

-- name: FindLanguageProgressBySlugAndUserID :one
SELECT * FROM "language_progress"
WHERE "language_slug" = $1 AND "user_id" = $2 LIMIT 1;

-- name: DeleteLanguageProgressByID :exec
DELETE FROM "language_progress"
WHERE "id" = $1;

-- name: AddLanguageProgressCompletedSeries :exec
UPDATE "language_progress" SET
  "completed_series" = "completed_series" + 1
WHERE "id" = $1;

-- name: DecrementLanguageProgressCompletedSeries :exec
UPDATE "language_progress" SET
  "completed_series" = "completed_series" - 1
WHERE "id" = $1;
