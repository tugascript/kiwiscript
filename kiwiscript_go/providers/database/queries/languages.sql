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

-- name: CreateLanguage :one
INSERT INTO "languages" (
  "name",
  "icon",
  "author_id"
) VALUES (
  $1,
  $2,
  $3
) RETURNING *;

-- name: UpdateLanguage :one
UPDATE "languages" SET
  "name" = $1,
  "icon" = $2
WHERE "id" = $3
RETURNING *;

-- name: FindLanguageById :one
SELECT * FROM "languages"
WHERE "id" = $1 LIMIT 1;

-- name: FindLanguageByName :one
SELECT * FROM "languages"
WHERE "name" = $1 LIMIT 1;

-- name: FindAllLanguages :many
SELECT * FROM "languages"
ORDER BY "name" ASC;

-- name: FindPaginatedLanguages :many
SELECT * FROM "languages"
ORDER BY "name" ASC
LIMIT $1 OFFSET $2;

-- name: CountLanguages :one
SELECT COUNT("id") FROM "languages";

-- name: FindFilteredPaginatedLanguages :many
SELECT * FROM "languages"
WHERE "name" ILIKE $1
ORDER BY "name" ASC
LIMIT $2 OFFSET $3;

-- name: CountFilteredLanguages :one
SELECT COUNT("id") FROM "languages"
WHERE "name" ILIKE $1;

-- name: DeleteLanguageById :exec
DELETE FROM "languages"
WHERE "id" = $1;

-- name: DeleteAllLanguages :exec
DELETE FROM "languages";