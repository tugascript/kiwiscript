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

type UploadLessonFileOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	LessonID     int32
	Name         string
	FileHeader   *multipart.FileHeader
}

func (s *Services) UploadLessonFile(
	ctx context.Context,
	opts UploadLessonFileOptions,
) (*db.LessonFile, *ServiceError) {
	log := s.
		log.
		WithGroup("services.lessonFiles.UploadLessonFile").
		With(
			"userId", opts.UserID,
			"languageSlug", opts.LanguageSlug,
			"seriesSlug", opts.SeriesSlug,
			"sectionId", opts.SectionID,
			"lessonId", opts.LessonID,
			"name", opts.Name,
		)
	log.InfoContext(ctx, "Uploading lesson file...")

	lessonOpts := AssertLessonOwnershipOptions{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SectionID:    opts.SectionID,
		LessonID:     opts.LessonID,
	}
	if _, serviceErr := s.AssertLessonOwnership(ctx, lessonOpts); serviceErr != nil {
		return nil, serviceErr
	}

	fileId, docExt, err := s.objStg.UploadDocument(ctx, opts.UserID, opts.FileHeader)
	if err != nil {
		log.ErrorContext(ctx, "Error uploading document", "error", err)
		return nil, NewServerError()
	}

	lessonFile, err := s.database.CreateLessonFile(ctx, db.CreateLessonFileParams{
		ID:       fileId,
		AuthorID: opts.UserID,
		LessonID: opts.LessonID,
		Ext:      docExt,
		Name:     opts.Name,
	})
	if err != nil {
		log.ErrorContext(ctx, "Error creating lesson file", "error", err)
		return nil, FromDBError(err)
	}

	return &lessonFile, nil
}

type DeleteLessonFileOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	LessonID     int32
	File         uuid.UUID
}

func (s *Services) DeleteLessonFile(
	ctx context.Context,
	opts DeleteLessonFileOptions,
) *ServiceError {
	log := s.
		log.
		WithGroup("services.lessonFiles.DeleteLessonFile").
		With("file", opts.File.String())
	log.InfoContext(ctx, "Deleting lesson file...")

	lessonOpts := AssertLessonOwnershipOptions{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SectionID:    opts.SectionID,
		LessonID:     opts.LessonID,
	}
	if _, serviceErr := s.AssertLessonOwnership(ctx, lessonOpts); serviceErr != nil {
		return serviceErr
	}

	lessonFile, err := s.database.FindLessonFileByIDAndLessonID(ctx, db.FindLessonFileByIDAndLessonIDParams{
		ID:       opts.File,
		LessonID: opts.LessonID,
	})
	if err != nil {
		log.WarnContext(ctx, "Error finding lesson file", "error", err)
		return FromDBError(err)
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err, nil)

	if err := qrs.DeleteLessonFile(ctx, lessonFile.ID); err != nil {
		log.ErrorContext(ctx, "Failed to delete lesson file", "error", err)
		return FromDBError(err)
	}
	if err := s.objStg.DeleteFile(ctx, opts.UserID, opts.File, lessonFile.Ext); err != nil {
		log.ErrorContext(ctx, "Failed to delete file from object storage", "error", err)
		return NewServerError()
	}

	return nil
}

type FindLessonFileOptions struct {
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	LessonID     int32
	File         uuid.UUID
	IsPublished  bool
}

func (s *Services) FindLessonFile(
	ctx context.Context,
	opts FindLessonFileOptions,
) (*db.LessonFile, *ServiceError) {
	log := s.
		log.
		WithGroup("services.lessonFiles.GetLessonFile").
		With("file", opts.File.String())
	log.InfoContext(ctx, "Getting lesson file...")

	lesson, serviceErr := s.FindLessonBySlugsAndIDs(ctx, FindLessonOptions{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SectionID:    opts.SectionID,
		LessonID:     opts.LessonID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	if opts.IsPublished && !lesson.IsPublished {
		log.WarnContext(ctx, "Cannot find file from unpublished lesson")
		return nil, NewValidationError("Cannot get file from published lesson")
	}

	lessonFile, err := s.database.FindLessonFileByIDAndLessonID(ctx, db.FindLessonFileByIDAndLessonIDParams{
		ID:       opts.File,
		LessonID: opts.LessonID,
	})
	if err != nil {
		log.WarnContext(ctx, "Error finding lesson file", "error", err)
		return nil, FromDBError(err)
	}

	return &lessonFile, nil
}

type FindLessonFilesOptions struct {
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	LessonID     int32
	IsPublished  bool
}

func (s *Services) FindLessonFiles(
	ctx context.Context,
	opts FindLessonFilesOptions,
) ([]db.LessonFile, *ServiceError) {
	log := s.log.WithGroup("services.lessonFiles.FindLessonFiles").With(
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"seriesPartId", opts.SectionID,
		"lessonId", opts.LessonID,
	)
	log.InfoContext(ctx, "Getting lesson files...")

	lesson, serviceErr := s.FindLessonBySlugsAndIDs(ctx, FindLessonOptions{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SectionID:    opts.SectionID,
		LessonID:     opts.LessonID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	if opts.IsPublished && !lesson.IsPublished {
		log.WarnContext(ctx, "Cannot find files from unpublished lesson")
		return nil, NewValidationError("Cannot get file from published lesson")
	}

	lessonFiles, err := s.database.FindLessonFilesByLessonID(ctx, opts.LessonID)
	if err != nil {
		log.ErrorContext(ctx, "Error finding lesson files", "error", err)
		return nil, FromDBError(err)
	}

	log.InfoContext(ctx, "Lesson files found")
	return lessonFiles, nil
}

func (s *Services) FindLessonFilesWithNoCheck(ctx context.Context, lessonID int32) ([]db.LessonFile, *ServiceError) {
	log := s.log.WithGroup("services_lessonFiles_FindLessonFilesWithNoCheck").With("lessonId", lessonID)
	log.InfoContext(ctx, "Finding lesson files...")

	lessonFiles, err := s.database.FindLessonFilesByLessonID(ctx, lessonID)
	if err != nil {
		log.ErrorContext(ctx, "Error finding lesson files", "error", err)
		return nil, FromDBError(err)
	}

	log.InfoContext(ctx, "Lesson files found")
	return lessonFiles, nil
}

type UpdateLessonFileOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	LessonID     int32
	File         uuid.UUID
	Name         string
}

func (s *Services) UpdateLessonFile(
	ctx context.Context,
	opts UpdateLessonFileOptions,
) (*db.LessonFile, *ServiceError) {
	log := s.
		log.
		WithGroup("services.lessonFiles.UpdateLessonFile").
		With("file", opts.File.String())
	log.InfoContext(ctx, "Updating lesson file...")

	lessonOpts := AssertLessonOwnershipOptions{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SectionID:    opts.SectionID,
		LessonID:     opts.LessonID,
	}
	if _, serviceErr := s.AssertLessonOwnership(ctx, lessonOpts); serviceErr != nil {
		return nil, serviceErr
	}

	lessonFile, err := s.database.FindLessonFileByIDAndLessonID(ctx, db.FindLessonFileByIDAndLessonIDParams{
		ID:       opts.File,
		LessonID: opts.LessonID,
	})
	if err != nil {
		log.WarnContext(ctx, "Error finding lesson file", "error", err)
		return nil, FromDBError(err)
	}

	lessonFile, err = s.database.UpdateLessonFile(ctx, db.UpdateLessonFileParams{
		ID:   lessonFile.ID,
		Name: opts.Name,
	})
	if err != nil {
		log.ErrorContext(ctx, "Error updating lesson file", "error", err)
		return nil, FromDBError(err)
	}

	return &lessonFile, nil
}
