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
	"github.com/kiwiscript/kiwiscript_go/exceptions"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"log/slog"
)

const lessonProgressLocation string = "lesson_progress"

type FindLessonProgressOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	LessonID     int32
}

func (s *Services) FindLessonProgressBySlugsAndIDs(
	ctx context.Context,
	opts FindLessonProgressOptions,
) (*db.LessonProgress, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, lessonProgressLocation, "FindLessonProgressBySlugsAndIDs").With(
		"userID", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"seriesPartID", opts.SectionID,
		"lessonID", opts.LessonID,
	)
	log.InfoContext(ctx, "Finding lesson progress by slugs and IDs...")

	lessonProgress, err := s.database.FindLessonProgressBySlugsIDsAndUserID(
		ctx,
		db.FindLessonProgressBySlugsIDsAndUserIDParams{
			UserID:       opts.UserID,
			LanguageSlug: opts.LanguageSlug,
			SeriesSlug:   opts.SeriesSlug,
			SectionID:    opts.SectionID,
			LessonID:     opts.LessonID,
		},
	)
	if err != nil {
		log.ErrorContext(ctx, "Failed to find lesson progress", "error", err)
		return nil, exceptions.FromDBError(err)
	}

	return &lessonProgress, nil
}

type createLessonProgressOptions struct {
	UserID             int32
	LanguageProgressID int32
	SeriesProgressID   int32
	SectionProgressID  int32
	LanguageSlug       string
	SeriesSlug         string
	SectionID          int32
	LessonID           int32
}

func (s *Services) createLessonProgress(
	ctx context.Context,
	log *slog.Logger,
	opts createLessonProgressOptions,
) (*db.LessonProgress, *exceptions.ServiceError) {
	log.InfoContext(ctx, "Creating lesson progress...")

	lessonProgress, err := s.database.CreateLessonProgress(ctx, db.CreateLessonProgressParams{
		UserID:             opts.UserID,
		LanguageProgressID: opts.LanguageProgressID,
		SeriesProgressID:   opts.SeriesProgressID,
		SectionProgressID:  opts.SectionProgressID,
		LessonID:           opts.LessonID,
		LanguageSlug:       opts.LanguageSlug,
		SeriesSlug:         opts.SeriesSlug,
		SectionID:          opts.SectionID,
	})
	if err != nil {
		return nil, exceptions.FromDBError(err)
	}

	return &lessonProgress, nil
}

type CreateOrUpdateLessonProgressOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	LessonID     int32
}

func (s *Services) CreateOrUpdateLessonProgress(
	ctx context.Context,
	opts CreateOrUpdateLessonProgressOptions,
) (*db.Lesson, *db.LessonProgress, bool, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, lessonProgressLocation, "CreateOrUpdateLessonProgress").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"sectionId", opts.SectionID,
		"lessonId", opts.LessonID,
	)
	log.InfoContext(ctx, "Creating or updating lesson progress...")

	lesson, serviceErr := s.FindPublishedLessonBySlugsAndIDs(ctx, FindLessonOptions{
		RequestID:    opts.RequestID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SectionID:    opts.SectionID,
		LessonID:     opts.LessonID,
	})
	if serviceErr != nil {
		return nil, nil, false, serviceErr
	}

	sectionProgress, serviceErr := s.FindSectionProgressBySlugsAndID(ctx, FindSectionProgressBySlugsAndIDOptions{
		RequestID:    opts.RequestID,
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SectionID:    opts.SectionID,
	})
	if serviceErr != nil {
		return nil, nil, false, serviceErr
	}

	lessonProgress, serviceErr := s.FindLessonProgressBySlugsAndIDs(ctx, FindLessonProgressOptions{
		RequestID:    opts.RequestID,
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SectionID:    opts.SectionID,
		LessonID:     opts.LessonID,
	})
	if serviceErr != nil {
		lessonProgress, serviceErr := s.createLessonProgress(ctx, log, createLessonProgressOptions{
			UserID:             opts.UserID,
			LanguageSlug:       opts.LanguageSlug,
			SeriesSlug:         opts.SeriesSlug,
			SectionID:          opts.SectionID,
			LessonID:           opts.LessonID,
			LanguageProgressID: sectionProgress.LanguageProgressID,
			SeriesProgressID:   sectionProgress.SeriesProgressID,
			SectionProgressID:  sectionProgress.ID,
		})
		if serviceErr != nil {
			return nil, nil, false, serviceErr
		}

		return lesson, lessonProgress, true, nil
	}

	if err := s.database.UpdateLanguageProgressViewedAt(ctx, lessonProgress.ID); err != nil {
		log.ErrorContext(ctx, "Failed to update lesson progress viewed at", "error", err)
		return nil, nil, false, exceptions.FromDBError(err)
	}

	return lesson, lessonProgress, false, nil
}

type CompleteLessonProgressOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	LessonID     int32
}

func (s *Services) CompleteLessonProgress(
	ctx context.Context,
	opts CompleteLessonProgressOptions,
) (*db.Lesson, *db.LessonProgress, *db.Certificate, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, lessonProgressLocation, "CompleteLessonProgress").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"sectionId", opts.SectionID,
		"lessonId", opts.LessonID,
	)
	log.InfoContext(ctx, "Completing lesson progress...")

	lesson, serviceErr := s.FindPublishedLessonBySlugsAndIDs(ctx, FindLessonOptions{
		RequestID:    opts.RequestID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SectionID:    opts.SectionID,
		LessonID:     opts.LessonID,
	})
	if serviceErr != nil {
		return nil, nil, nil, serviceErr
	}

	lessonProgress, serviceErr := s.FindLessonProgressBySlugsAndIDs(ctx, FindLessonProgressOptions(opts))
	if serviceErr != nil {
		return nil, nil, nil, serviceErr
	}

	if lessonProgress.CompletedAt.Valid {
		log.InfoContext(ctx, "Lesson progress already completed")
		return lesson, lessonProgress, nil, nil
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return nil, nil, nil, exceptions.FromDBError(err)
	}
	defer func() {
		log.DebugContext(ctx, "Finalizing transaction")
		s.database.FinalizeTx(ctx, txn, err, serviceErr)
	}()

	*lessonProgress, err = qrs.CompleteLessonProgress(ctx, lessonProgress.ID)
	if err != nil {
		log.ErrorContext(ctx, "Failed to complete progress", "error", err)
		serviceErr = exceptions.FromDBError(err)
		return nil, nil, nil, serviceErr
	}

	sectionProgress, err := qrs.IncrementSectionProgressCompletedLessons(ctx, lessonProgress.SectionProgressID)
	if err != nil {
		log.ErrorContext(ctx, "Failed to increment section progress completed lessons", "error", err)
		serviceErr = exceptions.FromDBError(err)
		return nil, nil, nil, serviceErr
	}

	if sectionProgress.CompletedAt.Valid {
		seriesProgress, err := qrs.IncrementSeriesProgressCompletedSections(ctx, sectionProgress.SeriesProgressID)
		if err != nil {
			log.ErrorContext(ctx, "Failed to increment series progress completed sections", "error", err)
			serviceErr = exceptions.FromDBError(err)
			return nil, nil, nil, serviceErr
		}

		if seriesProgress.CompletedAt.Valid {
			series, err := qrs.FindPublishedSeriesBySlugAndLanguageSlug(
				ctx,
				db.FindPublishedSeriesBySlugAndLanguageSlugParams{
					Slug:         opts.SeriesSlug,
					LanguageSlug: opts.LanguageSlug,
				},
			)
			if err != nil {
				log.ErrorContext(ctx, "Published series not found")
				serviceErr = exceptions.FromDBError(err)
				return nil, nil, nil, serviceErr
			}

			certificate, err := qrs.FindCertificateByUserIDAndSeriesSlug(
				ctx,
				db.FindCertificateByUserIDAndSeriesSlugParams{
					UserID:     opts.UserID,
					SeriesSlug: opts.SeriesSlug,
				},
			)
			if err != nil {
				certificate, err = qrs.CreateCertificate(ctx, db.CreateCertificateParams{
					ID:               uuid.New(),
					UserID:           opts.UserID,
					LanguageSlug:     opts.LanguageSlug,
					SeriesSlug:       opts.SeriesSlug,
					SeriesTitle:      series.Title,
					Lessons:          series.LessonsCount,
					WatchTimeSeconds: series.WatchTimeSeconds,
					ReadTimeSeconds:  series.ReadTimeSeconds,
				})
				if err != nil {
					log.ErrorContext(ctx, "Failed to create certificate")
					serviceErr = exceptions.FromDBError(err)
					return nil, nil, nil, serviceErr
				}
			}

			return lesson, lessonProgress, &certificate, nil
		}

		return lesson, lessonProgress, nil, nil
	}

	if err := qrs.IncrementSeriesProgressCompletedLessons(ctx, sectionProgress.SeriesProgressID); err != nil {
		log.ErrorContext(ctx, "Failed to increment series progress completed lessons", "error", err)
		serviceErr = exceptions.FromDBError(err)
		return nil, nil, nil, serviceErr
	}

	return lesson, lessonProgress, nil, nil
}

type DeleteLessonProgressOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	LessonID     int32
}

func (s *Services) DeleteLessonProgress(
	ctx context.Context,
	opts DeleteLessonProgressOptions,
) *exceptions.ServiceError {
	log := s.buildLogger(opts.RequestID, lessonProgressLocation, "DeleteLessonProgress").With(
		"userID", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"seriesPartID", opts.SectionID,
		"lessonID", opts.LessonID,
	)
	log.InfoContext(ctx, "Deleting lesson progress")

	lessonProgress, serviceErr := s.FindLessonProgressBySlugsAndIDs(ctx, FindLessonProgressOptions{
		RequestID:    opts.RequestID,
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SectionID:    opts.SectionID,
		LessonID:     opts.LessonID,
	})
	if serviceErr != nil {
		log.InfoContext(ctx, "Lesson progress not found")
		return serviceErr
	}

	if lessonProgress.CompletedAt.Valid {
		qrs, txn, err := s.database.BeginTx(ctx)
		if err != nil {
			log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
			return exceptions.FromDBError(err)
		}
		defer func() {
			log.DebugContext(ctx, "Finalizing transaction")
			s.database.FinalizeTx(ctx, txn, err, serviceErr)
		}()

		if err := qrs.DeleteLessonProgress(ctx, lessonProgress.ID); err != nil {
			log.ErrorContext(ctx, "Failed to delete lesson progress")
			serviceErr = exceptions.FromDBError(err)
			return serviceErr
		}

		sectionProgress, err := qrs.FindSectionProgressByID(ctx, lessonProgress.SectionProgressID)
		if err != nil {
			log.ErrorContext(ctx, "Failed to find section progress")
			serviceErr = exceptions.FromDBError(err)
			return serviceErr
		}

		if err := qrs.DecrementSectionProgressCompletedLessons(ctx, sectionProgress.ID); err != nil {
			log.ErrorContext(ctx, "Failed to decrement section progress completed lessons")
			serviceErr = exceptions.FromDBError(err)
			return serviceErr
		}

		if sectionProgress.CompletedAt.Valid {
			seriesProgressOpts := db.DecrementSeriesProgressCompletedSectionsParams{
				CompletedLessons: 1,
				ID:               lessonProgress.SeriesProgressID,
			}
			if err := qrs.DecrementSeriesProgressCompletedSections(ctx, seriesProgressOpts); err != nil {
				log.ErrorContext(ctx, "Failed to decrement series' completed sections")
				serviceErr = exceptions.FromDBError(err)
				return serviceErr
			}
		} else {
			if err := qrs.DecrementSeriesProgressCompletedLessons(ctx, lessonProgress.SectionProgressID); err != nil {
				log.ErrorContext(ctx, "Failed to decrement series progress completed lessons")
				serviceErr = exceptions.FromDBError(err)
				return serviceErr
			}
		}

		log.InfoContext(ctx, "Delete lesson progress successfully")
		return nil
	}

	if err := s.database.DeleteLessonProgress(ctx, lessonProgress.ID); err != nil {
		log.ErrorContext(ctx, "Failed to delete lesson progress")
		serviceErr = exceptions.FromDBError(err)
		return serviceErr
	}

	log.InfoContext(ctx, "Delete lesson progress successfully")
	return nil
}
