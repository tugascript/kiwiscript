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

-- name: CreateUserProfile :one
INSERT INTO "user_profiles" (
    "user_id",
    "bio",
    "github",
    "linkedin",
    "website"
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
) RETURNING *;

-- name: UpdateUserProfile :one
UPDATE "user_profiles" SET
  "bio" = $1,
  "github" = $2,
  "linkedin" = $3,
  "website" = $4,
  "updated_at" = NOW()
WHERE "id" = $5
RETURNING *;

-- name: DeleteUserProfile :exec
DELETE FROM "user_profiles"
WHERE "id" = $1;

-- name: FindUserProfileByUserID :one
SELECT * FROM "user_profiles"
WHERE "user_id" = $1
LIMIT 1;
