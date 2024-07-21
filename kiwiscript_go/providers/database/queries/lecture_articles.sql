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

-- name: CreateLectureArticle :one
INSERT INTO "lecture_articles" (
  "lecture_id",
  "content",
  "read_time_seconds"
) VALUES (
  $1,
  $2,
  $3
) RETURNING *;

-- name: UpdateLectureArticle :one
UPDATE "lecture_articles" SET
  "content" = $1,
  "read_time_seconds" = $2,
  "updated_at" = NOW()
WHERE "id" = $3
RETURNING *;

-- name: DeleteLectureArticle :exec
DELETE FROM "lecture_articles"
WHERE "id" = $1;

-- name: GetLectureArticleByLectureID :one
SELECT * FROM "lecture_articles"
WHERE "lecture_id" = $1
LIMIT 1;