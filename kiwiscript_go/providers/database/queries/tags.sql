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

-- name: DeleteTagById :exec
DELETE FROM "tags"
WHERE "id" = $1;

-- name: DeleteManyTags :exec
DELETE FROM "tags"
WHERE "ID" IN ($1::int[]);

-- name: FindOrCreateTag :one
INSERT INTO "tags" (
  "name",
  "author_id"
) VALUES (
  $1,
  $2
) ON CONFLICT ("name") DO UPDATE SET
  "name" = EXCLUDED."name"
RETURNING *;
