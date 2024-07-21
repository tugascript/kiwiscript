// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: series_parts.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const addSeriesPartReadTime = `-- name: AddSeriesPartReadTime :exec
UPDATE "series_parts" SET
  "read_time_seconds" = "read_time_seconds" + $1,
  "updated_at" = now()
WHERE "id" = $2
`

type AddSeriesPartReadTimeParams struct {
	ReadTimeSeconds int32
	ID              int32
}

func (q *Queries) AddSeriesPartReadTime(ctx context.Context, arg AddSeriesPartReadTimeParams) error {
	_, err := q.db.Exec(ctx, addSeriesPartReadTime, arg.ReadTimeSeconds, arg.ID)
	return err
}

const addSeriesPartWatchTime = `-- name: AddSeriesPartWatchTime :exec
UPDATE "series_parts" SET
  "watch_time_seconds" = "watch_time_seconds" + $1,
  "updated_at" = now()
WHERE "id" = $2
`

type AddSeriesPartWatchTimeParams struct {
	WatchTimeSeconds int32
	ID               int32
}

func (q *Queries) AddSeriesPartWatchTime(ctx context.Context, arg AddSeriesPartWatchTimeParams) error {
	_, err := q.db.Exec(ctx, addSeriesPartWatchTime, arg.WatchTimeSeconds, arg.ID)
	return err
}

const countSeriesPartsBySeriesSlug = `-- name: CountSeriesPartsBySeriesSlug :one
SELECT COUNT("id") AS "count" FROM "series_parts"
WHERE "series_slug" = $1 LIMIT 1
`

func (q *Queries) CountSeriesPartsBySeriesSlug(ctx context.Context, seriesSlug string) (int64, error) {
	row := q.db.QueryRow(ctx, countSeriesPartsBySeriesSlug, seriesSlug)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createSeriesPart = `-- name: CreateSeriesPart :one

INSERT INTO "series_parts" (
  "title",
  "language_slug",
  "series_slug",
  "description",
  "author_id",
  "position"
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  (
    SELECT COUNT("id") + 1 FROM "series_parts"
    WHERE "series_slug" = $3
  )
) RETURNING id, title, language_slug, series_slug, description, position, lectures_count, watch_time_seconds, read_time_seconds, is_published, author_id, created_at, updated_at
`

type CreateSeriesPartParams struct {
	Title        string
	LanguageSlug string
	SeriesSlug   string
	Description  string
	AuthorID     int32
}

// Copyright (C) 2024 Afonso Barracha
//
// This file is part of KiwiScript.
//
// KiwiScript is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// KiwiScript is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with KiwiScript.  If not, see <https://www.gnu.org/licenses/>.
func (q *Queries) CreateSeriesPart(ctx context.Context, arg CreateSeriesPartParams) (SeriesPart, error) {
	row := q.db.QueryRow(ctx, createSeriesPart,
		arg.Title,
		arg.LanguageSlug,
		arg.SeriesSlug,
		arg.Description,
		arg.AuthorID,
	)
	var i SeriesPart
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.LanguageSlug,
		&i.SeriesSlug,
		&i.Description,
		&i.Position,
		&i.LecturesCount,
		&i.WatchTimeSeconds,
		&i.ReadTimeSeconds,
		&i.IsPublished,
		&i.AuthorID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const decrementSeriesPartLecturesCount = `-- name: DecrementSeriesPartLecturesCount :exec
UPDATE "series_parts" SET
  "lectures_count" = "lectures_count" - 1,
  "watch_time_seconds" = "watch_time_seconds" - $2,
  "read_time_seconds" = "read_time_seconds" - $3,
  "updated_at" = now()
WHERE "id" = $1
`

type DecrementSeriesPartLecturesCountParams struct {
	ID               int32
	WatchTimeSeconds int32
	ReadTimeSeconds  int32
}

func (q *Queries) DecrementSeriesPartLecturesCount(ctx context.Context, arg DecrementSeriesPartLecturesCountParams) error {
	_, err := q.db.Exec(ctx, decrementSeriesPartLecturesCount, arg.ID, arg.WatchTimeSeconds, arg.ReadTimeSeconds)
	return err
}

const decrementSeriesPartPosition = `-- name: DecrementSeriesPartPosition :exec
UPDATE "series_parts" SET
  "position" = "position" - 1
WHERE 
  "series_slug" = $1 AND
  "position" > $2 AND 
  "position" <= $3
`

type DecrementSeriesPartPositionParams struct {
	SeriesSlug string
	Position   int16
	Position_2 int16
}

func (q *Queries) DecrementSeriesPartPosition(ctx context.Context, arg DecrementSeriesPartPositionParams) error {
	_, err := q.db.Exec(ctx, decrementSeriesPartPosition, arg.SeriesSlug, arg.Position, arg.Position_2)
	return err
}

const deleteSeriesPartById = `-- name: DeleteSeriesPartById :exec
DELETE FROM "series_parts"
WHERE "id" = $1
`

func (q *Queries) DeleteSeriesPartById(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, deleteSeriesPartById, id)
	return err
}

const findPaginatedPublishedSeriesPartsByLanguageSlugAndSeriesSlugWithLectures = `-- name: FindPaginatedPublishedSeriesPartsByLanguageSlugAndSeriesSlugWithLectures :many
WITH "series_parts" AS (
    SELECT id, title, language_slug, series_slug, description, position, lectures_count, watch_time_seconds, read_time_seconds, is_published, author_id, created_at, updated_at FROM "series_parts"
    WHERE 
        "series_parts"."language_slug" = $1 AND
        "series_parts"."series_slug" = $2 AND 
        "series_parts"."is_published" = true
    ORDER BY "series_parts"."position" ASC
    LIMIT $3 OFFSET $4
)
SELECT 
    series_parts.id, series_parts.title, series_parts.language_slug, series_parts.series_slug, series_parts.description, series_parts.position, series_parts.lectures_count, series_parts.watch_time_seconds, series_parts.read_time_seconds, series_parts.is_published, series_parts.author_id, series_parts.created_at, series_parts.updated_at, 
    "lectures"."id" AS "lecture_id", 
    "lectures"."title" AS "lecture_title",
    "lectures"."watch_time_seconds" AS "lecture_watch_time_seconds",
    "lectures"."read_time_seconds" AS "lecture_read_time_seconds",
    "lectures"."is_published" AS "lecture_is_published"
FROM "series_parts"
LEFT JOIN "lectures" ON (
    "series_parts"."id" = "lectures"."series_part_id" AND 
    "lectures"."is_published" = true
)
ORDER BY 
    "series_parts"."position" ASC,
    "lectures"."position" ASC
`

type FindPaginatedPublishedSeriesPartsByLanguageSlugAndSeriesSlugWithLecturesParams struct {
	LanguageSlug string
	SeriesSlug   string
	Limit        int32
	Offset       int32
}

type FindPaginatedPublishedSeriesPartsByLanguageSlugAndSeriesSlugWithLecturesRow struct {
	ID                      int32
	Title                   string
	LanguageSlug            string
	SeriesSlug              string
	Description             string
	Position                int16
	LecturesCount           int16
	WatchTimeSeconds        int32
	ReadTimeSeconds         int32
	IsPublished             bool
	AuthorID                int32
	CreatedAt               pgtype.Timestamp
	UpdatedAt               pgtype.Timestamp
	LectureID               pgtype.Int4
	LectureTitle            pgtype.Text
	LectureWatchTimeSeconds pgtype.Int4
	LectureReadTimeSeconds  pgtype.Int4
	LectureIsPublished      pgtype.Bool
}

func (q *Queries) FindPaginatedPublishedSeriesPartsByLanguageSlugAndSeriesSlugWithLectures(ctx context.Context, arg FindPaginatedPublishedSeriesPartsByLanguageSlugAndSeriesSlugWithLecturesParams) ([]FindPaginatedPublishedSeriesPartsByLanguageSlugAndSeriesSlugWithLecturesRow, error) {
	rows, err := q.db.Query(ctx, findPaginatedPublishedSeriesPartsByLanguageSlugAndSeriesSlugWithLectures,
		arg.LanguageSlug,
		arg.SeriesSlug,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []FindPaginatedPublishedSeriesPartsByLanguageSlugAndSeriesSlugWithLecturesRow{}
	for rows.Next() {
		var i FindPaginatedPublishedSeriesPartsByLanguageSlugAndSeriesSlugWithLecturesRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.LanguageSlug,
			&i.SeriesSlug,
			&i.Description,
			&i.Position,
			&i.LecturesCount,
			&i.WatchTimeSeconds,
			&i.ReadTimeSeconds,
			&i.IsPublished,
			&i.AuthorID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.LectureID,
			&i.LectureTitle,
			&i.LectureWatchTimeSeconds,
			&i.LectureReadTimeSeconds,
			&i.LectureIsPublished,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findPaginatedSeriesPartsByLanguageSlugAndSeriesSlugWithLectures = `-- name: FindPaginatedSeriesPartsByLanguageSlugAndSeriesSlugWithLectures :many
WITH "series_parts" AS (
    SELECT id, title, language_slug, series_slug, description, position, lectures_count, watch_time_seconds, read_time_seconds, is_published, author_id, created_at, updated_at FROM "series_parts"
    WHERE 
        "series_parts"."language_slug" = $1 AND
        "series_parts"."series_slug" = $2
    ORDER BY "series_parts"."position" ASC
    LIMIT $3 OFFSET $4
)
SELECT 
    series_parts.id, series_parts.title, series_parts.language_slug, series_parts.series_slug, series_parts.description, series_parts.position, series_parts.lectures_count, series_parts.watch_time_seconds, series_parts.read_time_seconds, series_parts.is_published, series_parts.author_id, series_parts.created_at, series_parts.updated_at, 
    "lectures"."id" AS "lecture_id", 
    "lectures"."title" AS "lecture_title",
    "lectures"."watch_time_seconds" AS "lecture_watch_time_seconds",
    "lectures"."read_time_seconds" AS "lecture_read_time_seconds",
    "lectures"."is_published" AS "lecture_is_published"
FROM "series_parts"
LEFT JOIN "lectures" ON ("series_parts"."id" = "lectures"."series_part_id")
ORDER BY 
    "series_parts"."position" ASC,
    "lectures"."position" ASC
`

type FindPaginatedSeriesPartsByLanguageSlugAndSeriesSlugWithLecturesParams struct {
	LanguageSlug string
	SeriesSlug   string
	Limit        int32
	Offset       int32
}

type FindPaginatedSeriesPartsByLanguageSlugAndSeriesSlugWithLecturesRow struct {
	ID                      int32
	Title                   string
	LanguageSlug            string
	SeriesSlug              string
	Description             string
	Position                int16
	LecturesCount           int16
	WatchTimeSeconds        int32
	ReadTimeSeconds         int32
	IsPublished             bool
	AuthorID                int32
	CreatedAt               pgtype.Timestamp
	UpdatedAt               pgtype.Timestamp
	LectureID               pgtype.Int4
	LectureTitle            pgtype.Text
	LectureWatchTimeSeconds pgtype.Int4
	LectureReadTimeSeconds  pgtype.Int4
	LectureIsPublished      pgtype.Bool
}

func (q *Queries) FindPaginatedSeriesPartsByLanguageSlugAndSeriesSlugWithLectures(ctx context.Context, arg FindPaginatedSeriesPartsByLanguageSlugAndSeriesSlugWithLecturesParams) ([]FindPaginatedSeriesPartsByLanguageSlugAndSeriesSlugWithLecturesRow, error) {
	rows, err := q.db.Query(ctx, findPaginatedSeriesPartsByLanguageSlugAndSeriesSlugWithLectures,
		arg.LanguageSlug,
		arg.SeriesSlug,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []FindPaginatedSeriesPartsByLanguageSlugAndSeriesSlugWithLecturesRow{}
	for rows.Next() {
		var i FindPaginatedSeriesPartsByLanguageSlugAndSeriesSlugWithLecturesRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.LanguageSlug,
			&i.SeriesSlug,
			&i.Description,
			&i.Position,
			&i.LecturesCount,
			&i.WatchTimeSeconds,
			&i.ReadTimeSeconds,
			&i.IsPublished,
			&i.AuthorID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.LectureID,
			&i.LectureTitle,
			&i.LectureWatchTimeSeconds,
			&i.LectureReadTimeSeconds,
			&i.LectureIsPublished,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findPaginatedSeriesPartsByLanguageSlugAndSeriesSlugWithPublishedLectures = `-- name: FindPaginatedSeriesPartsByLanguageSlugAndSeriesSlugWithPublishedLectures :many
WITH "series_parts" AS (
    SELECT id, title, language_slug, series_slug, description, position, lectures_count, watch_time_seconds, read_time_seconds, is_published, author_id, created_at, updated_at FROM "series_parts"
    WHERE 
        "series_parts"."language_slug" = $1 AND
        "series_parts"."series_slug" = $2
    ORDER BY "series_parts"."position" ASC
    LIMIT $3 OFFSET $4
)
SELECT 
    series_parts.id, series_parts.title, series_parts.language_slug, series_parts.series_slug, series_parts.description, series_parts.position, series_parts.lectures_count, series_parts.watch_time_seconds, series_parts.read_time_seconds, series_parts.is_published, series_parts.author_id, series_parts.created_at, series_parts.updated_at, 
    "lectures"."id" AS "lecture_id", 
    "lectures"."title" AS "lecture_title",
    "lectures"."watch_time_seconds" AS "lecture_watch_time_seconds",
    "lectures"."read_time_seconds" AS "lecture_read_time_seconds",
    "lectures"."is_published" AS "lecture_is_published"
FROM "series_parts"
LEFT JOIN "lectures" ON (
    "series_parts"."id" = "lectures"."series_part_id" AND 
    "lectures"."is_published" = true
)
ORDER BY 
    "series_parts"."position" ASC,
    "lectures"."position" ASC
`

type FindPaginatedSeriesPartsByLanguageSlugAndSeriesSlugWithPublishedLecturesParams struct {
	LanguageSlug string
	SeriesSlug   string
	Limit        int32
	Offset       int32
}

type FindPaginatedSeriesPartsByLanguageSlugAndSeriesSlugWithPublishedLecturesRow struct {
	ID                      int32
	Title                   string
	LanguageSlug            string
	SeriesSlug              string
	Description             string
	Position                int16
	LecturesCount           int16
	WatchTimeSeconds        int32
	ReadTimeSeconds         int32
	IsPublished             bool
	AuthorID                int32
	CreatedAt               pgtype.Timestamp
	UpdatedAt               pgtype.Timestamp
	LectureID               pgtype.Int4
	LectureTitle            pgtype.Text
	LectureWatchTimeSeconds pgtype.Int4
	LectureReadTimeSeconds  pgtype.Int4
	LectureIsPublished      pgtype.Bool
}

func (q *Queries) FindPaginatedSeriesPartsByLanguageSlugAndSeriesSlugWithPublishedLectures(ctx context.Context, arg FindPaginatedSeriesPartsByLanguageSlugAndSeriesSlugWithPublishedLecturesParams) ([]FindPaginatedSeriesPartsByLanguageSlugAndSeriesSlugWithPublishedLecturesRow, error) {
	rows, err := q.db.Query(ctx, findPaginatedSeriesPartsByLanguageSlugAndSeriesSlugWithPublishedLectures,
		arg.LanguageSlug,
		arg.SeriesSlug,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []FindPaginatedSeriesPartsByLanguageSlugAndSeriesSlugWithPublishedLecturesRow{}
	for rows.Next() {
		var i FindPaginatedSeriesPartsByLanguageSlugAndSeriesSlugWithPublishedLecturesRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.LanguageSlug,
			&i.SeriesSlug,
			&i.Description,
			&i.Position,
			&i.LecturesCount,
			&i.WatchTimeSeconds,
			&i.ReadTimeSeconds,
			&i.IsPublished,
			&i.AuthorID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.LectureID,
			&i.LectureTitle,
			&i.LectureWatchTimeSeconds,
			&i.LectureReadTimeSeconds,
			&i.LectureIsPublished,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findPublishedSeriesPartByLanguageSlugSeriesSlugAndIDWithLectures = `-- name: FindPublishedSeriesPartByLanguageSlugSeriesSlugAndIDWithLectures :many
SELECT 
    series_parts.id, series_parts.title, series_parts.language_slug, series_parts.series_slug, series_parts.description, series_parts.position, series_parts.lectures_count, series_parts.watch_time_seconds, series_parts.read_time_seconds, series_parts.is_published, series_parts.author_id, series_parts.created_at, series_parts.updated_at, 
    "lectures"."id" AS "lecture_id", 
    "lectures"."title" AS "lecture_title",
    "lectures"."watch_time_seconds" AS "lecture_watch_time_seconds",
    "lectures"."read_time_seconds" AS "lecture_read_time_seconds",
    "lectures"."is_published" AS "lecture_is_published"
FROM "series_parts"
LEFT JOIN "lectures" ON (
    "series_parts"."id" = "lectures"."series_part_id" AND 
    "lectures"."is_published" = true
)
WHERE
    "series_parts"."language_slug" = $1 AND
    "series_parts"."series_slug" = $2 AND 
    "series_parts"."id" = $3 AND
    "series_parts"."is_published" = true
ORDER BY "lectures"."position" ASC
`

type FindPublishedSeriesPartByLanguageSlugSeriesSlugAndIDWithLecturesParams struct {
	LanguageSlug string
	SeriesSlug   string
	ID           int32
}

type FindPublishedSeriesPartByLanguageSlugSeriesSlugAndIDWithLecturesRow struct {
	ID                      int32
	Title                   string
	LanguageSlug            string
	SeriesSlug              string
	Description             string
	Position                int16
	LecturesCount           int16
	WatchTimeSeconds        int32
	ReadTimeSeconds         int32
	IsPublished             bool
	AuthorID                int32
	CreatedAt               pgtype.Timestamp
	UpdatedAt               pgtype.Timestamp
	LectureID               pgtype.Int4
	LectureTitle            pgtype.Text
	LectureWatchTimeSeconds pgtype.Int4
	LectureReadTimeSeconds  pgtype.Int4
	LectureIsPublished      pgtype.Bool
}

func (q *Queries) FindPublishedSeriesPartByLanguageSlugSeriesSlugAndIDWithLectures(ctx context.Context, arg FindPublishedSeriesPartByLanguageSlugSeriesSlugAndIDWithLecturesParams) ([]FindPublishedSeriesPartByLanguageSlugSeriesSlugAndIDWithLecturesRow, error) {
	rows, err := q.db.Query(ctx, findPublishedSeriesPartByLanguageSlugSeriesSlugAndIDWithLectures, arg.LanguageSlug, arg.SeriesSlug, arg.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []FindPublishedSeriesPartByLanguageSlugSeriesSlugAndIDWithLecturesRow{}
	for rows.Next() {
		var i FindPublishedSeriesPartByLanguageSlugSeriesSlugAndIDWithLecturesRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.LanguageSlug,
			&i.SeriesSlug,
			&i.Description,
			&i.Position,
			&i.LecturesCount,
			&i.WatchTimeSeconds,
			&i.ReadTimeSeconds,
			&i.IsPublished,
			&i.AuthorID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.LectureID,
			&i.LectureTitle,
			&i.LectureWatchTimeSeconds,
			&i.LectureReadTimeSeconds,
			&i.LectureIsPublished,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findSeriesPartById = `-- name: FindSeriesPartById :one
SELECT id, title, language_slug, series_slug, description, position, lectures_count, watch_time_seconds, read_time_seconds, is_published, author_id, created_at, updated_at FROM "series_parts"
WHERE "id" = $1 LIMIT 1
`

func (q *Queries) FindSeriesPartById(ctx context.Context, id int32) (SeriesPart, error) {
	row := q.db.QueryRow(ctx, findSeriesPartById, id)
	var i SeriesPart
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.LanguageSlug,
		&i.SeriesSlug,
		&i.Description,
		&i.Position,
		&i.LecturesCount,
		&i.WatchTimeSeconds,
		&i.ReadTimeSeconds,
		&i.IsPublished,
		&i.AuthorID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const findSeriesPartByLanguageSlugSeriesSlugAndID = `-- name: FindSeriesPartByLanguageSlugSeriesSlugAndID :one
SELECT id, title, language_slug, series_slug, description, position, lectures_count, watch_time_seconds, read_time_seconds, is_published, author_id, created_at, updated_at FROM "series_parts"
WHERE
    "language_slug" = $1 AND
    "series_slug" = $2 AND
    "id" = $3 
LIMIT 1
`

type FindSeriesPartByLanguageSlugSeriesSlugAndIDParams struct {
	LanguageSlug string
	SeriesSlug   string
	ID           int32
}

func (q *Queries) FindSeriesPartByLanguageSlugSeriesSlugAndID(ctx context.Context, arg FindSeriesPartByLanguageSlugSeriesSlugAndIDParams) (SeriesPart, error) {
	row := q.db.QueryRow(ctx, findSeriesPartByLanguageSlugSeriesSlugAndID, arg.LanguageSlug, arg.SeriesSlug, arg.ID)
	var i SeriesPart
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.LanguageSlug,
		&i.SeriesSlug,
		&i.Description,
		&i.Position,
		&i.LecturesCount,
		&i.WatchTimeSeconds,
		&i.ReadTimeSeconds,
		&i.IsPublished,
		&i.AuthorID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const findSeriesPartByLanguageSlugSeriesSlugAndIDWithLectures = `-- name: FindSeriesPartByLanguageSlugSeriesSlugAndIDWithLectures :many
SELECT 
    series_parts.id, series_parts.title, series_parts.language_slug, series_parts.series_slug, series_parts.description, series_parts.position, series_parts.lectures_count, series_parts.watch_time_seconds, series_parts.read_time_seconds, series_parts.is_published, series_parts.author_id, series_parts.created_at, series_parts.updated_at, 
    "lectures"."id" AS "lecture_id", 
    "lectures"."title" AS "lecture_title",
    "lectures"."watch_time_seconds" AS "lecture_watch_time_seconds",
    "lectures"."read_time_seconds" AS "lecture_read_time_seconds",
    "lectures"."is_published" AS "lecture_is_published"
FROM "series_parts"
LEFT JOIN "lectures" ON ("series_parts"."id" = "lectures"."series_part_id")
WHERE 
    "series_parts"."language_slug" = $1 AND
    "series_parts"."series_slug" = $2 AND 
    "series_parts"."id" = $3
ORDER BY "lectures"."position" ASC
`

type FindSeriesPartByLanguageSlugSeriesSlugAndIDWithLecturesParams struct {
	LanguageSlug string
	SeriesSlug   string
	ID           int32
}

type FindSeriesPartByLanguageSlugSeriesSlugAndIDWithLecturesRow struct {
	ID                      int32
	Title                   string
	LanguageSlug            string
	SeriesSlug              string
	Description             string
	Position                int16
	LecturesCount           int16
	WatchTimeSeconds        int32
	ReadTimeSeconds         int32
	IsPublished             bool
	AuthorID                int32
	CreatedAt               pgtype.Timestamp
	UpdatedAt               pgtype.Timestamp
	LectureID               pgtype.Int4
	LectureTitle            pgtype.Text
	LectureWatchTimeSeconds pgtype.Int4
	LectureReadTimeSeconds  pgtype.Int4
	LectureIsPublished      pgtype.Bool
}

func (q *Queries) FindSeriesPartByLanguageSlugSeriesSlugAndIDWithLectures(ctx context.Context, arg FindSeriesPartByLanguageSlugSeriesSlugAndIDWithLecturesParams) ([]FindSeriesPartByLanguageSlugSeriesSlugAndIDWithLecturesRow, error) {
	rows, err := q.db.Query(ctx, findSeriesPartByLanguageSlugSeriesSlugAndIDWithLectures, arg.LanguageSlug, arg.SeriesSlug, arg.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []FindSeriesPartByLanguageSlugSeriesSlugAndIDWithLecturesRow{}
	for rows.Next() {
		var i FindSeriesPartByLanguageSlugSeriesSlugAndIDWithLecturesRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.LanguageSlug,
			&i.SeriesSlug,
			&i.Description,
			&i.Position,
			&i.LecturesCount,
			&i.WatchTimeSeconds,
			&i.ReadTimeSeconds,
			&i.IsPublished,
			&i.AuthorID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.LectureID,
			&i.LectureTitle,
			&i.LectureWatchTimeSeconds,
			&i.LectureReadTimeSeconds,
			&i.LectureIsPublished,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const incrementSeriesPartLecturesCount = `-- name: IncrementSeriesPartLecturesCount :exec
UPDATE "series_parts" SET
  "lectures_count" = "lectures_count" + 1,
  "watch_time_seconds" = "watch_time_seconds" + $2,
  "read_time_seconds" = "read_time_seconds" + $3,
  "updated_at" = now()
WHERE "id" = $1
`

type IncrementSeriesPartLecturesCountParams struct {
	ID               int32
	WatchTimeSeconds int32
	ReadTimeSeconds  int32
}

func (q *Queries) IncrementSeriesPartLecturesCount(ctx context.Context, arg IncrementSeriesPartLecturesCountParams) error {
	_, err := q.db.Exec(ctx, incrementSeriesPartLecturesCount, arg.ID, arg.WatchTimeSeconds, arg.ReadTimeSeconds)
	return err
}

const incrementSeriesPartPosition = `-- name: IncrementSeriesPartPosition :exec
UPDATE "series_parts" SET
  "position" = "position" + 1
WHERE
  "series_slug" = $1 AND 
  "position" < $2 AND
  "position" >= $3
`

type IncrementSeriesPartPositionParams struct {
	SeriesSlug string
	Position   int16
	Position_2 int16
}

func (q *Queries) IncrementSeriesPartPosition(ctx context.Context, arg IncrementSeriesPartPositionParams) error {
	_, err := q.db.Exec(ctx, incrementSeriesPartPosition, arg.SeriesSlug, arg.Position, arg.Position_2)
	return err
}

const updateSeriesPart = `-- name: UpdateSeriesPart :one
UPDATE "series_parts" SET
  "title" = $1,
  "description" = $2
WHERE "id" = $3
RETURNING id, title, language_slug, series_slug, description, position, lectures_count, watch_time_seconds, read_time_seconds, is_published, author_id, created_at, updated_at
`

type UpdateSeriesPartParams struct {
	Title       string
	Description string
	ID          int32
}

func (q *Queries) UpdateSeriesPart(ctx context.Context, arg UpdateSeriesPartParams) (SeriesPart, error) {
	row := q.db.QueryRow(ctx, updateSeriesPart, arg.Title, arg.Description, arg.ID)
	var i SeriesPart
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.LanguageSlug,
		&i.SeriesSlug,
		&i.Description,
		&i.Position,
		&i.LecturesCount,
		&i.WatchTimeSeconds,
		&i.ReadTimeSeconds,
		&i.IsPublished,
		&i.AuthorID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateSeriesPartIsPublished = `-- name: UpdateSeriesPartIsPublished :one
UPDATE "series_parts" SET
  "is_published" = $1,
  "updated_at" = now()
WHERE "id" = $2
RETURNING id, title, language_slug, series_slug, description, position, lectures_count, watch_time_seconds, read_time_seconds, is_published, author_id, created_at, updated_at
`

type UpdateSeriesPartIsPublishedParams struct {
	IsPublished bool
	ID          int32
}

func (q *Queries) UpdateSeriesPartIsPublished(ctx context.Context, arg UpdateSeriesPartIsPublishedParams) (SeriesPart, error) {
	row := q.db.QueryRow(ctx, updateSeriesPartIsPublished, arg.IsPublished, arg.ID)
	var i SeriesPart
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.LanguageSlug,
		&i.SeriesSlug,
		&i.Description,
		&i.Position,
		&i.LecturesCount,
		&i.WatchTimeSeconds,
		&i.ReadTimeSeconds,
		&i.IsPublished,
		&i.AuthorID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateSeriesPartWithPosition = `-- name: UpdateSeriesPartWithPosition :one
UPDATE "series_parts" SET
  "title" = $1,
  "description" = $2,
  "position" = $3,
  "updated_at" = now()
WHERE "id" = $4
RETURNING id, title, language_slug, series_slug, description, position, lectures_count, watch_time_seconds, read_time_seconds, is_published, author_id, created_at, updated_at
`

type UpdateSeriesPartWithPositionParams struct {
	Title       string
	Description string
	Position    int16
	ID          int32
}

func (q *Queries) UpdateSeriesPartWithPosition(ctx context.Context, arg UpdateSeriesPartWithPositionParams) (SeriesPart, error) {
	row := q.db.QueryRow(ctx, updateSeriesPartWithPosition,
		arg.Title,
		arg.Description,
		arg.Position,
		arg.ID,
	)
	var i SeriesPart
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.LanguageSlug,
		&i.SeriesSlug,
		&i.Description,
		&i.Position,
		&i.LecturesCount,
		&i.WatchTimeSeconds,
		&i.ReadTimeSeconds,
		&i.IsPublished,
		&i.AuthorID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
