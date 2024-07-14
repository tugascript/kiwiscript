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

-- name: CreateSeriesTag :exec
INSERT INTO "series_tags" (
    "series_id",
    "tag_id"
) VALUES (
    $1,
    $2
);

-- name: DeleteSeriesTagByIDs :exec
DELETE FROM "series_tags"
WHERE "series_id" = $1 AND "tag_id" = $2;

-- name: CountSeriesTagsBySeriesID :one
SELECT COUNT("tag_id") FROM "series_tags"
WHERE "series_id" = $1
LIMIT 1;

-- name: FindTagsBySeriesID :many
SELECT "tags".* FROM "series_tags"
JOIN "tags" ON "tags"."id" = "series_tags"."tag_id"
WHERE "series_tags"."series_id" = $1
ORDER BY "tags"."name" ASC;
