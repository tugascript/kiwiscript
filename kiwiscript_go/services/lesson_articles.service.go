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
	"math"
	"strings"

	db "github.com/kiwiscript/kiwiscript_go/providers/database"
)

const lessonArticlesLocation string = "lesson_articles"

type FindLessonArticleByLessonIDOptions struct {
	RequestID string
	LessonID  int32
}

func (s *Services) FindLessonArticleByLessonID(
	ctx context.Context,
	opts FindLessonArticleByLessonIDOptions,
) (*db.LessonArticle, *ServiceError) {
	log := s.buildLogger(opts.RequestID, lessonArticlesLocation, "FindLessonArticleByLessonID").With(
		"lessonId", opts.LessonID,
	)
	log.InfoContext(ctx, "Getting lesson article...")

	lessonArticle, err := s.database.GetLessonArticleByLessonID(ctx, opts.LessonID)
	if err != nil {
		log.WarnContext(ctx, "Lesson article not found", "error", err)
		return nil, FromDBError(err)
	}

	return &lessonArticle, nil
}

func CalculateReadingTime(content string) int32 {
	words := strings.Fields(content)
	return int32(math.Ceil((float64(len(words)) / 200.0) * 60.0))
}

type CreateLessonArticleOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	LessonID     int32
	Content      string
}

func (s *Services) CreateLessonArticle(ctx context.Context, opts CreateLessonArticleOptions) (*db.LessonArticle, *ServiceError) {
	log := s.buildLogger(opts.RequestID, lessonArticlesLocation, "CreateLessonArticle").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"seriesPartId", opts.SectionID,
		"lessonId", opts.LessonID,
	)
	log.InfoContext(ctx, "Creating lesson article...")

	lesson, serviceErr := s.AssertLessonOwnership(ctx, AssertLessonOwnershipOptions{
		RequestID:    opts.RequestID,
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SectionID:    opts.SectionID,
		LessonID:     opts.LessonID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	byIdOpts := FindLessonArticleByLessonIDOptions{
		RequestID: opts.RequestID,
		LessonID:  opts.LessonID,
	}
	if _, serviceErr := s.FindLessonArticleByLessonID(ctx, byIdOpts); serviceErr == nil {
		log.WarnContext(ctx, "Lesson article already exists")
		return nil, NewConflictError("Lesson article already exists")
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return nil, FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err, serviceErr)

	readTime := CalculateReadingTime(opts.Content)
	lessonArticle, err := qrs.CreateLessonArticle(ctx, db.CreateLessonArticleParams{
		LessonID:        opts.LessonID,
		Content:         opts.Content,
		ReadTimeSeconds: readTime,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to create lesson article", "error", err)
		return nil, FromDBError(err)
	}

	lecParams := db.UpdateLessonReadTimeSecondsParams{
		ID:              lesson.ID,
		ReadTimeSeconds: readTime,
	}
	if err := qrs.UpdateLessonReadTimeSeconds(ctx, lecParams); err != nil {
		log.ErrorContext(ctx, "Failed to update lesson read time", "error", err)
		return nil, FromDBError(err)
	}

	if lesson.IsPublished {
		seriesParams := db.AddSeriesReadTimeParams{
			ReadTimeSeconds: readTime,
			Slug:            opts.SeriesSlug,
		}
		if err := qrs.AddSeriesReadTime(ctx, seriesParams); err != nil {
			log.ErrorContext(ctx, "Failed to add series read time", "error", err)
			return nil, FromDBError(err)
		}

		seriesPartParams := db.AddSectionReadTimeParams{
			ReadTimeSeconds: readTime,
			ID:              opts.SectionID,
		}
		if err := qrs.AddSectionReadTime(ctx, seriesPartParams); err != nil {
			log.ErrorContext(ctx, "Failed to add series part read time", "error", err)
			return nil, FromDBError(err)
		}
	}

	return &lessonArticle, nil
}

type UpdateLessonArticleOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	LessonID     int32
	Content      string
}

func (s *Services) UpdateLessonArticle(
	ctx context.Context,
	opts UpdateLessonArticleOptions,
) (*db.LessonArticle, *ServiceError) {
	log := s.buildLogger(opts.RequestID, lessonArticlesLocation, "UpdateLessonArticle").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"seriesPartId", opts.SectionID,
		"lessonId", opts.LessonID,
	)
	log.InfoContext(ctx, "Updating lesson article...")

	lesson, serviceErr := s.AssertLessonOwnership(ctx, AssertLessonOwnershipOptions{
		RequestID:    opts.RequestID,
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SectionID:    opts.SectionID,
		LessonID:     opts.LessonID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	lessonArticle, serviceErr := s.FindLessonArticleByLessonID(ctx, FindLessonArticleByLessonIDOptions{
		RequestID: opts.RequestID,
		LessonID:  opts.LessonID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	oldReadTime := lessonArticle.ReadTimeSeconds
	readTime := CalculateReadingTime(opts.Content)

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return nil, FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err, serviceErr)

	*lessonArticle, err = qrs.UpdateLessonArticle(ctx, db.UpdateLessonArticleParams{
		ID:              lessonArticle.ID,
		Content:         opts.Content,
		ReadTimeSeconds: readTime,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to update lesson article", "error", err)
		return nil, FromDBError(err)
	}

	lecParams := db.UpdateLessonReadTimeSecondsParams{
		ID:              lesson.ID,
		ReadTimeSeconds: readTime,
	}
	if err := qrs.UpdateLessonReadTimeSeconds(ctx, lecParams); err != nil {
		log.ErrorContext(ctx, "Failed to update lesson read time", "error", err)
		return nil, FromDBError(err)
	}

	if lesson.IsPublished {
		readingTimeDiff := readTime - oldReadTime
		seriesParams := db.AddSeriesReadTimeParams{
			ReadTimeSeconds: readingTimeDiff,
			Slug:            opts.SeriesSlug,
		}
		if err := qrs.AddSeriesReadTime(ctx, seriesParams); err != nil {
			log.ErrorContext(ctx, "Failed to add series read time", "error", err)
			return nil, FromDBError(err)
		}

		seriesPartParams := db.AddSectionReadTimeParams{
			ReadTimeSeconds: readingTimeDiff,
			ID:              opts.SectionID,
		}
		if err := qrs.AddSectionReadTime(ctx, seriesPartParams); err != nil {
			log.ErrorContext(ctx, "Failed to add series part read time", "error", err)
			return nil, FromDBError(err)
		}
	}

	return lessonArticle, nil
}

type DeleteLessonArticleOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	LessonID     int32
}

func (s *Services) DeleteLessonArticle(ctx context.Context, opts DeleteLessonArticleOptions) *ServiceError {
	log := s.buildLogger(opts.RequestID, lessonArticlesLocation, "DeleteLessonArticle").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"seriesPartId", opts.SectionID,
		"lessonId", opts.LessonID,
	)
	log.InfoContext(ctx, "Deleting lesson article...")

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

	lessonArticle, serviceErr := s.FindLessonArticleByLessonID(ctx, FindLessonArticleByLessonIDOptions{
		RequestID: opts.RequestID,
		LessonID:  opts.LessonID,
	})
	if serviceErr != nil {
		return serviceErr
	}

	if lesson.IsPublished {
		log.WarnContext(ctx, "Cannot delete article from published lesson")
		return NewValidationError("Cannot delete article from published lesson")
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err, serviceErr)

	if err := qrs.DeleteLessonArticle(ctx, lessonArticle.ID); err != nil {
		log.ErrorContext(ctx, "Failed to delete lesson article", "error", err)
		return FromDBError(err)
	}

	lessonParams := db.UpdateLessonReadTimeSecondsParams{
		ID:              lesson.ID,
		ReadTimeSeconds: 0,
	}
	if err := qrs.UpdateLessonReadTimeSeconds(ctx, lessonParams); err != nil {
		log.ErrorContext(ctx, "Failed to update lesson read time", "error", err)
		return FromDBError(err)
	}

	return nil
}

type FindLessonArticleOptions struct {
	RequestID    string
	LanguageSlug string
	SeriesSlug   string
	SectionID    int32
	LessonID     int32
	IsPublished  bool
}

func (s *Services) FindLessonArticle(
	ctx context.Context,
	opts FindLessonArticleOptions,
) (*db.LessonArticle, *ServiceError) {
	log := s.buildLogger(opts.RequestID, lessonArticlesLocation, "FindLessonArticle").With(
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"seriesPartId", opts.SectionID,
		"lessonId", opts.LessonID,
	)
	log.InfoContext(ctx, "Getting lesson article...")

	lesson, serviceErr := s.FindLessonBySlugsAndIDs(ctx, FindLessonOptions{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SectionID:    opts.SectionID,
		LessonID:     opts.LessonID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	lessonArticle, serviceErr := s.FindLessonArticleByLessonID(ctx, FindLessonArticleByLessonIDOptions{
		RequestID: opts.RequestID,
		LessonID:  opts.LessonID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}
	if opts.IsPublished && !lesson.IsPublished {
		log.WarnContext(ctx, "Cannot find article from unpublished lesson")
		return nil, NewNotFoundError()
	}

	log.InfoContext(ctx, "Found lesson article")
	return lessonArticle, nil
}
