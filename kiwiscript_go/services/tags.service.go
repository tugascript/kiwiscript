package services

import (
	"context"

	db "github.com/kiwiscript/kiwiscript_go/providers/database"
)

func (s *Services) FindTagsBySeriesID(ctx context.Context, seriesID int32) ([]db.Tag, *ServiceError) {
	log := s.log.WithGroup("service.series.FindTagsBySeriesID").With("seriesID", seriesID)
	log.InfoContext(ctx, "Getting tags by series ID")

	tags, err := s.database.FindTagsBySeriesID(ctx, seriesID)
	if err != nil {
		log.ErrorContext(ctx, "Error getting tags", "error", err)
		return nil, FromDBError(err)
	}

	log.InfoContext(ctx, "Tags found")
	return tags, nil
}
