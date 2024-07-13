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

package services

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/utils"
)

type FindSeriesBySlugsOptions struct {
	LanguageSlug string
	SeriesSlug   string
}

func (s *Services) FindSeriesBySlugs(ctx context.Context, opts FindSeriesBySlugsOptions) (*db.Series, *ServiceError) {
	log := s.
		log.
		WithGroup("service.series.FindSeriesBySlug").
		With("laguageSlug", opts.LanguageSlug, "slug", opts.SeriesSlug)
	log.InfoContext(ctx, "Getting series by slug")

	language, serviceErr := s.FindLanguageBySlug(ctx, opts.LanguageSlug)
	if serviceErr != nil {
		log.WarnContext(ctx, "Language not found", "error", serviceErr)
		return nil, serviceErr
	}

	series, err := s.database.FindSeriesBySlugAndLanguageID(ctx, db.FindSeriesBySlugAndLanguageIDParams{
		Slug:       opts.SeriesSlug,
		LanguageID: language.ID,
	})
	if err != nil {
		log.Warn("Error getting series by slug", "error", err)
		return nil, FromDBError(err)
	}

	log.InfoContext(ctx, "Series found")
	return &series, nil
}

type findOrCreateTagOptions struct {
	Tags     []string
	SeriesID int32
	UserID   int32
}

func (s *Services) findOrCreateTags(ctx context.Context, qrs *db.Queries, opts findOrCreateTagOptions) ([]db.Tag, *ServiceError) {
	tags := make([]db.Tag, len(opts.Tags))

	for i, name := range opts.Tags {
		tag, err := qrs.FindOrCreateTag(ctx, db.FindOrCreateTagParams{
			Name:     name,
			AuthorID: opts.UserID,
		})
		if err != nil {
			return nil, FromDBError(err)
		}

		params := db.CreateSeriesTagParams{
			SeriesID: opts.SeriesID,
			TagID:    tag.ID,
		}
		if err := qrs.CreateSeriesTag(ctx, params); err != nil {
			return nil, FromDBError(err)
		}

		tags[i] = tag
	}

	return tags, nil
}

type CreateSeriesOptions struct {
	UserID       int32
	LanguageSlug string
	Title        string
	Description  string
	Tags         []string
}

func (s *Services) CreateSeries(ctx context.Context, options CreateSeriesOptions) (*db.Series, []db.Tag, *ServiceError) {
	log := s.log.WithGroup("service.series.CreateSeries")
	log.InfoContext(ctx, "create series", "title", options.Title)

	language, serviceErr := s.FindLanguageBySlug(ctx, options.LanguageSlug)
	if serviceErr != nil {
		return nil, nil, NewError(CodeNotFound, MessageNotFound)
	}

	slug := utils.Slugify(options.Title)
	params := db.FindSeriesBySlugAndLanguageIDParams{
		Slug:       slug,
		LanguageID: language.ID,
	}
	if _, err := s.database.FindSeriesBySlugAndLanguageID(ctx, params); err == nil {
		log.InfoContext(ctx, "series already exists", "slug", slug)
		return nil, nil, NewDuplicateKeyError("Series already exists")

	}

	if options.Tags == nil || len(options.Tags) == 0 {
		log.InfoContext(ctx, "create series without tags")
		series, err := s.database.CreateSeries(ctx, db.CreateSeriesParams{
			Title:       options.Title,
			Description: options.Description,
			AuthorID:    options.UserID,
			LanguageID:  language.ID,
			Slug:        slug,
		})

		if err != nil {
			return nil, nil, FromDBError(err)
		}

		return &series, nil, nil
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		return nil, nil, FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err)

	series, err := qrs.CreateSeries(ctx, db.CreateSeriesParams{
		Title:       options.Title,
		Description: options.Description,
		AuthorID:    options.UserID,
		LanguageID:  language.ID,
		Slug:        slug,
	})
	if err != nil {
		return nil, nil, FromDBError(err)
	}

	tags, serviceErr := s.findOrCreateTags(ctx, qrs, findOrCreateTagOptions{
		Tags:     options.Tags,
		SeriesID: series.ID,
		UserID:   series.AuthorID,
	})
	if serviceErr != nil {
		return nil, nil, serviceErr
	}

	return &series, tags, nil
}

type SeriesRow struct {
	ID              int32
	Title           string
	Slug            string
	Description     string
	PartsCount      int16
	LecturesCount   int16
	IsPublished     bool
	ReviewAvg       int16
	ReviewCount     int32
	AuthorID        int32
	AuthorFirstName pgtype.Text
	AuthorLastName  pgtype.Text
	TagName         pgtype.Text
}

type SeriesAuthor struct {
	ID        int32
	FirstName string
	LastName  string
}

type SeriesLanguage struct {
	Name string
	Slug string
}

type SeriesDto struct {
	ID          int32
	Title       string
	Slug        string
	Description string
	Parts       int16
	Lectures    int16
	IsPublished bool
	ReviewAvg   int16
	ReviewCount int32
	Author      SeriesAuthor
	Tags        []string
}

type seriesMapper struct {
	dto SeriesDto
	idx int
}

func mapSeriesRowsToDtos(rows []SeriesRow) []SeriesDto {
	dtos := make(map[int32]seriesMapper)

	for i, row := range rows {
		var dto SeriesDto
		idx := i - len(dtos)
		if m, ok := dtos[row.ID]; ok {
			dto = m.dto
			idx = m.idx
		} else {
			dto = SeriesDto{
				ID:          row.ID,
				Title:       row.Title,
				Slug:        row.Slug,
				Description: row.Description,
				Parts:       row.PartsCount,
				Lectures:    row.LecturesCount,
				IsPublished: row.IsPublished,
				ReviewAvg:   row.ReviewAvg,
				ReviewCount: row.ReviewCount,
				Author: SeriesAuthor{
					ID:        row.AuthorID,
					FirstName: row.AuthorFirstName.String,
					LastName:  row.AuthorLastName.String,
				},
				Tags: make([]string, 0),
			}
		}

		if row.TagName.Valid {
			dto.Tags = append(dto.Tags, row.TagName.String)
		}

		dtos[row.ID] = seriesMapper{
			dto: dto,
			idx: idx,
		}
	}

	result := make([]SeriesDto, len(dtos))

	for _, m := range dtos {
		result[m.idx] = m.dto
	}

	return result
}

type findPaginatedSeriesRowsOptions struct {
	LanguageID  int32
	IsPublished bool
	Search      string
	Tag         string
	Offset      int32
	Limit       int32
	SortBy      string
	Order       string
}

func (s *Services) findPaginatedSeriesRows(ctx context.Context, options findPaginatedSeriesRowsOptions) ([]SeriesRow, int64, *ServiceError) {
	log := s.log.WithGroup("service.series.GetSeries")
	log.InfoContext(ctx, "Getting series...")
	args := make([]interface{}, 5)
	args[0] = options.LanguageID
	idx := 1
	query := `SELECT
		"series"."id",
		"series"."title",
		"series"."slug",
		"series"."description",
		"series"."parts_count",
		"series"."lectures_count",
		"series"."is_published",
		"series"."review_avg",
		"series"."review_count",
		"series"."author_id",
		"users"."first_name" AS "author_first_name",
		"users"."last_name" AS "author_last_name",
		"tags"."name" AS "tag_name"
	FROM "series"
	LEFT JOIN "users" ON "series"."author_id" = "users"."id"
	LEFT JOIN "series_tags" ON "series"."id" = "series_tags"."series_id"
	LEFT JOIN "tags" ON "series_tags"."tag_id" = "tags"."id"
	WHERE "series"."language_id" = $1
	`
	countQueryWhere := `WHERE "series"."language_id" = $1 `
	countQueryJoin := ""

	if options.IsPublished {
		where := `AND "series"."is_published" = true `
		query += where
		countQueryWhere += where
	}
	if options.Search != "" {
		where := fmt.Sprintf(`AND "series"."title" ILIKE $%d `, idx+1)
		query += where
		countQueryWhere += where
		args[idx] = "%" + options.Search + "%"
		idx++
	}
	if options.Tag != "" {
		where := fmt.Sprintf(`AND "tags"."name" = $%d `, idx+1)
		query += where
		countQueryJoin += `
		LEFT JOIN "series_tags" ON "series"."id" = "series_tags"."series_id"
		LEFT JOIN "tags" ON "series_tags"."tag_id" = "tags"."id" 
		`
		countQueryWhere += where
		args[idx] = options.Tag
		idx++
	}

	countQuery := `SELECT COUNT("series"."id") FROM "series" ` + countQueryJoin + countQueryWhere + "LIMIT 1;"
	countRow := s.database.RawQueryRow(ctx, countQuery, args[:idx+1])
	var count int64
	if err := countRow.Scan(&count); err != nil {
		return nil, 0, FromDBError(err)
	}

	query += fmt.Sprintf(`ORDER BY "series"."%s" %s LIMIT $%d OFFSET $%d`, options.SortBy, options.Order, idx+1, idx+2)
	args[idx] = options.Limit
	args[idx+1] = options.Offset
	rows, err := s.database.RawQuery(ctx, query, args)
	if err != nil {
		log.ErrorContext(ctx, "Error getting series", "error", err)
		return nil, 0, FromDBError(err)
	}
	defer rows.Close()

	series := make([]SeriesRow, 0)
	for rows.Next() {
		var s SeriesRow
		if err := rows.Scan(
			&s.ID,
			&s.Title,
			&s.Slug,
			&s.Description,
			&s.PartsCount,
			&s.LecturesCount,
			&s.IsPublished,
			&s.ReviewAvg,
			&s.ReviewCount,
			&s.AuthorID,
			&s.AuthorFirstName,
			&s.AuthorLastName,
			&s.TagName,
		); err != nil {
			return nil, 0, FromDBError(err)
		}
		series = append(series, s)
	}
	if err := rows.Err(); err != nil {
		log.ErrorContext(ctx, "Error getting series", "error", err)
		return nil, 0, FromDBError(err)
	}

	return series, count, nil
}

type FindPaginatedSeriesOptions struct {
	LanguageSlug string
	IsPublished  bool
	Search       string
	Tag          string
	Offset       int32
	Limit        int32
	SortBy       string
	Order        string
}

func (s *Services) FindPaginatedSeries(ctx context.Context, options FindPaginatedSeriesOptions) ([]SeriesDto, int64, *ServiceError) {
	log := s.log.WithGroup("service.series.GetSeries")
	log.InfoContext(ctx, "Getting series...")

	language, serviceErr := s.FindLanguageBySlug(ctx, options.LanguageSlug)
	if serviceErr != nil {
		log.WarnContext(ctx, "Language not found", "error", serviceErr)
		return nil, 0, serviceErr
	}

	rows, count, serviceErr := s.findPaginatedSeriesRows(ctx, findPaginatedSeriesRowsOptions{
		LanguageID:  language.ID,
		IsPublished: options.IsPublished,
		Search:      options.Search,
		Tag:         options.Tag,
		Offset:      options.Offset,
		Limit:       options.Limit,
		SortBy:      options.SortBy,
		Order:       options.Order,
	})
	if serviceErr != nil {
		log.WarnContext(ctx, "Error getting series", "error", serviceErr)
		return nil, 0, serviceErr
	}

	return mapSeriesRowsToDtos(rows), count, nil
}

func mapSingleSeriesRowsToDto(rows []db.FindSeriesBySlugAndLanguageIDWithJoinsRow) *SeriesDto {
	dto := SeriesDto{
		ID:          rows[0].ID,
		Title:       rows[0].Title,
		Slug:        rows[0].Slug,
		Description: rows[0].Description,
		Parts:       rows[0].PartsCount,
		Lectures:    rows[0].LecturesCount,
		IsPublished: rows[0].IsPublished,
		ReviewAvg:   rows[0].ReviewAvg,
		ReviewCount: rows[0].ReviewCount,
		Author: SeriesAuthor{
			ID:        rows[0].AuthorID,
			FirstName: rows[0].AuthorFirstName.String,
			LastName:  rows[0].AuthorLastName.String,
		},
		Tags: make([]string, 0),
	}

	for _, row := range rows {
		if row.TagName.Valid {
			dto.Tags = append(dto.Tags, row.TagName.String)
		}
	}

	return &dto
}

type FindSeriesBySlugsWithJoinsOptions struct {
	LanguageSlug string
	SeriesSlug   string
	IsPublished  bool
}

func (s *Services) FindSeriesBySlugsWithJoins(ctx context.Context, opts FindSeriesBySlugsWithJoinsOptions) (*SeriesDto, *ServiceError) {
	log := s.log.WithGroup("service.series.FindSeriesBySlugWithJoins").With("slug", opts.SeriesSlug)
	log.InfoContext(ctx, "Getting series by slug")
	rows, err := s.database.FindSeriesBySlugAndLanguageIDWithJoins(ctx, db.FindSeriesBySlugAndLanguageIDWithJoinsParams{
		Slug:        opts.SeriesSlug,
		IsPublished: opts.IsPublished,
	})

	if err != nil {
		log.WarnContext(ctx, "Error getting series by slug", "error", err)
		return nil, FromDBError(err)
	}

	log.InfoContext(ctx, "Series found")
	return mapSingleSeriesRowsToDto(rows), nil
}

type UpdateSeriesOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	Title        string
	Description  string
	Tags         []string
}

// TODO: add tags update
func (s *Services) UpdateSeries(ctx context.Context, options UpdateSeriesOptions) (*db.Series, *ServiceError) {
	log := s.log.WithGroup("service.series.UpdateSeries").With("slug", options.SeriesSlug)
	log.InfoContext(ctx, "Updating series")

	series, serviceErr := s.FindSeriesBySlugs(ctx, FindSeriesBySlugsOptions{
		LanguageSlug: options.LanguageSlug,
		SeriesSlug:   options.SeriesSlug,
	})
	if serviceErr != nil {
		log.Warn("Series not found", "error", serviceErr)
		return nil, FromDBError(serviceErr)
	}
	if series.AuthorID != options.UserID {
		log.Warn("User is not the author of the series")
		return nil, NewForbiddenError()

	}

	updatedSeries, err := s.database.UpdateSeries(ctx, db.UpdateSeriesParams{
		ID:          series.ID,
		Title:       options.Title,
		Slug:        utils.Slugify(options.Title),
		Description: options.Description,
	})
	if err != nil {
		log.ErrorContext(ctx, "Error updating series", "error", err)
		return nil, FromDBError(err)
	}

	log.InfoContext(ctx, "Series updated")
	return &updatedSeries, nil
}

type DeleteSeriesOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
}

func (s *Services) DeleteSeries(ctx context.Context, opts DeleteSeriesOptions) *ServiceError {
	log := s.log.WithGroup("service.series.DeleteSeries").With("slug", opts.SeriesSlug)
	log.InfoContext(ctx, "Deleting series")

	series, serviceErr := s.FindSeriesBySlugs(ctx, FindSeriesBySlugsOptions{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
	})
	if serviceErr != nil {
		log.Warn("Series not found", "error", serviceErr)
		return FromDBError(serviceErr)
	}
	if series.AuthorID != opts.UserID {
		log.Warn("User is not the author of the series")
		return NewForbiddenError()
	}
	if series.PartsCount > 0 {
		log.Warn("Series has parts")
		// TODO: update this to be a constraint error
		return NewValidationError("series has parts")
	}

	if err := s.database.DeleteSeriesById(ctx, series.ID); err != nil {
		log.ErrorContext(ctx, "Error deleting series", "error", err)
		return FromDBError(err)
	}

	log.InfoContext(ctx, "Series deleted")
	return nil
}

type UpdateSeriesIsPublishedOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	IsPublished  bool
}

func (s *Services) UpdateSeriesIsPublished(ctx context.Context, options UpdateSeriesIsPublishedOptions) (*db.Series, *ServiceError) {
	log := s.log.WithGroup("service.series.UpdateSeriesIsPublished").With("slug", options.SeriesSlug, "isPublished", options.IsPublished)
	log.InfoContext(ctx, "Updating series is published...")

	series, serviceErr := s.FindSeriesBySlugs(ctx, FindSeriesBySlugsOptions{
		LanguageSlug: options.LanguageSlug,
		SeriesSlug:   options.SeriesSlug,
	})
	if serviceErr != nil {
		log.Warn("Series not found", "error", serviceErr)
		return nil, FromDBError(serviceErr)
	}
	if series.AuthorID != options.UserID {
		log.Warn("User is not the author of the series")
		return nil, NewForbiddenError()
	}

	if series.IsPublished == options.IsPublished {
		log.InfoContext(ctx, "Series already published")
		return series, nil
	}
	if series.PartsCount == 0 && options.IsPublished {
		errMsg := "Series must have parts to be published"
		log.Warn("Series has no parts", "error", errMsg)
		return nil, NewValidationError(errMsg)
	}

	// TODO: add constraints to prevent unpublishing series with students

	updateSeries, err := s.database.UpdateSeriesIsPublished(ctx, db.UpdateSeriesIsPublishedParams{
		ID:          series.ID,
		IsPublished: options.IsPublished,
	})
	if err != nil {
		log.ErrorContext(ctx, "Error publishing series", "error", err)
		return nil, FromDBError(err)
	}

	log.InfoContext(ctx, "Series isPublished updated")
	return &updateSeries, nil
}
