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

-- name: CreateSeries :one
INSERT INTO "series" (
  "title",
  "slug",
  "description",
  "language_slug",
  "author_id"
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5
) RETURNING *;

-- name: UpdateSeries :one
UPDATE "series" SET
  "title" = $1,
  "slug" = $2,
  "description" = $3,
  "updated_at" = now()
WHERE "id" = $4
RETURNING *;

-- name: UpdateSeriesIsPublished :one
UPDATE "series" SET
  "is_published" = $1,
  "updated_at" = now()
WHERE "id" = $2
RETURNING *;

-- name: FindPublishedSeriesBySlugsWithAuthor :one
SELECT
  "series".*,
  "users"."first_name" AS "author_first_name",
  "users"."last_name" AS "author_last_name",
  "series_pictures"."id" AS "picture_id",
  "series_pictures"."ext" AS "picture_ext"
FROM "series"
INNER JOIN "users" ON "series"."author_id" = "users"."id"
LEFT JOIN "series_pictures" ON "series"."id" = "series_pictures"."series_id"
WHERE
  "series"."slug" = $1 AND
  "series"."language_slug" = $2 AND
  "series"."is_published" = true
LIMIT 1;

-- name: FindPaginatedSeriesWithAuthorSortByID :many
SELECT
  "series".*,
  "users"."first_name" AS "author_first_name",
  "users"."last_name" AS "author_last_name",
  "series_pictures"."id" AS "picture_id",
  "series_pictures"."ext" AS "picture_ext"
FROM "series"
INNER JOIN "users" ON "series"."author_id" = "users"."id"
LEFT JOIN "series_pictures" ON "series"."id" = "series_pictures"."series_id"
WHERE "series"."language_slug" = $1
ORDER BY "series"."id" DESC
LIMIT $2 OFFSET $3;

-- name: FindPaginatedPublishedSeriesWithAuthorSortByID :many
SELECT
  "series".*,
  "users"."first_name" AS "author_first_name",
  "users"."last_name" AS "author_last_name",
  "series_pictures"."id" AS "picture_id",
  "series_pictures"."ext" AS "picture_ext"
FROM "series"
INNER JOIN "users" ON "series"."author_id" = "users"."id"
LEFT JOIN "series_pictures" ON "series"."id" = "series_pictures"."series_id"
WHERE
    "series"."language_slug" = $1 AND
    "series"."is_published" = true
ORDER BY "series"."id" DESC
LIMIT $2 OFFSET $3;

-- name: FindFilteredSeriesWithAuthorSortByID :many
SELECT
  "series".*,
  "users"."first_name" AS "author_first_name",
  "users"."last_name" AS "author_last_name",
  "series_pictures"."id" AS "picture_id",
  "series_pictures"."ext" AS "picture_ext"
FROM "series"
INNER JOIN "users" ON "series"."author_id" = "users"."id"
LEFT JOIN "series_pictures" ON "series"."id" = "series_pictures"."series_id"
WHERE
    "series"."language_slug" = $1 AND
    (
        "series"."title" ILIKE $2 OR
        "users"."first_name" ILIKE $2 OR
        "users"."last_name" ILIKE $2
    )
ORDER BY "series"."id" DESC
LIMIT $3 OFFSET $4;

-- name: FindFilteredPublishedSeriesWithAuthorSortByID :many
SELECT
  "series".*,
  "users"."first_name" AS "author_first_name",
  "users"."last_name" AS "author_last_name",
  "series_pictures"."id" AS "picture_id",
  "series_pictures"."ext" AS "picture_ext"
FROM "series"
INNER JOIN "users" ON "series"."author_id" = "users"."id"
LEFT JOIN "series_pictures" ON "series"."id" = "series_pictures"."series_id"
WHERE
    "series"."language_slug" = $1 AND
    "series"."is_published" = true AND
    (
        "series"."title" ILIKE $2 OR
        "users"."first_name" ILIKE $2 OR
        "users"."last_name" ILIKE $2
    )
ORDER BY "series"."id" DESC
LIMIT $3 OFFSET $4;

-- name: CountSeries :one
SELECT COUNT("id") FROM "series"
WHERE "language_slug" = $1;

-- name: CountPublishedSeries :one
SELECT COUNT("id") FROM "series"
WHERE
    "language_slug" = $1 AND
    "is_published" = true;

-- name: CountAllPublishedSeries :one
SELECT COUNT("id") AS "count" FROM "series"
WHERE "is_published" = true
LIMIT 1;

-- name: CountAllFilteredPublishedSeries :one
SELECT COUNT("series"."id") AS "count" FROM "series"
INNER JOIN "users" ON "series"."author_id" = "users"."id"
WHERE
    "series"."is_published" = true AND
    (
        "series"."title" ILIKE $1 OR
        "users"."first_name" ILIKE $1 OR
        "users"."last_name" ILIKE $1
    )
LIMIT 1;

-- name: CountFilteredSeries :one
SELECT COUNT("series"."id") FROM "series"
INNER JOIN "users" ON "series"."author_id" = "users"."id"
WHERE
    "series"."language_slug" = $1 AND
    (
        "series"."title" ILIKE $2 OR
        "users"."first_name" ILIKE $2 OR
        "users"."last_name" ILIKE $2
    );

-- name: CountFilteredPublishedSeries :one
SELECT COUNT("series"."id") FROM "series"
INNER JOIN "users" ON "series"."author_id" = "users"."id"
WHERE
    "series"."language_slug" = $1 AND
    "series"."is_published" = true AND
    (
        "series"."title" ILIKE $2 OR
        "users"."first_name" ILIKE $2 OR
        "users"."last_name" ILIKE $2
    );

-- name: FindPaginatedSeriesWithAuthorSortBySlug :many
SELECT
  "series".*,
  "users"."first_name" AS "author_first_name",
  "users"."last_name" AS "author_last_name",
  "series_pictures"."id" AS "picture_id",
  "series_pictures"."ext" AS "picture_ext"
FROM "series"
INNER JOIN "users" ON "series"."author_id" = "users"."id"
LEFT JOIN "series_pictures" ON "series"."id" = "series_pictures"."series_id"
WHERE "series"."language_slug" = $1
ORDER BY "series"."slug" ASC
LIMIT $2 OFFSET $3;

-- name: FindFilteredSeriesWithAuthorSortBySlug :many
SELECT
  "series".*,
  "users"."first_name" AS "author_first_name",
  "users"."last_name" AS "author_last_name",
  "series_pictures"."id" AS "picture_id",
  "series_pictures"."ext" AS "picture_ext"
FROM "series"
INNER JOIN "users" ON "series"."author_id" = "users"."id"
LEFT JOIN "series_pictures" ON "series"."id" = "series_pictures"."series_id"
WHERE
    "series"."language_slug" = $1 AND
    (
        "series"."title" ILIKE $2 OR
        "users"."first_name" ILIKE $2 OR
        "users"."last_name" ILIKE $2
    )
ORDER BY "series"."slug" ASC
LIMIT $3 OFFSET $4;

-- name: FindPaginatedPublishedSeriesWithAuthorSortBySlug :many
SELECT
  "series".*,
  "users"."first_name" AS "author_first_name",
  "users"."last_name" AS "author_last_name",
  "series_pictures"."id" AS "picture_id",
  "series_pictures"."ext" AS "picture_ext"
FROM "series"
INNER JOIN "users" ON "series"."author_id" = "users"."id"
LEFT JOIN "series_pictures" ON "series"."id" = "series_pictures"."series_id"
WHERE
    "series"."language_slug" = $1 AND
    "series"."is_published" = true
ORDER BY "series"."slug" ASC
LIMIT $2 OFFSET $3;

-- name: FindFilteredPublishedSeriesWithAuthorSortBySlug :many
SELECT
  "series".*,
  "users"."first_name" AS "author_first_name",
  "users"."last_name" AS "author_last_name",
  "series_pictures"."id" AS "picture_id",
  "series_pictures"."ext" AS "picture_ext"
FROM "series"
INNER JOIN "users" ON "series"."author_id" = "users"."id"
LEFT JOIN "series_pictures" ON "series"."id" = "series_pictures"."series_id"
WHERE
    "series"."language_slug" = $1 AND
    "series"."is_published" = true AND
    (
        "series"."title" ILIKE $2 OR
        "users"."first_name" ILIKE $2 OR
        "users"."last_name" ILIKE $2
    )
ORDER BY "series"."slug" ASC
LIMIT $3 OFFSET $4;

-- name: FindPaginatedPublishedSeriesWithAuthorAndProgressSortByID :many
SELECT
  "series".*,
  "users"."first_name" AS "author_first_name",
  "users"."last_name" AS "author_last_name",
  "series_progress"."id" AS "series_progress_id",
  "series_progress"."completed_sections" AS "series_progress_completed_sections",
  "series_progress"."completed_lessons" AS "series_progress_completed_lessons",
  "series_progress"."viewed_at" AS "series_progress_viewed_at",
  "series_progress"."completed_at" AS "series_progress_completed_at",
  "series_pictures"."id" AS "picture_id",
  "series_pictures"."ext" AS "picture_ext"
FROM "series"
INNER JOIN "users" ON "series"."author_id" = "users"."id"
LEFT JOIN "series_progress" ON (
    "series"."slug" = "series_progress"."series_slug" AND
    "series_progress"."user_id" = $1
)
LEFT JOIN "series_pictures" ON "series"."id" = "series_pictures"."series_id"
WHERE
  "series"."language_slug" = $2 AND
  "series"."is_published" = true
ORDER BY "series"."id" DESC
LIMIT $3 OFFSET $4;

-- name: FindPaginatedPublishedSeriesWithAuthorAndProgressSortBySlug :many
SELECT
  "series".*,
  "users"."first_name" AS "author_first_name",
  "users"."last_name" AS "author_last_name",
  "series_progress"."id" AS "series_progress_id",
  "series_progress"."completed_sections" AS "series_progress_completed_sections",
  "series_progress"."completed_lessons" AS "series_progress_completed_lessons",
  "series_progress"."viewed_at" AS "series_progress_viewed_at",
  "series_progress"."completed_at" AS "series_progress_completed_at",
  "series_pictures"."id" AS "picture_id",
  "series_pictures"."ext" AS "picture_ext"
FROM "series"
INNER JOIN "users" ON "series"."author_id" = "users"."id"
LEFT JOIN "series_progress" ON (
    "series"."slug" = "series_progress"."series_slug" AND
    "series_progress"."user_id" = $1
)
LEFT JOIN "series_pictures" ON "series"."id" = "series_pictures"."series_id"
WHERE
  "series"."language_slug" = $2 AND
  "series"."is_published" = true
ORDER BY "series"."slug" ASC
LIMIT $3 OFFSET $4;

-- name: FindFilteredPublishedSeriesWithAuthorAndProgressSortByID :many
SELECT
  "series".*,
  "users"."first_name" AS "author_first_name",
  "users"."last_name" AS "author_last_name",
  "series_progress"."id" AS "series_progress_id",
  "series_progress"."completed_sections" AS "series_progress_completed_sections",
  "series_progress"."completed_lessons" AS "series_progress_completed_lessons",
  "series_progress"."viewed_at" AS "series_progress_viewed_at",
  "series_progress"."completed_at" AS "series_progress_completed_at",
  "series_pictures"."id" AS "picture_id",
  "series_pictures"."ext" AS "picture_ext"
FROM "series"
INNER JOIN "users" ON "series"."author_id" = "users"."id"
LEFT JOIN "series_progress" ON (
    "series"."slug" = "series_progress"."series_slug" AND
    "series_progress"."user_id" = $1
)
LEFT JOIN "series_pictures" ON "series"."id" = "series_pictures"."series_id"
WHERE
    "series"."language_slug" = $3 AND
    "series"."is_published" = true AND
    (
        "series"."title" ILIKE $2 OR
        "users"."first_name" ILIKE $2 OR
        "users"."last_name" ILIKE $2
    )
ORDER BY "series"."id" DESC
LIMIT $4 OFFSET $5;

-- name: FindFilteredPublishedSeriesWithAuthorAndProgressSortBySlug :many
SELECT
  "series".*,
  "users"."first_name" AS "author_first_name",
  "users"."last_name" AS "author_last_name",
  "series_progress"."id" AS "series_progress_id",
  "series_progress"."completed_sections" AS "series_progress_completed_sections",
  "series_progress"."completed_lessons" AS "series_progress_completed_lessons",
  "series_progress"."viewed_at" AS "series_progress_viewed_at",
  "series_progress"."completed_at" AS "series_progress_completed_at",
  "series_pictures"."id" AS "picture_id",
  "series_pictures"."ext" AS "picture_ext"
FROM "series"
INNER JOIN "users" ON "series"."author_id" = "users"."id"
LEFT JOIN "series_progress" ON (
    "series"."slug" = "series_progress"."series_slug" AND
    "series_progress"."user_id" = $1
)
LEFT JOIN "series_pictures" ON "series"."id" = "series_pictures"."series_id"
WHERE
  "series"."language_slug" = $3 AND
  "series"."is_published" = true AND
  (
    "series"."title" ILIKE $2 OR
    "users"."first_name" ILIKE $2 OR
    "users"."last_name" ILIKE $2
  )
ORDER BY "series"."slug" ASC
LIMIT $4 OFFSET $5;

-- name: FindPaginatedPublishedSeriesWithAuthorAndInnerProgress :many
SELECT
  "series".*,
  "users"."first_name" AS "author_first_name",
  "users"."last_name" AS "author_last_name",
  "series_progress"."id" AS "series_progress_id",
  "series_progress"."completed_sections" AS "series_progress_completed_sections",
  "series_progress"."completed_lessons" AS "series_progress_completed_lessons",
  "series_progress"."viewed_at" AS "series_progress_viewed_at",
  "series_progress"."completed_at" AS "series_progress_completed_at",
  "series_pictures"."id" AS "picture_id",
  "series_pictures"."ext" AS "picture_ext"
FROM "series"
INNER JOIN "users" ON "series"."author_id" = "users"."id"
INNER JOIN "series_progress" ON (
    "series"."slug" = "series_progress"."series_slug" AND
    "series_progress"."user_id" = $1
)
LEFT JOIN "series_pictures" ON "series"."id" = "series_pictures"."series_id"
WHERE
  "series"."language_slug" = $2 AND
  "series"."is_published" = true
ORDER BY "series_progress"."viewed_at" DESC
LIMIT $3 OFFSET $4;

-- name: CountPublishedSeriesWithInnerProgress :one
SELECT COUNT("series"."id") AS "count" FROM "series"
INNER JOIN "series_progress" ON (
    "series"."slug" = "series_progress"."series_slug" AND
    "series_progress"."user_id" = $1
)
WHERE
    "series"."language_slug" = $2 AND
    "series"."is_published" = true
LIMIT 1;

-- name: AddSeriesSectionsCount :exec
UPDATE "series" SET
  "sections_count" = "sections_count" + 1,
  "lessons_count" = "lessons_count" + $2,
  "watch_time_seconds" = "watch_time_seconds" + $3,
  "read_time_seconds" = "read_time_seconds" + $4,
  "updated_at" = now()
WHERE "slug" = $1;

-- name: DecrementSeriesSectionsCount :exec
UPDATE "series" SET
  "sections_count" = "sections_count" - 1,
  "lessons_count" = "lessons_count" - $2,
  "watch_time_seconds" = "watch_time_seconds" - $3,
  "read_time_seconds" = "read_time_seconds" - $4,
  "updated_at" = now()
WHERE "slug" = $1;

-- name: IncrementSeriesLessonsCount :exec
UPDATE "series" SET
  "lessons_count" = "lessons_count" + 1,
  "watch_time_seconds" = "watch_time_seconds" + $2,
  "read_time_seconds" = "read_time_seconds" + $3,
  "updated_at" = now()
WHERE "slug" = $1;

-- name: DecrementSeriesLessonsCount :exec
UPDATE "series" SET
  "lessons_count" = "lessons_count" - 1,
  "watch_time_seconds" = "watch_time_seconds" + $2,
  "read_time_seconds" = "read_time_seconds" + $3,
  "updated_at" = now()
WHERE "slug" = $1;

-- name: AddSeriesWatchTime :exec
UPDATE "series" SET
  "watch_time_seconds" = "watch_time_seconds" + $1,
  "updated_at" = now()
WHERE "slug" = $2;

-- name: AddSeriesReadTime :exec
UPDATE "series" SET
  "read_time_seconds" = "read_time_seconds" + $1,
  "updated_at" = now()
WHERE "slug" = $2;

-- name: FindSeriesById :one
SELECT * FROM "series"
WHERE "id" = $1 LIMIT 1;

-- name: FindSeriesBySlugAndLanguageSlug :one
SELECT * FROM "series"
WHERE "slug" = $1 AND "language_slug" = $2
LIMIT 1;

-- name: FindPublishedSeriesBySlugAndLanguageSlug :one
SELECT * FROM "series"
WHERE
    "slug" = $1 AND
    "language_slug" = $2 AND
    "is_published" = true
LIMIT 1;

-- name: DeleteSeriesById :exec
DELETE FROM "series"
WHERE "id" = $1;

-- name: FindSeriesBySlugWithAuthor :one
SELECT
  "series".*,
  "users"."first_name" AS "author_first_name",
  "users"."last_name" AS "author_last_name",
  "series_pictures"."id" AS "picture_id",
  "series_pictures"."ext" AS "picture_ext"
FROM "series"
INNER JOIN "users" ON "series"."author_id" = "users"."id"
LEFT JOIN "series_pictures" ON "series"."id" = "series_pictures"."series_id"
WHERE
  "series"."slug" = $1 AND
  "series"."language_slug" = $2
LIMIT 1;

-- name: FindPublishedSeriesBySlugWithAuthorAndProgress :one
SELECT
    "series".*,
    "users"."first_name" AS "author_first_name",
    "users"."last_name" AS "author_last_name",
    "series_progress"."id" AS "series_progress_id",
    "series_progress"."completed_sections" AS "series_progress_completed_sections",
    "series_progress"."completed_lessons" AS "series_progress_completed_lessons",
    "series_progress"."viewed_at" AS "series_progress_viewed_at",
    "series_progress"."completed_at" AS "series_progress_completed_at",
    "series_pictures"."id" AS "picture_id",
    "series_pictures"."ext" AS "picture_ext"
FROM "series"
INNER JOIN "users" ON "series"."author_id" = "users"."id"
LEFT JOIN "series_progress" ON (
    "series"."slug" = "series_progress"."series_slug" AND
    "series_progress"."user_id" = $1
)
LEFT JOIN "series_pictures" ON "series"."id" = "series_pictures"."series_id"
WHERE
    "series"."slug" = $2 AND
    "series"."language_slug" = $3 AND
    "series"."is_published" = true
LIMIT 1;

-- name: FindPaginatedPublishedSeriesWithAuthorAndLanguage :many
SELECT
  "series".*,
  "users"."first_name" AS "author_first_name",
  "users"."last_name" AS "author_last_name",
  "series_pictures"."id" AS "picture_id",
  "series_pictures"."ext" AS "picture_ext",
  "languages"."name" AS "language_name",
  "languages"."icon" AS "language_icon"
FROM "series"
INNER JOIN "users" ON "series"."author_id" = "users"."id"
INNER JOIN "languages" ON "series"."language_slug" = "languages"."slug"
LEFT JOIN "series_pictures" ON "series"."id" = "series_pictures"."series_id"
WHERE
    "series"."is_published" = true
ORDER BY "series"."slug" ASC
LIMIT $1 OFFSET $2;

-- name: FindFilteredPublishedSeriesWithAuthorAndLanguage :many
SELECT
  "series".*,
  "users"."first_name" AS "author_first_name",
  "users"."last_name" AS "author_last_name",
  "series_pictures"."id" AS "picture_id",
  "series_pictures"."ext" AS "picture_ext",
  "languages"."name" AS "language_name",
  "languages"."icon" AS "language_icon"
FROM "series"
INNER JOIN "users" ON "series"."author_id" = "users"."id"
INNER JOIN "languages" ON "series"."language_slug" = "languages"."slug"
LEFT JOIN "series_pictures" ON "series"."id" = "series_pictures"."series_id"
WHERE
    "series"."is_published" = true AND
    (
        "series"."title" ILIKE $3 OR
        "users"."first_name" ILIKE $3 OR
        "users"."last_name" ILIKE $3
    )
ORDER BY "series"."slug" ASC
LIMIT $1 OFFSET $2;

-- name: FindPaginatedPublishedSeriesWithAuthorLanguageAndProgress :many
SELECT
    "series".*,
    "users"."first_name" AS "author_first_name",
    "users"."last_name" AS "author_last_name",
    "series_progress"."id" AS "series_progress_id",
    "series_progress"."completed_sections" AS "series_progress_completed_sections",
    "series_progress"."completed_lessons" AS "series_progress_completed_lessons",
    "series_progress"."viewed_at" AS "series_progress_viewed_at",
    "series_progress"."completed_at" AS "series_progress_completed_at",
    "series_pictures"."id" AS "picture_id",
    "series_pictures"."ext" AS "picture_ext",
    "languages"."name" AS "language_name",
    "languages"."icon" AS "language_icon"
FROM "series"
INNER JOIN "users" ON "series"."author_id" = "users"."id"
INNER JOIN "languages" ON "series"."language_slug" = "languages"."slug"
LEFT JOIN "series_progress" ON (
    "series"."slug" = "series_progress"."series_slug" AND
    "series_progress"."user_id" = $1
)
LEFT JOIN "series_pictures" ON "series"."id" = "series_pictures"."series_id"
WHERE
    "series"."is_published" = true
ORDER BY "series"."slug" ASC
LIMIT $2 OFFSET $3;

-- name: FindFilteredPublishedSeriesWithAuthorLanguageAndProgress :many
SELECT
    "series".*,
    "users"."first_name" AS "author_first_name",
    "users"."last_name" AS "author_last_name",
    "series_progress"."id" AS "series_progress_id",
    "series_progress"."completed_sections" AS "series_progress_completed_sections",
    "series_progress"."completed_lessons" AS "series_progress_completed_lessons",
    "series_progress"."viewed_at" AS "series_progress_viewed_at",
    "series_progress"."completed_at" AS "series_progress_completed_at",
    "series_pictures"."id" AS "picture_id",
    "series_pictures"."ext" AS "picture_ext",
    "languages"."name" AS "language_name",
    "languages"."icon" AS "language_icon"
FROM "series"
INNER JOIN "users" ON "series"."author_id" = "users"."id"
INNER JOIN "languages" ON "series"."language_slug" = "languages"."slug"
LEFT JOIN "series_progress" ON (
    "series"."slug" = "series_progress"."series_slug" AND
    "series_progress"."user_id" = $4
)
LEFT JOIN "series_pictures" ON "series"."id" = "series_pictures"."series_id"
WHERE
    "series"."is_published" = true AND
    (
        "series"."title" ILIKE $3 OR
        "users"."first_name" ILIKE $3 OR
        "users"."last_name" ILIKE $3
    )
ORDER BY "series"."slug" ASC
LIMIT $1 OFFSET $2;

-- name: DeleteAllLanguageSeries :exec
DELETE FROM "series"
WHERE "language_slug" = $1;

