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
	"mime/multipart"
)

func (s *Services) FindSeriesPictureBySeriesID(ctx context.Context, id int32) (*db.SeriesPicture, *ServiceError) {
	log := s.log.WithGroup("services.series_pictures.FindSeriesPictureBySeriesID").With("id", id)
	log.InfoContext(ctx, "Finding series picture...")

	seriesPicture, err := s.database.FindSeriesPictureBySeriesID(ctx, id)
	if err != nil {
		log.WarnContext(ctx, "Series picture not found", "error", err)
		return nil, FromDBError(err)
	}

	return &seriesPicture, nil
}

type UploadSeriesPictureOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	FileHeader   *multipart.FileHeader
}

func (s *Services) UploadSeriesPicture(
	ctx context.Context,
	opts UploadSeriesPictureOptions,
) (*db.SeriesPicture, *ServiceError) {
	log := s.log.WithGroup("services.series_pictures.UploadSeriesPicture").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
	)
	log.InfoContext(ctx, "Upload series picture...")

	series, serviceErr := s.AssertSeriesOwnership(ctx, AssertSeriesOwnershipOptions{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	seriesPicture, serviceErr := s.FindSeriesPictureBySeriesID(ctx, series.ID)
	if serviceErr != nil {
		return nil, serviceErr
	}

	fileId, ext, err := s.objStg.UploadImage(ctx, opts.UserID, opts.FileHeader)
	if err != nil {
		log.ErrorContext(ctx, "Error uploading picture", "error", err)
		return nil, FromDBError(err)
	}

	*seriesPicture, err = s.database.CreateSeriesPicture(ctx, db.CreateSeriesPictureParams{
		ID:       fileId,
		SeriesID: series.ID,
		AuthorID: opts.UserID,
		Ext:      ext,
	})
	if err != nil {
		log.ErrorContext(ctx, "Error creating series picture", "error", err)
		return nil, FromDBError(err)
	}

	return seriesPicture, nil
}

type DeletePictureOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
}

func (s *Services) DeleteSeriesPicture(
	ctx context.Context,
	opts DeletePictureOptions,
) *ServiceError {
	log := s.log.WithGroup("services.series_pictures.DeleteSeriesPicture").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
	)
	log.InfoContext(ctx, "Deleting series picture...")

	series, serviceErr := s.AssertSeriesOwnership(ctx, AssertSeriesOwnershipOptions{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
	})
	if serviceErr != nil {
		return serviceErr
	}

	seriesPicture, serviceErr := s.FindSeriesPictureBySeriesID(ctx, series.ID)
	if serviceErr != nil {
		return serviceErr
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err)

	if err := qrs.DeleteSeriesPicture(ctx, seriesPicture.ID); err != nil {
		log.ErrorContext(ctx, "Failed to delete series picture", "error", err)
		return FromDBError(err)
	}
	if err := s.objStg.DeleteFile(ctx, opts.UserID, seriesPicture.ID, seriesPicture.Ext); err != nil {
		log.ErrorContext(ctx, "Failed to delete file from object storage", "error", err)
		return NewServerError("Failed to delete file from object storage")
	}

	return nil
}
