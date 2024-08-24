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
	"log/slog"

	db "github.com/kiwiscript/kiwiscript_go/providers/database"
)

type CreateLessonOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	Title        string
}

func (s *Services) CreateLesson(ctx context.Context, opts CreateLessonOptions) (*db.Lesson, *ServiceError) {
	log := s.log.WithGroup("services.lessons.CreateLesson").With("title", opts.Title)
	log.InfoContext(ctx, "Creating lessons...")

	servicePart, serviceErr := s.AssertSectionOwnership(ctx, AssertSectionOwnershipOptions{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SectionID:    opts.SectionID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	lesson, err := s.database.CreateLesson(ctx, db.CreateLessonParams{
		SectionID:    servicePart.ID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		Title:        opts.Title,
		AuthorID:     opts.UserID,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to create lesson", "error", err)
		return nil, FromDBError(err)
	}

	log.InfoContext(ctx, "Lesson created successfully")
	return &lesson, nil
}

func (s *Services) FindLessonsBySectionID(ctx context.Context, seriesPartID int32) ([]db.Lesson, *ServiceError) {
	log := s.log.WithGroup("services.lessons.FindLessonsBySectionID").With("series_part_id", seriesPartID)
	log.InfoContext(ctx, "Finding lessons by series part ID...")

	lessons, err := s.database.FindLessonsBySectionID(ctx, seriesPartID)
	if err != nil {
		log.ErrorContext(ctx, "Failed to find lessons", "error", err)
		return nil, FromDBError(err)
	}

	log.InfoContext(ctx, "Lessons found successfully")
	return lessons, nil
}

type FindLessonOptions struct {
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	LessonID     int32
}

func (s *Services) FindLessonBySlugsAndIDs(ctx context.Context, opts FindLessonOptions) (*db.Lesson, *ServiceError) {
	log := s.
		log.
		WithGroup("services.lessons.FindLessonBySlugsAndIDs").
		With("lesson_id", opts.LessonID, "series_part_id", opts.SectionID)
	log.InfoContext(ctx, "Finding lesson...")

	lesson, err := s.database.FindLessonBySlugsAndIDs(ctx, db.FindLessonBySlugsAndIDsParams{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SectionID:    opts.SectionID,
		ID:           opts.LessonID,
	})
	if err != nil {
		log.WarnContext(ctx, "Failed to find lesson", "error", err)
		return nil, FromDBError(err)
	}

	log.InfoContext(ctx, "Lesson found successfully")
	return &lesson, nil
}

// TODO: fix by getting the series and sections first and check if it published

func (s *Services) FindPublishedLessonBySlugsAndIDs(
	ctx context.Context,
	opts FindLessonOptions,
) (*db.Lesson, *ServiceError) {
	log := s.
		log.
		WithGroup("services.lessons.FindPublishedLessonBySlugsAndIDs").
		With(
			"languageSlug", opts.LanguageSlug,
			"seriesSlug", opts.SeriesSlug,
			"sectionId", opts.SectionID,
			"lessonId", opts.LessonID,
		)
	log.InfoContext(ctx, "Finding published lesson...")

	sectionOpts := FindSectionBySlugsAndIDOptions{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SectionID:    opts.SectionID,
	}
	if _, serviceErr := s.FindPublishedSectionBySlugsAndID(ctx, sectionOpts); serviceErr != nil {
		return nil, serviceErr
	}

	lesson, serviceErr := s.FindLessonBySlugsAndIDs(ctx, opts)
	if serviceErr != nil {
		return nil, serviceErr
	}

	if !lesson.IsPublished {
		log.InfoContext(ctx, "Lesson is not published")
		return nil, NewNotFoundError()
	}

	return lesson, nil
}

type FindPaginatedLessonsOptions struct {
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	Offset       int32
	Limit        int32
}

func (s *Services) FindPaginatedLessons(
	ctx context.Context,
	opts FindPaginatedLessonsOptions,
) ([]db.Lesson, int64, *ServiceError) {
	log := s.log.WithGroup("services.lessons.FindLessons").With("series_slug", opts.SeriesSlug)
	log.InfoContext(ctx, "Finding lessons...")

	count, err := s.database.CountLessonsBySectionID(ctx, opts.SectionID)
	if err != nil {
		log.ErrorContext(ctx, "Failed to count lessons", "error", err)
		return nil, 0, FromDBError(err)
	}
	if count == 0 {
		return make([]db.Lesson, 0), 0, nil
	}

	lessons, err := s.database.FindPaginatedLessonsBySlugsAndSectionID(
		ctx,
		db.FindPaginatedLessonsBySlugsAndSectionIDParams{
			LanguageSlug: opts.LanguageSlug,
			SeriesSlug:   opts.SeriesSlug,
			SectionID:    opts.SectionID,
			Offset:       opts.Offset,
			Limit:        opts.Limit,
		},
	)
	if err != nil {
		log.ErrorContext(ctx, "Failed to find lessons", "error", err)
		return nil, 0, FromDBError(err)
	}

	log.InfoContext(ctx, "Lessons found successfully")
	return lessons, count, nil
}

type FindPaginatedPublishedLessonsWithProgressOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	Offset       int32
	Limit        int32
}

func (s *Services) findPublishedLessonsCount(
	ctx context.Context,
	log *slog.Logger,
	languageSlug,
	seriesSlug string,
	sectionID int32,
) (int64, *ServiceError) {
	seriesOpts := FindSectionBySlugsAndIDOptions{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SectionID:    sectionID,
	}
	if _, serviceErr := s.FindPublishedSectionBySlugsAndID(ctx, seriesOpts); serviceErr != nil {
		log.WarnContext(ctx, "Section not found", "error", serviceErr)
		return 0, nil
	}

	count, err := s.database.CountPublishedLessonsBySectionID(ctx, sectionID)
	if err != nil {
		log.ErrorContext(ctx, "Error count published lessons", "error", err)
		return 0, FromDBError(err)
	}

	return count, nil
}

func (s *Services) FindPaginatedPublishedLessonsWithProgress(
	ctx context.Context,
	opts FindPaginatedPublishedLessonsWithProgressOptions,
) ([]db.FindPaginatedPublishedLessonsBySlugsAndSectionIDWithProgressRow, int64, *ServiceError) {
	log := s.log.WithGroup("services.lessons.FindPaginatedPublishedLessonsWithProgress").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"sectionId", opts.SectionID,
		"offset", opts.Offset,
		"limit", opts.Limit,
	)
	log.InfoContext(ctx, "Finding paginated published lessons with progress...")

	count, serviceErr := s.findPublishedLessonsCount(ctx, log, opts.LanguageSlug, opts.SeriesSlug, opts.SectionID)
	if serviceErr != nil {
		return nil, 0, serviceErr
	}
	if count == 0 {
		return make([]db.FindPaginatedPublishedLessonsBySlugsAndSectionIDWithProgressRow, 0), 0, nil
	}

	lessons, err := s.database.FindPaginatedPublishedLessonsBySlugsAndSectionIDWithProgress(
		ctx,
		db.FindPaginatedPublishedLessonsBySlugsAndSectionIDWithProgressParams{
			UserID:       opts.UserID,
			LanguageSlug: opts.LanguageSlug,
			SeriesSlug:   opts.SeriesSlug,
			SectionID:    opts.SectionID,
			Offset:       opts.Offset,
			Limit:        opts.Limit,
		},
	)
	if err != nil {
		log.ErrorContext(ctx, "Failed to find paginated published lessons with progress", "error", err)
		return nil, 0, FromDBError(err)
	}

	return lessons, count, nil
}

func (s *Services) FindPaginatedPublishedLessons(
	ctx context.Context,
	opts FindPaginatedLessonsOptions,
) ([]db.Lesson, int64, *ServiceError) {
	log := s.log.WithGroup("services.lessons.FindPaginatedPublishedLessonsWithProgress").With(
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"sectionId", opts.SectionID,
		"offset", opts.Offset,
		"limit", opts.Limit,
	)
	log.InfoContext(ctx, "Finding paginated published lessons...")

	count, serviceErr := s.findPublishedLessonsCount(ctx, log, opts.LanguageSlug, opts.SeriesSlug, opts.SectionID)
	if serviceErr != nil {
		return nil, 0, serviceErr
	}
	if count == 0 {
		return make([]db.Lesson, 0), 0, nil
	}

	lessons, err := s.database.FindPaginatedPublishedLessonsBySlugsAndSectionID(
		ctx,
		db.FindPaginatedPublishedLessonsBySlugsAndSectionIDParams{
			LanguageSlug: opts.LanguageSlug,
			SeriesSlug:   opts.SeriesSlug,
			SectionID:    opts.SectionID,
			Offset:       opts.Offset,
			Limit:        opts.Limit,
		},
	)
	if err != nil {
		log.ErrorContext(ctx, "Failed to find paginated published lessons", "error", err)
		return nil, 0, FromDBError(err)
	}

	return lessons, count, nil
}

func (s *Services) FindLessonWithArticleAndVideo(
	ctx context.Context,
	opts FindLessonOptions,
) (*db.FindLessonBySlugsAndIDsWithArticleAndVideoRow, *ServiceError) {
	log := s.log.WithGroup("services.lessons.FindLessonWithArticleAndVideo").With(
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"sectionId", opts.SectionID,
		"lessonId", opts.LessonID,
	)
	log.InfoContext(ctx, "Finding lesson with article and video...")

	lesson, err := s.database.FindLessonBySlugsAndIDsWithArticleAndVideo(
		ctx,
		db.FindLessonBySlugsAndIDsWithArticleAndVideoParams{
			LanguageSlug: opts.LanguageSlug,
			SeriesSlug:   opts.SeriesSlug,
			SectionID:    opts.SectionID,
			ID:           opts.LessonID,
		},
	)
	if err != nil {
		log.WarnContext(ctx, "Lesson not found", "error", err)
		return nil, FromDBError(err)
	}

	return &lesson, nil
}

func (s *Services) FindPublishedLessonWithArticleAndVideo(
	ctx context.Context,
	opts FindLessonOptions,
) (*db.FindLessonBySlugsAndIDsWithArticleAndVideoRow, *ServiceError) {
	log := s.log.WithGroup("services.lessons.FindPublishedLessonWithArticleAndVideo").With(
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"sectionId", opts.SectionID,
		"lessonId", opts.LessonID,
	)
	log.InfoContext(ctx, "Finding published lesson with article and video...")

	sectionOpts := FindSectionBySlugsAndIDOptions{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SectionID:    opts.SectionID,
	}
	if _, serviceErr := s.FindPublishedSectionBySlugsAndID(ctx, sectionOpts); serviceErr != nil {
		return nil, serviceErr
	}

	lesson, serviceErr := s.FindLessonWithArticleAndVideo(ctx, opts)
	if serviceErr != nil {
		return nil, serviceErr
	}

	if !lesson.IsPublished {
		log.WarnContext(ctx, "Lesson is not published")
		return nil, NewNotFoundError()
	}

	return lesson, nil
}

type FindLessonWithProgressOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	LessonID     int32
}

func (s *Services) FindPublishedLessonWithProgressArticleAndVideo(
	ctx context.Context,
	opts FindLessonWithProgressOptions,
) (*db.FindPublishedLessonBySlugsAndIDsWithProgressArticleAndVideoRow, *ServiceError) {
	log := s.log.WithGroup("services.lessons.FindPublishedLessonWithProgressArticleAndVideo").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"sectionId", opts.SectionID,
		"lessonId", opts.LessonID,
	)
	log.InfoContext(ctx, "Finding lesson with article and video...")

	lesson, err := s.database.FindPublishedLessonBySlugsAndIDsWithProgressArticleAndVideo(
		ctx,
		db.FindPublishedLessonBySlugsAndIDsWithProgressArticleAndVideoParams{
			UserID:       opts.UserID,
			LanguageSlug: opts.LanguageSlug,
			SeriesSlug:   opts.SeriesSlug,
			SectionID:    opts.SectionID,
			ID:           opts.LessonID,
		},
	)
	if err != nil {
		log.WarnContext(ctx, "Lesson not found", "error", err)
		return nil, FromDBError(err)
	}

	return &lesson, nil
}

type AssertLessonOwnershipOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	LessonID     int32
}

func (s *Services) AssertLessonOwnership(ctx context.Context, opts AssertLessonOwnershipOptions) (*db.Lesson, *ServiceError) {
	log := s.log.WithGroup("services.lessons.AssertLessonOwnership").With("lesson_id", opts.LessonID)
	log.InfoContext(ctx, "Asserting lesson ownership...")

	lesson, serviceErr := s.FindLessonBySlugsAndIDs(ctx, FindLessonOptions{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SectionID:    opts.SectionID,
		LessonID:     opts.LessonID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	if lesson.AuthorID != opts.UserID {
		log.Warn("User is not the author of the series")
		return nil, NewForbiddenError()
	}

	return lesson, nil
}

type UpdateLessonOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	LessonID     int32
	Title        string
	Position     int16
}

func (s *Services) UpdateLesson(ctx context.Context, opts UpdateLessonOptions) (*db.Lesson, *ServiceError) {
	log := s.log.WithGroup("services.lessons.UpdateLessons").With("title", opts.Title)
	log.InfoContext(ctx, "Updating lessons...")

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

	if opts.Position == 0 || lesson.Position == opts.Position {
		var err error
		*lesson, err = s.database.UpdateLesson(ctx, db.UpdateLessonParams{
			ID:    lesson.ID,
			Title: opts.Title,
		})
		if err != nil {
			log.ErrorContext(ctx, "Failed to update lesson", "error", err)
			return nil, FromDBError(err)
		}

		log.InfoContext(ctx, "Lesson updated successfully")
		return lesson, nil
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return nil, FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err, serviceErr)

	oldPosition := lesson.Position
	*lesson, err = qrs.UpdateLessonWithPosition(ctx, db.UpdateLessonWithPositionParams{
		ID:       lesson.ID,
		Title:    opts.Title,
		Position: opts.Position,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to update lesson with position", "error", err)
		return nil, FromDBError(err)
	}

	if oldPosition < opts.Position {
		params := db.DecrementLessonPositionParams{
			SectionID:  opts.SectionID,
			Position:   oldPosition,
			Position_2: opts.Position,
		}
		if err := qrs.DecrementLessonPosition(ctx, params); err != nil {
			log.ErrorContext(ctx, "Failed to decrement lesson position", "error", err)
			return nil, FromDBError(err)
		}
	} else {
		params := db.IncrementLessonPositionParams{
			SectionID:  opts.SectionID,
			Position:   opts.Position,
			Position_2: oldPosition,
		}
		if err := qrs.IncrementLessonPosition(ctx, params); err != nil {
			log.ErrorContext(ctx, "Failed to increment lesson position", "error", err)
			return nil, FromDBError(err)
		}
	}

	log.InfoContext(ctx, "Lesson updated successfully")
	return lesson, nil
}

type DeleteLessonOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	LessonID     int32
}

func (s *Services) DeleteLesson(ctx context.Context, opts DeleteLessonOptions) *ServiceError {
	log := s.log.WithGroup("services.lessons.DeleteLessons").With("lesson_id", opts.LessonID)
	log.InfoContext(ctx, "Deleting lessons...")

	lesson, serviceErr := s.AssertLessonOwnership(ctx, AssertLessonOwnershipOptions(opts))
	if serviceErr != nil {
		return serviceErr
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err, serviceErr)

	if err := qrs.DeleteLessonByID(ctx, lesson.ID); err != nil {
		log.ErrorContext(ctx, "Failed to delete lesson", "error", err)
		return FromDBError(err)
	}

	params := db.DecrementLessonPositionParams{
		SectionID:  opts.SectionID,
		Position:   lesson.Position,
		Position_2: 1,
	}
	if err := qrs.DecrementLessonPosition(ctx, params); err != nil {
		log.ErrorContext(ctx, "Failed to decrement lesson position", "error", err)
		return FromDBError(err)
	}

	log.InfoContext(ctx, "Lesson deleted successfully")
	return nil
}

type UpdateLessonIsPublishedOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	LessonID     int32
	IsPublished  bool
}

func (s *Services) UpdateLessonIsPublished(ctx context.Context, opts UpdateLessonIsPublishedOptions) (*db.Lesson, *ServiceError) {
	log := s.log.WithGroup("services.lessons.UpdateLessonIsPublished").With("lesson_id", opts.LessonID)
	log.InfoContext(ctx, "Updating lesson article is published...")

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

	if lesson.IsPublished == opts.IsPublished {
		log.InfoContext(ctx, "Lesson article is already published")
		return lesson, nil
	}
	if opts.IsPublished && lesson.ReadTimeSeconds == 0 && lesson.WatchTimeSeconds == 0 {
		log.WarnContext(ctx, "Cannot publish lesson article without read time or watch time")
		return nil, NewValidationError("Cannot publish lesson without content")
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return nil, FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err, serviceErr)

	*lesson, err = qrs.UpdateLessonIsPublished(ctx, db.UpdateLessonIsPublishedParams{
		ID:          lesson.ID,
		IsPublished: opts.IsPublished,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to update lesson article is published", "error", err)
		return nil, FromDBError(err)
	}

	if opts.IsPublished {
		incPartParams := db.IncrementSectionLessonsCountParams{
			ID:               opts.SectionID,
			WatchTimeSeconds: lesson.WatchTimeSeconds,
			ReadTimeSeconds:  lesson.ReadTimeSeconds,
		}
		if err := qrs.IncrementSectionLessonsCount(ctx, incPartParams); err != nil {
			log.ErrorContext(ctx, "Failed to increment series part lessons count", "error", err)
			return nil, FromDBError(err)
		}

		incSeriesParams := db.IncrementSeriesLessonsCountParams{
			Slug:             opts.SeriesSlug,
			WatchTimeSeconds: lesson.WatchTimeSeconds,
			ReadTimeSeconds:  lesson.ReadTimeSeconds,
		}
		if err := qrs.IncrementSeriesLessonsCount(ctx, incSeriesParams); err != nil {
			log.ErrorContext(ctx, "Failed to increment series lessons count", "error", err)
			return nil, FromDBError(err)
		}
	} else {
		decPartParams := db.DecrementSectionLessonsCountParams{
			ID:               opts.SectionID,
			WatchTimeSeconds: lesson.WatchTimeSeconds,
			ReadTimeSeconds:  lesson.ReadTimeSeconds,
		}
		if err := qrs.DecrementSectionLessonsCount(ctx, decPartParams); err != nil {
			log.ErrorContext(ctx, "Failed to decrement series part lessons count", "error", err)
			return nil, FromDBError(err)
		}

		decSeriesParams := db.DecrementSeriesLessonsCountParams{
			Slug:             opts.SeriesSlug,
			WatchTimeSeconds: lesson.WatchTimeSeconds,
			ReadTimeSeconds:  lesson.ReadTimeSeconds,
		}
		if err := qrs.DecrementSeriesLessonsCount(ctx, decSeriesParams); err != nil {
			log.ErrorContext(ctx, "Failed to decrement series lessons count", "error", err)
			return nil, FromDBError(err)
		}
	}

	return lesson, nil
}
