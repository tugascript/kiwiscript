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
	objStg "github.com/kiwiscript/kiwiscript_go/providers/object_storage"
	"mime/multipart"
)

const seriesPicturesLocation string = "series_pictures"

type FindSeriesPictureBySeriesIDOptions struct {
	RequestID string
	SeriesID  int32
}

func (s *Services) FindSeriesPictureBySeriesID(
	ctx context.Context,
	opts FindSeriesPictureBySeriesIDOptions,
) (*db.SeriesPicture, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, seriesPicturesLocation, "FindSeriesPictureBySeriesID").With(
		"seriesId", opts.SeriesID,
	)
	log.InfoContext(ctx, "Finding series picture...")

	seriesPicture, err := s.database.FindSeriesPictureBySeriesID(ctx, opts.SeriesID)
	if err != nil {
		log.WarnContext(ctx, "Series picture not found", "error", err)
		return nil, exceptions.FromDBError(err)
	}

	return &seriesPicture, nil
}

type UploadSeriesPictureOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	FileHeader   *multipart.FileHeader
}

func (s *Services) UploadSeriesPicture(
	ctx context.Context,
	opts UploadSeriesPictureOptions,
) (*db.SeriesPicture, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, seriesPicturesLocation, "UploadSeriesPicture").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
	)
	log.InfoContext(ctx, "Upload series picture...")

	series, serviceErr := s.AssertSeriesOwnership(ctx, AssertSeriesOwnershipOptions{
		RequestID:    opts.RequestID,
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	findOpts := FindSeriesPictureBySeriesIDOptions{
		RequestID: opts.RequestID,
		SeriesID:  series.ID,
	}
	if _, serviceErr = s.FindSeriesPictureBySeriesID(ctx, findOpts); serviceErr == nil {
		log.WarnContext(ctx, "Series picture already exists")
		return nil, exceptions.NewConflictError("Series picture already exists")
	}

	fileId, ext, err := s.objStg.UploadImage(ctx, objStg.UploadImageOptions{
		RequestID: opts.RequestID,
		UserID:    opts.UserID,
		FH:        opts.FileHeader,
	})
	if err != nil {
		if err.Error() == "mime type not supported" {
			log.WarnContext(ctx, "Invalid file type", "error", err)
			return nil, exceptions.NewValidationError("Invalid file type")
		}

		log.ErrorContext(ctx, "Error uploading picture", "error", err)
		return nil, exceptions.NewServerError()
	}

	seriesPicture, err := s.database.CreateSeriesPicture(ctx, db.CreateSeriesPictureParams{
		ID:       fileId,
		SeriesID: series.ID,
		AuthorID: opts.UserID,
		Ext:      ext,
	})
	if err != nil {
		log.ErrorContext(ctx, "Error creating series picture", "error", err)
		return nil, exceptions.FromDBError(err)
	}

	return &seriesPicture, nil
}

type DeletePictureOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
}

func (s *Services) DeleteSeriesPicture(
	ctx context.Context,
	opts DeletePictureOptions,
) *exceptions.ServiceError {
	log := s.buildLogger(opts.RequestID, seriesPicturesLocation, "DeleteSeriesPicture").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
	)
	log.InfoContext(ctx, "Deleting series picture...")

	series, serviceErr := s.AssertSeriesOwnership(ctx, AssertSeriesOwnershipOptions{
		RequestID:    opts.RequestID,
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
	})
	if serviceErr != nil {
		return serviceErr
	}

	seriesPicture, serviceErr := s.FindSeriesPictureBySeriesID(ctx, FindSeriesPictureBySeriesIDOptions{
		RequestID: opts.RequestID,
		SeriesID:  series.ID,
	})
	if serviceErr != nil {
		return serviceErr
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return exceptions.FromDBError(err)
	}
	defer func() {
		log.DebugContext(ctx, "Finalizing transaction")
		s.database.FinalizeTx(ctx, txn, err, serviceErr)
	}()

	if err := qrs.DeleteSeriesPicture(ctx, seriesPicture.ID); err != nil {
		log.ErrorContext(ctx, "Failed to delete series picture", "error", err)
		serviceErr = exceptions.FromDBError(err)
		return serviceErr
	}
	if err := s.objStg.DeleteFile(ctx, opts.UserID, seriesPicture.ID, seriesPicture.Ext); err != nil {
		log.ErrorContext(ctx, "Failed to delete file from object storage", "error", err)
		serviceErr = exceptions.NewServerError()
		return serviceErr
	}

	return nil
}
