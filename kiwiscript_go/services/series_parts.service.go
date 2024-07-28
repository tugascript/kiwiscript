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

type CreateSeriesPartOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	Title        string
	Description  string
}

func (s *Services) CreateSeriesPart(ctx context.Context, opts CreateSeriesPartOptions) (*db.SeriesPart, *ServiceError) {
	log := s.
		log.
		WithGroup("services.series_parts.CreateSeriesPart").
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

	seriesPart, err := s.database.CreateSeriesPart(ctx, db.CreateSeriesPartParams{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   series.Slug,
		Title:        opts.Title,
		Description:  opts.Description,
		AuthorID:     opts.UserID,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to create series part", "error", err)
		return nil, FromDBError(err)
	}

	log.InfoContext(ctx, "Series part created", "id", seriesPart.ID)
	return &seriesPart, nil
}

type FindSeriesPartBySlugsAndIDOptions struct {
	LanguageSlug string
	SeriesSlug   string
	SeriesPartID int32
}

func (s *Services) FindSeriesPartBySlugsAndID(
	ctx context.Context,
	opts FindSeriesPartBySlugsAndIDOptions,
) (*db.SeriesPart, *ServiceError) {
	log := s.
		log.
		WithGroup("services.series_parts.FindSeriesPartBySlugsAndID").
		With(
			"languageSlug", opts.LanguageSlug,
			"seriesSlug", opts.SeriesSlug,
			"seriesPartId", opts.SeriesPartID,
		)
	log.InfoContext(ctx, "Finding series part...")

	seriesPart, err := s.database.FindSeriesPartBySlugsAndID(ctx, db.FindSeriesPartBySlugsAndIDParams{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		ID:           opts.SeriesPartID,
	})
	if err != nil {
		log.WarnContext(ctx, "Failed to find series part", "error", err)
		return nil, FromDBError(err)
	}

	return &seriesPart, nil
}

func (s *Services) FindPublishedSeriesPartBySlugsAndID(
	ctx context.Context,
	opts FindSeriesPartBySlugsAndIDOptions,
) (*db.SeriesPart, *ServiceError) {
	log := s.
		log.
		WithGroup("services.series_parts.FindPublishedSeriesPartBySlugsAndID").
		With(
			"languageSlug", opts.LanguageSlug,
			"seriesSlug", opts.SeriesSlug,
			"seriesPartId", opts.SeriesPartID,
		)
	log.InfoContext(ctx, "Finding published series part...")

	seriesPart, err := s.database.FindPublishedSeriesPartBySlugsAndID(ctx, db.FindPublishedSeriesPartBySlugsAndIDParams{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		ID:           opts.SeriesPartID,
	})
	if err != nil {
		log.WarnContext(ctx, "Failed to find published series part", "error", err)
		return nil, FromDBError(err)
	}

	return &seriesPart, nil
}

type FindPublishedSeriesPartBySlugsAndIDWithProgressOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SeriesPartID int32
}

func (s *Services) FindPublishedSeriesPartBySlugsAndIDWithProgress(
	ctx context.Context,
	opts FindPublishedSeriesPartBySlugsAndIDWithProgressOptions,
) (*db.FindPublishedSeriesPartBySlugsAndIDWithProgressRow, *ServiceError) {
	log := s.log.WithGroup("services.series_parts.FindPublishedSeriesPartBySlugsAndIDWithProgress").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"seriesPartId", opts.SeriesPartID,
	)
	log.InfoContext(ctx, "Finding published series part with progress...")

	seriesPart, err := s.database.FindPublishedSeriesPartBySlugsAndIDWithProgress(
		ctx,
		db.FindPublishedSeriesPartBySlugsAndIDWithProgressParams{
			UserID:       opts.UserID,
			LanguageSlug: opts.LanguageSlug,
			SeriesSlug:   opts.SeriesSlug,
			ID:           opts.SeriesPartID,
		},
	)
	if err != nil {
		log.WarnContext(ctx, "Published series part not found")
		return nil, FromDBError(err)
	}

	return &seriesPart, nil
}

type FindPaginatedSeriesPartsBySlugsOptions struct {
	LanguageSlug string
	SeriesSlug   string
	Limit        int32
	Offset       int32
}

func (s *Services) FindPaginatedSeriesPartsBySlugs(
	ctx context.Context,
	opts FindPaginatedSeriesPartsBySlugsOptions,
) ([]db.SeriesPart, int64, *ServiceError) {
	log := s.log.WithGroup("services.series.FindPaginatedSeriesPartsBySlugs").With(
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

	count, err := s.database.CountSeriesPartsBySeriesSlug(ctx, opts.SeriesSlug)
	if err != nil {
		log.ErrorContext(ctx, "Error counting series parts", "error", err)
		return nil, 0, FromDBError(err)
	}
	if count == 0 {
		return make([]db.SeriesPart, 0), 0, nil
	}

	seriesParts, err := s.database.FindPaginatedSeriesPartsBySlugs(ctx, db.FindPaginatedSeriesPartsBySlugsParams{
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
	return seriesParts, count, nil
}

func (s *Services) findPublishedPartsCount(
	ctx context.Context,
	log *slog.Logger,
	languageSlug,
	seriesSlug string,
) (int64, *ServiceError) {
	seriesOpts := FindSeriesBySlugsOptions{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
	}
	if _, serviceErr := s.FindSeriesBySlugs(ctx, seriesOpts); serviceErr != nil {
		log.WarnContext(ctx, "Series not found", "error", serviceErr)
		return 0, serviceErr
	}

	count, err := s.database.CountPublishedSeriesPartsBySeriesSlug(ctx, seriesSlug)
	if err != nil {
		log.ErrorContext(ctx, "Error counting published series parts", "error", err)
		return 0, FromDBError(err)
	}

	return count, nil
}

func (s *Services) FindPaginatedPublishedSeriesPartsBySlugs(
	ctx context.Context,
	opts FindPaginatedSeriesPartsBySlugsOptions,
) ([]db.SeriesPart, int64, *ServiceError) {
	log := s.log.WithGroup("services.series.FindPaginatedPublishedSeriesPartsBySlugs").With(
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
	)
	log.InfoContext(ctx, "Finding paginated series parts...")

	count, serviceErr := s.findPublishedPartsCount(ctx, log, opts.LanguageSlug, opts.SeriesSlug)
	if serviceErr != nil {
		return nil, 0, serviceErr
	}
	if count == 0 {
		return make([]db.SeriesPart, 0), 0, nil
	}

	seriesParts, err := s.database.FindPaginatedPublishedSeriesPartsBySlugs(
		ctx,
		db.FindPaginatedPublishedSeriesPartsBySlugsParams{
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

	return seriesParts, count, nil
}

type FindSeriesPartBySlugsAndIDWithProgressOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	Limit        int32
	Offset       int32
}

func (s *Services) FindPaginatedPublishedSeriesPartsBySlugsWithProgress(
	ctx context.Context,
	opts FindSeriesPartBySlugsAndIDWithProgressOptions,
) ([]db.FindPaginatedPublishedSeriesPartsBySlugsWithProgressRow, int64, *ServiceError) {
	log := s.log.WithGroup("services.series.FindPaginatedPublishedSeriesPartsBySlugsWithProgress").With(
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
	)
	log.InfoContext(ctx, "Finding paginated series parts with progress...")

	count, serviceErr := s.findPublishedPartsCount(ctx, log, opts.LanguageSlug, opts.SeriesSlug)
	if serviceErr != nil {
		return nil, 0, serviceErr
	}
	if count == 0 {
		return make([]db.FindPaginatedPublishedSeriesPartsBySlugsWithProgressRow, 0), 0, nil
	}

	seriesParts, err := s.database.FindPaginatedPublishedSeriesPartsBySlugsWithProgress(
		ctx, db.FindPaginatedPublishedSeriesPartsBySlugsWithProgressParams{
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

	return seriesParts, count, nil
}

type AssertSeriesPartOwnershipOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SeriesPartID int32
}

func (s *Services) AssertSeriesPartOwnership(ctx context.Context, opts AssertSeriesPartOwnershipOptions) (*db.SeriesPart, *ServiceError) {
	log := s.
		log.
		WithGroup("services.series_parts.AssertSeriesOwnership").
		With(
			"languageSlug", opts.LanguageSlug,
			"seriesSlug", opts.SeriesSlug,
			"seriesPartId", opts.SeriesPartID,
		)
	log.InfoContext(ctx, "Asserting series part ownership...")

	seriesPart, err := s.FindSeriesPartBySlugsAndID(ctx, FindSeriesPartBySlugsAndIDOptions{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SeriesPartID: opts.SeriesPartID,
	})
	if err != nil {
		log.WarnContext(ctx, "Series part not found", "error", err)
		return nil, FromDBError(err)
	}

	if seriesPart.AuthorID != opts.UserID {
		log.WarnContext(ctx, "User is not the author of the series", "user_id", opts.UserID)
		return nil, NewForbiddenError()
	}

	log.InfoContext(ctx, "Series part ownership asserted")
	return seriesPart, nil
}

type UpdateSeriesPartOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SeriesPartID int32
	Title        string
	Description  string
	Position     int16
}

func (s *Services) UpdateSeriesPart(ctx context.Context, opts UpdateSeriesPartOptions) (*db.SeriesPart, *ServiceError) {
	log := s.
		log.
		WithGroup("services.series_parts.UpdateSeriesPart").
		With("series_slug", opts.SeriesSlug, "series_part_id", opts.SeriesPartID)
	log.InfoContext(ctx, "Updating series part...")

	seriesPart, serviceErr := s.AssertSeriesPartOwnership(ctx, AssertSeriesPartOwnershipOptions{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SeriesPartID: opts.SeriesPartID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	if opts.Position == 0 || opts.Position == seriesPart.Position {
		seriesPart, err := s.database.UpdateSeriesPart(ctx, db.UpdateSeriesPartParams{
			ID:          seriesPart.ID,
			Title:       opts.Title,
			Description: opts.Description,
		})

		if err != nil {
			log.ErrorContext(ctx, "Failed to update series part", "error", err)
			return nil, FromDBError(err)
		}

		log.InfoContext(ctx, "Series part updated", "id", seriesPart.ID)
		return &seriesPart, nil
	}

	count, err := s.database.CountSeriesPartsBySeriesSlug(ctx, opts.SeriesSlug)
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
	defer s.database.FinalizeTx(ctx, txn, err)

	oldPosition := seriesPart.Position
	*seriesPart, err = qrs.UpdateSeriesPartWithPosition(ctx, db.UpdateSeriesPartWithPositionParams{
		ID:          seriesPart.ID,
		Title:       opts.Title,
		Description: opts.Description,
		Position:    opts.Position,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to update series part", "error", err)
		return nil, FromDBError(err)
	}

	if oldPosition < opts.Position {
		params := db.DecrementSeriesPartPositionParams{
			SeriesSlug: opts.SeriesSlug,
			Position:   oldPosition,
			Position_2: opts.Position,
		}
		if err := qrs.DecrementSeriesPartPosition(ctx, params); err != nil {
			log.ErrorContext(ctx, "Failed to decrement series part position", "error", err)
			return nil, FromDBError(err)
		}
	} else {
		params := db.IncrementSeriesPartPositionParams{
			SeriesSlug: opts.SeriesSlug,
			Position:   oldPosition,
			Position_2: opts.Position,
		}
		if err := qrs.IncrementSeriesPartPosition(ctx, params); err != nil {
			log.ErrorContext(ctx, "Failed to increment series part position", "error", err)
			return nil, FromDBError(err)
		}
	}

	log.InfoContext(ctx, "Series part updated", "id", seriesPart.ID)
	return seriesPart, nil
}

type UpdateSeriesPartIsPublishedOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SeriesPartID int32
	IsPublished  bool
}

func (s *Services) UpdateSeriesPartIsPublished(ctx context.Context, opts UpdateSeriesPartIsPublishedOptions) (*db.SeriesPart, *ServiceError) {
	log := s.
		log.
		WithGroup("services.series_parts.UpdateSeriesPartIsPublished").
		With("series_slug", opts.SeriesSlug, "series_part_id", opts.SeriesPartID)
	log.InfoContext(ctx, "Updating series part is published...")

	seriesPart, serviceErr := s.AssertSeriesPartOwnership(ctx, AssertSeriesPartOwnershipOptions{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SeriesPartID: opts.SeriesPartID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	if seriesPart.IsPublished == opts.IsPublished {
		log.InfoContext(ctx, "Series part is already published", "is_published", opts.IsPublished)
		return seriesPart, nil
	}
	if opts.IsPublished && seriesPart.LecturesCount == 0 {
		log.WarnContext(ctx, "Cannot publish series part without lectures", "lectures_count", seriesPart.LecturesCount)
		return nil, NewValidationError("Cannot publish series part without lectures")
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return nil, FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err)

	*seriesPart, err = qrs.UpdateSeriesPartIsPublished(ctx, db.UpdateSeriesPartIsPublishedParams{
		ID:          seriesPart.ID,
		IsPublished: opts.IsPublished,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to update series part is published", "error", err)
		return nil, FromDBError(err)
	}
	if opts.IsPublished {
		params := db.AddSeriesPartsCountParams{
			Slug:             opts.SeriesSlug,
			LecturesCount:    seriesPart.LecturesCount,
			ReadTimeSeconds:  seriesPart.ReadTimeSeconds,
			WatchTimeSeconds: seriesPart.WatchTimeSeconds,
		}
		if err := qrs.AddSeriesPartsCount(ctx, params); err != nil {
			log.ErrorContext(ctx, "Failed to add series parts count", "error", err)
			return nil, FromDBError(err)
		}
	} else {
		// TODO: add constraints
		params := db.DecrementSeriesPartsCountParams{
			Slug:             opts.SeriesSlug,
			LecturesCount:    seriesPart.LecturesCount,
			ReadTimeSeconds:  seriesPart.ReadTimeSeconds,
			WatchTimeSeconds: seriesPart.WatchTimeSeconds,
		}
		if err := qrs.DecrementSeriesPartsCount(ctx, params); err != nil {
			log.ErrorContext(ctx, "Failed to decrement series parts count", "error", err)
			return nil, FromDBError(err)
		}
	}

	log.InfoContext(ctx, "Series part is published updated", "id", seriesPart.ID)
	return seriesPart, nil
}

type DeleteSeriesPartOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SeriesPartID int32
}

func (s *Services) DeleteSeriesPart(ctx context.Context, opts DeleteSeriesPartOptions) *ServiceError {
	log := s.
		log.
		WithGroup("services.series_parts.DeleteSeriesPart").
		With("series_slug", opts.SeriesSlug, "series_part_id", opts.SeriesPartID)
	log.InfoContext(ctx, "Deleting series part...")

	seriesPart, serviceErr := s.AssertSeriesPartOwnership(ctx, AssertSeriesPartOwnershipOptions(opts))
	if serviceErr != nil {
		return serviceErr
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err)

	if err := qrs.DeleteSeriesPartById(ctx, opts.SeriesPartID); err != nil {
		log.ErrorContext(ctx, "Failed to delete series part", "error", err)
		return FromDBError(err)
	}

	posParams := db.DecrementSeriesPartPositionParams{
		SeriesSlug: opts.SeriesSlug,
		Position:   seriesPart.Position,
		Position_2: 1,
	}
	if err := qrs.DecrementSeriesPartPosition(ctx, posParams); err != nil {
		log.ErrorContext(ctx, "Failed to decrement series part position", "error", err)
		return FromDBError(err)
	}

	countParams := db.DecrementSeriesPartsCountParams{
		Slug:             opts.SeriesSlug,
		LecturesCount:    seriesPart.LecturesCount,
		ReadTimeSeconds:  seriesPart.ReadTimeSeconds,
		WatchTimeSeconds: seriesPart.WatchTimeSeconds,
	}
	if err := qrs.DecrementSeriesPartsCount(ctx, countParams); err != nil {
		log.ErrorContext(ctx, "Failed to decrement series parts count", "error", err)
		return FromDBError(err)
	}

	log.InfoContext(ctx, "Series part deleted", "id", seriesPart.ID)
	return nil
}
