// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: user_pictures.sql

package db

import (
	"context"

	"github.com/google/uuid"
)

const createUserPicture = `-- name: CreateUserPicture :one

INSERT INTO "user_pictures" (
    "id",
    "user_id",
    "ext"
) VALUES (
    $1,
    $2,
    $3
) RETURNING id, user_id, ext, created_at, updated_at
`

type CreateUserPictureParams struct {
	ID     uuid.UUID
	UserID int32
	Ext    string
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
func (q *Queries) CreateUserPicture(ctx context.Context, arg CreateUserPictureParams) (UserPicture, error) {
	row := q.db.QueryRow(ctx, createUserPicture, arg.ID, arg.UserID, arg.Ext)
	var i UserPicture
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Ext,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteUserPicture = `-- name: DeleteUserPicture :exec
DELETE FROM "user_pictures"
WHERE "id" = $1
`

func (q *Queries) DeleteUserPicture(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, deleteUserPicture, id)
	return err
}

const findUserPictureByUserID = `-- name: FindUserPictureByUserID :one
SELECT id, user_id, ext, created_at, updated_at FROM "user_pictures"
WHERE "user_id" = $1
LIMIT 1
`

func (q *Queries) FindUserPictureByUserID(ctx context.Context, userID int32) (UserPicture, error) {
	row := q.db.QueryRow(ctx, findUserPictureByUserID, userID)
	var i UserPicture
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Ext,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}