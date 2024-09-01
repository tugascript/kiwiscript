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
	"log/slog"

	db "github.com/kiwiscript/kiwiscript_go/providers/database"
)

const sectionsLocation string = "sections"

type CreateSectionOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	Title        string
	Description  string
}

func (s *Services) CreateSection(ctx context.Context, opts CreateSectionOptions) (*db.Section, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, sectionsLocation, "CreateSection").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"title", opts.Title,
	)
	log.InfoContext(ctx, "Creating series part...")

	series, serviceErr := s.AssertSeriesOwnership(ctx, AssertSeriesOwnershipOptions{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
	})
	if serviceErr != nil {
		log.WarnContext(ctx, "Series not found", "error", serviceErr)
		return nil, serviceErr
	}

	section, err := s.database.CreateSection(ctx, db.CreateSectionParams{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   series.Slug,
		Title:        opts.Title,
		AuthorID:     opts.UserID,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to create series part", "error", err)
		return nil, exceptions.FromDBError(err)
	}

	log.InfoContext(ctx, "Series part created", "id", section.ID)
	return &section, nil
}

type FindSectionBySlugsAndIDOptions struct {
	RequestID    string
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
}

func (s *Services) FindSectionBySlugsAndID(
	ctx context.Context,
	opts FindSectionBySlugsAndIDOptions,
) (*db.Section, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, sectionsLocation, "FindSectionBySlugsAndID").With(
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"sectionId", opts.SectionID,
	)
	log.InfoContext(ctx, "Finding section...")

	section, err := s.database.FindSectionBySlugsAndID(ctx, db.FindSectionBySlugsAndIDParams{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		ID:           opts.SectionID,
	})
	if err != nil {
		log.WarnContext(ctx, "Failed to find series part", "error", err)
		return nil, exceptions.FromDBError(err)
	}

	return &section, nil
}

func (s *Services) FindPublishedSectionBySlugsAndID(
	ctx context.Context,
	opts FindSectionBySlugsAndIDOptions,
) (*db.Section, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, sectionsLocation, "FindPublishedSectionBySlugsAndID").With(
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"sectionId", opts.SectionID,
	)
	log.InfoContext(ctx, "Finding published series part...")

	seriesOpts := FindSeriesBySlugsOptions{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
	}
	if _, serviceErr := s.FindPublishedSeriesBySlugs(ctx, seriesOpts); serviceErr != nil {
		return nil, serviceErr
	}

	section, serviceErr := s.FindSectionBySlugsAndID(ctx, opts)
	if serviceErr != nil {
		return nil, serviceErr
	}

	if !section.IsPublished {
		log.WarnContext(ctx, "Section is not published")
		return nil, exceptions.NewNotFoundError()
	}

	return section, nil
}

type FindPublishedSectionBySlugsAndIDWithProgressOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
}

func (s *Services) FindPublishedSectionBySlugsAndIDWithProgress(
	ctx context.Context,
	opts FindPublishedSectionBySlugsAndIDWithProgressOptions,
) (*db.FindPublishedSectionBySlugsAndIDWithProgressRow, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, sectionsLocation, "FindPublishedSectionBySlugsAndIDWithProgress").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"sectionId", opts.SectionID,
	)
	log.InfoContext(ctx, "Finding published series part with progress...")

	seriesOpts := FindSeriesBySlugsOptions{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
	}
	if _, serviceErr := s.FindPublishedSeriesBySlugs(ctx, seriesOpts); serviceErr != nil {
		log.WarnContext(ctx, "Published series not found")
		return nil, serviceErr
	}

	section, err := s.database.FindPublishedSectionBySlugsAndIDWithProgress(
		ctx,
		db.FindPublishedSectionBySlugsAndIDWithProgressParams{
			UserID:       opts.UserID,
			LanguageSlug: opts.LanguageSlug,
			SeriesSlug:   opts.SeriesSlug,
			ID:           opts.SectionID,
		},
	)
	if err != nil {
		log.WarnContext(ctx, "Published series part not found")
		return nil, exceptions.FromDBError(err)
	}

	return &section, nil
}

type FindPaginatedSectionsBySlugsOptions struct {
	RequestID    string
	LanguageSlug string
	SeriesSlug   string
	Limit        int32
	Offset       int32
}

func (s *Services) FindPaginatedSectionsBySlugs(
	ctx context.Context,
	opts FindPaginatedSectionsBySlugsOptions,
) ([]db.Section, int64, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, sectionsLocation, "FindPaginatedSectionsBySlugs").With(
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
	)
	log.InfoContext(ctx, "Finding paginated series parts...")

	seriesOpts := FindSeriesBySlugsOptions{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
	}
	if _, serviceErr := s.FindSeriesBySlugs(ctx, seriesOpts); serviceErr != nil {
		log.WarnContext(ctx, "Series not found", "error", serviceErr)
		return nil, 0, serviceErr
	}

	count, err := s.database.CountSectionsBySeriesSlug(ctx, opts.SeriesSlug)
	if err != nil {
		log.ErrorContext(ctx, "Error counting series parts", "error", err)
		return nil, 0, exceptions.FromDBError(err)
	}
	if count == 0 {
		return make([]db.Section, 0), 0, nil
	}

	sections, err := s.database.FindPaginatedSectionsBySlugs(ctx, db.FindPaginatedSectionsBySlugsParams{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		Limit:        opts.Limit,
		Offset:       opts.Offset,
	})
	if err != nil {
		log.ErrorContext(ctx, "Error finding series parts", "error", err)
		return nil, 0, exceptions.FromDBError(err)
	}

	log.InfoContext(ctx, "Series parts found")
	return sections, count, nil
}

func (s *Services) findPublishedSectionsCount(
	ctx context.Context,
	log *slog.Logger,
	languageSlug,
	seriesSlug string,
) (int64, *exceptions.ServiceError) {
	seriesOpts := FindSeriesBySlugsOptions{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
	}
	if _, serviceErr := s.FindPublishedSeriesBySlugs(ctx, seriesOpts); serviceErr != nil {
		log.WarnContext(ctx, "Series not found", "error", serviceErr)
		return 0, serviceErr
	}

	count, err := s.database.CountPublishedSectionsBySeriesSlug(ctx, seriesSlug)
	if err != nil {
		log.ErrorContext(ctx, "Error counting published sections", "error", err)
		return 0, exceptions.FromDBError(err)
	}

	return count, nil
}

func (s *Services) FindPaginatedPublishedSectionsBySlugs(
	ctx context.Context,
	opts FindPaginatedSectionsBySlugsOptions,
) ([]db.Section, int64, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, sectionsLocation, "FindPaginatedPublishedSectionsBySlugs").With(
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
	)
	log.InfoContext(ctx, "Finding paginated series parts...")

	count, serviceErr := s.findPublishedSectionsCount(ctx, log, opts.LanguageSlug, opts.SeriesSlug)
	if serviceErr != nil {
		return nil, 0, serviceErr
	}
	if count == 0 {
		return make([]db.Section, 0), 0, nil
	}

	sections, err := s.database.FindPaginatedPublishedSectionsBySlugs(
		ctx,
		db.FindPaginatedPublishedSectionsBySlugsParams{
			LanguageSlug: opts.LanguageSlug,
			SeriesSlug:   opts.SeriesSlug,
			Limit:        opts.Limit,
			Offset:       opts.Offset,
		},
	)
	if err != nil {
		log.ErrorContext(ctx, "Error finding series parts", "error", err)
		return nil, 0, exceptions.FromDBError(err)
	}

	return sections, count, nil
}

type FindSectionBySlugsAndIDWithProgressOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	Limit        int32
	Offset       int32
}

func (s *Services) FindPaginatedPublishedSectionsBySlugsWithProgress(
	ctx context.Context,
	opts FindSectionBySlugsAndIDWithProgressOptions,
) ([]db.FindPaginatedPublishedSectionsBySlugsWithProgressRow, int64, *exceptions.ServiceError) {
	log := s.buildLogger(
		opts.RequestID,
		sectionsLocation,
		"FindPaginatedPublishedSectionsBySlugsWithProgress",
	).With(
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
	)
	log.InfoContext(ctx, "Finding paginated series parts with progress...")

	count, serviceErr := s.findPublishedSectionsCount(ctx, log, opts.LanguageSlug, opts.SeriesSlug)
	if serviceErr != nil {
		return nil, 0, serviceErr
	}
	if count == 0 {
		return make([]db.FindPaginatedPublishedSectionsBySlugsWithProgressRow, 0), 0, nil
	}

	sections, err := s.database.FindPaginatedPublishedSectionsBySlugsWithProgress(
		ctx, db.FindPaginatedPublishedSectionsBySlugsWithProgressParams{
			UserID:       opts.UserID,
			LanguageSlug: opts.LanguageSlug,
			SeriesSlug:   opts.SeriesSlug,
			Limit:        opts.Limit,
			Offset:       opts.Offset,
		},
	)
	if err != nil {
		log.ErrorContext(ctx, "Error finding series parts with progress", "error", serviceErr)
		return nil, 0, exceptions.FromDBError(err)
	}

	return sections, count, nil
}

type AssertSectionOwnershipOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
}

func (s *Services) AssertSectionOwnership(ctx context.Context, opts AssertSectionOwnershipOptions) (*db.Section, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, sectionsLocation, "AssertSectionOwnership").With(
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"sectionId", opts.SectionID,
	)
	log.InfoContext(ctx, "Asserting section ownership...")

	section, serviceErr := s.FindSectionBySlugsAndID(ctx, FindSectionBySlugsAndIDOptions{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SectionID:    opts.SectionID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	if section.AuthorID != opts.UserID {
		log.WarnContext(ctx, "User is not the author of the series", "user_id", opts.UserID)
		return nil, exceptions.NewForbiddenError()
	}

	log.InfoContext(ctx, "Series part ownership asserted")
	return section, nil
}

type UpdateSectionOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	Title        string
	Description  string
	Position     int16
}

func (s *Services) UpdateSection(ctx context.Context, opts UpdateSectionOptions) (*db.Section, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, sectionsLocation, "UpdateSection").With(
		"seriesSlug", opts.SeriesSlug,
		"sectionId", opts.SectionID,
		"title", opts.Title,
		"position", opts.Position,
	)
	log.InfoContext(ctx, "Updating section...")

	section, serviceErr := s.AssertSectionOwnership(ctx, AssertSectionOwnershipOptions{
		RequestID:    opts.RequestID,
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SectionID:    opts.SectionID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	if opts.Position == 0 || opts.Position == section.Position {
		section, err := s.database.UpdateSection(ctx, db.UpdateSectionParams{
			ID:          section.ID,
			Title:       opts.Title,
			Description: opts.Description,
		})

		if err != nil {
			log.ErrorContext(ctx, "Failed to update series part", "error", err)
			return nil, exceptions.FromDBError(err)
		}

		log.InfoContext(ctx, "Series part updated", "id", section.ID)
		return &section, nil
	}

	count, err := s.database.CountSectionsBySeriesSlug(ctx, opts.SeriesSlug)
	if err != nil {
		log.ErrorContext(ctx, "Failed to count series parts", "error", err)
		return nil, exceptions.FromDBError(err)
	}
	if int64(opts.Position) > count {
		log.WarnContext(ctx, "Position is out of range", "position", opts.Position, "count", count)
		return nil, exceptions.NewValidationError("Position is out of range")
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return nil, exceptions.FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err, serviceErr)

	oldPosition := section.Position
	if oldPosition < opts.Position {
		params := db.DecrementSectionPositionParams{
			SeriesSlug: opts.SeriesSlug,
			Position:   oldPosition,
			Position_2: opts.Position,
		}
		if err := qrs.DecrementSectionPosition(ctx, params); err != nil {
			log.ErrorContext(ctx, "Failed to decrement series part position", "error", err)
			return nil, exceptions.FromDBError(err)
		}
	} else {
		params := db.IncrementSectionPositionParams{
			SeriesSlug: opts.SeriesSlug,
			Position:   oldPosition,
			Position_2: opts.Position,
		}
		if err := qrs.IncrementSectionPosition(ctx, params); err != nil {
			log.ErrorContext(ctx, "Failed to increment series part position", "error", err)
			return nil, exceptions.FromDBError(err)
		}
	}

	*section, err = qrs.UpdateSectionWithPosition(ctx, db.UpdateSectionWithPositionParams{
		ID:          section.ID,
		Title:       opts.Title,
		Description: opts.Description,
		Position:    opts.Position,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to update series part", "error", err)
		return nil, exceptions.FromDBError(err)
	}

	log.InfoContext(ctx, "Series part updated", "id", section.ID)
	return section, nil
}

type UpdateSectionIsPublishedOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	IsPublished  bool
}

func (s *Services) UpdateSectionIsPublished(ctx context.Context, opts UpdateSectionIsPublishedOptions) (*db.Section, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, sectionsLocation, "UpdateSectionIsPublished").With(
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"sectionId", opts.SectionID,
		"isPublished", opts.IsPublished,
	)
	log.InfoContext(ctx, "Updating series part is published...")

	section, serviceErr := s.AssertSectionOwnership(ctx, AssertSectionOwnershipOptions{
		RequestID:    opts.RequestID,
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SectionID:    opts.SectionID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	if section.IsPublished == opts.IsPublished {
		log.InfoContext(ctx, "Series part is already published", "is_published", opts.IsPublished)
		return section, nil
	}
	if opts.IsPublished && section.LessonsCount == 0 {
		log.WarnContext(
			ctx,
			"Cannot publish series part without lessons",
			"lessonsCount", section.LessonsCount,
		)
		return nil, exceptions.NewValidationError("Cannot publish series part without lessons")
	}
	if !opts.IsPublished && section.IsPublished {
		if section.IsPublished {
			progressCount, err := s.database.CountSectionProgress(ctx, section.ID)
			if err != nil {
				log.ErrorContext(ctx, "Failed to count sections progress", "error", err)
				return nil, exceptions.FromDBError(err)
			}

			if progressCount > 0 {
				log.WarnContext(ctx, "Section is published and has progress")
				return nil, exceptions.NewConflictError("Section has students")
			}
		}
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return nil, exceptions.FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err, serviceErr)

	*section, err = qrs.UpdateSectionIsPublished(ctx, db.UpdateSectionIsPublishedParams{
		ID:          section.ID,
		IsPublished: opts.IsPublished,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to update series part is published", "error", err)
		return nil, exceptions.FromDBError(err)
	}
	if opts.IsPublished {
		params := db.AddSeriesSectionsCountParams{
			Slug:             opts.SeriesSlug,
			LessonsCount:     section.LessonsCount,
			ReadTimeSeconds:  section.ReadTimeSeconds,
			WatchTimeSeconds: section.WatchTimeSeconds,
		}
		if err := qrs.AddSeriesSectionsCount(ctx, params); err != nil {
			log.ErrorContext(ctx, "Failed to add series parts count", "error", err)
			return nil, exceptions.FromDBError(err)
		}
	} else {
		params := db.DecrementSeriesSectionsCountParams{
			Slug:             opts.SeriesSlug,
			LessonsCount:     section.LessonsCount,
			ReadTimeSeconds:  section.ReadTimeSeconds,
			WatchTimeSeconds: section.WatchTimeSeconds,
		}
		if err := qrs.DecrementSeriesSectionsCount(ctx, params); err != nil {
			log.ErrorContext(ctx, "Failed to decrement series parts count", "error", err)
			return nil, exceptions.FromDBError(err)
		}
	}

	log.InfoContext(ctx, "Series part is published updated", "id", section.ID)
	return section, nil
}

type DeleteSectionOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
}

func (s *Services) DeleteSection(ctx context.Context, opts DeleteSectionOptions) *exceptions.ServiceError {
	log := s.buildLogger(opts.RequestID, sectionsLocation, "DeleteSection").With(
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"sectionId", opts.SectionID,
	)
	log.InfoContext(ctx, "Deleting series part...")

	section, serviceErr := s.AssertSectionOwnership(ctx, AssertSectionOwnershipOptions(opts))
	if serviceErr != nil {
		return serviceErr
	}

	if section.IsPublished {
		progressCount, err := s.database.CountSectionProgress(ctx, section.ID)
		if err != nil {
			log.ErrorContext(ctx, "Failed to count sections progress", "error", err)
			return exceptions.FromDBError(err)
		}

		if progressCount > 0 {
			log.WarnContext(ctx, "Section is published and has progress")
			return exceptions.NewConflictError("Section has students")
		}
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return exceptions.FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err, serviceErr)

	if err := qrs.DeleteSectionById(ctx, opts.SectionID); err != nil {
		log.ErrorContext(ctx, "Failed to delete series part", "error", err)
		return exceptions.FromDBError(err)
	}

	posParams := db.DecrementSectionPositionParams{
		SeriesSlug: opts.SeriesSlug,
		Position:   section.Position,
		Position_2: 1,
	}
	if err := qrs.DecrementSectionPosition(ctx, posParams); err != nil {
		log.ErrorContext(ctx, "Failed to decrement series part position", "error", err)
		return exceptions.FromDBError(err)
	}

	countParams := db.DecrementSeriesSectionsCountParams{
		Slug:             opts.SeriesSlug,
		LessonsCount:     section.LessonsCount,
		ReadTimeSeconds:  section.ReadTimeSeconds,
		WatchTimeSeconds: section.WatchTimeSeconds,
	}
	if err := qrs.DecrementSeriesSectionsCount(ctx, countParams); err != nil {
		log.ErrorContext(ctx, "Failed to decrement series parts count", "error", err)
		return exceptions.FromDBError(err)
	}

	log.InfoContext(ctx, "Series part deleted", "id", section.ID)
	return nil
}
