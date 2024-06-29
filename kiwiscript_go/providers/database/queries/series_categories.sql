-- name: CreateSeriesCategory :one
INSERT INTO "series_categories" (
  "series_id",
  "category_id"
) VALUES (
  $1,
  $2
) RETURNING *;

-- name: DeleteSeriesCategory :exec
DELETE FROM "series_categories"
WHERE "series_id" = $1 AND "category_id" = $2;

-- name: FindCategorySeries :many
SELECT "series".* FROM "series_categories"
LEFT JOIN "series" ON "series"."id" = "series_categories"."series_id"
WHERE "category_id" = $1;

-- name: FindSeriesCategories :many
SELECT "categories".* FROM "series_categories"
LEFT JOIN "categories" ON "categories"."id" = "series_categories"."category_id"
WHERE "series_id" = $1;
