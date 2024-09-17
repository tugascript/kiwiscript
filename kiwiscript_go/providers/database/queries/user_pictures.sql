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

-- name: CreateUserPicture :one
INSERT INTO "user_pictures" (
    "id",
    "user_id",
    "ext"
) VALUES (
    $1,
    $2,
    $3
) RETURNING *;

-- name: DeleteUserPicture :exec
DELETE FROM "user_pictures"
WHERE "id" = $1;

-- name: FindUserPictureByUserID :one
SELECT * FROM "user_pictures"
WHERE "user_id" = $1
LIMIT 1;