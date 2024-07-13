// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: series.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const addSeriesPartsCount = `-- name: AddSeriesPartsCount :exec
UPDATE "series" SET
  "series_parts_count" = "series_parts_count" + 1,
  "lectures_count" = "lectures_count" + $2
WHERE "id" = $1
`

type AddSeriesPartsCountParams struct {
	ID            int32
	LecturesCount int16
}

func (q *Queries) AddSeriesPartsCount(ctx context.Context, arg AddSeriesPartsCountParams) error {
	_, err := q.db.Exec(ctx, addSeriesPartsCount, arg.ID, arg.LecturesCount)
	return err
}

const createSeries = `-- name: CreateSeries :one

INSERT INTO "series" (
  "title",
  "slug",
  "description",
  "language_id",
  "author_id"
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5
) RETURNING id, title, slug, description, parts_count, lectures_count, total_duration_seconds, review_avg, review_count, is_published, language_id, author_id, created_at, updated_at
`

type CreateSeriesParams struct {
	Title       string
	Slug        string
	Description string
	LanguageID  int32
	AuthorID    int32
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
func (q *Queries) CreateSeries(ctx context.Context, arg CreateSeriesParams) (Series, error) {
	row := q.db.QueryRow(ctx, createSeries,
		arg.Title,
		arg.Slug,
		arg.Description,
		arg.LanguageID,
		arg.AuthorID,
	)
	var i Series
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Slug,
		&i.Description,
		&i.PartsCount,
		&i.LecturesCount,
		&i.TotalDurationSeconds,
		&i.ReviewAvg,
		&i.ReviewCount,
		&i.IsPublished,
		&i.LanguageID,
		&i.AuthorID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const decrementSeriesLecturesCount = `-- name: DecrementSeriesLecturesCount :exec
UPDATE "series" SET
  "lectures_count" = "lectures_count" - 1
WHERE "id" = $1
`

func (q *Queries) DecrementSeriesLecturesCount(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, decrementSeriesLecturesCount, id)
	return err
}

const decrementSeriesPartsCount = `-- name: DecrementSeriesPartsCount :exec
UPDATE "series" SET
  "series_parts_count" = "series_parts_count" - 1,
  "lectures_count" = "lectures_count" - $2
WHERE "id" = $1
`

type DecrementSeriesPartsCountParams struct {
	ID            int32
	LecturesCount int16
}

func (q *Queries) DecrementSeriesPartsCount(ctx context.Context, arg DecrementSeriesPartsCountParams) error {
	_, err := q.db.Exec(ctx, decrementSeriesPartsCount, arg.ID, arg.LecturesCount)
	return err
}

const deleteSeriesById = `-- name: DeleteSeriesById :exec
DELETE FROM "series"
WHERE "id" = $1
`

func (q *Queries) DeleteSeriesById(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, deleteSeriesById, id)
	return err
}

const findSeriesById = `-- name: FindSeriesById :one
SELECT id, title, slug, description, parts_count, lectures_count, total_duration_seconds, review_avg, review_count, is_published, language_id, author_id, created_at, updated_at FROM "series"
WHERE "id" = $1 LIMIT 1
`

func (q *Queries) FindSeriesById(ctx context.Context, id int32) (Series, error) {
	row := q.db.QueryRow(ctx, findSeriesById, id)
	var i Series
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Slug,
		&i.Description,
		&i.PartsCount,
		&i.LecturesCount,
		&i.TotalDurationSeconds,
		&i.ReviewAvg,
		&i.ReviewCount,
		&i.IsPublished,
		&i.LanguageID,
		&i.AuthorID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const findSeriesBySlugAndLanguageID = `-- name: FindSeriesBySlugAndLanguageID :one
SELECT id, title, slug, description, parts_count, lectures_count, total_duration_seconds, review_avg, review_count, is_published, language_id, author_id, created_at, updated_at FROM "series"
WHERE "slug" = $1 AND "language_id" = $2
LIMIT 1
`

type FindSeriesBySlugAndLanguageIDParams struct {
	Slug       string
	LanguageID int32
}

func (q *Queries) FindSeriesBySlugAndLanguageID(ctx context.Context, arg FindSeriesBySlugAndLanguageIDParams) (Series, error) {
	row := q.db.QueryRow(ctx, findSeriesBySlugAndLanguageID, arg.Slug, arg.LanguageID)
	var i Series
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Slug,
		&i.Description,
		&i.PartsCount,
		&i.LecturesCount,
		&i.TotalDurationSeconds,
		&i.ReviewAvg,
		&i.ReviewCount,
		&i.IsPublished,
		&i.LanguageID,
		&i.AuthorID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const findSeriesBySlugAndLanguageIDWithJoins = `-- name: FindSeriesBySlugAndLanguageIDWithJoins :many
SELECT 
  series.id, series.title, series.slug, series.description, series.parts_count, series.lectures_count, series.total_duration_seconds, series.review_avg, series.review_count, series.is_published, series.language_id, series.author_id, series.created_at, series.updated_at,
  "tags"."name" AS "tag_name",
  "users"."first_name" AS "author_first_name",
  "users"."last_name" AS "author_last_name"
FROM "series"
LEFT JOIN "users" ON "series"."author_id" = "users"."id"
LEFT JOIN "series_tags" ON "series"."id" = "series_tags"."series_id"
  LEFT JOIN "tags" ON "series_tags"."tag_id" = "tags"."id"
WHERE 
  "series"."is_published" = $1 AND 
  "series"."slug" = $2 AND
  "series"."language_id" = $3
LIMIT 1
`

type FindSeriesBySlugAndLanguageIDWithJoinsParams struct {
	IsPublished bool
	Slug        string
	LanguageID  int32
}

type FindSeriesBySlugAndLanguageIDWithJoinsRow struct {
	ID                   int32
	Title                string
	Slug                 string
	Description          string
	PartsCount           int16
	LecturesCount        int16
	TotalDurationSeconds int32
	ReviewAvg            int16
	ReviewCount          int32
	IsPublished          bool
	LanguageID           int32
	AuthorID             int32
	CreatedAt            pgtype.Timestamp
	UpdatedAt            pgtype.Timestamp
	TagName              pgtype.Text
	AuthorFirstName      pgtype.Text
	AuthorLastName       pgtype.Text
}

func (q *Queries) FindSeriesBySlugAndLanguageIDWithJoins(ctx context.Context, arg FindSeriesBySlugAndLanguageIDWithJoinsParams) ([]FindSeriesBySlugAndLanguageIDWithJoinsRow, error) {
	rows, err := q.db.Query(ctx, findSeriesBySlugAndLanguageIDWithJoins, arg.IsPublished, arg.Slug, arg.LanguageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []FindSeriesBySlugAndLanguageIDWithJoinsRow{}
	for rows.Next() {
		var i FindSeriesBySlugAndLanguageIDWithJoinsRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Slug,
			&i.Description,
			&i.PartsCount,
			&i.LecturesCount,
			&i.TotalDurationSeconds,
			&i.ReviewAvg,
			&i.ReviewCount,
			&i.IsPublished,
			&i.LanguageID,
			&i.AuthorID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.TagName,
			&i.AuthorFirstName,
			&i.AuthorLastName,
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

const incrementSeriesLecturesCount = `-- name: IncrementSeriesLecturesCount :exec
UPDATE "series" SET
  "lectures_count" = "lectures_count" + 1
WHERE "id" = $1
`

func (q *Queries) IncrementSeriesLecturesCount(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, incrementSeriesLecturesCount, id)
	return err
}

const incrementSeriesReviewCount = `-- name: IncrementSeriesReviewCount :exec
UPDATE "series" SET
  "review_count" = "review_count" + 1
WHERE "id" = $1
`

func (q *Queries) IncrementSeriesReviewCount(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, incrementSeriesReviewCount, id)
	return err
}

const updateSeries = `-- name: UpdateSeries :one
UPDATE "series" SET
  "title" = $1,
  "slug" = $2,
  "description" = $3
WHERE "id" = $4
RETURNING id, title, slug, description, parts_count, lectures_count, total_duration_seconds, review_avg, review_count, is_published, language_id, author_id, created_at, updated_at
`

type UpdateSeriesParams struct {
	Title       string
	Slug        string
	Description string
	ID          int32
}

func (q *Queries) UpdateSeries(ctx context.Context, arg UpdateSeriesParams) (Series, error) {
	row := q.db.QueryRow(ctx, updateSeries,
		arg.Title,
		arg.Slug,
		arg.Description,
		arg.ID,
	)
	var i Series
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Slug,
		&i.Description,
		&i.PartsCount,
		&i.LecturesCount,
		&i.TotalDurationSeconds,
		&i.ReviewAvg,
		&i.ReviewCount,
		&i.IsPublished,
		&i.LanguageID,
		&i.AuthorID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateSeriesIsPublished = `-- name: UpdateSeriesIsPublished :one
UPDATE "series" SET
  "is_published" = $1
WHERE "id" = $2
RETURNING id, title, slug, description, parts_count, lectures_count, total_duration_seconds, review_avg, review_count, is_published, language_id, author_id, created_at, updated_at
`

type UpdateSeriesIsPublishedParams struct {
	IsPublished bool
	ID          int32
}

func (q *Queries) UpdateSeriesIsPublished(ctx context.Context, arg UpdateSeriesIsPublishedParams) (Series, error) {
	row := q.db.QueryRow(ctx, updateSeriesIsPublished, arg.IsPublished, arg.ID)
	var i Series
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Slug,
		&i.Description,
		&i.PartsCount,
		&i.LecturesCount,
		&i.TotalDurationSeconds,
		&i.ReviewAvg,
		&i.ReviewCount,
		&i.IsPublished,
		&i.LanguageID,
		&i.AuthorID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateSeriesReviewAvg = `-- name: UpdateSeriesReviewAvg :exec
UPDATE "series" SET
  "review_avg" = $1
WHERE "id" = $2
`

type UpdateSeriesReviewAvgParams struct {
	ReviewAvg int16
	ID        int32
}

func (q *Queries) UpdateSeriesReviewAvg(ctx context.Context, arg UpdateSeriesReviewAvgParams) error {
	_, err := q.db.Exec(ctx, updateSeriesReviewAvg, arg.ReviewAvg, arg.ID)
	return err
}
