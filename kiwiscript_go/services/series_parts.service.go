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

	series, serviceErr := s.FindSeriesBySlugs(ctx, FindSeriesBySlugsOptions{
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

type SeriesPartLecture struct {
	ID               int32
	Title            string
	WatchTimeSeconds int32
	ReadTimeSeconds  int32
	IsPublished      bool
}

type SeriesPartDto struct {
	ID               int32
	Title            string
	Description      string
	Position         int16
	LecturesCount    int16
	ReadTimeSeconds  int32
	WatchTimeSeconds int32
	IsPublished      bool
	Lectures         []SeriesPartLecture
}

func mapSingleSeriesPartToDto(parts []db.FindPublishedSeriesPartByLanguageSlugSeriesSlugAndIDWithLecturesRow) *SeriesPartDto {
	dto := SeriesPartDto{
		ID:               parts[0].ID,
		Title:            parts[0].Title,
		Description:      parts[0].Description,
		Position:         parts[0].Position,
		LecturesCount:    parts[0].LecturesCount,
		ReadTimeSeconds:  parts[0].ReadTimeSeconds,
		WatchTimeSeconds: parts[0].WatchTimeSeconds,
		IsPublished:      parts[0].IsPublished,
		Lectures:         make([]SeriesPartLecture, len(parts)),
	}

	for i, part := range parts {
		dto.Lectures[i] = SeriesPartLecture{
			ID:               part.LectureID.Int32,
			Title:            part.LectureTitle.String,
			WatchTimeSeconds: part.LectureWatchTimeSeconds.Int32,
			ReadTimeSeconds:  part.LectureReadTimeSeconds.Int32,
			IsPublished:      part.LectureIsPublished.Bool,
		}
	}

	return &dto
}

type FindSeriesPartBySlugsAndIDOptions struct {
	LanguageSlug string
	SeriesSlug   string
	SeriesPartID int32
	IsPublished  bool
}

func (s *Services) FindSeriesPartBySlugAndID(ctx context.Context, opts FindSeriesPartBySlugsAndIDOptions) (*SeriesPartDto, *ServiceError) {
	log := s.
		log.
		WithGroup("services.series_parts.FindSeriesPartBySlugAndID").
		With("series_slug", opts.SeriesSlug, "series_part_id", opts.SeriesPartID)
	log.InfoContext(ctx, "Finding series part...")

	var parts []db.FindPublishedSeriesPartByLanguageSlugSeriesSlugAndIDWithLecturesRow
	if opts.IsPublished {
		var err error
		parts, err = s.database.FindPublishedSeriesPartByLanguageSlugSeriesSlugAndIDWithLectures(ctx, db.FindPublishedSeriesPartByLanguageSlugSeriesSlugAndIDWithLecturesParams{
			LanguageSlug: opts.LanguageSlug,
			SeriesSlug:   opts.SeriesSlug,
			ID:           opts.SeriesPartID,
		})
		if err != nil {
			log.ErrorContext(ctx, "Failed to find published series part", "error", err)
			return nil, FromDBError(err)
		}
	} else {
		notPublishedParts, err := s.database.FindSeriesPartByLanguageSlugSeriesSlugAndIDWithLectures(ctx, db.FindSeriesPartByLanguageSlugSeriesSlugAndIDWithLecturesParams{
			LanguageSlug: opts.LanguageSlug,
			SeriesSlug:   opts.SeriesSlug,
			ID:           opts.SeriesPartID,
		})
		if err != nil {
			log.ErrorContext(ctx, "Failed to find series part", "error", err)
			return nil, FromDBError(err)
		}

		parts = make([]db.FindPublishedSeriesPartByLanguageSlugSeriesSlugAndIDWithLecturesRow, len(notPublishedParts))
		for i, part := range notPublishedParts {
			parts[i] = db.FindPublishedSeriesPartByLanguageSlugSeriesSlugAndIDWithLecturesRow(part)
		}
	}

	if len(parts) == 0 {
		log.WarnContext(ctx, "Series part not found", "id", opts.SeriesPartID)
		return nil, NewError(CodeNotFound, MessageNotFound)
	}

	log.InfoContext(ctx, "Series part found", "id", parts[0].ID)
	return mapSingleSeriesPartToDto(parts), nil
}

type seriesPartMapper struct {
	dto SeriesPartDto
	idx int
}

func mapSeriesPartsToDtos(rows []db.FindPaginatedPublishedSeriesPartsByLanguageSlugAndSeriesSlugWithLecturesRow) []SeriesPartDto {
	rowMapper := make(map[int32]seriesPartMapper)
	for i, row := range rows {
		var dto SeriesPartDto
		idx := i - len(rowMapper)
		if m, ok := rowMapper[row.ID]; ok {
			dto = m.dto
			idx = m.idx
		} else {
			dto = SeriesPartDto{
				ID:               row.ID,
				Title:            row.Title,
				Description:      row.Description,
				Position:         row.Position,
				LecturesCount:    row.LecturesCount,
				ReadTimeSeconds:  row.ReadTimeSeconds,
				WatchTimeSeconds: row.WatchTimeSeconds,
				IsPublished:      row.IsPublished,
				Lectures:         make([]SeriesPartLecture, 0),
			}
		}

		if row.LectureID.Valid && row.LectureTitle.Valid {
			dto.Lectures = append(dto.Lectures, SeriesPartLecture{
				ID:               row.LectureID.Int32,
				Title:            row.LectureTitle.String,
				WatchTimeSeconds: row.LectureWatchTimeSeconds.Int32,
				ReadTimeSeconds:  row.LectureReadTimeSeconds.Int32,
				IsPublished:      row.LectureIsPublished.Bool,
			})
		}

		rowMapper[row.ID] = seriesPartMapper{
			dto: dto,
			idx: idx,
		}
	}

	dtos := make([]SeriesPartDto, len(rowMapper))

	for _, m := range rowMapper {
		dtos[m.idx] = m.dto
	}

	return dtos
}

type FindSeriesPartsBySlugsAndIDOptions struct {
	LanguageSlug      string
	SeriesSlug        string
	IsPublished       bool
	PublishedLectures bool
	Limit             int32
	Offset            int32
}

func (s *Services) FindPaginatedSeriesPartsBySlugsAndID(ctx context.Context, opts FindSeriesPartsBySlugsAndIDOptions) ([]SeriesPartDto, int64, *ServiceError) {
	log := s.
		log.
		WithGroup("services.series_parts.FindPaginatedSeriesPartsBySlugAndId").
		With("seriesSlug", opts.SeriesSlug)
	log.InfoContext(ctx, "Finding series parts...")

	params := db.FindPaginatedPublishedSeriesPartsByLanguageSlugAndSeriesSlugWithLecturesParams{
		SeriesSlug:   opts.SeriesSlug,
		LanguageSlug: opts.LanguageSlug,
		Limit:        opts.Limit,
		Offset:       opts.Offset,
	}
	var parts []db.FindPaginatedPublishedSeriesPartsByLanguageSlugAndSeriesSlugWithLecturesRow
	if opts.IsPublished {
		var err error
		parts, err = s.database.FindPaginatedPublishedSeriesPartsByLanguageSlugAndSeriesSlugWithLectures(
			ctx,
			params,
		)

		if err != nil {
			log.ErrorContext(ctx, "Failed to find published series parts", "error", err)
			return nil, 0, FromDBError(err)
		}
	} else {
		if opts.PublishedLectures {
			notPublishedParts, err := s.database.FindPaginatedSeriesPartsByLanguageSlugAndSeriesSlugWithPublishedLectures(
				ctx,
				db.FindPaginatedSeriesPartsByLanguageSlugAndSeriesSlugWithPublishedLecturesParams(params),
			)

			if err != nil {
				log.ErrorContext(ctx, "Failed to find series parts", "error", err)
				return nil, 0, FromDBError(err)
			}

			parts = make([]db.FindPaginatedPublishedSeriesPartsByLanguageSlugAndSeriesSlugWithLecturesRow, len(notPublishedParts))
			for i, part := range notPublishedParts {
				parts[i] = db.FindPaginatedPublishedSeriesPartsByLanguageSlugAndSeriesSlugWithLecturesRow(part)
			}
		} else {
			notPublishedParts, err := s.database.FindPaginatedSeriesPartsByLanguageSlugAndSeriesSlugWithLectures(
				ctx,
				db.FindPaginatedSeriesPartsByLanguageSlugAndSeriesSlugWithLecturesParams(params),
			)

			if err != nil {
				log.ErrorContext(ctx, "Failed to find series parts", "error", err)
				return nil, 0, FromDBError(err)
			}

			parts = make([]db.FindPaginatedPublishedSeriesPartsByLanguageSlugAndSeriesSlugWithLecturesRow, len(notPublishedParts))
			for i, part := range notPublishedParts {
				parts[i] = db.FindPaginatedPublishedSeriesPartsByLanguageSlugAndSeriesSlugWithLecturesRow(part)
			}
		}
	}

	count, err := s.database.CountSeriesPartsBySeriesSlug(ctx, opts.SeriesSlug)
	if err != nil {
		log.ErrorContext(ctx, "Failed to count series parts", "error", err)
		return nil, 0, FromDBError(err)
	}

	log.InfoContext(ctx, "Series parts found", "count", count)
	return mapSeriesPartsToDtos(parts), count, nil
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
		With("series_slug", opts.SeriesSlug, "series_part_id", opts.SeriesPartID)
	log.InfoContext(ctx, "Asserting series part ownership...")

	seriesPart, err := s.database.FindSeriesPartByLanguageSlugSeriesSlugAndID(ctx, db.FindSeriesPartByLanguageSlugSeriesSlugAndIDParams{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		ID:           opts.SeriesPartID,
	})
	if err != nil {
		log.WarnContext(ctx, "Series part not found", "error", err)
		return nil, FromDBError(err)
	}

	if seriesPart.AuthorID != opts.UserID {
		log.WarnContext(ctx, "User is not the author of the series", "user_id", opts.UserID)
		return nil, NewForbiddenError()
	}

	log.InfoContext(ctx, "Series part ownership asserted", "id", seriesPart.ID)
	return &seriesPart, nil
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
