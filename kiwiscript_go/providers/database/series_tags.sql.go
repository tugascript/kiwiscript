// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: series_tags.sql

package db

import (
	"context"
)

const createSeriesTag = `-- name: CreateSeriesTag :exec

INSERT INTO "series_tags" (
    "series_id",
    "tag_id"
) VALUES (
    $1,
    $2
)
`

type CreateSeriesTagParams struct {
	SeriesID int32
	TagID    int32
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
func (q *Queries) CreateSeriesTag(ctx context.Context, arg CreateSeriesTagParams) error {
	_, err := q.db.Exec(ctx, createSeriesTag, arg.SeriesID, arg.TagID)
	return err
}

const deleteSeriesTagByIds = `-- name: DeleteSeriesTagByIds :exec
DELETE FROM "series_tags"
WHERE "series_id" = $1 AND "tag_id" = $2
`

type DeleteSeriesTagByIdsParams struct {
	SeriesID int32
	TagID    int32
}

func (q *Queries) DeleteSeriesTagByIds(ctx context.Context, arg DeleteSeriesTagByIdsParams) error {
	_, err := q.db.Exec(ctx, deleteSeriesTagByIds, arg.SeriesID, arg.TagID)
	return err
}