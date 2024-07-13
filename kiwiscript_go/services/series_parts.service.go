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

func (s *Services) CreateSeriesPart(ctx context.Context, opts CreateSeriesPartOptions) (db.SeriesPart, *ServiceError) {
	log := s.
		log.
		WithGroup("services.series_parts.CreateSeriesPart").
		With("series_slug", opts.SeriesSlug, "title", opts.Title)
	log.InfoContext(ctx, "Creating series part...")
	var seriesPart db.SeriesPart

	series, serviceErr := s.FindSeriesBySlugs(ctx, FindSeriesBySlugsOptions{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
	})
	if serviceErr != nil {
		log.WarnContext(ctx, "Series not found", "error", serviceErr)
		return seriesPart, serviceErr
	}

	seriesPart, err := s.database.CreateSeriesPart(ctx, db.CreateSeriesPartParams{
		SeriesID:    series.ID,
		Title:       opts.Title,
		Description: opts.Description,
		AuthorID:    opts.UserID,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to create series part", "error", err)
		return seriesPart, FromDBError(err)
	}

	log.InfoContext(ctx, "Series part created", "id", seriesPart.ID)
	return seriesPart, nil
}

type FindSeriesPartByIDsOptions struct {
	SeriesID     int32
	SeriesPartID int32
}

func (s *Services) FindSeriesPartByIDs(ctx context.Context, opts FindSeriesPartByIDsOptions) (db.SeriesPart, *ServiceError) {
	log := s.
		log.
		WithGroup("services.series_parts.FindSeriesPart").
		With("series_id", opts.SeriesID, "series_part_id", opts.SeriesPartID)
	log.InfoContext(ctx, "Finding series part...")

	seriesPart, err := s.database.FindSeriesPartBySeriesIDAndID(ctx, db.FindSeriesPartBySeriesIDAndIDParams{
		SeriesID: opts.SeriesID,
		ID:       opts.SeriesPartID,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to find series part", "error", err)
		return seriesPart, FromDBError(err)
	}

	log.InfoContext(ctx, "Series part found", "id", seriesPart.ID)
	return seriesPart, nil
}

type SeriesLecture struct {
	ID    int32
	Title string
}

type SeriesPartDto struct {
	ID                   int32
	Title                string
	Description          string
	Position             int16
	LecturesCount        int16
	TotalDurationSeconds int32
	Lectures             []SeriesLecture
}

func mapSingleSeriesPartToDto(parts []db.FindSeriesPartBySeriesIDAndIDWithLecturesRow) SeriesPartDto {
	dto := SeriesPartDto{
		ID:                   parts[0].ID,
		Title:                parts[0].Title,
		Description:          parts[0].Description,
		Position:             parts[0].Position,
		LecturesCount:        parts[0].LecturesCount,
		TotalDurationSeconds: parts[0].TotalDurationSeconds,
		Lectures:             make([]SeriesLecture, len(parts)),
	}

	for i, part := range parts {
		dto.Lectures[i] = SeriesLecture{
			ID:    part.LectureID.Int32,
			Title: part.LectureTitle.String,
		}
	}

	return dto
}

type FindSeriesPartBySlugsAndIDOptions struct {
	LanguageSlug string
	SeriesSlug   string
	SeriesPartID int32
}

func (s *Services) FindSeriesPartBySlugAndID(ctx context.Context, opts FindSeriesPartBySlugsAndIDOptions) (SeriesPartDto, *ServiceError) {
	log := s.
		log.
		WithGroup("services.series_parts.FindSeriesPartBySlugAndID").
		With("series_slug", opts.SeriesSlug, "series_part_id", opts.SeriesPartID)
	log.InfoContext(ctx, "Finding series part...")

	series, serviceErr := s.FindSeriesBySlugs(ctx, FindSeriesBySlugsOptions{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
	})
	if serviceErr != nil {
		log.WarnContext(ctx, "Series not found", "error", serviceErr)
		return SeriesPartDto{}, serviceErr
	}

	parts, err := s.database.FindSeriesPartBySeriesIDAndIDWithLectures(ctx, db.FindSeriesPartBySeriesIDAndIDWithLecturesParams{
		SeriesID: series.ID,
		ID:       opts.SeriesPartID,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to find series part", "error", err)
		return SeriesPartDto{}, FromDBError(err)
	}
	if len(parts) == 0 {
		log.WarnContext(ctx, "Series part not found", "id", opts.SeriesPartID)
		return SeriesPartDto{}, NewError(CodeNotFound, MessageNotFound)
	}

	log.InfoContext(ctx, "Series part found", "id", parts[0].ID)
	return mapSingleSeriesPartToDto(parts), nil
}

type seriesPartMapper struct {
	dto SeriesPartDto
	idx int
}

func mapSeriesPartsToDtos(rows []db.FindPaginatedSeriesPartsBySeriesIdWithLecturesRow) []SeriesPartDto {
	rowMapper := make(map[int32]seriesPartMapper)
	for i, row := range rows {
		var dto SeriesPartDto
		idx := i - len(rowMapper)
		if m, ok := rowMapper[row.ID]; ok {
			dto = m.dto
			idx = m.idx
		} else {
			dto = SeriesPartDto{
				ID:                   row.ID,
				Title:                row.Title,
				Description:          row.Description,
				Position:             row.Position,
				LecturesCount:        row.LecturesCount,
				TotalDurationSeconds: row.TotalDurationSeconds,
				Lectures:             make([]SeriesLecture, 0),
			}
		}

		if row.LectureID.Valid && row.LectureTitle.Valid {
			dto.Lectures = append(dto.Lectures, SeriesLecture{
				ID:    row.LectureID.Int32,
				Title: row.LectureTitle.String,
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

type FindSeriesPartsOptions struct {
	LanguageSlug string
	SeriesSlug   string
	Limit        int32
	Offset       int32
}

func (s *Services) FindPaginatedSeriesPartsBySlugsAndId(ctx context.Context, opts FindSeriesPartsOptions) ([]SeriesPartDto, int64, *ServiceError) {
	log := s.
		log.
		WithGroup("services.series_parts.FindPaginatedSeriesPartsBySlugAndId").
		With("seriesSlug", opts.SeriesSlug)
	log.InfoContext(ctx, "Finding series parts...")

	series, serviceErr := s.FindSeriesBySlugs(ctx, FindSeriesBySlugsOptions{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
	})
	if serviceErr != nil {
		log.WarnContext(ctx, "Series not found", "error", serviceErr)
		return nil, 0, serviceErr
	}

	parts, err := s.database.FindPaginatedSeriesPartsBySeriesIdWithLectures(ctx, db.FindPaginatedSeriesPartsBySeriesIdWithLecturesParams{
		SeriesID: series.ID,
		Limit:    opts.Limit,
		Offset:   opts.Offset,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to find series parts", "error", err)
		return nil, 0, FromDBError(err)
	}
	count, err := s.database.CountSeriesPartsBySeriesId(ctx, series.ID)
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

func (s *Services) AssertSeriesOwnership(ctx context.Context, opts AssertSeriesPartOwnershipOptions) (db.SeriesPart, *ServiceError) {
	log := s.
		log.
		WithGroup("services.series_parts.AssertSeriesOwnership").
		With("series_slug", opts.SeriesSlug, "series_part_id", opts.SeriesPartID)
	log.InfoContext(ctx, "Asserting series part ownership...")

	log.InfoContext(ctx, "Finding series...")
	series, serviceErr := s.FindSeriesBySlugs(ctx, FindSeriesBySlugsOptions{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
	})
	var seriesPart db.SeriesPart

	if serviceErr != nil {
		log.WarnContext(ctx, "Series not found", "error", serviceErr)
		return seriesPart, serviceErr
	}
	if series.AuthorID != opts.UserID {
		log.WarnContext(ctx, "User is not the author of the series", "user_id", opts.UserID)
		return seriesPart, NewForbiddenError()
	}

	seriesPart, serviceErr = s.FindSeriesPartByIDs(ctx, FindSeriesPartByIDsOptions{
		SeriesID:     series.ID,
		SeriesPartID: opts.SeriesPartID,
	})
	if serviceErr != nil {
		log.WarnContext(ctx, "Series part not found", "error", serviceErr)
		return db.SeriesPart{}, serviceErr
	}

	log.InfoContext(ctx, "Series part ownership asserted", "id", seriesPart.ID)
	return seriesPart, nil
}

type UpdateSeriesPartOptions struct {
	UserID       int32
	SeriesSlug   string
	SeriesPartID int32
	Title        string
	Description  string
	Position     int16
}

func (s *Services) UpdateSeriesPart(ctx context.Context, opts UpdateSeriesPartOptions) (db.SeriesPart, *ServiceError) {
	log := s.
		log.
		WithGroup("services.series_parts.UpdateSeriesPart").
		With("series_slug", opts.SeriesSlug, "series_part_id", opts.SeriesPartID)
	log.InfoContext(ctx, "Updating series part...")

	seriesPart, serviceErr := s.AssertSeriesOwnership(ctx, AssertSeriesPartOwnershipOptions{
		UserID:       opts.UserID,
		SeriesSlug:   opts.SeriesSlug,
		SeriesPartID: opts.SeriesPartID,
	})
	if serviceErr != nil {
		return seriesPart, serviceErr
	}

	if opts.Position == 0 || opts.Position == seriesPart.Position {
		seriesPart, err := s.database.UpdateSeriesPart(ctx, db.UpdateSeriesPartParams{
			ID:          seriesPart.ID,
			Title:       opts.Title,
			Description: opts.Description,
		})

		if err != nil {
			log.ErrorContext(ctx, "Failed to update series part", "error", err)
			return seriesPart, FromDBError(err)
		}

		log.InfoContext(ctx, "Series part updated", "id", seriesPart.ID)
		return seriesPart, nil
	}

	count, err := s.database.CountSeriesPartsBySeriesId(ctx, seriesPart.SeriesID)
	if err != nil {
		log.ErrorContext(ctx, "Failed to count series parts", "error", err)
		return db.SeriesPart{}, FromDBError(err)
	}
	if int64(opts.Position) > count {
		log.WarnContext(ctx, "Position is out of range", "position", opts.Position, "count", count)
		return db.SeriesPart{}, NewValidationError("Position is out of range")
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return db.SeriesPart{}, FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err)

	oldPosition := seriesPart.Position
	seriesPart, err = qrs.UpdateSeriesPartWithPosition(ctx, db.UpdateSeriesPartWithPositionParams{
		ID:          seriesPart.ID,
		Title:       opts.Title,
		Description: opts.Description,
		Position:    opts.Position,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to update series part", "error", err)
		return seriesPart, FromDBError(err)
	}

	if oldPosition < opts.Position {
		params := db.DecrementSeriesPartPositionParams{
			SeriesID:   seriesPart.SeriesID,
			Position:   oldPosition,
			Position_2: opts.Position,
		}
		if err := qrs.DecrementSeriesPartPosition(ctx, params); err != nil {
			log.ErrorContext(ctx, "Failed to decrement series part position", "error", err)
			return db.SeriesPart{}, FromDBError(err)
		}
	} else {
		params := db.IncrementSeriesPartPositionParams{
			SeriesID:   seriesPart.SeriesID,
			Position:   oldPosition,
			Position_2: opts.Position,
		}
		if err := qrs.IncrementSeriesPartPosition(ctx, params); err != nil {
			log.ErrorContext(ctx, "Failed to increment series part position", "error", err)
			return db.SeriesPart{}, FromDBError(err)
		}
	}

	log.InfoContext(ctx, "Series part updated", "id", seriesPart.ID)
	return seriesPart, nil
}

type UpdateSeriesPartIsPublishedOptions struct {
	UserID       int32
	SeriesSlug   string
	SeriesPartID int32
	IsPublished  bool
}

func (s *Services) UpdateSeriesPartIsPublished(ctx context.Context, opts UpdateSeriesPartIsPublishedOptions) (db.SeriesPart, *ServiceError) {
	log := s.
		log.
		WithGroup("services.series_parts.UpdateSeriesPartIsPublished").
		With("series_slug", opts.SeriesSlug, "series_part_id", opts.SeriesPartID)
	log.InfoContext(ctx, "Updating series part is published...")

	seriesPart, serviceErr := s.AssertSeriesOwnership(ctx, AssertSeriesPartOwnershipOptions{
		UserID:       opts.UserID,
		SeriesSlug:   opts.SeriesSlug,
		SeriesPartID: opts.SeriesPartID,
	})
	if serviceErr != nil {
		return db.SeriesPart{}, serviceErr
	}

	if seriesPart.IsPublished == opts.IsPublished {
		log.InfoContext(ctx, "Series part is already published", "is_published", opts.IsPublished)
		return seriesPart, nil
	}
	if opts.IsPublished && seriesPart.LecturesCount == 0 {
		log.WarnContext(ctx, "Cannot publish series part without lectures", "lectures_count", seriesPart.LecturesCount)
		return db.SeriesPart{}, NewValidationError("Cannot publish series part without lectures")
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return db.SeriesPart{}, FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err)

	seriesPart, err = qrs.UpdateSeriesPartIsPublished(ctx, db.UpdateSeriesPartIsPublishedParams{
		ID:          seriesPart.ID,
		IsPublished: opts.IsPublished,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to update series part is published", "error", err)
		return seriesPart, FromDBError(err)
	}
	if opts.IsPublished {
		params := db.AddSeriesPartsCountParams{
			ID:            seriesPart.SeriesID,
			LecturesCount: seriesPart.LecturesCount,
		}
		if err := qrs.AddSeriesPartsCount(ctx, params); err != nil {
			log.ErrorContext(ctx, "Failed to add series parts count", "error", err)
			return db.SeriesPart{}, FromDBError(err)
		}
	} else {
		// TODO: add constraints
		params := db.DecrementSeriesPartsCountParams{
			ID:            seriesPart.SeriesID,
			LecturesCount: seriesPart.LecturesCount,
		}
		if err := qrs.DecrementSeriesPartsCount(ctx, params); err != nil {
			log.ErrorContext(ctx, "Failed to decrement series parts count", "error", err)
			return db.SeriesPart{}, FromDBError(err)
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

	seriesPart, serviceErr := s.AssertSeriesOwnership(ctx, AssertSeriesPartOwnershipOptions(opts))
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

	params := db.DecrementSeriesPartPositionParams{
		SeriesID:   seriesPart.SeriesID,
		Position:   seriesPart.Position,
		Position_2: 1,
	}
	if err := qrs.DecrementSeriesPartPosition(ctx, params); err != nil {
		log.ErrorContext(ctx, "Failed to decrement series part position", "error", err)
		return FromDBError(err)
	}

	log.InfoContext(ctx, "Series part deleted", "id", seriesPart.ID)
	return nil
}
