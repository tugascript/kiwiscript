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

