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

-- name: CreateLectureVideo :one
INSERT INTO "lecture_videos" (
  "lecture_id",
  "url",
  "watch_time_seconds"
) VALUES (
  $1,
  $2,
  $3
) RETURNING *;

-- name: UpdateLectureVideo :one
UPDATE "lecture_videos" SET
  "url" = $1,
  "watch_time_seconds" = $2,
  "updated_at" = NOW()
WHERE "id" = $3
RETURNING *;

-- name: DeleteLectureVideo :exec
DELETE FROM "lecture_videos"
WHERE "id" = $1;

-- name: GetLectureVideoByLectureID :one
SELECT * FROM "lecture_videos"
WHERE "lecture_id" = $1
LIMIT 1;