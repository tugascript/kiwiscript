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

type CreateLecturesOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SeriesPartID int32
	Title        string
}

func (s *Services) CreateLectures(ctx context.Context, opts CreateLecturesOptions) (*db.Lecture, *ServiceError) {
	log := s.log.WithGroup("services.lectures.CreateLectures").With("title", opts.Title)
	log.InfoContext(ctx, "Creating lectures...")

	servicePart, serviceErr := s.AssertSeriesPartOwnership(ctx, AssertSeriesPartOwnershipOptions{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SeriesPartID: opts.SeriesPartID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	lecture, err := s.database.CreateLecture(ctx, db.CreateLectureParams{
		SeriesPartID: servicePart.ID,
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		Title:        opts.Title,
		AuthorID:     opts.UserID,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to create lecture", "error", err)
		return nil, FromDBError(err)
	}

	log.InfoContext(ctx, "Lecture created successfully")
	return &lecture, nil
}

func (s *Services) FindLecturesBySeriesPartID(ctx context.Context, seriesPartID int32) ([]db.Lecture, *ServiceError) {
	log := s.log.WithGroup("services.lectures.FindLecturesBySeriesPartID").With("series_part_id", seriesPartID)
	log.InfoContext(ctx, "Finding lectures by series part ID...")

	lectures, err := s.database.FindLecturesBySeriesPartID(ctx, seriesPartID)
	if err != nil {
		log.ErrorContext(ctx, "Failed to find lectures", "error", err)
		return nil, FromDBError(err)
	}

	log.InfoContext(ctx, "Lectures found successfully")
	return lectures, nil
}

type FindLectureOptions struct {
	LanguageSlug string
	SeriesSlug   string
	SeriesPartID int32
	LectureID    int32
}

func (s *Services) FindLectureBySlugsAndIDs(ctx context.Context, opts FindLectureOptions) (*db.Lecture, *ServiceError) {
	log := s.
		log.
		WithGroup("services.lectures.FindLecture").
		With("lecture_id", opts.LectureID, "series_part_id", opts.SeriesPartID)
	log.InfoContext(ctx, "Finding lecture...")

	lecture, err := s.database.FindLectureBySlugsAndIDs(ctx, db.FindLectureBySlugsAndIDsParams{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SeriesPartID: opts.SeriesPartID,
		ID:           opts.LectureID,
	})
	if err != nil {
		log.WarnContext(ctx, "Failed to find lecture", "error", err)
		return nil, FromDBError(err)
	}

	log.InfoContext(ctx, "Lecture found successfully")
	return &lecture, nil
}

func (s *Services) FindLectureWithArticleAndVideo(ctx context.Context, opts FindLectureOptions) (*db.FindPaginatedPublishedLecturesBySeriesPartIDWithArticleAndVideoRow, *ServiceError) {
	log := s.
		log.
		WithGroup("services.lectures.FindLecture").
		With("lecture_id", opts.LectureID, "series_part_id", opts.SeriesPartID)
	log.InfoContext(ctx, "Finding lecture...")

	lecture, err := s.database.FindLectureBySlugsAndIDsWithArticleAndVideo(ctx, db.FindLectureBySlugsAndIDsWithArticleAndVideoParams{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SeriesPartID: opts.SeriesPartID,
		ID:           opts.LectureID,
	})
	if err != nil {
		log.WarnContext(ctx, "Failed to find lecture", "error", err)
		return nil, FromDBError(err)
	}

	log.InfoContext(ctx, "Lecture found successfully")
	convLec := db.FindPaginatedPublishedLecturesBySeriesPartIDWithArticleAndVideoRow(lecture)
	return &convLec, nil
}

type FindPaginatedLecturesOptions struct {
	IsPublished  bool
	LanguageSlug string
	SeriesSlug   string
	SeriesPartID int32
	Offset       int32
	Limit        int32
}

func (s *Services) FindPaginatedLectures(ctx context.Context, opts FindPaginatedLecturesOptions) ([]db.FindPaginatedPublishedLecturesBySeriesPartIDWithArticleAndVideoRow, int64, *ServiceError) {
	log := s.log.WithGroup("services.lectures.FindLectures").With("series_slug", opts.SeriesSlug)
	log.InfoContext(ctx, "Finding lectures...")

	var count int64
	var lectures []db.FindPaginatedPublishedLecturesBySeriesPartIDWithArticleAndVideoRow
	if opts.IsPublished {
		var err error
		lectures, err = s.database.FindPaginatedPublishedLecturesBySeriesPartIDWithArticleAndVideo(ctx, db.FindPaginatedPublishedLecturesBySeriesPartIDWithArticleAndVideoParams{
			LanguageSlug: opts.LanguageSlug,
			SeriesSlug:   opts.SeriesSlug,
			SeriesPartID: opts.SeriesPartID,
			Offset:       opts.Offset,
			Limit:        opts.Limit,
		})
		if err != nil {
			log.ErrorContext(ctx, "Failed to find published lectures", "error", err)
			return nil, 0, FromDBError(err)
		}

		count, err = s.database.CountPublishedLecturesBySeriesPartID(ctx, opts.SeriesPartID)
		if err != nil {
			log.ErrorContext(ctx, "Failed to count published lectures", "error", err)
			return nil, 0, FromDBError(err)
		}
	} else {
		lecs, err := s.database.FindPaginatedLecturesBySeriesPartIDWithArticleAndVideo(ctx, db.FindPaginatedLecturesBySeriesPartIDWithArticleAndVideoParams{
			LanguageSlug: opts.LanguageSlug,
			SeriesSlug:   opts.SeriesSlug,
			SeriesPartID: opts.SeriesPartID,
			Offset:       opts.Offset,
			Limit:        opts.Limit,
		})
		if err != nil {
			log.ErrorContext(ctx, "Failed to find lectures", "error", err)
			return nil, 0, FromDBError(err)
		}

		lectures = make([]db.FindPaginatedPublishedLecturesBySeriesPartIDWithArticleAndVideoRow, len(lecs))
		for i, l := range lecs {
			lectures[i] = db.FindPaginatedPublishedLecturesBySeriesPartIDWithArticleAndVideoRow(l)
		}

		count, err = s.database.CountLecturesBySeriesPartID(ctx, opts.SeriesPartID)
		if err != nil {
			log.ErrorContext(ctx, "Failed to count lectures", "error", err)
			return nil, 0, FromDBError(err)
		}
	}

	log.InfoContext(ctx, "Lectures found successfully")
	return lectures, count, nil
}

type AssertLectureOwnershipOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SeriesPartID int32
	LectureID    int32
}

func (s *Services) AssertLectureOwnership(ctx context.Context, opts AssertLectureOwnershipOptions) (*db.Lecture, *ServiceError) {
	log := s.log.WithGroup("services.lectures.AssertLectureOwnership").With("lecture_id", opts.LectureID)
	log.InfoContext(ctx, "Asserting lecture ownership...")

	lecture, serviceErr := s.FindLectureBySlugsAndIDs(ctx, FindLectureOptions{
		LanguageSlug: opts.LanguageSlug,
		SeriesSlug:   opts.SeriesSlug,
		SeriesPartID: opts.SeriesPartID,
		LectureID:    opts.LectureID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	if lecture.AuthorID != opts.UserID {
		log.Warn("User is not the author of the series")
		return nil, NewForbiddenError()
	}

	return lecture, nil
}

type UpdateLectureOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SeriesPartID int32
	LectureID    int32
	Title        string
	Position     int16
}

func (s *Services) UpdateLecture(ctx context.Context, opts UpdateLectureOptions) (*db.Lecture, *ServiceError) {
	log := s.log.WithGroup("services.lectures.UpdateLectures").With("title", opts.Title)
	log.InfoContext(ctx, "Updating lectures...")

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

	if opts.Position == 0 || lecture.Position == opts.Position {
		var err error
		*lecture, err = s.database.UpdateLecture(ctx, db.UpdateLectureParams{
			ID:    lecture.ID,
			Title: opts.Title,
		})
		if err != nil {
			log.ErrorContext(ctx, "Failed to update lecture", "error", err)
			return nil, FromDBError(err)
		}

		log.InfoContext(ctx, "Lecture updated successfully")
		return lecture, nil
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return nil, FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err)

	oldPosition := lecture.Position
	*lecture, err = qrs.UpdateLectureWithPosition(ctx, db.UpdateLectureWithPositionParams{
		ID:       lecture.ID,
		Title:    opts.Title,
		Position: opts.Position,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to update lecture with position", "error", err)
		return nil, FromDBError(err)
	}

	if oldPosition < opts.Position {
		params := db.DecrementLecturePositionParams{
			SeriesPartID: opts.SeriesPartID,
			Position:     oldPosition,
			Position_2:   opts.Position,
		}
		if err := qrs.DecrementLecturePosition(ctx, params); err != nil {
			log.ErrorContext(ctx, "Failed to decrement lecture position", "error", err)
			return nil, FromDBError(err)
		}
	} else {
		params := db.IncrementLecturePositionParams{
			SeriesPartID: opts.SeriesPartID,
			Position:     opts.Position,
			Position_2:   oldPosition,
		}
		if err := qrs.IncrementLecturePosition(ctx, params); err != nil {
			log.ErrorContext(ctx, "Failed to increment lecture position", "error", err)
			return nil, FromDBError(err)
		}
	}

	log.InfoContext(ctx, "Lecture updated successfully")
	return lecture, nil
}

type DeleteLectureOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SeriesPartID int32
	LectureID    int32
}

func (s *Services) DeleteLecture(ctx context.Context, opts DeleteLectureOptions) *ServiceError {
	log := s.log.WithGroup("services.lectures.DeleteLectures").With("lecture_id", opts.LectureID)
	log.InfoContext(ctx, "Deleting lectures...")

	lecture, serviceErr := s.AssertLectureOwnership(ctx, AssertLectureOwnershipOptions(opts))
	if serviceErr != nil {
		return serviceErr
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err)

	if err := qrs.DeleteLectureByID(ctx, lecture.ID); err != nil {
		log.ErrorContext(ctx, "Failed to delete lecture", "error", err)
		return FromDBError(err)
	}

	params := db.DecrementLecturePositionParams{
		SeriesPartID: opts.SeriesPartID,
		Position:     lecture.Position,
		Position_2:   1,
	}
	if err := qrs.DecrementLecturePosition(ctx, params); err != nil {
		log.ErrorContext(ctx, "Failed to decrement lecture position", "error", err)
		return FromDBError(err)
	}

	log.InfoContext(ctx, "Lecture deleted successfully")
	return nil
}

type UpdateLectureIsPublishedOptions struct {
	UserID       int32
	LanguageSlug string
	SeriesSlug   string
	SeriesPartID int32
	LectureID    int32
	IsPublished  bool
}

func (s *Services) UpdateLectureIsPublished(ctx context.Context, opts UpdateLectureIsPublishedOptions) (*db.Lecture, *ServiceError) {
	log := s.log.WithGroup("services.lectures.UpdateLectureIsPublished").With("lecture_id", opts.LectureID)
	log.InfoContext(ctx, "Updating lecture article is published...")

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

	if lecture.IsPublished == opts.IsPublished {
		log.InfoContext(ctx, "Lecture article is already published")
		return lecture, nil
	}
	if opts.IsPublished && lecture.ReadTimeSeconds == 0 && lecture.WatchTimeSeconds == 0 {
		log.WarnContext(ctx, "Cannot publish lecture article without read time or watch time")
		return nil, NewValidationError("Cannot publish lecture without content")
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return nil, FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err)

	*lecture, err = qrs.UpdateLectureIsPublished(ctx, db.UpdateLectureIsPublishedParams{
		ID:          lecture.ID,
		IsPublished: opts.IsPublished,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to update lecture article is published", "error", err)
		return nil, FromDBError(err)
	}

	if opts.IsPublished {
		incPartParams := db.IncrementSeriesPartLecturesCountParams{
			ID:               opts.SeriesPartID,
			WatchTimeSeconds: lecture.WatchTimeSeconds,
			ReadTimeSeconds:  lecture.ReadTimeSeconds,
		}
		if err := qrs.IncrementSeriesPartLecturesCount(ctx, incPartParams); err != nil {
			log.ErrorContext(ctx, "Failed to increment series part lectures count", "error", err)
			return nil, FromDBError(err)
		}

		incSeriesParams := db.IncrementSeriesLecturesCountParams{
			Slug:             opts.SeriesSlug,
			WatchTimeSeconds: lecture.WatchTimeSeconds,
			ReadTimeSeconds:  lecture.ReadTimeSeconds,
		}
		if err := qrs.IncrementSeriesLecturesCount(ctx, incSeriesParams); err != nil {
			log.ErrorContext(ctx, "Failed to increment series lectures count", "error", err)
			return nil, FromDBError(err)
		}
	} else {
		decPartParams := db.DecrementSeriesPartLecturesCountParams{
			ID:               opts.SeriesPartID,
			WatchTimeSeconds: lecture.WatchTimeSeconds,
			ReadTimeSeconds:  lecture.ReadTimeSeconds,
		}
		if err := qrs.DecrementSeriesPartLecturesCount(ctx, decPartParams); err != nil {
			log.ErrorContext(ctx, "Failed to decrement series part lectures count", "error", err)
			return nil, FromDBError(err)
		}

		decSeriesParams := db.DecrementSeriesLecturesCountParams{
			Slug:             opts.SeriesSlug,
			WatchTimeSeconds: lecture.WatchTimeSeconds,
			ReadTimeSeconds:  lecture.ReadTimeSeconds,
		}
		if err := qrs.DecrementSeriesLecturesCount(ctx, decSeriesParams); err != nil {
			log.ErrorContext(ctx, "Failed to decrement series lectures count", "error", err)
			return nil, FromDBError(err)
		}
	}

	return lecture, nil
}
