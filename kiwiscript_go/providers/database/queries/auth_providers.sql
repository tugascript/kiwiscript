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

-- name: CreateAuthProvider :exec
INSERT INTO "auth_providers" (
  "email",
  "provider"
) VALUES (
  $1,
  $2
);

-- name: FindAuthProviderByEmailAndProvider :one
SELECT * FROM "auth_providers"
WHERE 
  "email" = $1 AND 
  "provider" = $2
LIMIT 1;

-- name: DeleteProviderByEmailAndNotProvider :exec
DELETE FROM "auth_providers"
WHERE "email" = $1 AND "provider" <> $2;