package services

import (
	"context"

	db "github.com/kiwiscript/kiwiscript_go/providers/database"
)

type CreateSeriesPartOptions struct {
	UserID      int32
	SeriesSlug  string
	Title       string
	Description string
}

func (s *Services) CreateSeriesPart(ctx context.Context, opts CreateSeriesPartOptions) (db.SeriesPart, *ServiceError) {
	log := s.
		log.
		WithGroup("services.series_parts.CreateSeriesPart").
		With("series_slug", opts.SeriesSlug, "title", opts.Title)
	log.InfoContext(ctx, "Creating series part...")
	var seriesPart db.SeriesPart

	series, serviceErr := s.FindSeriesBySlug(ctx, opts.SeriesSlug)
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
