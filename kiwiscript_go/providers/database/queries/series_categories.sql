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

-- name: CreateSeriesCategory :one
INSERT INTO "series_categories" (
  "series_id",
  "category_id"
) VALUES (
  $1,
  $2
) RETURNING *;

-- name: DeleteSeriesCategory :exec
DELETE FROM "series_categories"
WHERE "series_id" = $1 AND "category_id" = $2;

-- name: FindCategorySeries :many
SELECT "series".* FROM "series_categories"
LEFT JOIN "series" ON "series"."id" = "series_categories"."series_id"
WHERE "category_id" = $1;

-- name: FindSeriesCategories :many
SELECT "categories".* FROM "series_categories"
LEFT JOIN "categories" ON "categories"."id" = "series_categories"."category_id"
WHERE "series_id" = $1;
