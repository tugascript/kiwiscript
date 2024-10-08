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

-- name: CreateUserWithPassword :one
INSERT INTO "users" (
  "first_name",
  "last_name",
  "location",
  "email",
  "password",
  "is_confirmed"
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  false
) RETURNING *;

-- name: CreateUserWithoutPassword :one
INSERT INTO "users" (
  "first_name",
  "last_name",
  "location",
  "email",
  "is_confirmed"
) VALUES (
  $1,
  $2,
  $3,
  $4,
  true
) RETURNING *;

-- name: UpdateUserEmail :one
UPDATE "users" SET
  "email" = $1,
  "version" = "version" + 1
WHERE "id" = $2
RETURNING *;

-- name: UpdateUserPassword :one
UPDATE "users" SET
  "password" = $1,
  "version" = "version" + 1
WHERE "id" = $2
RETURNING *;

-- name: UpdateUser :one
UPDATE "users" SET
  "first_name" = $1,
  "last_name" = $2,
  "location" = $3
WHERE "id" = $4
RETURNING *;

-- name: FindUserByEmail :one
SELECT * FROM "users"
WHERE "email" = $1 LIMIT 1;

-- name: FindUserById :one
SELECT * FROM "users"
WHERE "id" = $1 LIMIT 1;

-- name: FindStaffUserByIdWithProfileAndPicture :one
SELECT
    "users".*,
    "user_profiles"."id" AS "profile_id",
    "user_profiles"."bio" AS "profile_bio",
    "user_profiles"."github" AS "profile_github",
    "user_profiles"."linkedin" AS "profile_linkedin",
    "user_profiles"."website" AS "profile_website",
    "user_pictures"."id" AS "picture_id",
    "user_pictures"."ext" AS "picture_ext"
FROM "users"
LEFT JOIN "user_profiles" ON "users"."id" = "user_profiles"."user_id"
LEFT JOIN "user_pictures" ON "users"."id" = "user_pictures"."user_id"
WHERE "users"."id" = $1 AND "users"."is_staff" = true LIMIT 1;

-- name: DeleteUserById :exec
DELETE FROM "users"
WHERE "id" = $1;

-- name: ConfirmUser :one
UPDATE "users" SET
  "is_confirmed" = true,
  "version" = "version" + 1
WHERE "id" = $1
RETURNING *;

-- name: UpdateUserIsStaff :exec
UPDATE "users" SET
  "is_staff" = $1,
  "version" = "version" + 1
WHERE "id" = $2;

-- name: DeleteAllUsers :exec
DELETE FROM "users";
