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

func (s *Services) FindLectureArticleByLectureID(ctx context.Context, lectureID int32) (*db.LectureArticle, *ServiceError) {
	log := s.log.WithGroup("services.lectures.FindLectureArticleByLectureID").With("lectureId", lectureID)
	log.InfoContext(ctx, "Getting lecture article...")

	lectureArticle, err := s.database.GetLectureArticleByLectureID(ctx, lectureID)
	if err != nil {
		return nil, FromDBError(err)
	}

	return &lectureArticle, nil
}

func CalculateReadingTime(content string) int32 {
	words := strings.Fields(content)
	return int32(math.Ceil((float64(len(words)) / 200.0) * 60.0))
}

type CreateLectureArticleOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SeriesPartID int32
	LectureID    int32
	Content      string
}

func (s *Services) CreateLectureArticle(ctx context.Context, opts CreateLectureArticleOptions) (*db.LectureArticle, *ServiceError) {
	log := s.log.WithGroup("services.lectures.CreateLectureArticle").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"seriesPartId", opts.SeriesPartID,
		"lectureId", opts.LectureID,
	)
	log.InfoContext(ctx, "Creating lecture article...")

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

	if _, serviceErr := s.FindLectureArticleByLectureID(ctx, opts.LectureID); serviceErr == nil {
		return nil, NewDuplicateKeyError("Lecture article already exists")
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return nil, FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err)

	readTime := CalculateReadingTime(opts.Content)
	lectureArticle, err := qrs.CreateLectureArticle(ctx, db.CreateLectureArticleParams{
		LectureID:       opts.LectureID,
		Content:         opts.Content,
		ReadTimeSeconds: readTime,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to create lecture article", "error", err)
		return nil, FromDBError(err)
	}

	lecParams := db.UpdateLectureReadTimeSecondsParams{
		ID:              lecture.ID,
		ReadTimeSeconds: readTime,
	}
	if err := qrs.UpdateLectureReadTimeSeconds(ctx, lecParams); err != nil {
		log.ErrorContext(ctx, "Failed to update lecture read time", "error", err)
		return nil, FromDBError(err)
	}

	if lecture.IsPublished {
		seriesParams := db.AddSeriesReadTimeParams{
			ReadTimeSeconds: readTime,
			Slug:            opts.SeriesSlug,
		}
		if err := qrs.AddSeriesReadTime(ctx, seriesParams); err != nil {
			log.ErrorContext(ctx, "Failed to add series read time", "error", err)
			return nil, FromDBError(err)
		}

		seriesPartParams := db.AddSeriesPartReadTimeParams{
			ReadTimeSeconds: readTime,
			ID:              opts.SeriesPartID,
		}
		if err := qrs.AddSeriesPartReadTime(ctx, seriesPartParams); err != nil {
			log.ErrorContext(ctx, "Failed to add series part read time", "error", err)
			return nil, FromDBError(err)
		}
	}

	return &lectureArticle, nil
}

type UpdateLectureArticleOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SeriesPartID int32
	LectureID    int32
	Content      string
}

func (s *Services) UpdateLectureArticle(
	ctx context.Context,
	opts UpdateLectureArticleOptions,
) (*db.LectureArticle, *ServiceError) {
	log := s.log.WithGroup("services.lectures.UpdateLectureArticle").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"seriesPartId", opts.SeriesPartID,
		"lectureId", opts.LectureID,
	)
	log.InfoContext(ctx, "Updating lecture article...")

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

	lectureArticle, serviceErr := s.FindLectureArticleByLectureID(ctx, opts.LectureID)
	if serviceErr != nil {
		return nil, serviceErr
	}

	oldReadTime := lectureArticle.ReadTimeSeconds
	readTime := CalculateReadingTime(opts.Content)

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return nil, FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err)

	*lectureArticle, err = qrs.UpdateLectureArticle(ctx, db.UpdateLectureArticleParams{
		ID:              lectureArticle.ID,
		Content:         opts.Content,
		ReadTimeSeconds: readTime,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to update lecture article", "error", err)
		return nil, FromDBError(err)
	}

	lecParams := db.UpdateLectureReadTimeSecondsParams{
		ID:              lecture.ID,
		ReadTimeSeconds: readTime,
	}
	if err := qrs.UpdateLectureReadTimeSeconds(ctx, lecParams); err != nil {
		log.ErrorContext(ctx, "Failed to update lecture read time", "error", err)
		return nil, FromDBError(err)
	}

	if lecture.IsPublished {
		readingTimeDiff := readTime - oldReadTime
		seriesParams := db.AddSeriesReadTimeParams{
			ReadTimeSeconds: readingTimeDiff,
			Slug:            opts.SeriesSlug,
		}
		if err := qrs.AddSeriesReadTime(ctx, seriesParams); err != nil {
			log.ErrorContext(ctx, "Failed to add series read time", "error", err)
			return nil, FromDBError(err)
		}

		seriesPartParams := db.AddSeriesPartReadTimeParams{
			ReadTimeSeconds: readingTimeDiff,
			ID:              opts.SeriesPartID,
		}
		if err := qrs.AddSeriesPartReadTime(ctx, seriesPartParams); err != nil {
			log.ErrorContext(ctx, "Failed to add series part read time", "error", err)
			return nil, FromDBError(err)
		}
	}

	return lectureArticle, nil
}

type DeleteLectureArticleOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SeriesPartID int32
	LectureID    int32
}

func (s *Services) DeleteLectureArticle(ctx context.Context, opts DeleteLectureArticleOptions) *ServiceError {
	log := s.log.WithGroup("services.lectures.DeleteLectureArticle").With(
		"userId", opts.UserID,
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"seriesPartId", opts.SeriesPartID,
		"lectureId", opts.LectureID,
	)
	log.InfoContext(ctx, "Deleting lecture article...")

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

	lectureArticle, serviceErr := s.FindLectureArticleByLectureID(ctx, opts.LectureID)
	if serviceErr != nil {
		return serviceErr
	}

	if lecture.IsPublished {
		log.WarnContext(ctx, "Cannot delete article from published lecture")
		return NewValidationError("Cannot delete article from published lecture")
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err)

	if err := qrs.DeleteLectureArticle(ctx, lectureArticle.ID); err != nil {
		log.ErrorContext(ctx, "Failed to delete lecture article", "error", err)
		return FromDBError(err)
	}

	lecParams := db.UpdateLectureReadTimeSecondsParams{
		ID:              lecture.ID,
		ReadTimeSeconds: 0,
	}
	if err := qrs.UpdateLectureReadTimeSeconds(ctx, lecParams); err != nil {
		log.ErrorContext(ctx, "Failed to update lecture read time", "error", err)
		return FromDBError(err)
	}

	return nil
}

type FindLectureArticleOptions struct {
	LanguageSlug string
	SeriesSlug   string
	SeriesPartID int32
	LectureID    int32
	IsPublished  bool
}

func (s *Services) FindLectureArticle(ctx context.Context, opts FindLectureArticleOptions) (*db.LectureArticle, *ServiceError) {
	log := s.log.WithGroup("services.lectures.GetLectureArticle").With(
		"languageSlug", opts.LanguageSlug,
		"seriesSlug", opts.SeriesSlug,
		"seriesPartId", opts.SeriesPartID,
		"lectureId", opts.LectureID,
	)
	log.InfoContext(ctx, "Getting lecture article...")

	lecture, serviceErr := s.FindLectureBySlugsAndIDs(ctx, FindLectureOptions{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SeriesPartID: opts.SeriesPartID,
		LectureID:    opts.LectureID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	lectureArticle, serviceErr := s.FindLectureArticleByLectureID(ctx, opts.LectureID)
	if serviceErr != nil {
		return nil, serviceErr
	}
	if opts.IsPublished && !lecture.IsPublished {
		log.WarnContext(ctx, "Cannot find article from unpublished lecture")
		return nil, NewNotFoundError()
	}

	log.InfoContext(ctx, "Found lecture article")
	return lectureArticle, nil
}
