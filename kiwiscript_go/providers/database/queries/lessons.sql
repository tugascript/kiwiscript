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

-- name: CreateLesson :one
INSERT INTO "lessons" (
  "title",
  "author_id",
  "section_id",
  "language_slug",
  "series_slug",
  "position"
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  (
    SELECT COUNT("id") + 1 FROM "lessons"
    WHERE "section_id" = $3
  )
) RETURNING *;

-- name: UpdateLesson :one
UPDATE "lessons" SET
  "title" = $1
WHERE "id" = $2
RETURNING *;

-- name: UpdateLessonWithPosition :one
UPDATE "lessons" SET
  "title" = $1,
  "position" = $2
WHERE "id" = $3
RETURNING *;

-- name: UpdateLessonIsPublished :one
UPDATE "lessons" SET
  "is_published" = $1
WHERE "id" = $2
RETURNING *;

-- name: UpdateLessonPosition :one
UPDATE "lessons" SET
  "position" = $1
WHERE "id" = $2
RETURNING *;

-- name: IncrementLessonPosition :exec
UPDATE "lessons" SET
  "position" = "position" + 1
WHERE
  "section_id" = $1 AND
  "position" < $2 AND
  "position" >= $3;

-- name: DecrementLessonPosition :exec
UPDATE "lessons" SET
  "position" = "position" - 1
WHERE 
    "section_id" = $1 AND
    "position" > $2 AND 
    "position" <= $3;

-- name: FindLessonBySlugsAndIDs :one
SELECT * FROM "lessons"
WHERE
  "language_slug" = $1 AND
  "series_slug" = $2 AND
  "section_id" = $3 AND
  "id" = $4
LIMIT 1;

-- name: FindPublishedLessonBySlugsAndIDs :one
SELECT * FROM "lessons"
WHERE
  "language_slug" = $1 AND
  "series_slug" = $2 AND
  "section_id" = $3 AND
  "id" = $4 AND
  "is_published" = true
LIMIT 1;

-- name: FindPaginatedLessonsBySlugsAndSectionID :many
SELECT * FROM "lessons"
WHERE
  "language_slug" = $1 AND
  "series_slug" = $2 AND
  "section_id" = $3
ORDER BY "position" ASC
LIMIT $4 OFFSET $5;

-- name: FindPaginatedPublishedLessonsBySlugsAndSectionIDWithProgress :many
SELECT
    "lessons".*,
    "lesson_progress"."completed_at" AS "lesson_progress_completed_at"
FROM "lessons"
LEFT JOIN "lesson_progress" ON (
    "lessons"."id" = "lesson_progress"."lesson_id" AND
    "lesson_progress"."user_id" = $1
)
WHERE
    "lessons"."language_slug" = $2 AND
    "lessons"."series_slug" = $3 AND
    "lessons"."section_id" = $4 AND
    "lessons"."is_published" = true
ORDER BY "lessons"."position" ASC
LIMIT $5 OFFSET $6;

-- name: FindPaginatedPublishedLessonsBySlugsAndSectionID :many
SELECT * FROM "lessons"
WHERE
    "language_slug" = $1 AND
    "series_slug" = $2 AND
    "section_id" = $3 AND
    "is_published" = true
ORDER BY "position" ASC
LIMIT $4 OFFSET $5;


-- name: FindPublishedLessonBySlugsAndIDsWithProgressArticleAndVideo :one
SELECT
    "lessons".*,
    "lesson_progress"."completed_at" AS "lesson_progress_completed_at",
    "lesson_articles"."id" AS "lesson_acticle_id",
    "lesson_articles"."content" AS "lesson_article_content",
    "lesson_videos"."id" AS "lesson_video_id",
    "lesson_videos"."url" AS "lesson_video_url"
FROM "lessons"
LEFT JOIN "lesson_progress" ON (
    "lessons"."id" = "lesson_progress"."lesson_id" AND
    "lesson_progress"."user_id" = $1
)
LEFT JOIN "lesson_articles" ON "lessons"."id" = "lesson_articles"."lesson_id"
LEFT JOIN "lesson_videos" ON "lessons"."id" = "lesson_videos"."lesson_id"
WHERE
    "lessons"."language_slug" = $2 AND
    "lessons"."series_slug" = $3 AND
    "lessons"."section_id" = $4 AND
    "lessons"."id" = $5 AND
    "lessons"."is_published" = true
LIMIT 1;

-- name: FindLessonBySlugsAndIDsWithArticleAndVideo :one
SELECT
    "lessons".*,
    "lesson_articles"."id" AS "lesson_acticle_id",
    "lesson_articles"."content" AS "lesson_article_content",
    "lesson_videos"."id" AS "lesson_video_id",
    "lesson_videos"."url" AS "lesson_video_url"
FROM "lessons"
LEFT JOIN "lesson_articles" ON "lessons"."id" = "lesson_articles"."lesson_id"
LEFT JOIN "lesson_videos" ON "lessons"."id" = "lesson_videos"."lesson_id"
WHERE
    "lessons"."language_slug" = $1 AND
    "lessons"."series_slug" = $2 AND
    "lessons"."section_id" = $3 AND
    "lessons"."id" = $4
LIMIT 1;

-- name: FindLessonsBySectionID :many
SELECT * FROM "lessons"
WHERE "section_id" = $1
ORDER BY "position" ASC;

-- name: UpdateLessonReadTimeSeconds :exec
UPDATE "lessons" SET
  "read_time_seconds" = $1
WHERE "id" = $2;

-- name: UpdateLessonWatchTimeSeconds :exec
UPDATE "lessons" SET
  "watch_time_seconds" = $1
WHERE "id" = $2;

-- name: CountLessonsBySectionID :one
SELECT COUNT("id") FROM "lessons"
WHERE "section_id" = $1
LIMIT 1;

-- name: CountPublishedLessonsBySectionID :one
SELECT COUNT("id") FROM "lessons"
WHERE "section_id" = $1 AND "is_published" = true
LIMIT 1;

-- name: DeleteLessonByID :exec
DELETE FROM "lessons"
WHERE "id" = $1;