// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: lesson_articles.sql

package db

import (
	"context"
)

const createLessonArticle = `-- name: CreateLessonArticle :one

INSERT INTO "lesson_articles" (
    "lesson_id",
    "author_id",
    "content",
    "read_time_seconds"
) VALUES (
    $1,
    $2,
    $3,
    $4
) RETURNING id, lesson_id, author_id, content, read_time_seconds, created_at, updated_at
`

type CreateLessonArticleParams struct {
	LessonID        int32
	AuthorID        int32
	Content         string
	ReadTimeSeconds int32
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
func (q *Queries) CreateLessonArticle(ctx context.Context, arg CreateLessonArticleParams) (LessonArticle, error) {
	row := q.db.QueryRow(ctx, createLessonArticle,
		arg.LessonID,
		arg.AuthorID,
		arg.Content,
		arg.ReadTimeSeconds,
	)
	var i LessonArticle
	err := row.Scan(
		&i.ID,
		&i.LessonID,
		&i.AuthorID,
		&i.Content,
		&i.ReadTimeSeconds,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteLessonArticle = `-- name: DeleteLessonArticle :exec
DELETE FROM "lesson_articles"
WHERE "id" = $1
`

func (q *Queries) DeleteLessonArticle(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, deleteLessonArticle, id)
	return err
}

const getLessonArticleByLessonID = `-- name: GetLessonArticleByLessonID :one
SELECT id, lesson_id, author_id, content, read_time_seconds, created_at, updated_at FROM "lesson_articles"
WHERE "lesson_id" = $1
LIMIT 1
`

func (q *Queries) GetLessonArticleByLessonID(ctx context.Context, lessonID int32) (LessonArticle, error) {
	row := q.db.QueryRow(ctx, getLessonArticleByLessonID, lessonID)
	var i LessonArticle
	err := row.Scan(
		&i.ID,
		&i.LessonID,
		&i.AuthorID,
		&i.Content,
		&i.ReadTimeSeconds,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateLessonArticle = `-- name: UpdateLessonArticle :one
UPDATE "lesson_articles" SET
  "content" = $1,
  "read_time_seconds" = $2,
  "updated_at" = NOW()
WHERE "id" = $3
RETURNING id, lesson_id, author_id, content, read_time_seconds, created_at, updated_at
`

type UpdateLessonArticleParams struct {
	Content         string
	ReadTimeSeconds int32
	ID              int32
}

func (q *Queries) UpdateLessonArticle(ctx context.Context, arg UpdateLessonArticleParams) (LessonArticle, error) {
	row := q.db.QueryRow(ctx, updateLessonArticle, arg.Content, arg.ReadTimeSeconds, arg.ID)
	var i LessonArticle
	err := row.Scan(
		&i.ID,
		&i.LessonID,
		&i.AuthorID,
		&i.Content,
		&i.ReadTimeSeconds,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
