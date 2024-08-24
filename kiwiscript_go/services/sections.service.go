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
	"log/slog"

	db "github.com/kiwiscript/kiwiscript_go/providers/database"
)

type CreateSectionOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	Title        string
	Description  string
}

func (s *Services) CreateSection(ctx context.Context, opts CreateSectionOptions) (*db.Section, *ServiceError) {
	log := s.
		log.
		WithGroup("services.series_parts.CreateSection").
		With("series_slug", opts.SeriesSlug, "title", opts.Title)
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

	count, err := s.database.CountSectionsBySeriesSlug(ctx, opts.SeriesSlug)
	if err != nil {
		log.ErrorContext(ctx, "Failed to count series by slug", "error", err)
		return nil, FromDBError(err)
	}

	section, err := s.database.CreateSection(ctx, db.CreateSectionParams{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   series.Slug,
		Title:        opts.Title,
		Description:  opts.Description,
		Position:     int16(count) + 1,
		AuthorID:     opts.UserID,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to create series part", "error", err)
		return nil, FromDBError(err)
	}

	log.InfoContext(ctx, "Series part created", "id", section.ID)
	return &section, nil
}

type FindSectionBySlugsAndIDOptions struct {
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
}

func (s *Services) FindSectionBySlugsAndID(
	ctx context.Context,
	opts FindSectionBySlugsAndIDOptions,
) (*db.Section, *ServiceError) {
	log := s.
		log.
		WithGroup("services.series_parts.FindSectionBySlugsAndID").
		With(
			"languageSlug", opts.LanguageSlug,
			"seriesSlug", opts.SeriesSlug,
			"sectionId", opts.SectionID,
		)
	log.InfoContext(ctx, "Finding series part...")

	section, err := s.database.FindSectionBySlugsAndID(ctx, db.FindSectionBySlugsAndIDParams{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		ID:           opts.SectionID,
	})
	if err != nil {
		log.WarnContext(ctx, "Failed to find series part", "error", err)
		return nil, FromDBError(err)
	}

	return &section, nil
}

func (s *Services) FindPublishedSectionBySlugsAndID(
	ctx context.Context,
	opts FindSectionBySlugsAndIDOptions,
) (*db.Section, *ServiceError) {
	log := s.
		log.
		WithGroup("services.series_parts.FindPublishedSectionBySlugsAndID").
		With(
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
		return nil, NewNotFoundError()
	}

	return section, nil
}

type FindPublishedSectionBySlugsAndIDWithProgressOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
}

func (s *Services) FindPublishedSectionBySlugsAndIDWithProgress(
	ctx context.Context,
	opts FindPublishedSectionBySlugsAndIDWithProgressOptions,
) (*db.FindPublishedSectionBySlugsAndIDWithProgressRow, *ServiceError) {
	log := s.log.WithGroup("services.series_parts.FindPublishedSectionBySlugsAndIDWithProgress").With(
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
		return nil, FromDBError(err)
	}

	return &section, nil
}

type FindPaginatedSectionsBySlugsOptions struct {
	LanguageSlug string
	SeriesSlug   string
	Limit        int32
	Offset       int32
}

func (s *Services) FindPaginatedSectionsBySlugs(
	ctx context.Context,
	opts FindPaginatedSectionsBySlugsOptions,
) ([]db.Section, int64, *ServiceError) {
	log := s.log.WithGroup("services.series.FindPaginatedSectionsBySlugs").With(
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
		return nil, 0, FromDBError(err)
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
		return nil, 0, FromDBError(err)
	}

	log.InfoContext(ctx, "Series parts found")
	return sections, count, nil
}

func (s *Services) findPublishedSectionsCount(
	ctx context.Context,
	log *slog.Logger,
	languageSlug,
	seriesSlug string,
) (int64, *ServiceError) {
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
		return 0, FromDBError(err)
	}

	return count, nil
}

func (s *Services) FindPaginatedPublishedSectionsBySlugs(
	ctx context.Context,
	opts FindPaginatedSectionsBySlugsOptions,
) ([]db.Section, int64, *ServiceError) {
	log := s.log.WithGroup("services.series.FindPaginatedPublishedSectionsBySlugs").With(
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
		return nil, 0, FromDBError(err)
	}

	return sections, count, nil
}

type FindSectionBySlugsAndIDWithProgressOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	Limit        int32
	Offset       int32
}

func (s *Services) FindPaginatedPublishedSectionsBySlugsWithProgress(
	ctx context.Context,
	opts FindSectionBySlugsAndIDWithProgressOptions,
) ([]db.FindPaginatedPublishedSectionsBySlugsWithProgressRow, int64, *ServiceError) {
	log := s.log.WithGroup("services.series.FindPaginatedPublishedSectionsBySlugsWithProgress").With(
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
		return nil, 0, FromDBError(err)
	}

	return sections, count, nil
}

type AssertSectionOwnershipOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
}

func (s *Services) AssertSectionOwnership(ctx context.Context, opts AssertSectionOwnershipOptions) (*db.Section, *ServiceError) {
	log := s.
		log.
		WithGroup("services.series_parts.AssertSeriesOwnership").
		With(
			"languageSlug", opts.LanguageSlug,
			"seriesSlug", opts.SeriesSlug,
			"sectionId", opts.SectionID,
		)
	log.InfoContext(ctx, "Asserting series part ownership...")

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
		return nil, NewForbiddenError()
	}

	log.InfoContext(ctx, "Series part ownership asserted")
	return section, nil
}

type UpdateSectionOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	Title        string
	Description  string
	Position     int16
}

func (s *Services) UpdateSection(ctx context.Context, opts UpdateSectionOptions) (*db.Section, *ServiceError) {
	log := s.
		log.
		WithGroup("services.series_parts.UpdateSection").
		With("series_slug", opts.SeriesSlug, "series_part_id", opts.SectionID)
	log.InfoContext(ctx, "Updating series part...")

	section, serviceErr := s.AssertSectionOwnership(ctx, AssertSectionOwnershipOptions{
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
			return nil, FromDBError(err)
		}

		log.InfoContext(ctx, "Series part updated", "id", section.ID)
		return &section, nil
	}

	count, err := s.database.CountSectionsBySeriesSlug(ctx, opts.SeriesSlug)
	if err != nil {
		log.ErrorContext(ctx, "Failed to count series parts", "error", err)
		return nil, FromDBError(err)
	}
	if int64(opts.Position) > count {
		log.WarnContext(ctx, "Position is out of range", "position", opts.Position, "count", count)
		return nil, NewValidationError("Position is out of range")
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return nil, FromDBError(err)
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
			return nil, FromDBError(err)
		}
	} else {
		params := db.IncrementSectionPositionParams{
			SeriesSlug: opts.SeriesSlug,
			Position:   oldPosition,
			Position_2: opts.Position,
		}
		if err := qrs.IncrementSectionPosition(ctx, params); err != nil {
			log.ErrorContext(ctx, "Failed to increment series part position", "error", err)
			return nil, FromDBError(err)
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
		return nil, FromDBError(err)
	}

	log.InfoContext(ctx, "Series part updated", "id", section.ID)
	return section, nil
}

type UpdateSectionIsPublishedOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	IsPublished  bool
}

func (s *Services) UpdateSectionIsPublished(ctx context.Context, opts UpdateSectionIsPublishedOptions) (*db.Section, *ServiceError) {
	log := s.
		log.
		WithGroup("services.series_parts.UpdateSectionIsPublished").
		With("series_slug", opts.SeriesSlug, "series_part_id", opts.SectionID)
	log.InfoContext(ctx, "Updating series part is published...")

	section, serviceErr := s.AssertSectionOwnership(ctx, AssertSectionOwnershipOptions{
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
		log.WarnContext(ctx, "Cannot publish series part without lectures", "lectures_count", section.LessonsCount)
		return nil, NewValidationError("Cannot publish series part without lectures")
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return nil, FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err, serviceErr)

	*section, err = qrs.UpdateSectionIsPublished(ctx, db.UpdateSectionIsPublishedParams{
		ID:          section.ID,
		IsPublished: opts.IsPublished,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to update series part is published", "error", err)
		return nil, FromDBError(err)
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
			return nil, FromDBError(err)
		}
	} else {
		// TODO: add constraints
		params := db.DecrementSeriesSectionsCountParams{
			Slug:             opts.SeriesSlug,
			LessonsCount:     section.LessonsCount,
			ReadTimeSeconds:  section.ReadTimeSeconds,
			WatchTimeSeconds: section.WatchTimeSeconds,
		}
		if err := qrs.DecrementSeriesSectionsCount(ctx, params); err != nil {
			log.ErrorContext(ctx, "Failed to decrement series parts count", "error", err)
			return nil, FromDBError(err)
		}
	}

	log.InfoContext(ctx, "Series part is published updated", "id", section.ID)
	return section, nil
}

type DeleteSectionOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
}

func (s *Services) DeleteSection(ctx context.Context, opts DeleteSectionOptions) *ServiceError {
	log := s.
		log.
		WithGroup("services.series_parts.DeleteSection").
		With("series_slug", opts.SeriesSlug, "series_part_id", opts.SectionID)
	log.InfoContext(ctx, "Deleting series part...")

	section, serviceErr := s.AssertSectionOwnership(ctx, AssertSectionOwnershipOptions(opts))
	if serviceErr != nil {
		return serviceErr
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err, serviceErr)

	if err := qrs.DeleteSectionById(ctx, opts.SectionID); err != nil {
		log.ErrorContext(ctx, "Failed to delete series part", "error", err)
		return FromDBError(err)
	}

	posParams := db.DecrementSectionPositionParams{
		SeriesSlug: opts.SeriesSlug,
		Position:   section.Position,
		Position_2: 1,
	}
	if err := qrs.DecrementSectionPosition(ctx, posParams); err != nil {
		log.ErrorContext(ctx, "Failed to decrement series part position", "error", err)
		return FromDBError(err)
	}

	countParams := db.DecrementSeriesSectionsCountParams{
		Slug:             opts.SeriesSlug,
		LessonsCount:     section.LessonsCount,
		ReadTimeSeconds:  section.ReadTimeSeconds,
		WatchTimeSeconds: section.WatchTimeSeconds,
	}
	if err := qrs.DecrementSeriesSectionsCount(ctx, countParams); err != nil {
		log.ErrorContext(ctx, "Failed to decrement series parts count", "error", err)
		return FromDBError(err)
	}

	log.InfoContext(ctx, "Series part deleted", "id", section.ID)
	return nil
}
