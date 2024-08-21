// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: sections.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const addSectionReadTime = `-- name: AddSectionReadTime :exec
UPDATE "sections" SET
  "read_time_seconds" = "read_time_seconds" + $1,
  "updated_at" = now()
WHERE "id" = $2
`

type AddSectionReadTimeParams struct {
	ReadTimeSeconds int32
	ID              int32
}

func (q *Queries) AddSectionReadTime(ctx context.Context, arg AddSectionReadTimeParams) error {
	_, err := q.db.Exec(ctx, addSectionReadTime, arg.ReadTimeSeconds, arg.ID)
	return err
}

const addSectionWatchTime = `-- name: AddSectionWatchTime :exec
UPDATE "sections" SET
  "watch_time_seconds" = "watch_time_seconds" + $1,
  "updated_at" = now()
WHERE "id" = $2
`

type AddSectionWatchTimeParams struct {
	WatchTimeSeconds int32
	ID               int32
}

func (q *Queries) AddSectionWatchTime(ctx context.Context, arg AddSectionWatchTimeParams) error {
	_, err := q.db.Exec(ctx, addSectionWatchTime, arg.WatchTimeSeconds, arg.ID)
	return err
}

const countPublishedSectionsBySeriesSlug = `-- name: CountPublishedSectionsBySeriesSlug :one
SELECT COUNT("id") AS "count" FROM "sections"
WHERE "series_slug" = $1 AND "is_published" = true LIMIT 1
`

func (q *Queries) CountPublishedSectionsBySeriesSlug(ctx context.Context, seriesSlug string) (int64, error) {
	row := q.db.QueryRow(ctx, countPublishedSectionsBySeriesSlug, seriesSlug)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const countSectionsBySeriesSlug = `-- name: CountSectionsBySeriesSlug :one
SELECT COUNT("id") AS "count" FROM "sections"
WHERE "series_slug" = $1 LIMIT 1
`

func (q *Queries) CountSectionsBySeriesSlug(ctx context.Context, seriesSlug string) (int64, error) {
	row := q.db.QueryRow(ctx, countSectionsBySeriesSlug, seriesSlug)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createSection = `-- name: CreateSection :one

INSERT INTO "sections" (
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
  $6
) RETURNING id, title, language_slug, series_slug, description, position, lessons_count, watch_time_seconds, read_time_seconds, is_published, author_id, created_at, updated_at
`

type CreateSectionParams struct {
	Title        string
	LanguageSlug string
	SeriesSlug   string
	Description  string
	AuthorID     int32
	Position     int16
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
func (q *Queries) CreateSection(ctx context.Context, arg CreateSectionParams) (Section, error) {
	row := q.db.QueryRow(ctx, createSection,
		arg.Title,
		arg.LanguageSlug,
		arg.SeriesSlug,
		arg.Description,
		arg.AuthorID,
		arg.Position,
	)
	var i Section
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.LanguageSlug,
		&i.SeriesSlug,
		&i.Description,
		&i.Position,
		&i.LessonsCount,
		&i.WatchTimeSeconds,
		&i.ReadTimeSeconds,
		&i.IsPublished,
		&i.AuthorID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const decrementSectionLessonsCount = `-- name: DecrementSectionLessonsCount :exec
UPDATE "sections" SET
  "lessons_count" = "lessons_count" - 1,
  "watch_time_seconds" = "watch_time_seconds" - $2,
  "read_time_seconds" = "read_time_seconds" - $3,
  "updated_at" = now()
WHERE "id" = $1
`

type DecrementSectionLessonsCountParams struct {
	ID               int32
	WatchTimeSeconds int32
	ReadTimeSeconds  int32
}

func (q *Queries) DecrementSectionLessonsCount(ctx context.Context, arg DecrementSectionLessonsCountParams) error {
	_, err := q.db.Exec(ctx, decrementSectionLessonsCount, arg.ID, arg.WatchTimeSeconds, arg.ReadTimeSeconds)
	return err
}

const decrementSectionPosition = `-- name: DecrementSectionPosition :exec
UPDATE "sections" SET
  "position" = "position" - 1
WHERE 
  "series_slug" = $1 AND
  "position" > $2 AND 
  "position" <= $3
`

type DecrementSectionPositionParams struct {
	SeriesSlug string
	Position   int16
	Position_2 int16
}

func (q *Queries) DecrementSectionPosition(ctx context.Context, arg DecrementSectionPositionParams) error {
	_, err := q.db.Exec(ctx, decrementSectionPosition, arg.SeriesSlug, arg.Position, arg.Position_2)
	return err
}

const deleteSectionById = `-- name: DeleteSectionById :exec
DELETE FROM "sections"
WHERE "id" = $1
`

func (q *Queries) DeleteSectionById(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, deleteSectionById, id)
	return err
}

const findPaginatedPublishedSectionsBySlugs = `-- name: FindPaginatedPublishedSectionsBySlugs :many
SELECT id, title, language_slug, series_slug, description, position, lessons_count, watch_time_seconds, read_time_seconds, is_published, author_id, created_at, updated_at FROM "sections"
WHERE
    "language_slug" = $1 AND
    "series_slug" = $2 AND
    "is_published" = true
ORDER BY "position" ASC
LIMIT $3 OFFSET $4
`

type FindPaginatedPublishedSectionsBySlugsParams struct {
	LanguageSlug string
	SeriesSlug   string
	Limit        int32
	Offset       int32
}

func (q *Queries) FindPaginatedPublishedSectionsBySlugs(ctx context.Context, arg FindPaginatedPublishedSectionsBySlugsParams) ([]Section, error) {
	rows, err := q.db.Query(ctx, findPaginatedPublishedSectionsBySlugs,
		arg.LanguageSlug,
		arg.SeriesSlug,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Section{}
	for rows.Next() {
		var i Section
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.LanguageSlug,
			&i.SeriesSlug,
			&i.Description,
			&i.Position,
			&i.LessonsCount,
			&i.WatchTimeSeconds,
			&i.ReadTimeSeconds,
			&i.IsPublished,
			&i.AuthorID,
			&i.CreatedAt,
			&i.UpdatedAt,
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

const findPaginatedPublishedSectionsBySlugsWithProgress = `-- name: FindPaginatedPublishedSectionsBySlugsWithProgress :many
SELECT
    sections.id, sections.title, sections.language_slug, sections.series_slug, sections.description, sections.position, sections.lessons_count, sections.watch_time_seconds, sections.read_time_seconds, sections.is_published, sections.author_id, sections.created_at, sections.updated_at,
    "section_progress"."completed_lessons" AS "section_progress_completed_lessons",
    "section_progress"."completed_at" AS "section_progress_completed_at",
    "section_progress"."viewed_at" AS "section_progress_viewed_at"
FROM "sections"
LEFT JOIN "section_progress" ON (
    "sections"."id" = "section_progress"."section_id" AND
    "section_progress"."user_id" = $1
)
WHERE
    "sections"."language_slug" = $2 AND
    "sections"."series_slug" = $3 AND
    "sections"."is_published" = true
ORDER BY "sections"."position" ASC
LIMIT $4 OFFSET $5
`

type FindPaginatedPublishedSectionsBySlugsWithProgressParams struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	Limit        int32
	Offset       int32
}

type FindPaginatedPublishedSectionsBySlugsWithProgressRow struct {
	ID                              int32
	Title                           string
	LanguageSlug                    string
	SeriesSlug                      string
	Description                     string
	Position                        int16
	LessonsCount                    int16
	WatchTimeSeconds                int32
	ReadTimeSeconds                 int32
	IsPublished                     bool
	AuthorID                        int32
	CreatedAt                       pgtype.Timestamp
	UpdatedAt                       pgtype.Timestamp
	SectionProgressCompletedLessons pgtype.Int2
	SectionProgressCompletedAt      pgtype.Timestamp
	SectionProgressViewedAt         pgtype.Timestamp
}

func (q *Queries) FindPaginatedPublishedSectionsBySlugsWithProgress(ctx context.Context, arg FindPaginatedPublishedSectionsBySlugsWithProgressParams) ([]FindPaginatedPublishedSectionsBySlugsWithProgressRow, error) {
	rows, err := q.db.Query(ctx, findPaginatedPublishedSectionsBySlugsWithProgress,
		arg.UserID,
		arg.LanguageSlug,
		arg.SeriesSlug,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []FindPaginatedPublishedSectionsBySlugsWithProgressRow{}
	for rows.Next() {
		var i FindPaginatedPublishedSectionsBySlugsWithProgressRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.LanguageSlug,
			&i.SeriesSlug,
			&i.Description,
			&i.Position,
			&i.LessonsCount,
			&i.WatchTimeSeconds,
			&i.ReadTimeSeconds,
			&i.IsPublished,
			&i.AuthorID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.SectionProgressCompletedLessons,
			&i.SectionProgressCompletedAt,
			&i.SectionProgressViewedAt,
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

const findPaginatedSectionsBySlugs = `-- name: FindPaginatedSectionsBySlugs :many
SELECT id, title, language_slug, series_slug, description, position, lessons_count, watch_time_seconds, read_time_seconds, is_published, author_id, created_at, updated_at FROM "sections"
WHERE
    "sections"."language_slug" = $1 AND
    "sections"."series_slug" = $2
ORDER BY "sections"."position" ASC
LIMIT $3 OFFSET $4
`

type FindPaginatedSectionsBySlugsParams struct {
	LanguageSlug string
	SeriesSlug   string
	Limit        int32
	Offset       int32
}

func (q *Queries) FindPaginatedSectionsBySlugs(ctx context.Context, arg FindPaginatedSectionsBySlugsParams) ([]Section, error) {
	rows, err := q.db.Query(ctx, findPaginatedSectionsBySlugs,
		arg.LanguageSlug,
		arg.SeriesSlug,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Section{}
	for rows.Next() {
		var i Section
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.LanguageSlug,
			&i.SeriesSlug,
			&i.Description,
			&i.Position,
			&i.LessonsCount,
			&i.WatchTimeSeconds,
			&i.ReadTimeSeconds,
			&i.IsPublished,
			&i.AuthorID,
			&i.CreatedAt,
			&i.UpdatedAt,
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

const findPublishedSectionBySlugsAndIDWithProgress = `-- name: FindPublishedSectionBySlugsAndIDWithProgress :one
SELECT
    sections.id, sections.title, sections.language_slug, sections.series_slug, sections.description, sections.position, sections.lessons_count, sections.watch_time_seconds, sections.read_time_seconds, sections.is_published, sections.author_id, sections.created_at, sections.updated_at,
    "section_progress"."completed_lessons" AS "section_progress_completed_lessons",
    "section_progress"."completed_at" AS "section_progress_completed_at",
    "section_progress"."viewed_at" AS "section_progress_viewed_at"
FROM "sections"
LEFT JOIN "section_progress" ON (
    "sections"."id" = "section_progress"."section_id" AND
    "section_progress"."user_id" = $1
)
WHERE
    "sections"."language_slug" = $2 AND
    "sections"."series_slug" = $3 AND
    "sections"."id" = $4 AND
    "sections"."is_published" = true
LIMIT 1
`

type FindPublishedSectionBySlugsAndIDWithProgressParams struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	ID           int32
}

type FindPublishedSectionBySlugsAndIDWithProgressRow struct {
	ID                              int32
	Title                           string
	LanguageSlug                    string
	SeriesSlug                      string
	Description                     string
	Position                        int16
	LessonsCount                    int16
	WatchTimeSeconds                int32
	ReadTimeSeconds                 int32
	IsPublished                     bool
	AuthorID                        int32
	CreatedAt                       pgtype.Timestamp
	UpdatedAt                       pgtype.Timestamp
	SectionProgressCompletedLessons pgtype.Int2
	SectionProgressCompletedAt      pgtype.Timestamp
	SectionProgressViewedAt         pgtype.Timestamp
}

func (q *Queries) FindPublishedSectionBySlugsAndIDWithProgress(ctx context.Context, arg FindPublishedSectionBySlugsAndIDWithProgressParams) (FindPublishedSectionBySlugsAndIDWithProgressRow, error) {
	row := q.db.QueryRow(ctx, findPublishedSectionBySlugsAndIDWithProgress,
		arg.UserID,
		arg.LanguageSlug,
		arg.SeriesSlug,
		arg.ID,
	)
	var i FindPublishedSectionBySlugsAndIDWithProgressRow
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.LanguageSlug,
		&i.SeriesSlug,
		&i.Description,
		&i.Position,
		&i.LessonsCount,
		&i.WatchTimeSeconds,
		&i.ReadTimeSeconds,
		&i.IsPublished,
		&i.AuthorID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.SectionProgressCompletedLessons,
		&i.SectionProgressCompletedAt,
		&i.SectionProgressViewedAt,
	)
	return i, err
}

const findSectionById = `-- name: FindSectionById :one
SELECT id, title, language_slug, series_slug, description, position, lessons_count, watch_time_seconds, read_time_seconds, is_published, author_id, created_at, updated_at FROM "sections"
WHERE "id" = $1 LIMIT 1
`

func (q *Queries) FindSectionById(ctx context.Context, id int32) (Section, error) {
	row := q.db.QueryRow(ctx, findSectionById, id)
	var i Section
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.LanguageSlug,
		&i.SeriesSlug,
		&i.Description,
		&i.Position,
		&i.LessonsCount,
		&i.WatchTimeSeconds,
		&i.ReadTimeSeconds,
		&i.IsPublished,
		&i.AuthorID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const findSectionBySlugsAndID = `-- name: FindSectionBySlugsAndID :one
SELECT id, title, language_slug, series_slug, description, position, lessons_count, watch_time_seconds, read_time_seconds, is_published, author_id, created_at, updated_at FROM "sections"
WHERE
    "language_slug" = $1 AND
    "series_slug" = $2 AND
    "id" = $3 
LIMIT 1
`

type FindSectionBySlugsAndIDParams struct {
	LanguageSlug string
	SeriesSlug   string
	ID           int32
}

func (q *Queries) FindSectionBySlugsAndID(ctx context.Context, arg FindSectionBySlugsAndIDParams) (Section, error) {
	row := q.db.QueryRow(ctx, findSectionBySlugsAndID, arg.LanguageSlug, arg.SeriesSlug, arg.ID)
	var i Section
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.LanguageSlug,
		&i.SeriesSlug,
		&i.Description,
		&i.Position,
		&i.LessonsCount,
		&i.WatchTimeSeconds,
		&i.ReadTimeSeconds,
		&i.IsPublished,
		&i.AuthorID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const incrementSectionLessonsCount = `-- name: IncrementSectionLessonsCount :exec
UPDATE "sections" SET
  "lessons_count" = "lessons_count" + 1,
  "watch_time_seconds" = "watch_time_seconds" + $2,
  "read_time_seconds" = "read_time_seconds" + $3,
  "updated_at" = now()
WHERE "id" = $1
`

type IncrementSectionLessonsCountParams struct {
	ID               int32
	WatchTimeSeconds int32
	ReadTimeSeconds  int32
}

func (q *Queries) IncrementSectionLessonsCount(ctx context.Context, arg IncrementSectionLessonsCountParams) error {
	_, err := q.db.Exec(ctx, incrementSectionLessonsCount, arg.ID, arg.WatchTimeSeconds, arg.ReadTimeSeconds)
	return err
}

const incrementSectionPosition = `-- name: IncrementSectionPosition :exec
UPDATE "sections" SET
  "position" = "position" + 1
WHERE
  "series_slug" = $1 AND 
  "position" < $2 AND
  "position" >= $3
`

type IncrementSectionPositionParams struct {
	SeriesSlug string
	Position   int16
	Position_2 int16
}

func (q *Queries) IncrementSectionPosition(ctx context.Context, arg IncrementSectionPositionParams) error {
	_, err := q.db.Exec(ctx, incrementSectionPosition, arg.SeriesSlug, arg.Position, arg.Position_2)
	return err
}

const updateSection = `-- name: UpdateSection :one
UPDATE "sections" SET
  "title" = $1,
  "description" = $2
WHERE "id" = $3
RETURNING id, title, language_slug, series_slug, description, position, lessons_count, watch_time_seconds, read_time_seconds, is_published, author_id, created_at, updated_at
`

type UpdateSectionParams struct {
	Title       string
	Description string
	ID          int32
}

func (q *Queries) UpdateSection(ctx context.Context, arg UpdateSectionParams) (Section, error) {
	row := q.db.QueryRow(ctx, updateSection, arg.Title, arg.Description, arg.ID)
	var i Section
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.LanguageSlug,
		&i.SeriesSlug,
		&i.Description,
		&i.Position,
		&i.LessonsCount,
		&i.WatchTimeSeconds,
		&i.ReadTimeSeconds,
		&i.IsPublished,
		&i.AuthorID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateSectionIsPublished = `-- name: UpdateSectionIsPublished :one
UPDATE "sections" SET
  "is_published" = $1,
  "updated_at" = now()
WHERE "id" = $2
RETURNING id, title, language_slug, series_slug, description, position, lessons_count, watch_time_seconds, read_time_seconds, is_published, author_id, created_at, updated_at
`

type UpdateSectionIsPublishedParams struct {
	IsPublished bool
	ID          int32
}

func (q *Queries) UpdateSectionIsPublished(ctx context.Context, arg UpdateSectionIsPublishedParams) (Section, error) {
	row := q.db.QueryRow(ctx, updateSectionIsPublished, arg.IsPublished, arg.ID)
	var i Section
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.LanguageSlug,
		&i.SeriesSlug,
		&i.Description,
		&i.Position,
		&i.LessonsCount,
		&i.WatchTimeSeconds,
		&i.ReadTimeSeconds,
		&i.IsPublished,
		&i.AuthorID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateSectionWithPosition = `-- name: UpdateSectionWithPosition :one
UPDATE "sections" SET
  "title" = $1,
  "description" = $2,
  "position" = $3,
  "updated_at" = now()
WHERE "id" = $4
RETURNING id, title, language_slug, series_slug, description, position, lessons_count, watch_time_seconds, read_time_seconds, is_published, author_id, created_at, updated_at
`

type UpdateSectionWithPositionParams struct {
	Title       string
	Description string
	Position    int16
	ID          int32
}

func (q *Queries) UpdateSectionWithPosition(ctx context.Context, arg UpdateSectionWithPositionParams) (Section, error) {
	row := q.db.QueryRow(ctx, updateSectionWithPosition,
		arg.Title,
		arg.Description,
		arg.Position,
		arg.ID,
	)
	var i Section
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.LanguageSlug,
		&i.SeriesSlug,
		&i.Description,
		&i.Position,
		&i.LessonsCount,
		&i.WatchTimeSeconds,
		&i.ReadTimeSeconds,
		&i.IsPublished,
		&i.AuthorID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
