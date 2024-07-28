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

-- name: CreateLectureProgress :one
-- INSERT INTO "lecture_progress" (
--   "language_slug",
--   "series_slug",
--   "lecture_slug",
--   "language_progress_id",
--   "user_id",
--   "is_current"
-- ) VALUES (
--   $1,
--   $2,
--   $3,
--   $4,
--   $5,
--   true
-- ) RETURNING *;