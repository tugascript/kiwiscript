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

-- name: CreateLessonFile :one
INSERT INTO "lesson_files" (
    "id",
    "lesson_id",
    "author_id",
    "ext",
    "name"
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
) RETURNING *;

-- name: DeleteLessonFile :exec
DELETE FROM "lesson_files"
WHERE "id" = $1;

-- name: FindLessonFilesByLessonID :many
SELECT * FROM "lesson_files"
WHERE "lesson_id" = $1
ORDER BY "created_at" ASC;

-- name: FindLessonFileByIDAndLessonID :one
SELECT * FROM "lesson_files"
WHERE "id" = $1 AND "lesson_id" = $2
LIMIT 1;

-- name: UpdateLessonFile :one
UPDATE "lesson_files" SET
    "name" = $1
WHERE "id" = $2
RETURNING *;