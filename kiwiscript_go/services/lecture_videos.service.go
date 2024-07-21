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

func (s *Services) FindLectureVideoByLectureID(ctx context.Context, lectureID int32) (*db.LectureVideo, *ServiceError) {
	log := s.log.WithGroup("services.lectures.FindLectureVideoByLectureID").With("lectureId", lectureID)
	log.InfoContext(ctx, "Getting lecture video...")

	lectureVideo, err := s.database.GetLectureVideoByLectureID(ctx, lectureID)
	if err != nil {
		return nil, FromDBError(err)
	}

	return &lectureVideo, nil
}

type CreateLectureVideoOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SeriesPartID int32
	LectureID    int32
	Url          string
	WatchTime    int32
}

func (s *Services) CreateLectureVideo(
	ctx context.Context,
	opts CreateLectureVideoOptions,
) (*db.LectureVideo, *ServiceError) {
	log := s.log.WithGroup("services.lectures.CreateLectureVideo").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"seriesPartId", opts.SeriesPartID,
		"lectureId", opts.LectureID,
	)
	log.InfoContext(ctx, "Creating lecture video...")

	lecture, serviceErr := s.AssertLectureOwnership(ctx, AssertLectureOwnershipOptions{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SeriesPartID: opts.SeriesPartID,
		LectureID:    opts.LectureID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	if _, serviceErr := s.FindLectureVideoByLectureID(ctx, opts.LectureID); serviceErr == nil {
		return nil, NewDuplicateKeyError("Lecture video already exists")
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return nil, FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err)

	lectureVideo, err := qrs.CreateLectureVideo(ctx, db.CreateLectureVideoParams{
		LectureID:        lecture.ID,
		Url:              opts.Url,
		WatchTimeSeconds: opts.WatchTime,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to create lecture video", "error", err)
		return nil, FromDBError(err)
	}

	lecParams := db.UpdateLectureWatchTimeSecondsParams{
		ID:               lecture.ID,
		WatchTimeSeconds: opts.WatchTime,
	}
	if err := qrs.UpdateLectureWatchTimeSeconds(ctx, lecParams); err != nil {
		log.ErrorContext(ctx, "Failed to update lecture watch time", "error", err)
		return nil, FromDBError(err)
	}

	return &lectureVideo, nil
}

type UpdateLectureVideoOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SeriesPartID int32
	LectureID    int32
	Url          string
	WatchTime    int32
}

func (s *Services) UpdateLectureVideo(
	ctx context.Context,
	opts UpdateLectureVideoOptions,
) (*db.LectureVideo, *ServiceError) {
	log := s.log.WithGroup("services.lectures.UpdateLectureVideo").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"seriesPartId", opts.SeriesPartID,
		"lectureId", opts.LectureID,
	)
	log.InfoContext(ctx, "Updating lecture video...")

	lecture, serviceErr := s.AssertLectureOwnership(ctx, AssertLectureOwnershipOptions{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SeriesPartID: opts.SeriesPartID,
		LectureID:    opts.LectureID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	lectureVideo, serviceErr := s.FindLectureVideoByLectureID(ctx, opts.LectureID)
	if serviceErr != nil {
		return nil, serviceErr
	}

	oldWatchTime := lectureVideo.WatchTimeSeconds

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return nil, FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err)

	*lectureVideo, err = qrs.UpdateLectureVideo(ctx, db.UpdateLectureVideoParams{
		ID:               lectureVideo.ID,
		Url:              opts.Url,
		WatchTimeSeconds: opts.WatchTime,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to update lecture video", "error", err)
		return nil, FromDBError(err)
	}

	lecParams := db.UpdateLectureWatchTimeSecondsParams{
		ID:               lecture.ID,
		WatchTimeSeconds: opts.WatchTime,
	}
	if err := qrs.UpdateLectureWatchTimeSeconds(ctx, lecParams); err != nil {
		log.ErrorContext(ctx, "Failed to update lecture watch time", "error", err)
		return nil, FromDBError(err)
	}

	if lecture.IsPublished {
		watchTimeDiff := opts.WatchTime - oldWatchTime
		seriesParams := db.AddSeriesWatchTimeParams{
			Slug:             opts.SeriesSlug,
			WatchTimeSeconds: watchTimeDiff,
		}
		if err := qrs.AddSeriesWatchTime(ctx, seriesParams); err != nil {
			log.ErrorContext(ctx, "Failed to update series watch time", "error", err)
			return nil, FromDBError(err)
		}

		seriesPartParams := db.AddSeriesPartWatchTimeParams{
			ID:               opts.SeriesPartID,
			WatchTimeSeconds: watchTimeDiff,
		}
		if err := qrs.AddSeriesPartWatchTime(ctx, seriesPartParams); err != nil {
			log.ErrorContext(ctx, "Failed to update series part watch time", "error", err)
			return nil, FromDBError(err)
		}
	}

	return lectureVideo, nil
}

type DeleteLectureVideoOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SeriesPartID int32
	LectureID    int32
}

func (s *Services) DeleteLectureVideo(
	ctx context.Context,
	opts DeleteLectureVideoOptions,
) *ServiceError {
	log := s.log.WithGroup("services.lectures.DeleteLectureVideo").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"seriesPartId", opts.SeriesPartID,
		"lectureId", opts.LectureID,
	)
	log.InfoContext(ctx, "Deleting lecture video...")

	lecture, serviceErr := s.AssertLectureOwnership(ctx, AssertLectureOwnershipOptions{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SeriesPartID: opts.SeriesPartID,
		LectureID:    opts.LectureID,
	})
	if serviceErr != nil {
		return serviceErr
	}

	lectureVideo, serviceErr := s.FindLectureVideoByLectureID(ctx, opts.LectureID)
	if serviceErr != nil {
		return serviceErr
	}

	if lecture.IsPublished {
		log.WarnContext(ctx, "Cannot delete article from published lecture")
		return NewValidationError("Cannot delete video from published lecture")
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err)

	if err := qrs.DeleteLectureVideo(ctx, lectureVideo.ID); err != nil {
		log.ErrorContext(ctx, "Failed to delete lecture video", "error", err)
		return FromDBError(err)
	}

	lecParams := db.UpdateLectureWatchTimeSecondsParams{
		ID:               lecture.ID,
		WatchTimeSeconds: 0,
	}
	if err := qrs.UpdateLectureWatchTimeSeconds(ctx, lecParams); err != nil {
		log.ErrorContext(ctx, "Failed to update lecture watch time", "error", err)
		return FromDBError(err)
	}

	return nil
}

type FindLectureVideoOptions struct {
	LanguageSlug string
	SeriesSlug   string
	SeriesPartID int32
	LectureID    int32
	IsPublished  bool
}

func (s *Services) FindLectureVideo(
	ctx context.Context,
	opts FindLectureVideoOptions,
) (*db.LectureVideo, *ServiceError) {
	log := s.log.WithGroup("services.lectures.FindLectureVideo").With(
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"seriesPartId", opts.SeriesPartID,
		"lectureId", opts.LectureID,
	)
	log.InfoContext(ctx, "Getting lecture video...")

	lecture, serviceErr := s.FindLectureBySlugsAndIDs(ctx, FindLectureOptions{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SeriesPartID: opts.SeriesPartID,
		LectureID:    opts.LectureID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	lectureVideo, serviceErr := s.FindLectureVideoByLectureID(ctx, opts.LectureID)
	if serviceErr != nil {
		return nil, serviceErr
	}
	if opts.IsPublished && !lecture.IsPublished {
		log.WarnContext(ctx, "Cannot get unpublished video from published lecture")
		return nil, NewNotFoundError()
	}

	return lectureVideo, nil
}
