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

-- name: FindCertificateByUserIDAndSeriesSlug :one
SELECT * FROM "certificates"
WHERE "user_id" = $1 AND "series_slug" = $2
LIMIT 1;

-- name: CreateCertificate :one
INSERT INTO "certificates" (
    "id",
    "user_id",
    "language_slug",
    "series_slug",
    "series_title",
    "lessons",
    "watch_time_seconds",
    "read_time_seconds",
    "completed_at"
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    now()
) RETURNING *;

-- name: FindPaginatedCertificatesByUserID :many
SELECT
    "certificates".*,
    "languages"."id" AS "language_id",
    "languages"."name" AS "language_name"
FROM "certificates"
INNER JOIN "languages" ON "certificates"."language_slug" = "languages"."slug"
WHERE "certificates"."user_id" = $1
ORDER BY "certificates"."created_at" ASC
LIMIT $2 OFFSET $3;

-- name: CountCertificatesByUserID :one
SELECT COUNT("id") AS "count"
FROM "certificates"
WHERE "user_id" = $1
LIMIT 1;

-- name: FindCertificateByIDWithUserAndLanguage :one
SELECT
    "certificates".*,
    "languages"."id" AS "language_id",
    "languages"."name" AS "language_name",
    "users"."first_name" AS "author_first_name",
    "users"."last_name" AS "author_last_name"
FROM "certificates"
INNER JOIN "languages" ON "certificates"."language_slug" = "languages"."slug"
INNER JOIN "users" ON "certificates"."user_id" = "users"."id"
WHERE "certificates"."id" = $1
LIMIT 1;