-- name: CreateLanguageCategory :one
INSERT INTO "language_categories" (
  "language_id",
  "category_id"
) VALUES (
  $1,
  $2
) RETURNING *;

-- name: DeleteLanguageCategory :exec
DELETE FROM "language_categories"
WHERE "language_id" = $1 AND "category_id" = $2;

-- name: FindCategoryLanguages :many
SELECT "languages".* FROM "language_categories"
LEFT JOIN "languages" ON "languages"."id" = "language_categories"."language_id"
WHERE "category_id" = $1;

-- name: FindLanguageCategories :many
SELECT "categories".* FROM "language_categories"
LEFT JOIN "categories" ON "categories"."id" = "language_categories"."category_id"
WHERE "language_id" = $1;