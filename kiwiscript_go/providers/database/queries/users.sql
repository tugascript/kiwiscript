-- name: CreateUserWithPassword :one
INSERT INTO "users" (
  "first_name",
  "last_name",
  "location",
  "email",
  "birth_date",
  "password",
  "is_confirmed"
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  false
) RETURNING *;

-- name: CreateUserWithoutPassword :one
INSERT INTO "users" (
  "first_name",
  "last_name",
  "location",
  "email",
  "birth_date",
  "is_confirmed"
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
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

-- name: UpdateUser :exec
UPDATE "users" SET
  "first_name" = $1,
  "last_name" = $2,
  "location" = $3
WHERE "id" = $4;

-- name: FindUserByEmail :one
SELECT * FROM "users"
WHERE "email" = $1 LIMIT 1;

-- name: FindUserById :one
SELECT * FROM "users"
WHERE "id" = $1 LIMIT 1;

-- name: DeleteUserById :exec
DELETE FROM "users"
WHERE "id" = $1;

-- name: ConfirmUser :one
UPDATE "users" SET
  "is_confirmed" = true,
  "version" = "version" + 1
WHERE "id" = $1
RETURNING *;