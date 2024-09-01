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
)

const sectionProgressLocation string = "section_progress"

type FindSectionProgressBySlugsAndIDOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
}

func (s *Services) FindSectionProgressBySlugsAndID(
	ctx context.Context,
	opts FindSectionProgressBySlugsAndIDOptions,
) (*db.SectionProgress, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, sectionProgressLocation, "FindSectionProgressBySlugsAndID").With(
		"userID", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"seriesPartID", opts.SectionID,
	)
	log.InfoContext(ctx, "Finding series part progress by slugs and ID...")

	seriesPartProgress, err := s.database.FindSectionProgressBySlugsAndUserID(
		ctx,
		db.FindSectionProgressBySlugsAndUserIDParams{
			UserID:       opts.UserID,
			LanguageSlug: opts.LanguageSlug,
			SeriesSlug:   opts.SeriesSlug,
			SectionID:    opts.SectionID,
		},
	)
	if err != nil {
		log.ErrorContext(ctx, "Failed to find series part progress", "error", err)
		return nil, exceptions.FromDBError(err)
	}

	return &seriesPartProgress, nil
}

type createSectionProgressOptions struct {
	RequestID          string
	UserID             int32
	LanguageProgressID int32
	SeriesProgressID   int32
	LanguageSlug       string
	SeriesSlug         string
	SectionID          int32
}

func (s *Services) createSectionProgress(
	ctx context.Context,
	opts createSectionProgressOptions,
) (*db.SectionProgress, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, sectionProgressLocation, "createSectionProgress").With(
		"userID", opts.UserID,
		"languageProgressID", opts.LanguageProgressID,
		"seriesProgressID", opts.SeriesProgressID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"seriesPartID", opts.SectionID,
	)
	log.InfoContext(ctx, "Creating series part progress...")

	seriesPartProgress, err := s.database.CreateSectionProgress(ctx, db.CreateSectionProgressParams{
		UserID:             opts.UserID,
		LanguageProgressID: opts.LanguageProgressID,
		SeriesProgressID:   opts.SeriesProgressID,
		SectionID:          opts.SectionID,
		LanguageSlug:       opts.LanguageSlug,
		SeriesSlug:         opts.SeriesSlug,
	})
	if err != nil {
		return nil, exceptions.FromDBError(err)
	}

	return &seriesPartProgress, nil
}

type CreateOrUpdateSectionProgressOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
}

func (s *Services) CreateOrUpdateSectionProgress(
	ctx context.Context,
	opts CreateOrUpdateSectionProgressOptions,
) (*db.Section, *db.SectionProgress, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, sectionProgressLocation, "CreateOrUpdateSectionProgress").With(
		"userID", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"seriesPartID", opts.SectionID,
	)
	log.InfoContext(ctx, "Creating or updating series part progress...")

	seriesPart, serviceErr := s.FindPublishedSectionBySlugsAndID(ctx, FindSectionBySlugsAndIDOptions{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SectionID:    opts.SectionID,
	})
	if serviceErr != nil {
		log.InfoContext(ctx, "Series part not found")
		return nil, nil, serviceErr
	}

	seriesProgress, serviceErr := s.FindSeriesProgress(ctx, FindSeriesProgressOptions{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
	})
	if serviceErr != nil {
		log.InfoContext(ctx, "Series progress not found")
		return nil, nil, serviceErr
	}

	seriesPartProgress, serviceErr := s.FindSectionProgressBySlugsAndID(
		ctx,
		FindSectionProgressBySlugsAndIDOptions{
			UserID:       opts.UserID,
			LanguageSlug: opts.LanguageSlug,
			SeriesSlug:   opts.SeriesSlug,
			SectionID:    opts.SectionID,
		},
	)
	if serviceErr != nil {
		seriesPartProgress, serviceErr := s.createSectionProgress(ctx, createSectionProgressOptions{
			RequestID:          opts.RequestID,
			UserID:             opts.UserID,
			LanguageProgressID: seriesProgress.LanguageProgressID,
			SeriesProgressID:   seriesProgress.ID,
			LanguageSlug:       opts.LanguageSlug,
			SeriesSlug:         opts.SeriesSlug,
			SectionID:          opts.SectionID,
		})
		if serviceErr != nil {
			return nil, nil, serviceErr
		}

		return seriesPart, seriesPartProgress, nil
	}

	if err := s.database.UpdateSectionProgressViewedAt(ctx, seriesPartProgress.ID); err != nil {
		log.ErrorContext(ctx, "Failed to update series part progress viewed at", "error", err)
		return nil, nil, exceptions.FromDBError(err)
	}

	log.InfoContext(ctx, "Series part progress updated")
	return seriesPart, seriesPartProgress, nil
}

type DeleteSectionProgressOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
}

func (s *Services) DeleteSectionProgress(
	ctx context.Context,
	opts DeleteSectionProgressOptions,
) *exceptions.ServiceError {
	log := s.buildLogger(opts.RequestID, sectionProgressLocation, "DeleteSectionProgress").With(
		"userID", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"seriesPartID", opts.SectionID,
	)
	log.InfoContext(ctx, "Deleting series part progress")

	seriesPartProgress, serviceErr := s.FindSectionProgressBySlugsAndID(
		ctx,
		FindSectionProgressBySlugsAndIDOptions{
			UserID:       opts.UserID,
			LanguageSlug: opts.LanguageSlug,
			SeriesSlug:   opts.SeriesSlug,
			SectionID:    opts.SectionID,
		},
	)
	if serviceErr != nil {
		log.InfoContext(ctx, "Series part progress not found")
		return serviceErr
	}

	if seriesPartProgress.CompletedLessons > 0 {
		qrs, txn, err := s.database.BeginTx(ctx)
		if err != nil {
			log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
			return exceptions.FromDBError(err)
		}
		defer s.database.FinalizeTx(ctx, txn, err, serviceErr)

		if err := qrs.DeleteSectionProgress(ctx, seriesPartProgress.ID); err != nil {
			log.ErrorContext(ctx, "Failed to delete series part progress", "error", err)
			return exceptions.FromDBError(err)
		}

		if seriesPartProgress.CompletedAt.Valid {
			params := db.DecrementSeriesProgressCompletedSectionsParams{
				CompletedLessons: seriesPartProgress.CompletedLessons,
				ID:               seriesPartProgress.ID,
			}
			if err := qrs.DecrementSeriesProgressCompletedSections(ctx, params); err != nil {
				log.ErrorContext(ctx, "Failed to decrement series progress completed sections", "error", err)
				return exceptions.FromDBError(err)
			}

			return nil
		}

		params := db.RemoveSeriesProgressCompletedLessonsParams{
			CompletedLessons: seriesPartProgress.CompletedLessons,
			ID:               seriesPartProgress.ID,
		}
		if err := qrs.RemoveSeriesProgressCompletedLessons(ctx, params); err != nil {
			log.ErrorContext(ctx, "Failed to remove series progress completed lessons", "error", err)
			return exceptions.FromDBError(err)
		}

		return nil
	}

	if err := s.database.DeleteSectionProgress(ctx, seriesPartProgress.ID); err != nil {
		log.ErrorContext(ctx, "Failed to delete series part progress", "error", err)
		return exceptions.FromDBError(err)
	}

	return nil
}
