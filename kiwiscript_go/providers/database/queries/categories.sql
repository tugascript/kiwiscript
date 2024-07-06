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

-- name: CreateCategory :one
INSERT INTO "categories" (
  "title",
  "slug",
  "description",
  "author_id"
) VALUES (
  $1,
  $2,
  $3,
  $4
) RETURNING *;

-- name: UpdateCategory :one
UPDATE "categories" SET
  "title" = $1,
  "slug" = $2,
  "description" = $3
WHERE "id" = $4
RETURNING *;

-- name: FindCategoryById :one
SELECT * FROM "categories"
WHERE "id" = $1 LIMIT 1;

-- name: FindCategoryBySlug :one
SELECT * FROM "categories"
WHERE "slug" = $1 LIMIT 1;

-- name: FindAllCategories :many
SELECT * FROM "categories"
ORDER BY "slug" ASC;

-- name: FindPaginatedCategories :many
SELECT * FROM "categories"
ORDER BY "slug" ASC
LIMIT $1 OFFSET $2;

-- name: DeleteCategoryById :exec
DELETE FROM "categories"
WHERE "id" = $1;

