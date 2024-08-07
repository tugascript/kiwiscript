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

-- name: CreateLessonProgress :one
INSERT INTO "lesson_progress" (
    "language_slug",
    "series_slug",
    "section_id",
    "lesson_id",
    "language_progress_id",
    "series_progress_id",
    "section_progress_id",
    "user_id"
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
) RETURNING *;

-- name: FindLessonProgressBySlugsIDsAndUserID :one
SELECT * FROM "lesson_progress"
WHERE
    "language_slug" = $1 AND
    "series_slug" = $2 AND
    "section_id" = $3 AND
    "lesson_id" = $4 AND
    "user_id" = $5
LIMIT 1;


-- name: UpdateLessonProgressViewedAt :exec
UPDATE "lesson_progress"
SET "viewed_at" = now()
WHERE "id" = $1;

-- name: CompleteLessonProgress :one
UPDATE "lesson_progress"
SET "completed_at" = now()
WHERE "id" = $1
RETURNING *;

-- name: RemovedLessonProgressCompletedAt :exec
UPDATE "lesson_progress"
SET "completed_at" = NULL
WHERE "id" = $1;

-- name: DeleteLessonProgress :exec
DELETE FROM "lesson_progress"
WHERE "id" = $1;