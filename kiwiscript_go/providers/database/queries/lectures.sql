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

-- name: CreateLecture :one
INSERT INTO "lectures" (
  "title",
  "author_id",
  "series_part_id",
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
    SELECT COUNT("id") + 1 FROM "lectures"
    WHERE "series_part_id" = $3
  )
) RETURNING *;

-- name: UpdateLecture :one
UPDATE "lectures" SET
  "title" = $1
WHERE "id" = $2
RETURNING *;

-- name: UpdateLectureWithPosition :one
UPDATE "lectures" SET
  "title" = $1,
  "position" = $2
WHERE "id" = $3
RETURNING *;

-- name: UpdateLectureIsPublished :one
UPDATE "lectures" SET
  "is_published" = $1
WHERE "id" = $2
RETURNING *;

-- name: UpdateLecturePosition :one
UPDATE "lectures" SET
  "position" = $1
WHERE "id" = $2
RETURNING *;

-- name: IncrementLecturePosition :exec
UPDATE "lectures" SET
  "position" = "position" + 1
WHERE
  "series_part_id" = $1 AND
  "position" < $2 AND
  "position" >= $3;

-- name: DecrementLecturePosition :exec
UPDATE "lectures" SET
  "position" = "position" - 1
WHERE 
    "series_part_id" = $1 AND 
    "position" > $2 AND 
    "position" <= $3;

-- name: IncrementLectureCommentsCount :exec
UPDATE "lectures" SET
  "comments_count" = "comments_count" + 1
WHERE "id" = $1;

-- name: DecrementLectureCommentsCount :exec
UPDATE "lectures" SET
  "comments_count" = "comments_count" - 1
WHERE "id" = $1;

-- name: UpdateLectureReadTimeSeconds :exec
UPDATE "lectures" SET
  "read_time_seconds" = $1
WHERE "id" = $2;

-- name: UpdateLectureWatchTimeSeconds :exec
UPDATE "lectures" SET
  "watch_time_seconds" = $1
WHERE "id" = $2;

-- name: FindLectureBySlugsAndIDs :one
SELECT * FROM "lectures"
WHERE
  "language_slug" = $1 AND
  "series_slug" = $2 AND
  "series_part_id" = $3 AND
  "id" = $4
LIMIT 1;

-- name: FindLectureBySlugsAndIDsWithArticleAndVideo :one
SELECT 
  "lectures".*,
  "lecture_articles"."id" AS "article_id",
  "lecture_articles"."content" AS "article_content",
  "lecture_videos"."id" AS "video_id",
  "lecture_videos"."url" AS "video_url"
FROM "lectures"
LEFT JOIN "lecture_articles" ON "lecture_articles"."lecture_id" = "lectures"."id"
LEFT JOIN "lecture_videos" ON "lecture_videos"."lecture_id" = "lectures"."id"
WHERE
  "lectures"."language_slug" = $1 AND
  "lectures"."series_slug" = $2 AND
  "lectures"."series_part_id" = $3 AND
  "lectures"."id" = $4
LIMIT 1;

-- name: FindPaginatedLecturesBySeriesPartIDWithArticleAndVideo :many
SELECT 
  "lectures".*,
  "lecture_articles"."id" AS "article_id",
  "lecture_articles"."content" AS "article_content",
  "lecture_videos"."id" AS "video_id",
  "lecture_videos"."url" AS "video_url"
FROM "lectures"
LEFT JOIN "lecture_articles" ON "lecture_articles"."lecture_id" = "lectures"."id"
LEFT JOIN "lecture_videos" ON "lecture_videos"."lecture_id" = "lectures"."id"
WHERE
  "lectures"."language_slug" = $1 AND
  "lectures"."series_slug" = $2 AND
  "lectures"."series_part_id" = $3
ORDER BY "position" ASC
LIMIT $4 OFFSET $5;

-- name: FindPaginatedPublishedLecturesBySeriesPartIDWithArticleAndVideo :many
SELECT 
  "lectures".*,
  "lecture_articles"."id" AS "article_id",
  "lecture_articles"."content" AS "article_content",
  "lecture_videos"."id" AS "video_id",
  "lecture_videos"."url" AS "video_url"
FROM "lectures"
LEFT JOIN "lecture_articles" ON "lecture_articles"."lecture_id" = "lectures"."id"
LEFT JOIN "lecture_videos" ON "lecture_videos"."lecture_id" = "lectures"."id"
WHERE
  "lectures"."language_slug" = $1 AND
  "lectures"."series_slug" = $2 AND
  "lectures"."series_part_id" = $3 AND
  "lectures"."is_published" = true
ORDER BY "position" ASC
LIMIT $4 OFFSET $5;

-- name: FindLecturesBySeriesPartID :many
SELECT * FROM "lectures"
WHERE "series_part_id" = $1
ORDER BY "position" ASC;

-- name: CountLecturesBySeriesPartID :one
SELECT COUNT("id") FROM "lectures"
WHERE "series_part_id" = $1
LIMIT 1;

-- name: CountPublishedLecturesBySeriesPartID :one
SELECT COUNT("id") FROM "lectures"
WHERE "series_part_id" = $1 AND "is_published" = true
LIMIT 1;

-- name: DeleteLectureByID :exec
DELETE FROM "lectures"
WHERE "id" = $1;