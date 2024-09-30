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
	"github.com/kiwiscript/kiwiscript_go/exceptions"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/utils"
	"log/slog"
)

const seriesLocation string = "series"

type FindSeriesBySlugsOptions struct {
	RequestID    string
	LanguageSlug string
	SeriesSlug   string
}

func (s *Services) FindSeriesBySlugs(ctx context.Context, opts FindSeriesBySlugsOptions) (*db.Series, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, seriesLocation, "FindSeriesBySlugs").With(
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
	)
	log.InfoContext(ctx, "Finding series by slugs...")

	series, err := s.database.FindSeriesBySlugAndLanguageSlug(ctx, db.FindSeriesBySlugAndLanguageSlugParams{
		Slug:         opts.SeriesSlug,
		LanguageSlug: opts.LanguageSlug,
	})
	if err != nil {
		log.WarnContext(ctx, "Error getting series by slug", "error", err)
		return nil, exceptions.FromDBError(err)
	}

	log.InfoContext(ctx, "Series found")
	return &series, nil
}

func (s *Services) FindPublishedSeriesBySlugs(
	ctx context.Context,
	opts FindSeriesBySlugsOptions,
) (*db.Series, *exceptions.ServiceError) {
	log := s.
		log.
		WithGroup("service.series.FindSeriesBySlug").
		With("languageSlug", opts.LanguageSlug, "slug", opts.SeriesSlug)
	log.InfoContext(ctx, "Getting published series by slugs")

	series, serviceErr := s.FindSeriesBySlugs(ctx, opts)
	if serviceErr != nil {
		return nil, serviceErr
	}
	if !series.IsPublished {
		log.WarnContext(ctx, "Series is not published")
		return nil, exceptions.NewNotFoundError()
	}

	return series, nil
}

func (s *Services) FindPublishedSeriesBySlugsWithAuthor(
	ctx context.Context,
	opts FindSeriesBySlugsOptions,
) (*db.FindPublishedSeriesBySlugsWithAuthorRow, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, seriesLocation, "FindPublishedSeriesBySlugsWithAuthor").With(
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
	)
	log.InfoContext(ctx, "Getting published series by slug")

	series, err := s.database.FindPublishedSeriesBySlugsWithAuthor(ctx, db.FindPublishedSeriesBySlugsWithAuthorParams{
		Slug:         opts.SeriesSlug,
		LanguageSlug: opts.LanguageSlug,
	})
	if err != nil {
		log.Warn("Error getting series by slug", "error", err)
		return nil, exceptions.FromDBError(err)
	}

	log.InfoContext(ctx, "Series found")
	return &series, nil
}

type FindSeriesBySlugsWithProgressOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
}

func (s *Services) FindPublishedSeriesBySlugsWithProgress(
	ctx context.Context,
	opts FindSeriesBySlugsWithProgressOptions,
) (*db.FindPublishedSeriesBySlugWithAuthorAndProgressRow, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, seriesLocation, "FindPublishedSeriesBySlugsWithProgress").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
	)
	log.InfoContext(ctx, "Finding published series with progress...")

	series, err := s.database.FindPublishedSeriesBySlugWithAuthorAndProgress(
		ctx,
		db.FindPublishedSeriesBySlugWithAuthorAndProgressParams{
			UserID:       opts.UserID,
			LanguageSlug: opts.LanguageSlug,
			Slug:         opts.SeriesSlug,
		},
	)
	if err != nil {
		log.Warn("Error finding series by slug", "error", err)
		return nil, exceptions.FromDBError(err)
	}

	log.InfoContext(ctx, "Series found")
	return &series, nil
}

type CreateSeriesOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	Title        string
	Description  string
}

func (s *Services) CreateSeries(ctx context.Context, options CreateSeriesOptions) (*db.Series, *exceptions.ServiceError) {
	log := s.buildLogger(options.RequestID, seriesLocation, "CreateSeries").With(
		"userId", options.UserID,
		"languageSlug", options.LanguageSlug,
		"title", options.Title,
	)
	log.InfoContext(ctx, "Creating series...")

	language, serviceErr := s.FindLanguageBySlug(ctx, options.LanguageSlug)
	if serviceErr != nil {
		return nil, serviceErr
	}

	slug := utils.Slugify(options.Title)
	params := db.FindSeriesBySlugAndLanguageSlugParams{
		Slug:         slug,
		LanguageSlug: options.LanguageSlug,
	}
	if _, err := s.database.FindSeriesBySlugAndLanguageSlug(ctx, params); err == nil {
		log.InfoContext(ctx, "series already exists", "slug", slug)
		return nil, exceptions.NewConflictError("Series already exists")

	}

	series, err := s.database.CreateSeries(ctx, db.CreateSeriesParams{
		Title:        options.Title,
		Description:  options.Description,
		AuthorID:     options.UserID,
		LanguageSlug: language.Slug,
		Slug:         slug,
	})
	if err != nil {
		return nil, exceptions.FromDBError(err)
	}

	return &series, nil
}

type FindPaginatedSeriesOptions struct {
	RequestID    string
	LanguageSlug string
	Offset       int32
	Limit        int32
	SortBySlug   bool
}

func (s *Services) findPublishedCount(
	ctx context.Context,
	log *slog.Logger,
	languageSlug string,
) (int64, *exceptions.ServiceError) {
	if _, serviceErr := s.FindLanguageBySlug(ctx, languageSlug); serviceErr != nil {
		log.WarnContext(ctx, "Language not found", "error", serviceErr)
		return 0, serviceErr
	}

	count, err := s.database.CountPublishedSeries(ctx, languageSlug)
	if err != nil {
		log.ErrorContext(ctx, "Error counting series", "error", err)
		return 0, exceptions.FromDBError(err)
	}

	return count, nil
}

func (s *Services) FindPaginatedPublishedSeries(
	ctx context.Context,
	opts FindPaginatedSeriesOptions,
) ([]db.SeriesModel, int64, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, seriesLocation, "FindPaginatedPublishedSeries").With(
		"languageSlug", opts.LanguageSlug,
		"offset", opts.Offset,
		"limit", opts.Limit,
		"sortBySlug", opts.SortBySlug,
	)
	log.InfoContext(ctx, "Getting published series...")

	count, serviceErr := s.findPublishedCount(ctx, log, opts.LanguageSlug)
	if serviceErr != nil {
		return nil, 0, serviceErr
	}

	seriesDTOs := make([]db.SeriesModel, 0)
	if count == 0 {
		return seriesDTOs, 0, nil
	}

	if opts.SortBySlug {
		series, err := s.database.FindPaginatedPublishedSeriesWithAuthorSortBySlug(
			ctx,
			db.FindPaginatedPublishedSeriesWithAuthorSortBySlugParams{
				LanguageSlug: opts.LanguageSlug,
				Limit:        opts.Limit,
				Offset:       opts.Offset,
			},
		)
		if err != nil {
			log.ErrorContext(ctx, "Error getting series", "error", err)
			return nil, 0, exceptions.FromDBError(err)
		}

		for _, row := range series {
			seriesDTOs = append(seriesDTOs, *row.ToSeriesModel())
		}

		return seriesDTOs, count, nil
	}

	series, err := s.database.FindPaginatedPublishedSeriesWithAuthorSortByID(
		ctx,
		db.FindPaginatedPublishedSeriesWithAuthorSortByIDParams{
			LanguageSlug: opts.LanguageSlug,
			Limit:        opts.Limit,
			Offset:       opts.Offset,
		},
	)
	if err != nil {
		log.ErrorContext(ctx, "Error getting series", "error", err)
		return nil, 0, exceptions.FromDBError(err)
	}

	for _, row := range series {
		seriesDTOs = append(seriesDTOs, *row.ToSeriesModel())
	}

	return seriesDTOs, count, nil
}

type FindPaginatedSeriesWithProgressOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	Offset       int32
	Limit        int32
	SortBySlug   bool
}

func (s *Services) FindPaginatedPublishedSeriesWithProgress(
	ctx context.Context,
	opts FindPaginatedSeriesWithProgressOptions,
) ([]db.SeriesModel, int64, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, seriesLocation, "FindPaginatedPublishedSeriesWithProgress").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"offset", opts.Offset,
		"limit", opts.Limit,
		"sortBySlug", opts.SortBySlug,
	)
	log.InfoContext(ctx, "Getting published series with progress...")

	count, serviceErr := s.findPublishedCount(ctx, log, opts.LanguageSlug)
	if serviceErr != nil {
		return nil, 0, serviceErr
	}

	seriesDTOs := make([]db.SeriesModel, 0)
	if count == 0 {
		return seriesDTOs, 0, nil
	}

	if opts.SortBySlug {
		series, err := s.database.FindPaginatedPublishedSeriesWithAuthorAndProgressSortBySlug(
			ctx,
			db.FindPaginatedPublishedSeriesWithAuthorAndProgressSortBySlugParams{
				UserID:       opts.UserID,
				LanguageSlug: opts.LanguageSlug,
				Limit:        opts.Limit,
				Offset:       opts.Offset,
			},
		)
		if err != nil {
			log.ErrorContext(ctx, "Error getting series", "error", err)
			return nil, 0, exceptions.FromDBError(err)
		}

		for _, row := range series {
			seriesDTOs = append(seriesDTOs, *row.ToSeriesModel())
		}

		return seriesDTOs, count, nil
	}

	series, err := s.database.FindPaginatedPublishedSeriesWithAuthorAndProgressSortByID(
		ctx,
		db.FindPaginatedPublishedSeriesWithAuthorAndProgressSortByIDParams{
			UserID:       opts.UserID,
			LanguageSlug: opts.LanguageSlug,
			Limit:        opts.Limit,
			Offset:       opts.Offset,
		},
	)
	if err != nil {
		log.ErrorContext(ctx, "Error getting series", "error", err)
		return nil, 0, exceptions.FromDBError(err)
	}

	for _, row := range series {
		seriesDTOs = append(seriesDTOs, *row.ToSeriesModel())
	}

	return seriesDTOs, count, nil
}

func (s *Services) FindPaginatedSeries(
	ctx context.Context,
	opts FindPaginatedSeriesOptions,
) ([]db.SeriesModel, int64, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, seriesLocation, "FindPaginatedSeries").With(
		"languageSlug", opts.LanguageSlug,
		"offset", opts.Offset,
		"limit", opts.Limit,
		"sortBySlug", opts.SortBySlug,
	)
	log.InfoContext(ctx, "Finding paginated series...")

	if _, serviceErr := s.FindLanguageBySlug(ctx, opts.LanguageSlug); serviceErr != nil {
		log.WarnContext(ctx, "Language not found", "error", serviceErr)
		return nil, 0, serviceErr
	}

	count, err := s.database.CountSeries(ctx, opts.LanguageSlug)
	if err != nil {
		log.ErrorContext(ctx, "Error counting series", "error", err)
		return nil, 0, exceptions.FromDBError(err)
	}

	seriesDTOs := make([]db.SeriesModel, 0)
	if count == 0 {
		return seriesDTOs, 0, nil
	}

	if opts.SortBySlug {
		series, err := s.database.FindPaginatedSeriesWithAuthorSortBySlug(ctx, db.FindPaginatedSeriesWithAuthorSortBySlugParams{
			LanguageSlug: opts.LanguageSlug,
			Limit:        opts.Limit,
			Offset:       opts.Offset,
		})
		if err != nil {
			log.ErrorContext(ctx, "Error getting series", "error", err)
			return nil, 0, exceptions.FromDBError(err)
		}

		for _, row := range series {
			seriesDTOs = append(seriesDTOs, *row.ToSeriesModel())
		}

		return seriesDTOs, count, nil
	}

	series, err := s.database.FindPaginatedSeriesWithAuthorSortByID(ctx, db.FindPaginatedSeriesWithAuthorSortByIDParams{
		LanguageSlug: opts.LanguageSlug,
		Limit:        opts.Limit,
		Offset:       opts.Offset,
	})
	if err != nil {
		log.ErrorContext(ctx, "Error getting series", "error", err)
		return nil, 0, exceptions.FromDBError(err)
	}

	for _, row := range series {
		seriesDTOs = append(seriesDTOs, *row.ToSeriesModel())
	}

	return seriesDTOs, count, nil
}

type FindFilteredSeriesOptions struct {
	RequestID    string
	Search       string
	LanguageSlug string
	Offset       int32
	Limit        int32
	SortBySlug   bool
}

func (s *Services) findFilteredPublishedCount(
	ctx context.Context,
	log *slog.Logger,
	languageSlug,
	dbSearch string,
) (int64, *exceptions.ServiceError) {
	if _, serviceErr := s.FindLanguageBySlug(ctx, languageSlug); serviceErr != nil {
		log.WarnContext(ctx, "Language not found", "error", serviceErr)
		return 0, serviceErr
	}

	count, err := s.database.CountFilteredPublishedSeries(ctx, db.CountFilteredPublishedSeriesParams{
		LanguageSlug: languageSlug,
		Title:        dbSearch,
	})
	if err != nil {
		log.ErrorContext(ctx, "Error counting series", "error", err)
		return 0, exceptions.FromDBError(err)
	}

	return count, nil
}

func (s *Services) FindFilteredPublishedSeries(
	ctx context.Context,
	opts FindFilteredSeriesOptions,
) ([]db.SeriesModel, int64, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, seriesLocation, "FindFilteredPublishedSeries").With(
		"search", opts.Search,
		"languageSlug", opts.LanguageSlug,
		"offset", opts.Offset,
		"limit", opts.Limit,
		"sortBySlug", opts.SortBySlug,
	)
	log.InfoContext(ctx, "Find filtered published series...")

	dbSearch := utils.DbSearch(opts.Search)
	count, serviceErr := s.findFilteredPublishedCount(ctx, log, opts.LanguageSlug, dbSearch)
	if serviceErr != nil {
		return nil, 0, serviceErr
	}

	seriesDTOs := make([]db.SeriesModel, 0)
	if count == 0 {
		return seriesDTOs, 0, nil
	}

	if opts.SortBySlug {
		series, err := s.database.FindFilteredPublishedSeriesWithAuthorSortBySlug(
			ctx,
			db.FindFilteredPublishedSeriesWithAuthorSortBySlugParams{
				LanguageSlug: opts.LanguageSlug,
				Title:        dbSearch,
				Limit:        opts.Limit,
				Offset:       opts.Offset,
			},
		)
		if err != nil {
			log.ErrorContext(ctx, "Error getting series", "error", err)
			return nil, 0, exceptions.FromDBError(err)
		}

		for _, row := range series {
			seriesDTOs = append(seriesDTOs, *row.ToSeriesModel())
		}

		return seriesDTOs, count, nil
	}

	series, err := s.database.FindFilteredPublishedSeriesWithAuthorSortByID(
		ctx,
		db.FindFilteredPublishedSeriesWithAuthorSortByIDParams{
			LanguageSlug: opts.LanguageSlug,
			Title:        dbSearch,
			Limit:        opts.Limit,
			Offset:       opts.Offset,
		},
	)
	if err != nil {
		log.ErrorContext(ctx, "Error getting series", "error", err)
		return nil, 0, exceptions.FromDBError(err)
	}

	for _, row := range series {
		seriesDTOs = append(seriesDTOs, *row.ToSeriesModel())
	}

	return seriesDTOs, count, nil
}

type FindFilteredPublishedSeriesWithProgressOptions struct {
	UserID       int32
	RequestID    string
	Search       string
	LanguageSlug string
	Offset       int32
	Limit        int32
	SortBySlug   bool
}

func (s *Services) FindFilteredPublishedSeriesWithProgress(
	ctx context.Context,
	opts FindFilteredPublishedSeriesWithProgressOptions,
) ([]db.SeriesModel, int64, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, seriesLocation, "FindFilteredPublishedSeriesWithProgress").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"search", opts.Search,
		"offset", opts.Offset,
		"limit", opts.Limit,
		"sortBySlug", opts.SortBySlug,
	)
	log.InfoContext(ctx, "Finding filtered published series...")

	dbSearch := utils.DbSearch(opts.Search)
	count, serviceErr := s.findFilteredPublishedCount(ctx, log, opts.LanguageSlug, dbSearch)
	if serviceErr != nil {
		return nil, 0, serviceErr
	}

	seriesDTOs := make([]db.SeriesModel, 0)
	if count == 0 {
		return seriesDTOs, 0, nil
	}

	if opts.SortBySlug {
		series, err := s.database.FindFilteredPublishedSeriesWithAuthorAndProgressSortBySlug(
			ctx,
			db.FindFilteredPublishedSeriesWithAuthorAndProgressSortBySlugParams{
				UserID:       opts.UserID,
				LanguageSlug: opts.LanguageSlug,
				Title:        dbSearch,
				Limit:        opts.Limit,
				Offset:       opts.Offset,
			},
		)
		if err != nil {
			log.ErrorContext(ctx, "Error getting series", "error", err)
			return nil, 0, exceptions.FromDBError(err)
		}

		for _, row := range series {
			seriesDTOs = append(seriesDTOs, *row.ToSeriesModel())
		}

		return seriesDTOs, count, nil
	}

	series, err := s.database.FindFilteredPublishedSeriesWithAuthorAndProgressSortByID(
		ctx,
		db.FindFilteredPublishedSeriesWithAuthorAndProgressSortByIDParams{
			UserID:       opts.UserID,
			LanguageSlug: opts.LanguageSlug,
			Title:        dbSearch,
			Limit:        opts.Limit,
			Offset:       opts.Offset,
		},
	)
	if err != nil {
		log.ErrorContext(ctx, "Error getting series", "error", err)
		return nil, 0, exceptions.FromDBError(err)
	}

	for _, row := range series {
		seriesDTOs = append(seriesDTOs, *row.ToSeriesModel())
	}

	return seriesDTOs, count, nil
}

func (s *Services) FindFilteredSeries(
	ctx context.Context,
	opts FindFilteredSeriesOptions,
) ([]db.SeriesModel, int64, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, seriesLocation, "FindFilteredSeries").With(
		"search", opts.Search,
		"languageSlug", opts.LanguageSlug,
		"offset", opts.Offset,
		"limit", opts.Limit,
		"sortBySlug", opts.SortBySlug,
	)
	log.InfoContext(ctx, "Finding filtered series...")

	if _, serviceErr := s.FindLanguageBySlug(ctx, opts.LanguageSlug); serviceErr != nil {
		log.WarnContext(ctx, "Language not found", "error", serviceErr)
		return nil, 0, serviceErr
	}

	dbSearch := utils.DbSearch(opts.Search)
	count, err := s.database.CountFilteredSeries(ctx, db.CountFilteredSeriesParams{
		LanguageSlug: opts.LanguageSlug,
		Title:        dbSearch,
	})
	if err != nil {
		log.ErrorContext(ctx, "Error counting series", "error", err)
		return nil, 0, exceptions.FromDBError(err)
	}

	seriesDTOs := make([]db.SeriesModel, 0)
	if count == 0 {
		return seriesDTOs, 0, nil
	}

	if opts.SortBySlug {
		series, err := s.database.FindFilteredSeriesWithAuthorSortBySlug(
			ctx,
			db.FindFilteredSeriesWithAuthorSortBySlugParams{
				LanguageSlug: opts.LanguageSlug,
				Title:        dbSearch,
				Limit:        opts.Limit,
				Offset:       opts.Offset,
			},
		)
		if err != nil {
			log.ErrorContext(ctx, "Error getting series", "error", err)
			return nil, 0, exceptions.FromDBError(err)
		}

		for _, row := range series {
			seriesDTOs = append(seriesDTOs, *row.ToSeriesModel())
		}

		return seriesDTOs, count, nil
	}

	series, err := s.database.FindFilteredSeriesWithAuthorSortByID(
		ctx,
		db.FindFilteredSeriesWithAuthorSortByIDParams{
			LanguageSlug: opts.LanguageSlug,
			Title:        dbSearch,
			Limit:        opts.Limit,
			Offset:       opts.Offset,
		},
	)
	if err != nil {
		log.ErrorContext(ctx, "Error getting series", "error", err)
		return nil, 0, exceptions.FromDBError(err)
	}

	for _, row := range series {
		seriesDTOs = append(seriesDTOs, *row.ToSeriesModel())
	}

	return seriesDTOs, count, nil
}

type AssertSeriesOwnershipOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
}

func (s *Services) AssertSeriesOwnership(ctx context.Context, opts AssertSeriesOwnershipOptions) (*db.Series, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, seriesLocation, "AssertSeriesOwnership").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
	)
	log.InfoContext(ctx, "Asserting series ownership...")

	series, serviceErr := s.FindSeriesBySlugs(ctx, FindSeriesBySlugsOptions{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
	})
	if serviceErr != nil {
		log.Warn("Series not found", "error", serviceErr)
		return nil, serviceErr
	}
	if series.AuthorID != opts.UserID {
		log.Warn("User is not the author of the series")
		return nil, exceptions.NewForbiddenError()
	}

	log.InfoContext(ctx, "Series ownership asserted")
	return series, nil
}

type UpdateSeriesOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	Title        string
	Description  string
}

func (s *Services) UpdateSeries(ctx context.Context, opts UpdateSeriesOptions) (*db.Series, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, seriesLocation, "UpdateSeries").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
	)
	log.InfoContext(ctx, "Updating series")

	series, serviceErr := s.AssertSeriesOwnership(ctx, AssertSeriesOwnershipOptions{
		RequestID:    opts.RequestID,
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	var err error
	*series, err = s.database.UpdateSeries(ctx, db.UpdateSeriesParams{
		ID:          series.ID,
		Title:       opts.Title,
		Slug:        utils.Slugify(opts.Title),
		Description: opts.Description,
	})
	if err != nil {
		log.ErrorContext(ctx, "Error updating series", "error", err)
		return nil, exceptions.FromDBError(err)
	}

	log.InfoContext(ctx, "Series updated")
	return series, nil
}

type DeleteSeriesOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
}

func (s *Services) DeleteSeries(ctx context.Context, opts DeleteSeriesOptions) *exceptions.ServiceError {
	log := s.buildLogger(opts.RequestID, seriesLocation, "DeleteSeries").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
	)
	log.InfoContext(ctx, "Deleting series...")

	series, serviceErr := s.AssertSeriesOwnership(ctx, AssertSeriesOwnershipOptions(opts))
	if serviceErr != nil {
		return serviceErr
	}

	if series.IsPublished {
		progressCount, err := s.database.CountSeriesProgressBySeriesSlug(ctx, series.Slug)
		if err != nil {
			log.ErrorContext(ctx, "Failed to count series progress", "error", err)
			return exceptions.FromDBError(err)
		}

		if progressCount > 0 {
			log.WarnContext(ctx, "Series is published and has progress")
			return exceptions.NewConflictError("Series has students")
		}
	}

	if err := s.database.DeleteSeriesById(ctx, series.ID); err != nil {
		log.ErrorContext(ctx, "Error deleting series", "error", err)
		return exceptions.FromDBError(err)
	}

	log.InfoContext(ctx, "Series deleted")
	return nil
}

type UpdateSeriesIsPublishedOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	IsPublished  bool
}

func (s *Services) UpdateSeriesIsPublished(
	ctx context.Context,
	opts UpdateSeriesIsPublishedOptions,
) (*db.Series, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, seriesLocation, "UpdateSeriesIsPublished").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"isPublished", opts.IsPublished,
	)
	log.InfoContext(ctx, "Updating series is published...")

	series, serviceErr := s.FindSeriesBySlugs(ctx, FindSeriesBySlugsOptions{
		RequestID:    opts.RequestID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}
	if series.AuthorID != opts.UserID {
		log.Warn("User is not the author of the series")
		return nil, exceptions.NewForbiddenError()
	}

	if series.IsPublished == opts.IsPublished {
		log.InfoContext(ctx, "Series already published")
		return series, nil
	}
	if series.SectionsCount == 0 && opts.IsPublished {
		log.WarnContext(ctx, "Series has no sections")
		return nil, exceptions.NewValidationError("Series must have sections to be published")
	}
	if series.IsPublished && !opts.IsPublished {
		progressCount, err := s.database.CountSeriesProgressBySeriesSlug(ctx, series.Slug)
		if err != nil {
			log.ErrorContext(ctx, "Failed to count series progress", "error", err)
			return nil, exceptions.FromDBError(err)
		}

		if progressCount > 0 {
			log.WarnContext(ctx, "Series is published and has progress")
			return nil, exceptions.NewConflictError("Series has students")
		}
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return nil, exceptions.FromDBError(err)
	}
	defer func() {
		log.DebugContext(ctx, "Finalizing transaction")
		s.database.FinalizeTx(ctx, txn, err, serviceErr)
	}()

	*series, err = qrs.UpdateSeriesIsPublished(ctx, db.UpdateSeriesIsPublishedParams{
		ID:          series.ID,
		IsPublished: opts.IsPublished,
	})
	if err != nil {
		log.ErrorContext(ctx, "Error publishing series", "error", err)
		serviceErr = exceptions.FromDBError(err)
		return nil, serviceErr
	}

	if opts.IsPublished {
		if err := qrs.IncrementLanguageSeriesCount(ctx, opts.LanguageSlug); err != nil {
			log.ErrorContext(ctx, "Error incrementing language series count", "error", err)
			serviceErr = exceptions.FromDBError(err)
			return nil, serviceErr
		}
	} else {
		if err := qrs.DecrementLanguageSeriesCount(ctx, opts.LanguageSlug); err != nil {
			log.ErrorContext(ctx, "Error decrementing language series count", "error", err)
			serviceErr = exceptions.FromDBError(err)
			return nil, serviceErr
		}
	}

	log.InfoContext(ctx, "Series isPublished updated")
	return series, nil
}

type FindPaginatedViewedSeriesWithProgressOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	Offset       int32
	Limit        int32
}

func (s *Services) FindPaginatedViewedSeriesWithProgress(
	ctx context.Context,
	opts FindPaginatedViewedSeriesWithProgressOptions,
) ([]db.SeriesModel, int64, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, seriesLocation, "FindPaginatedViewedSeriesWithProgress").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"offset", opts.Offset,
		"limit", opts.Limit,
	)
	log.InfoContext(ctx, "Finding paginated viewed series with progress...")

	count, err := s.database.CountPublishedSeriesWithInnerProgress(ctx, db.CountPublishedSeriesWithInnerProgressParams{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to count viewed series")
		return nil, 0, exceptions.FromDBError(err)
	}

	seriesModels := make([]db.SeriesModel, 0)
	if count == 0 {
		log.DebugContext(ctx, "No viewed series found", "count", count)
		return seriesModels, 0, nil
	}

	series, err := s.database.FindPaginatedPublishedSeriesWithAuthorAndInnerProgress(
		ctx,
		db.FindPaginatedPublishedSeriesWithAuthorAndInnerProgressParams{
			UserID:       opts.UserID,
			LanguageSlug: opts.LanguageSlug,
			Limit:        opts.Limit,
			Offset:       opts.Offset,
		},
	)
	if err != nil {
		log.ErrorContext(ctx, "Failed to find viewed series", "error", err)
		return nil, 0, exceptions.FromDBError(err)
	}

	for _, ss := range series {
		seriesModels = append(seriesModels, *ss.ToSeriesModel())
	}

	return seriesModels, count, nil
}

func (s *Services) findAllPublishedSeriesCount(ctx context.Context, log *slog.Logger) (int64, *exceptions.ServiceError) {
	count, err := s.database.CountAllPublishedSeries(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Error counting all published series", "error", err)
		return 0, exceptions.FromDBError(err)
	}

	return count, nil
}

func (s *Services) findAllFilteredPublishedSeriesCount(
	ctx context.Context,
	log *slog.Logger,
	search string,
) (int64, *exceptions.ServiceError) {
	count, err := s.database.CountAllFilteredPublishedSeries(ctx, search)
	if err != nil {
		log.ErrorContext(ctx, "Error counting all filtered published series", "error", err)
		return 0, exceptions.FromDBError(err)
	}

	return count, nil
}

type FindPaginatedPublishedSeriesWithLanguageOptions struct {
	RequestID string
	Offset    int32
	Limit     int32
}

func (s *Services) FindPaginatedPublishedSeriesWithLanguage(
	ctx context.Context,
	opts FindPaginatedPublishedSeriesWithLanguageOptions,
) ([]db.FindPaginatedPublishedSeriesWithAuthorAndLanguageRow, int64, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, seriesLocation, "FindPaginatedPublishedSeriesWithLanguage").With(
		"offset", opts.Offset,
		"limit", opts.Limit,
	)
	log.InfoContext(ctx, "Finding published series with language...")

	count, serviceErr := s.findAllPublishedSeriesCount(ctx, log)
	if serviceErr != nil {
		return nil, 0, serviceErr
	}

	if count == 0 {
		log.DebugContext(ctx, "No series found with language", "count", count)
		return make([]db.FindPaginatedPublishedSeriesWithAuthorAndLanguageRow, 0), 0, nil
	}

	series, err := s.database.FindPaginatedPublishedSeriesWithAuthorAndLanguage(
		ctx,
		db.FindPaginatedPublishedSeriesWithAuthorAndLanguageParams{
			Limit:  opts.Limit,
			Offset: opts.Offset,
		},
	)
	if err != nil {
		log.ErrorContext(ctx, "Failed to find published series with language", "error", err)
		return nil, 0, exceptions.FromDBError(err)
	}

	return series, count, nil
}

type FindFilteredPublishedSeriesWithLanguageOptions struct {
	RequestID string
	Search    string
	Offset    int32
	Limit     int32
}

func (s *Services) FindFilteredPublishedSeriesWithLanguage(
	ctx context.Context,
	opts FindFilteredPublishedSeriesWithLanguageOptions,
) ([]db.FindFilteredPublishedSeriesWithAuthorAndLanguageRow, int64, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, seriesLocation, "FindFilteredPublishedSeriesWithLanguage").With(
		"search", opts.Search,
		"offset", opts.Offset,
		"limit", opts.Limit,
	)
	log.InfoContext(ctx, "Finding filtered published series with language...")

	dbSearch := utils.DbSearch(opts.Search)
	count, serviceErr := s.findAllFilteredPublishedSeriesCount(ctx, log, opts.Search)
	if serviceErr != nil {
		return nil, 0, serviceErr
	}

	if count == 0 {
		log.DebugContext(ctx, "No series found with language", "count", count)
		return make([]db.FindFilteredPublishedSeriesWithAuthorAndLanguageRow, 0), 0, nil
	}

	series, err := s.database.FindFilteredPublishedSeriesWithAuthorAndLanguage(
		ctx,
		db.FindFilteredPublishedSeriesWithAuthorAndLanguageParams{
			Title:  dbSearch,
			Limit:  opts.Limit,
			Offset: opts.Offset,
		},
	)
	if err != nil {
		log.ErrorContext(ctx, "Failed to find published series with language", "error", err)
		return nil, 0, exceptions.FromDBError(err)
	}

	return series, count, nil
}

type FindPaginatedPublishedSeriesWithLanguageAndProgressOptions struct {
	RequestID string
	UserID    int32
	Offset    int32
	Limit     int32
}

func (s *Services) FindPaginatedPublishedSeriesWithLanguageAndProgress(
	ctx context.Context,
	opts FindPaginatedLanguagesWithProgressOptions,
) ([]db.FindPaginatedPublishedSeriesWithAuthorLanguageAndProgressRow, int64, *exceptions.ServiceError) {
	log := s.buildLogger(
		opts.RequestID,
		seriesLocation,
		"FindPaginatedPublishedSeriesWithLanguageAndProgress",
	).With("offset", opts.Offset, "limit", opts.Limit)
	log.InfoContext(ctx, "Finding published series with language and progress...")

	count, serviceErr := s.findAllPublishedSeriesCount(ctx, log)
	if serviceErr != nil {
		return nil, 0, serviceErr
	}

	if count == 0 {
		log.DebugContext(ctx, "No series found with language and progress", "count", count)
		return make([]db.FindPaginatedPublishedSeriesWithAuthorLanguageAndProgressRow, 0), 0, nil
	}

	series, err := s.database.FindPaginatedPublishedSeriesWithAuthorLanguageAndProgress(
		ctx,
		db.FindPaginatedPublishedSeriesWithAuthorLanguageAndProgressParams{
			UserID: opts.UserID,
			Limit:  opts.Limit,
			Offset: opts.Offset,
		},
	)
	if err != nil {
		log.ErrorContext(ctx, "Failed to find published series with language and progress", "error", err)
		return nil, 0, exceptions.FromDBError(err)
	}

	return series, count, nil
}

type FindFilteredPublishedSeriesWithLanguageAndProgressOptions struct {
	RequestID string
	UserID    int32
	Search    string
	Offset    int32
	Limit     int32
}

func (s *Services) FindFilteredPublishedSeriesWithLanguageAndProgress(
	ctx context.Context,
	opts FindFilteredPublishedSeriesWithLanguageAndProgressOptions,
) ([]db.FindFilteredPublishedSeriesWithAuthorLanguageAndProgressRow, int64, *exceptions.ServiceError) {
	log := s.buildLogger(
		opts.RequestID,
		seriesLocation,
		"FindFilteredPublishedSeriesWithLanguageAndProgress",
	).With("userId", opts.UserID, "offset", opts.Offset, "limit", opts.Limit)
	log.InfoContext(ctx, "Finding filtered published series with language and progress...")

	dbSearch := utils.DbSearch(opts.Search)
	count, serviceErr := s.findAllFilteredPublishedSeriesCount(ctx, log, opts.Search)
	if serviceErr != nil {
		return nil, 0, serviceErr
	}

	if count == 0 {
		log.DebugContext(ctx, "No series found with language and progress", "count", count)
		return make([]db.FindFilteredPublishedSeriesWithAuthorLanguageAndProgressRow, 0), 0, nil
	}

	series, err := s.database.FindFilteredPublishedSeriesWithAuthorLanguageAndProgress(
		ctx,
		db.FindFilteredPublishedSeriesWithAuthorLanguageAndProgressParams{
			Limit:  opts.Limit,
			Offset: opts.Offset,
			Title:  dbSearch,
			UserID: opts.UserID,
		},
	)
	if err != nil {
		log.ErrorContext(ctx, "Failed to find published series with language and progress", "error", err)
		return nil, 0, exceptions.FromDBError(err)
	}

	return series, count, nil
}
