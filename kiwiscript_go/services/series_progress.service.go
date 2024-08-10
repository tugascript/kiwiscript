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
)

type FindSeriesProgressOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
}

func (s *Services) FindSeriesProgress(ctx context.Context, opts FindSeriesProgressOptions) (*db.SeriesProgress, *ServiceError) {
	log := s.log.WithGroup("service.series.FindSeriesProgress").With(
		"userID", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
	)
	log.InfoContext(ctx, "Finding series progress")

	seriesProgress, err := s.database.FindSeriesProgressBySlugAndUserID(ctx, db.FindSeriesProgressBySlugAndUserIDParams{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
	})
	if err != nil {
		log.ErrorContext(ctx, "Error finding series progress", "error", err)
		return nil, FromDBError(err)
	}

	log.InfoContext(ctx, "Series progress found")
	return &seriesProgress, nil
}

type createSeriesProgressOptions struct {
	UserID             int32
	LanguageProgressID int32
	LanguageSlug       string
	SeriesSlug         string
}

func (s *Services) createSeriesProgress(
	ctx context.Context,
	opts createSeriesProgressOptions,
) (*db.SeriesProgress, *ServiceError) {
	log := s.log.WithGroup("service.series.createSeriesProgress").With(
		"userId", opts.UserID,
		"languageProgressId", opts.LanguageProgressID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
	)
	log.InfoContext(ctx, "Creating series progress...")

	seriesProgress, err := s.database.CreateSeriesProgress(ctx, db.CreateSeriesProgressParams{
		UserID:             opts.UserID,
		LanguageSlug:       opts.LanguageSlug,
		SeriesSlug:         opts.SeriesSlug,
		LanguageProgressID: opts.LanguageProgressID,
	})
	if err != nil {
		log.ErrorContext(ctx, "Error creating series progress", "error", err)
		return nil, FromDBError(err)
	}

	return &seriesProgress, nil
}

type CreateOrUpdateSeriesProgressOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
}

func (s *Services) CreateOrUpdateSeriesProgress(
	ctx context.Context,
	opts CreateOrUpdateSeriesProgressOptions,
) (*db.FindPublishedSeriesBySlugsWithAuthorRow, *db.SeriesProgress, *ServiceError) {
	log := s.log.WithGroup("service.series.CreateOrUpdateSeriesProgress").With(
		"userID", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
	)
	log.InfoContext(ctx, "Creating series progress")

	series, serviceErr := s.FindPublishedSeriesBySlugsWithAuthor(ctx, FindSeriesBySlugsOptions{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
	})
	if serviceErr != nil {
		log.WarnContext(ctx, "Series not found", "slug", opts.SeriesSlug)
		return nil, nil, serviceErr
	}

	languageProgress, serviceErr := s.FindLanguageProgressBySlug(ctx, FindLanguageProgressOptions{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
	})
	if serviceErr != nil {
		log.WarnContext(ctx, "Language progress not found", "slug", opts.LanguageSlug)
		return nil, nil, serviceErr
	}

	seriesProgress, serviceErr := s.FindSeriesProgress(ctx, FindSeriesProgressOptions{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
	})
	if serviceErr != nil {
		seriesProgress, serviceErr = s.createSeriesProgress(ctx, createSeriesProgressOptions{
			UserID:             opts.UserID,
			LanguageProgressID: languageProgress.ID,
			LanguageSlug:       opts.LanguageSlug,
			SeriesSlug:         opts.SeriesSlug,
		})
		if serviceErr != nil {
			log.ErrorContext(ctx, "Error creating series progress", "error", serviceErr)
			return nil, nil, serviceErr
		}

		return series, seriesProgress, nil
	}

	if err := s.database.UpdateSeriesProgressViewedAt(ctx, seriesProgress.ID); err != nil {
		log.ErrorContext(ctx, "Error updating series progress viewed at", "error", err)
		return nil, nil, FromDBError(err)
	}

	log.InfoContext(ctx, "Series progress updated")
	return series, seriesProgress, nil
}

type SetSeriesProgressIsCurrentOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
}

type DeleteSeriesProgressOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
}

func (s *Services) DeleteSeriesProgress(ctx context.Context, opts DeleteSeriesProgressOptions) *ServiceError {
	log := s.log.WithGroup("service.series.DeleteSeriesProgress").With(
		"userID", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
	)
	log.InfoContext(ctx, "Deleting series progress")

	seriesProgress, serviceError := s.FindSeriesProgress(ctx, FindSeriesProgressOptions{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
	})
	if serviceError != nil {
		log.InfoContext(ctx, "Series progress not found")
		return serviceError
	}

	if seriesProgress.CompletedAt.Valid {
		qrs, txn, err := s.database.BeginTx(ctx)
		if err != nil {
			log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
			return FromDBError(err)
		}
		defer s.database.FinalizeTx(ctx, txn, err)

		if err := qrs.DeleteSeriesProgress(ctx, seriesProgress.ID); err != nil {
			log.ErrorContext(ctx, "Error deleting series progress", "error", err)
			return FromDBError(err)
		}
		if err := qrs.DecrementLanguageProgressCompletedSeries(ctx, seriesProgress.LanguageProgressID); err != nil {
			log.ErrorContext(ctx, "Error decrementing language progress completed series", "error", err)
			return FromDBError(err)
		}

		log.InfoContext(ctx, "Series progress deleted")
		return nil
	}

	if err := s.database.DeleteSeriesProgress(ctx, seriesProgress.ID); err != nil {
		log.ErrorContext(ctx, "Error deleting series progress", "error", err)
		return FromDBError(err)
	}

	log.InfoContext(ctx, "Series progress deleted")
	return nil
}
