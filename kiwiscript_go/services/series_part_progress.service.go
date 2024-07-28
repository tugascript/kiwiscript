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

type createSeriesPartProgressOptions struct {
	UserID             int32
	LanguageProgressID int32
	SeriesProgressID   int32
	LanguageSlug       string
	SeriesSlug         string
	SeriesPartID       int32
}

func (s *Services) createSeriesPartProgress(
	ctx context.Context,
	opts createSeriesPartProgressOptions,
) (*db.SeriesPartProgress, *ServiceError) {
	log := s.log.WithGroup("services.series_part_progress.createSeriesPartProgress").With(
		"userID", opts.UserID,
		"languageProgressID", opts.LanguageProgressID,
		"seriesProgressID", opts.SeriesProgressID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"seriesPartID", opts.SeriesPartID,
	)
	log.InfoContext(ctx, "Creating series part progress...")

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return nil, FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err)

	seriesPartProgress, err := qrs.CreateSeriesPartProgress(ctx, db.CreateSeriesPartProgressParams{
		UserID:             opts.UserID,
		LanguageProgressID: opts.LanguageProgressID,
		SeriesProgressID:   opts.SeriesProgressID,
		SeriesPartID:       opts.SeriesPartID,
		LanguageSlug:       opts.LanguageSlug,
		SeriesSlug:         opts.SeriesSlug,
	})
	if err != nil {
		return nil, FromDBError(err)
	}

	params := db.SetSeriesPartProgressIsCurrentFalseParams{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SeriesPartID: opts.SeriesPartID,
	}
	if err := qrs.SetSeriesPartProgressIsCurrentFalse(ctx, params); err != nil {
		log.ErrorContext(ctx, "Failed to set series part progress is current false", "error", err)
		return nil, FromDBError(err)
	}

	return &seriesPartProgress, nil
}

//type CreateOrUpdateSeriesPartProgressOptions struct {
//	UserID       int32
//	LanguageSlug string
//	SeriesSlug   string
//	SeriesPartID int32
//}
//
//func (s *Services) CreateOrUpdateSeriesPartProgress(
//	ctx context.Context,
//	opts CreateOrUpdateSeriesProgressOptions,
//) (db.ToSeriesPartDTOWithProgress, *db.SeriesPartProgress, *ServiceError) {
//	log := s.log.WithGroup("services.series_part_progress.CreateOrUpdateSeriesPartProgress").With(
//		"userID", opts.UserID,
//		"languageSlug", opts.LanguageSlug,
//		"seriesSlug", opts.SeriesSlug,
//	)
//	log.InfoContext(ctx, "Creating or updating series part progress...")
//
//	seriesPart, serviceErr := s.FindSeriesPartBySlugsAndID(ctx, FindSeriesPartBySlugsAndIDOptions{
//		LanguageSlug: opts.LanguageSlug,
//
//	})
//}
