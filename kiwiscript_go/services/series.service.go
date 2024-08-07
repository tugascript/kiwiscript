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
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/utils"
	"log/slog"
)

type FindSeriesBySlugsOptions struct {
	LanguageSlug string
	SeriesSlug   string
}

func (s *Services) FindSeriesBySlugs(ctx context.Context, opts FindSeriesBySlugsOptions) (*db.Series, *ServiceError) {
	log := s.
		log.
		WithGroup("service.series.FindSeriesBySlug").
		With("languageSlug", opts.LanguageSlug, "slug", opts.SeriesSlug)
	log.InfoContext(ctx, "Getting series by slug")

	series, err := s.database.FindSeriesBySlugAndLanguageSlug(ctx, db.FindSeriesBySlugAndLanguageSlugParams{
		Slug:         opts.SeriesSlug,
		LanguageSlug: opts.LanguageSlug,
	})
	if err != nil {
		log.Warn("Error getting series by slug", "error", err)
		return nil, FromDBError(err)
	}

	log.InfoContext(ctx, "Series found")
	return &series, nil
}

func (s *Services) FindPublishedSeriesBySlugs(
	ctx context.Context,
	opts FindSeriesBySlugsOptions,
) (*db.Series, *ServiceError) {
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
		return nil, NewNotFoundError()
	}

	return series, nil
}

func (s *Services) FindPublishedSeriesBySlugsWithAuthor(
	ctx context.Context,
	opts FindSeriesBySlugsOptions,
) (*db.FindPublishedSeriesBySlugsWithAuthorRow, *ServiceError) {
	log := s.
		log.
		WithGroup("service.series.FindPublishedSeriesBySlugsWithAuthor").
		With(
			"laguageSlug", opts.LanguageSlug,
			"slug", opts.SeriesSlug,
		)
	log.InfoContext(ctx, "Getting published series by slug")

	series, err := s.database.FindPublishedSeriesBySlugsWithAuthor(ctx, db.FindPublishedSeriesBySlugsWithAuthorParams{
		Slug:         opts.SeriesSlug,
		LanguageSlug: opts.LanguageSlug,
	})
	if err != nil {
		log.Warn("Error getting series by slug", "error", err)
		return nil, FromDBError(err)
	}

	log.InfoContext(ctx, "Series found")
	return &series, nil
}

type FindSeriesBySlugsWithProgressOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
}

func (s *Services) FindPublishedSeriesBySlugsWithProgress(
	ctx context.Context,
	opts FindSeriesBySlugsWithProgressOptions,
) (*db.SeriesModel, *ServiceError) {
	log := s.log.WithGroup("services.series.FindPublishedSeriesBySlugsWithProgress").With(
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
		return nil, FromDBError(err)
	}

	log.InfoContext(ctx, "Series found")
	return series.ToSeriesModel(), nil
}

type CreateSeriesOptions struct {
	UserID       int32
	LanguageSlug string
	Title        string
	Description  string
}

func (s *Services) CreateSeries(ctx context.Context, options CreateSeriesOptions) (*db.Series, *ServiceError) {
	log := s.log.WithGroup("service.series.CreateSeries")
	log.InfoContext(ctx, "create series", "title", options.Title)

	language, serviceErr := s.FindLanguageBySlug(ctx, options.LanguageSlug)
	if serviceErr != nil {
		return nil, NewError(CodeNotFound, MessageNotFound)
	}

	slug := utils.Slugify(options.Title)
	params := db.FindSeriesBySlugAndLanguageSlugParams{
		Slug:         slug,
		LanguageSlug: options.LanguageSlug,
	}
	if _, err := s.database.FindSeriesBySlugAndLanguageSlug(ctx, params); err == nil {
		log.InfoContext(ctx, "series already exists", "slug", slug)
		return nil, NewDuplicateKeyError("Series already exists")

	}

	series, err := s.database.CreateSeries(ctx, db.CreateSeriesParams{
		Title:        options.Title,
		Description:  options.Description,
		AuthorID:     options.UserID,
		LanguageSlug: language.Slug,
		Slug:         slug,
	})
	if err != nil {
		return nil, FromDBError(err)
	}

	return &series, nil
}

type FindPaginatedSeriesOptions struct {
	LanguageSlug string
	Offset       int32
	Limit        int32
	SortBySlug   bool
}

func (s *Services) findPublishedCount(
	ctx context.Context,
	log *slog.Logger,
	languageSlug string,
) (int64, *ServiceError) {
	if _, serviceErr := s.FindLanguageBySlug(ctx, languageSlug); serviceErr != nil {
		log.WarnContext(ctx, "Language not found", "error", serviceErr)
		return 0, serviceErr
	}

	count, err := s.database.CountPublishedSeries(ctx, languageSlug)
	if err != nil {
		log.ErrorContext(ctx, "Error counting series", "error", err)
		return 0, FromDBError(err)
	}

	return count, nil
}

func (s *Services) FindPaginatedPublishedSeries(
	ctx context.Context,
	opts FindPaginatedSeriesOptions,
) ([]db.SeriesModel, int64, *ServiceError) {
	log := s.log.WithGroup("service.series.findPaginatedPublishedSeries")
	log.InfoContext(ctx, "Getting published series...")

	count, serviceErr := s.findPublishedCount(ctx, log, opts.LanguageSlug)
	if serviceErr != nil {
		return nil, 0, serviceErr
	}

	seriesDTOs := make([]db.SeriesModel, 0, count)
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
			return nil, 0, FromDBError(err)
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
		return nil, 0, FromDBError(err)
	}

	for _, row := range series {
		seriesDTOs = append(seriesDTOs, *row.ToSeriesModel())
	}

	return seriesDTOs, count, nil
}

func (s *Services) FindPaginatedPublishedSeriesWithProgress(
	ctx context.Context,
	opts FindPaginatedSeriesOptions,
) ([]db.SeriesModel, int64, *ServiceError) {
	log := s.log.WithGroup("service.series.findPaginatedPublishedSeriesWithProgress")
	log.InfoContext(ctx, "Getting published series with progress...")

	count, serviceErr := s.findPublishedCount(ctx, log, opts.LanguageSlug)
	if serviceErr != nil {
		return nil, 0, serviceErr
	}

	seriesDTOs := make([]db.SeriesModel, 0, count)
	if count == 0 {
		return seriesDTOs, 0, nil
	}

	if opts.SortBySlug {
		series, err := s.database.FindPaginatedPublishedSeriesWithAuthorAndProgressSortBySlug(
			ctx,
			db.FindPaginatedPublishedSeriesWithAuthorAndProgressSortBySlugParams{
				LanguageSlug: opts.LanguageSlug,
				Limit:        opts.Limit,
				Offset:       opts.Offset,
			},
		)
		if err != nil {
			log.ErrorContext(ctx, "Error getting series", "error", err)
			return nil, 0, FromDBError(err)
		}

		for _, row := range series {
			seriesDTOs = append(seriesDTOs, *row.ToSeriesModel())
		}

		return seriesDTOs, count, nil
	}

	series, err := s.database.FindPaginatedPublishedSeriesWithAuthorAndProgressSortByID(
		ctx,
		db.FindPaginatedPublishedSeriesWithAuthorAndProgressSortByIDParams{
			LanguageSlug: opts.LanguageSlug,
			Limit:        opts.Limit,
			Offset:       opts.Offset,
		},
	)
	if err != nil {
		log.ErrorContext(ctx, "Error getting series", "error", err)
		return nil, 0, FromDBError(err)
	}

	for _, row := range series {
		seriesDTOs = append(seriesDTOs, *row.ToSeriesModel())
	}

	return seriesDTOs, count, nil
}

func (s *Services) FindPaginatedSeries(
	ctx context.Context,
	opts FindPaginatedSeriesOptions,
) ([]db.SeriesModel, int64, *ServiceError) {
	log := s.log.WithGroup("service.series.GetSeries")
	log.InfoContext(ctx, "Getting series...")

	if _, serviceErr := s.FindLanguageBySlug(ctx, opts.LanguageSlug); serviceErr != nil {
		log.WarnContext(ctx, "Language not found", "error", serviceErr)
		return nil, 0, serviceErr
	}

	count, err := s.database.CountSeries(ctx, opts.LanguageSlug)
	if err != nil {
		log.ErrorContext(ctx, "Error counting series", "error", err)
		return nil, 0, FromDBError(err)
	}

	seriesDTOs := make([]db.SeriesModel, 0, count)
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
			return nil, 0, FromDBError(err)
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
		return nil, 0, FromDBError(err)
	}

	for _, row := range series {
		seriesDTOs = append(seriesDTOs, *row.ToSeriesModel())
	}

	return seriesDTOs, count, nil
}

type FindFilteredSeriesOptions struct {
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
) (int64, *ServiceError) {
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
		return 0, FromDBError(err)
	}

	return count, nil
}

func (s *Services) FindFilteredPublishedSeries(
	ctx context.Context,
	opts FindFilteredSeriesOptions,
) ([]db.SeriesModel, int64, *ServiceError) {
	log := s.log.WithGroup("service.series.FindFilteredPublishedSeries").With(
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

	seriesDTOs := make([]db.SeriesModel, 0, count)
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
			return nil, 0, FromDBError(err)
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
		return nil, 0, FromDBError(err)
	}

	for _, row := range series {
		seriesDTOs = append(seriesDTOs, *row.ToSeriesModel())
	}

	return seriesDTOs, count, nil
}

func (s *Services) FindFilteredPublishedSeriesWithProgress(
	ctx context.Context,
	opts FindFilteredSeriesOptions,
) ([]db.SeriesModel, int64, *ServiceError) {
	log := s.log.WithGroup("service.series.FindFilteredPublishedSeriesWithProgress").With(
		"search", opts.Search,
		"languageSlug", opts.LanguageSlug,
		"offset", opts.Offset,
		"limit", opts.Limit,
		"sortBySlug", opts.SortBySlug,
	)
	log.InfoContext(ctx, "Find filtered published series with progress...")

	dbSearch := utils.DbSearch(opts.Search)
	count, serviceErr := s.findFilteredPublishedCount(ctx, log, opts.LanguageSlug, dbSearch)
	if serviceErr != nil {
		return nil, 0, serviceErr
	}

	seriesDTOs := make([]db.SeriesModel, 0, count)
	if count == 0 {
		return seriesDTOs, 0, nil
	}

	if opts.SortBySlug {
		series, err := s.database.FindFilteredPublishedSeriesWithAuthorAndProgressSortBySlug(
			ctx,
			db.FindFilteredPublishedSeriesWithAuthorAndProgressSortBySlugParams{
				LanguageSlug: opts.LanguageSlug,
				Title:        dbSearch,
				Limit:        opts.Limit,
				Offset:       opts.Offset,
			},
		)
		if err != nil {
			log.ErrorContext(ctx, "Error getting series", "error", err)
			return nil, 0, FromDBError(err)
		}

		for _, row := range series {
			seriesDTOs = append(seriesDTOs, *row.ToSeriesModel())
		}

		return seriesDTOs, count, nil
	}

	series, err := s.database.FindFilteredPublishedSeriesWithAuthorAndProgressSortByID(
		ctx,
		db.FindFilteredPublishedSeriesWithAuthorAndProgressSortByIDParams{
			LanguageSlug: opts.LanguageSlug,
			Title:        dbSearch,
			Limit:        opts.Limit,
			Offset:       opts.Offset,
		},
	)
	if err != nil {
		log.ErrorContext(ctx, "Error getting series", "error", err)
		return nil, 0, FromDBError(err)
	}

	for _, row := range series {
		seriesDTOs = append(seriesDTOs, *row.ToSeriesModel())
	}

	return seriesDTOs, count, nil
}

func (s *Services) FindFilteredSeries(
	ctx context.Context,
	opts FindFilteredSeriesOptions,
) ([]db.SeriesModel, int64, *ServiceError) {
	log := s.log.WithGroup("service.series.FindFilteredSeries").With(
		"search", opts.Search,
		"languageSlug", opts.LanguageSlug,
		"offset", opts.Offset,
		"limit", opts.Limit,
		"sortBySlug", opts.SortBySlug,
	)
	log.InfoContext(ctx, "Find filtered published series...")

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
		return nil, 0, FromDBError(err)
	}

	seriesDTOs := make([]db.SeriesModel, 0, count)
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
			return nil, 0, FromDBError(err)
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
		return nil, 0, FromDBError(err)
	}

	for _, row := range series {
		seriesDTOs = append(seriesDTOs, *row.ToSeriesModel())
	}

	return seriesDTOs, count, nil
}

type AssertSeriesOwnershipOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
}

func (s *Services) AssertSeriesOwnership(ctx context.Context, opts AssertSeriesOwnershipOptions) (*db.Series, *ServiceError) {
	log := s.log.WithGroup("service.series.AssertSeriesOwnership").With("slug", opts.SeriesSlug)
	log.InfoContext(ctx, "Asserting series ownership...")

	series, serviceErr := s.FindSeriesBySlugs(ctx, FindSeriesBySlugsOptions{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
	})
	if serviceErr != nil {
		log.Warn("Series not found", "error", serviceErr)
		return nil, FromDBError(serviceErr)
	}
	if series.AuthorID != opts.UserID {
		log.Warn("User is not the author of the series")
		return nil, NewForbiddenError()
	}

	log.InfoContext(ctx, "Series ownership asserted")
	return series, nil
}

type UpdateSeriesOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	Title        string
	Description  string
}

func (s *Services) UpdateSeries(ctx context.Context, opts UpdateSeriesOptions) (*db.Series, *ServiceError) {
	log := s.log.WithGroup("service.series.UpdateSeries").With("slug", opts.SeriesSlug)
	log.InfoContext(ctx, "Updating series")

	series, serviceErr := s.AssertSeriesOwnership(ctx, AssertSeriesOwnershipOptions{
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
		return nil, FromDBError(err)
	}

	log.InfoContext(ctx, "Series updated")
	return series, nil
}

type DeleteSeriesOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
}

func (s *Services) DeleteSeries(ctx context.Context, opts DeleteSeriesOptions) *ServiceError {
	log := s.log.WithGroup("service.series.DeleteSeries").With("slug", opts.SeriesSlug)
	log.InfoContext(ctx, "Deleting series")

	series, serviceErr := s.AssertSeriesOwnership(ctx, AssertSeriesOwnershipOptions(opts))
	if serviceErr != nil {
		return serviceErr
	}

	if series.SectionsCount > 0 {
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
	if series.SectionsCount == 0 && options.IsPublished {
		errMsg := "Series must have parts to be published"
		log.Warn("Series has no parts", "error", errMsg)
		return nil, NewValidationError(errMsg)
	}

	// TODO: add constraints to prevent unpublishing series with students

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return nil, FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err)

	*series, err = qrs.UpdateSeriesIsPublished(ctx, db.UpdateSeriesIsPublishedParams{
		ID:          series.ID,
		IsPublished: options.IsPublished,
	})
	if err != nil {
		log.ErrorContext(ctx, "Error publishing series", "error", err)
		return nil, FromDBError(err)
	}

	if options.IsPublished {
		if err := qrs.IncrementLanguageSeriesCount(ctx, options.LanguageSlug); err != nil {
			log.ErrorContext(ctx, "Error incrementing language series count", "error", err)
			return nil, FromDBError(err)
		}
	} else {
		if err := qrs.DecrementLanguageSeriesCount(ctx, options.LanguageSlug); err != nil {
			log.ErrorContext(ctx, "Error decrementing language series count", "error", err)
			return nil, FromDBError(err)
		}
	}

	log.InfoContext(ctx, "Series isPublished updated")
	return series, nil
}
