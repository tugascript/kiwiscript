// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: lesson_files.sql

package db

import (
	"context"

	"github.com/google/uuid"
)

const createLessonFile = `-- name: CreateLessonFile :one

INSERT INTO "lesson_files" (
    "id",
    "lesson_id",
    "author_id",
    "ext",
    "name"
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
) RETURNING id, lesson_id, author_id, ext, name, created_at, updated_at
`

type CreateLessonFileParams struct {
	ID       uuid.UUID
	LessonID int32
	AuthorID int32
	Ext      string
	Name     string
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
func (q *Queries) CreateLessonFile(ctx context.Context, arg CreateLessonFileParams) (LessonFile, error) {
	row := q.db.QueryRow(ctx, createLessonFile,
		arg.ID,
		arg.LessonID,
		arg.AuthorID,
		arg.Ext,
		arg.Name,
	)
	var i LessonFile
	err := row.Scan(
		&i.ID,
		&i.LessonID,
		&i.AuthorID,
		&i.Ext,
		&i.Name,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteLessonFile = `-- name: DeleteLessonFile :exec
DELETE FROM "lesson_files"
WHERE "id" = $1
`

func (q *Queries) DeleteLessonFile(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, deleteLessonFile, id)
	return err
}

const findLessonFileByIDAndLessonID = `-- name: FindLessonFileByIDAndLessonID :one
SELECT id, lesson_id, author_id, ext, name, created_at, updated_at FROM "lesson_files"
WHERE "id" = $1 AND "lesson_id" = $2
LIMIT 1
`

type FindLessonFileByIDAndLessonIDParams struct {
	ID       uuid.UUID
	LessonID int32
}

func (q *Queries) FindLessonFileByIDAndLessonID(ctx context.Context, arg FindLessonFileByIDAndLessonIDParams) (LessonFile, error) {
	row := q.db.QueryRow(ctx, findLessonFileByIDAndLessonID, arg.ID, arg.LessonID)
	var i LessonFile
	err := row.Scan(
		&i.ID,
		&i.LessonID,
		&i.AuthorID,
		&i.Ext,
		&i.Name,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const findLessonFilesByLessonID = `-- name: FindLessonFilesByLessonID :many
SELECT id, lesson_id, author_id, ext, name, created_at, updated_at FROM "lesson_files"
WHERE "lesson_id" = $1
ORDER BY "created_at" ASC
`

func (q *Queries) FindLessonFilesByLessonID(ctx context.Context, lessonID int32) ([]LessonFile, error) {
	rows, err := q.db.Query(ctx, findLessonFilesByLessonID, lessonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []LessonFile{}
	for rows.Next() {
		var i LessonFile
		if err := rows.Scan(
			&i.ID,
			&i.LessonID,
			&i.AuthorID,
			&i.Ext,
			&i.Name,
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

const updateLessonFile = `-- name: UpdateLessonFile :one
UPDATE "lesson_files" SET
    "name" = $1
WHERE "id" = $2
RETURNING id, lesson_id, author_id, ext, name, created_at, updated_at
`

type UpdateLessonFileParams struct {
	Name string
	ID   uuid.UUID
}

func (q *Queries) UpdateLessonFile(ctx context.Context, arg UpdateLessonFileParams) (LessonFile, error) {
	row := q.db.QueryRow(ctx, updateLessonFile, arg.Name, arg.ID)
	var i LessonFile
	err := row.Scan(
		&i.ID,
		&i.LessonID,
		&i.AuthorID,
		&i.Ext,
		&i.Name,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
