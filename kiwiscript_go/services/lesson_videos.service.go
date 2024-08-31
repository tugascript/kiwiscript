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

const lessonVideosLocation string = "lesson_videos"

type FindLessonVideoByLessonIDOptions struct {
	RequestID string
	LessonID  int32
}

func (s *Services) FindLessonVideoByLessonID(
	ctx context.Context,
	opts FindLessonVideoByLessonIDOptions,
) (*db.LessonVideo, *ServiceError) {
	log := s.buildLogger(opts.RequestID, lessonVideosLocation, "FindLessonVideoByLessonID").With(
		"lessonId", opts.LessonID,
	)
	log.InfoContext(ctx, "Getting lesson video...")

	lessonVideo, err := s.database.GetLessonVideoByLessonID(ctx, opts.LessonID)
	if err != nil {
		return nil, FromDBError(err)
	}

	return &lessonVideo, nil
}

type CreateLessonVideoOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	LessonID     int32
	Url          string
	WatchTime    int32
}

func (s *Services) CreateLessonVideo(
	ctx context.Context,
	opts CreateLessonVideoOptions,
) (*db.LessonVideo, *ServiceError) {
	log := s.buildLogger(opts.RequestID, lessonVideosLocation, "CreateLessonVideo").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"sectionId", opts.SectionID,
		"lessonId", opts.LessonID,
	)
	log.InfoContext(ctx, "Creating lesson video...")

	lesson, serviceErr := s.AssertLessonOwnership(ctx, AssertLessonOwnershipOptions{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SectionID:    opts.SectionID,
		LessonID:     opts.LessonID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	findOpts := FindLessonVideoByLessonIDOptions{
		RequestID: opts.RequestID,
		LessonID:  opts.LessonID,
	}
	if _, serviceErr := s.FindLessonVideoByLessonID(ctx, findOpts); serviceErr == nil {
		return nil, NewConflictError("Lesson video already exists")
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return nil, FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err, serviceErr)

	lessonVideo, err := qrs.CreateLessonVideo(ctx, db.CreateLessonVideoParams{
		LessonID:         lesson.ID,
		Url:              opts.Url,
		WatchTimeSeconds: opts.WatchTime,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to create lesson video", "error", err)
		return nil, FromDBError(err)
	}

	lecParams := db.UpdateLessonWatchTimeSecondsParams{
		ID:               lesson.ID,
		WatchTimeSeconds: opts.WatchTime,
	}
	if err := qrs.UpdateLessonWatchTimeSeconds(ctx, lecParams); err != nil {
		log.ErrorContext(ctx, "Failed to update lesson watch time", "error", err)
		return nil, FromDBError(err)
	}

	return &lessonVideo, nil
}

type UpdateLessonVideoOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	LessonID     int32
	Url          string
	WatchTime    int32
}

func (s *Services) UpdateLessonVideo(
	ctx context.Context,
	opts UpdateLessonVideoOptions,
) (*db.LessonVideo, *ServiceError) {
	log := s.buildLogger(opts.RequestID, lessonVideosLocation, "UpdateLessonVideo").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"sectionId", opts.SectionID,
		"lessonId", opts.LessonID,
		"url", opts.Url,
		"watchTime", opts.WatchTime,
	)
	log.InfoContext(ctx, "Updating lesson video...")

	lesson, serviceErr := s.AssertLessonOwnership(ctx, AssertLessonOwnershipOptions{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SectionID:    opts.SectionID,
		LessonID:     opts.LessonID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	lessonVideo, serviceErr := s.FindLessonVideoByLessonID(ctx, FindLessonVideoByLessonIDOptions{
		RequestID: opts.RequestID,
		LessonID:  opts.LessonID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	oldWatchTime := lessonVideo.WatchTimeSeconds

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return nil, FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err, serviceErr)

	*lessonVideo, err = qrs.UpdateLessonVideo(ctx, db.UpdateLessonVideoParams{
		ID:               lessonVideo.ID,
		Url:              opts.Url,
		WatchTimeSeconds: opts.WatchTime,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to update lesson video", "error", err)
		return nil, FromDBError(err)
	}

	lecParams := db.UpdateLessonWatchTimeSecondsParams{
		ID:               lesson.ID,
		WatchTimeSeconds: opts.WatchTime,
	}
	if err := qrs.UpdateLessonWatchTimeSeconds(ctx, lecParams); err != nil {
		log.ErrorContext(ctx, "Failed to update lesson watch time", "error", err)
		return nil, FromDBError(err)
	}

	if lesson.IsPublished {
		watchTimeDiff := opts.WatchTime - oldWatchTime
		seriesParams := db.AddSeriesWatchTimeParams{
			Slug:             opts.SeriesSlug,
			WatchTimeSeconds: watchTimeDiff,
		}
		if err := qrs.AddSeriesWatchTime(ctx, seriesParams); err != nil {
			log.ErrorContext(ctx, "Failed to update series watch time", "error", err)
			return nil, FromDBError(err)
		}

		sectionParams := db.AddSectionWatchTimeParams{
			ID:               opts.SectionID,
			WatchTimeSeconds: watchTimeDiff,
		}
		if err := qrs.AddSectionWatchTime(ctx, sectionParams); err != nil {
			log.ErrorContext(ctx, "Failed to update series part watch time", "error", err)
			return nil, FromDBError(err)
		}
	}

	return lessonVideo, nil
}

type DeleteLessonVideoOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	LessonID     int32
}

func (s *Services) DeleteLessonVideo(
	ctx context.Context,
	opts DeleteLessonVideoOptions,
) *ServiceError {
	log := s.buildLogger(opts.RequestID, lessonVideosLocation, "DeleteLessonVideo").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"sectionId", opts.SectionID,
		"lessonId", opts.LessonID,
	)
	log.InfoContext(ctx, "Deleting lesson video...")

	lesson, serviceErr := s.AssertLessonOwnership(ctx, AssertLessonOwnershipOptions{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SectionID:    opts.SectionID,
		LessonID:     opts.LessonID,
	})
	if serviceErr != nil {
		return serviceErr
	}

	lessonVideo, serviceErr := s.FindLessonVideoByLessonID(ctx, FindLessonVideoByLessonIDOptions{
		RequestID: opts.RequestID,
		LessonID:  opts.LessonID,
	})
	if serviceErr != nil {
		return serviceErr
	}

	if lesson.IsPublished {
		log.WarnContext(ctx, "Cannot delete article from published lesson")
		return NewValidationError("Cannot delete video from published lesson")
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err, serviceErr)

	if err := qrs.DeleteLessonVideo(ctx, lessonVideo.ID); err != nil {
		log.ErrorContext(ctx, "Failed to delete lesson video", "error", err)
		return FromDBError(err)
	}

	lecParams := db.UpdateLessonWatchTimeSecondsParams{
		ID:               lesson.ID,
		WatchTimeSeconds: 0,
	}
	if err := qrs.UpdateLessonWatchTimeSeconds(ctx, lecParams); err != nil {
		log.ErrorContext(ctx, "Failed to update lesson watch time", "error", err)
		return FromDBError(err)
	}

	return nil
}

type FindLessonVideoOptions struct {
	RequestID    string
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	LessonID     int32
	IsPublished  bool
}

func (s *Services) FindLessonVideo(
	ctx context.Context,
	opts FindLessonVideoOptions,
) (*db.LessonVideo, *ServiceError) {
	log := s.buildLogger(opts.RequestID, lessonVideosLocation, "FindLessonVideo").With(
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"sectionId", opts.SectionID,
		"lessonId", opts.LessonID,
	)
	log.InfoContext(ctx, "Getting lesson video...")

	lesson, serviceErr := s.FindLessonBySlugsAndIDs(ctx, FindLessonOptions{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SectionID:    opts.SectionID,
		LessonID:     opts.LessonID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	lessonVideo, serviceErr := s.FindLessonVideoByLessonID(ctx, FindLessonVideoByLessonIDOptions{
		RequestID: opts.RequestID,
		LessonID:  opts.LessonID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}
	if opts.IsPublished && !lesson.IsPublished {
		log.WarnContext(ctx, "Cannot get unpublished video from published lesson")
		return nil, NewNotFoundError()
	}

	return lessonVideo, nil
}
