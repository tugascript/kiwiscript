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
	"github.com/google/uuid"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"mime/multipart"
)

type UploadLectureFileOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SeriesPartID int32
	LectureID    int32
	FileName     string
	FileHeader   *multipart.FileHeader
}

func (s *Services) UploadLectureFile(
	ctx context.Context,
	opts UploadLectureFileOptions,
) (*db.LectureFile, *ServiceError) {
	log := s.
		log.
		WithGroup("services.lecture_files.UploadLectureFile").
		With("fileName", opts.FileName)
	log.InfoContext(ctx, "Uploading lecture file...")

	lecOpts := AssertLectureOwnershipOptions{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SeriesPartID: opts.SeriesPartID,
		LectureID:    opts.LectureID,
	}
	if _, serviceErr := s.AssertLectureOwnership(ctx, lecOpts); serviceErr != nil {
		return nil, serviceErr
	}

	fileId, docExt, err := s.objStg.UploadDocument(ctx, opts.UserID, opts.FileHeader)
	if err != nil {
		log.ErrorContext(ctx, "Error uploading document", "error", err)
		return nil, NewServerError("Error uploading document")
	}

	lectureFile, err := s.database.CreateLectureFile(ctx, db.CreateLectureFileParams{
		AuthorID:  opts.UserID,
		LectureID: opts.LectureID,
		File:      fileId,
		Ext:       docExt,
		Filename:  opts.FileName,
	})
	if err != nil {
		log.ErrorContext(ctx, "Error creating lecture file", "error", err)
		return nil, FromDBError(err)
	}

	return &lectureFile, nil
}

type DeleteLectureFileOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SeriesPartID int32
	LectureID    int32
	File         uuid.UUID
}

func (s *Services) DeleteLectureFile(
	ctx context.Context,
	opts DeleteLectureFileOptions,
) *ServiceError {
	log := s.
		log.
		WithGroup("services.lecture_files.DeleteLectureFile").
		With("file", opts.File.String())
	log.InfoContext(ctx, "Deleting lecture file...")

	lecOpts := AssertLectureOwnershipOptions{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SeriesPartID: opts.SeriesPartID,
		LectureID:    opts.LectureID,
	}
	if _, serviceErr := s.AssertLectureOwnership(ctx, lecOpts); serviceErr != nil {
		return serviceErr
	}

	lectureFile, err := s.database.FindLectureFileByFileAndLectureID(ctx, db.FindLectureFileByFileAndLectureIDParams{
		File:      opts.File,
		LectureID: opts.LectureID,
	})
	if err != nil {
		log.WarnContext(ctx, "Error finding lecture file", "error", err)
		return FromDBError(err)
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err)

	if err := qrs.DeleteLectureFile(ctx, lectureFile.ID); err != nil {
		log.ErrorContext(ctx, "Failed to delete lecture file", "error", err)
		return FromDBError(err)
	}
	if err := s.objStg.DeleteFile(ctx, opts.UserID, opts.File, lectureFile.Ext); err != nil {
		log.ErrorContext(ctx, "Failed to delete file from object storage", "error", err)
		return NewServerError("Failed to delete file from object storage")
	}

	return nil
}

type FindLectureFileOptions struct {
	LanguageSlug string
	SeriesSlug   string
	SeriesPartID int32
	LectureID    int32
	File         uuid.UUID
	IsPublished  bool
}

func (s *Services) FindLectureFile(
	ctx context.Context,
	opts FindLectureFileOptions,
) (*db.LectureFile, *ServiceError) {
	log := s.
		log.
		WithGroup("services.lecture_files.GetLectureFile").
		With("file", opts.File.String())
	log.InfoContext(ctx, "Getting lecture file...")

	lecture, serviceErr := s.FindLectureBySlugsAndIDs(ctx, FindLectureOptions{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SeriesPartID: opts.SeriesPartID,
		LectureID:    opts.LectureID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	if opts.IsPublished && !lecture.IsPublished {
		log.WarnContext(ctx, "Cannot find file from unpublished lecture")
		return nil, NewValidationError("Cannot get file from published lecture")
	}

	lectureFile, err := s.database.FindLectureFileByFileAndLectureID(ctx, db.FindLectureFileByFileAndLectureIDParams{
		File:      opts.File,
		LectureID: opts.LectureID,
	})
	if err != nil {
		log.WarnContext(ctx, "Error finding lecture file", "error", err)
		return nil, FromDBError(err)
	}

	return &lectureFile, nil
}

type FindLectureFilesOptions struct {
	LanguageSlug string
	SeriesSlug   string
	SeriesPartID int32
	LectureID    int32
	File         uuid.UUID
	IsPublished  bool
}

func (s *Services) FindLectureFiles(
	ctx context.Context,
	opts FindLectureFilesOptions,
) ([]db.LectureFile, *ServiceError) {
	log := s.log.WithGroup("services.lecture_files.GetLectureFiles").With(
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"seriesPartId", opts.SeriesPartID,
		"lectureId", opts.LectureID,
	)
	log.InfoContext(ctx, "Getting lecture files...")

	lecture, serviceErr := s.FindLectureBySlugsAndIDs(ctx, FindLectureOptions{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SeriesPartID: opts.SeriesPartID,
		LectureID:    opts.LectureID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	if opts.IsPublished && !lecture.IsPublished {
		log.WarnContext(ctx, "Cannot find files from unpublished lecture")
		return nil, NewValidationError("Cannot get file from published lecture")
	}

	lectureFiles, err := s.database.FindLectureFilesByLectureID(ctx, opts.LectureID)
	if err != nil {
		log.ErrorContext(ctx, "Error finding lecture files", "error", err)
		return nil, FromDBError(err)
	}

	log.InfoContext(ctx, "Lecture files found")
	return lectureFiles, nil
}
