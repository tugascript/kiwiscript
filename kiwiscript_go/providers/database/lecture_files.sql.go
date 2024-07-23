// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: lecture_files.sql

package db

import (
	"context"

	"github.com/google/uuid"
)

const createLectureFile = `-- name: CreateLectureFile :one

INSERT INTO "lecture_files" (
    "lecture_id",
    "author_id",
    "file",
    "ext",
    "filename"
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
) RETURNING id, lecture_id, author_id, file, ext, filename, created_at, updated_at
`

type CreateLectureFileParams struct {
	LectureID int32
	AuthorID  int32
	File      uuid.UUID
	Ext       string
	Filename  string
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
func (q *Queries) CreateLectureFile(ctx context.Context, arg CreateLectureFileParams) (LectureFile, error) {
	row := q.db.QueryRow(ctx, createLectureFile,
		arg.LectureID,
		arg.AuthorID,
		arg.File,
		arg.Ext,
		arg.Filename,
	)
	var i LectureFile
	err := row.Scan(
		&i.ID,
		&i.LectureID,
		&i.AuthorID,
		&i.File,
		&i.Ext,
		&i.Filename,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteLectureFile = `-- name: DeleteLectureFile :exec
DELETE FROM "lecture_files"
WHERE "id" = $1
`

func (q *Queries) DeleteLectureFile(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, deleteLectureFile, id)
	return err
}

const findLectureFileByFileAndLectureID = `-- name: FindLectureFileByFileAndLectureID :one
SELECT id, lecture_id, author_id, file, ext, filename, created_at, updated_at FROM "lecture_files"
WHERE "file" = $1 AND "lecture_id" = $2
LIMIT 1
`

type FindLectureFileByFileAndLectureIDParams struct {
	File      uuid.UUID
	LectureID int32
}

func (q *Queries) FindLectureFileByFileAndLectureID(ctx context.Context, arg FindLectureFileByFileAndLectureIDParams) (LectureFile, error) {
	row := q.db.QueryRow(ctx, findLectureFileByFileAndLectureID, arg.File, arg.LectureID)
	var i LectureFile
	err := row.Scan(
		&i.ID,
		&i.LectureID,
		&i.AuthorID,
		&i.File,
		&i.Ext,
		&i.Filename,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const findLectureFilesByLectureID = `-- name: FindLectureFilesByLectureID :many
SELECT id, lecture_id, author_id, file, ext, filename, created_at, updated_at FROM "lecture_files"
WHERE "lecture_id" = $1
ORDER BY "id" ASC
`

func (q *Queries) FindLectureFilesByLectureID(ctx context.Context, lectureID int32) ([]LectureFile, error) {
	rows, err := q.db.Query(ctx, findLectureFilesByLectureID, lectureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []LectureFile{}
	for rows.Next() {
		var i LectureFile
		if err := rows.Scan(
			&i.ID,
			&i.LectureID,
			&i.AuthorID,
			&i.File,
			&i.Ext,
			&i.Filename,
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

const updateLectureFile = `-- name: UpdateLectureFile :one
UPDATE "lecture_files" SET
    "filename" = $1
WHERE "id" = $2
RETURNING id, lecture_id, author_id, file, ext, filename, created_at, updated_at
`

type UpdateLectureFileParams struct {
	Filename string
	ID       int32
}

func (q *Queries) UpdateLectureFile(ctx context.Context, arg UpdateLectureFileParams) (LectureFile, error) {
	row := q.db.QueryRow(ctx, updateLectureFile, arg.Filename, arg.ID)
	var i LectureFile
	err := row.Scan(
		&i.ID,
		&i.LectureID,
		&i.AuthorID,
		&i.File,
		&i.Ext,
		&i.Filename,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
