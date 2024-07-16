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
	SeriesSlug   string
	SeriesPartID int32
	Title        string
	Description  string
}

func (s *Services) CreateLectures(ctx context.Context, opts CreateLecturesOptions) (db.Lecture, *ServiceError) {
	log := s.log.WithGroup("services.lectures.CreateLectures").With("title", opts.Title)
	log.InfoContext(ctx, "Creating lectures...")

	servicePart, serviceErr := s.AssertSeriesPartOwnership(ctx, AssertSeriesPartOwnershipOptions{
		UserID:       opts.UserID,
		SeriesSlug:   opts.SeriesSlug,
		SeriesPartID: opts.SeriesPartID,
	})
	if serviceErr != nil {
		return db.Lecture{}, serviceErr
	}

	lecture, err := s.database.CreateLecture(ctx, db.CreateLectureParams{
		SeriesPartID: servicePart.ID,
		Title:        opts.Title,
		Description:  opts.Description,
		AuthorID:     opts.UserID,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to create lecture", "error", err)
		return db.Lecture{}, FromDBError(err)
	}

	log.InfoContext(ctx, "Lecture created successfully")
	return lecture, nil
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
	SeriesSlug   string
	SeriesPartID int32
	LectureID    int32
}

func (s *Services) FindLecture(ctx context.Context, opts FindLectureOptions) (db.Lecture, *ServiceError) {
	log := s.
		log.
		WithGroup("services.lectures.FindLecture").
		With("lecture_id", opts.LectureID, "series_part_id", opts.SeriesPartID)
	log.InfoContext(ctx, "Finding lecture...")

	servicePart, serviceErr := s.FindSeriesPartBySlugAndID(ctx, FindSeriesPartBySlugsAndIDOptions{
		SeriesSlug:   opts.SeriesSlug,
		SeriesPartID: opts.SeriesPartID,
	})
	if serviceErr != nil {
		return db.Lecture{}, serviceErr
	}

	lecture, err := s.database.FindLectureByIDs(ctx, db.FindLectureByIDsParams{
		ID:           opts.LectureID,
		SeriesPartID: servicePart.ID,
	})
	if err != nil {
		log.WarnContext(ctx, "Failed to find lecture", "error", err)
		return db.Lecture{}, FromDBError(err)
	}

	log.InfoContext(ctx, "Lecture found successfully")
	return lecture, nil
}

func (s *Services) FindPublishedLecture(ctx context.Context, opts FindLectureOptions) (db.Lecture, *ServiceError) {
	log := s.
		log.
		WithGroup("services.lectures.FindPublishedLecture").
		With("lecture_id", opts.LectureID, "series_part_id", opts.SeriesPartID)
	log.InfoContext(ctx, "Finding published lecture...")

	servicePart, serviceErr := s.FindSeriesPartBySlugAndID(ctx, FindSeriesPartBySlugsAndIDOptions{
		SeriesSlug:   opts.SeriesSlug,
		SeriesPartID: opts.SeriesPartID,
	})
	if serviceErr != nil {
		return db.Lecture{}, serviceErr
	}

	lecture, err := s.database.FindPublishedLectureByIDs(ctx, db.FindPublishedLectureByIDsParams{
		ID:           opts.LectureID,
		SeriesPartID: servicePart.ID,
	})
	if err != nil {
		log.WarnContext(ctx, "Failed to find published lecture", "error", err)
		return db.Lecture{}, FromDBError(err)
	}

	log.InfoContext(ctx, "Published lecture found successfully")
	return lecture, nil
}

type FindPaginatedLecturesOptions struct {
	UserID       int32
	SeriesSlug   string
	SeriesPartID int32
	Offset       int32
	Limit        int32
}

func (s *Services) FindPaginatedLectures(ctx context.Context, opts FindPaginatedLecturesOptions) ([]db.Lecture, int64, *ServiceError) {
	log := s.log.WithGroup("services.lectures.FindLectures").With("series_slug", opts.SeriesSlug)
	log.InfoContext(ctx, "Finding lectures...")

	servicePart, serviceErr := s.FindSeriesPartBySlugAndID(ctx, FindSeriesPartBySlugsAndIDOptions{
		SeriesSlug:   opts.SeriesSlug,
		SeriesPartID: opts.SeriesPartID,
	})
	if serviceErr != nil {
		return nil, 0, serviceErr
	}

	count, err := s.database.CountLecturesBySeriesPartID(ctx, opts.SeriesPartID)
	if err != nil {
		log.ErrorContext(ctx, "Failed to count lectures", "error", err)
		return nil, 0, FromDBError(err)
	}

	lectures, err := s.database.FindPaginatedLecturesBySeriesPartID(ctx, db.FindPaginatedLecturesBySeriesPartIDParams{
		SeriesPartID: servicePart.ID,
		Offset:       opts.Offset,
		Limit:        opts.Limit,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to find lectures", "error", err)
		return nil, 0, FromDBError(err)
	}

	log.InfoContext(ctx, "Lectures found successfully")
	return lectures, count, nil
}

type AssertLectureOwnershipOptions struct {
	UserID       int32
	SeriesSlug   string
	SeriesPartID int32
	LectureID    int32
}

func (s *Services) AssertLectureOwnership(ctx context.Context, opts AssertLectureOwnershipOptions) (db.Lecture, *ServiceError) {
	log := s.log.WithGroup("services.lectures.AssertLectureOwnership").With("lecture_id", opts.LectureID)
	log.InfoContext(ctx, "Asserting lecture ownership...")

	servicePart, serviceErr := s.AssertSeriesPartOwnership(ctx, AssertSeriesPartOwnershipOptions{
		UserID:       opts.UserID,
		SeriesSlug:   opts.SeriesSlug,
		SeriesPartID: opts.SeriesPartID,
	})
	if serviceErr != nil {
		return db.Lecture{}, serviceErr
	}

	lecture, err := s.database.FindLectureByIDs(ctx, db.FindLectureByIDsParams{
		ID:           opts.LectureID,
		SeriesPartID: servicePart.ID,
	})
	if err != nil {
		log.WarnContext(ctx, "Failed to find lecture", "error", err)
		return db.Lecture{}, FromDBError(err)
	}

	return lecture, nil
}

type UpdateLecturesOptions struct {
	UserID       int32
	SeriesSlug   string
	SeriesPartID int32
	LectureID    int32
	Title        string
	Description  string
	Position     int16
}

func (s *Services) UpdateLectures(ctx context.Context, opts UpdateLecturesOptions) (db.Lecture, *ServiceError) {
	log := s.log.WithGroup("services.lectures.UpdateLectures").With("title", opts.Title)
	log.InfoContext(ctx, "Updating lectures...")

	lecture, serviceErr := s.AssertLectureOwnership(ctx, AssertLectureOwnershipOptions{
		UserID:       opts.UserID,
		SeriesSlug:   opts.SeriesSlug,
		SeriesPartID: opts.SeriesPartID,
		LectureID:    opts.LectureID,
	})
	if serviceErr != nil {
		return db.Lecture{}, serviceErr
	}

	if opts.Position == 0 || lecture.Position == opts.Position {
		lecture, err := s.database.UpdateLecture(ctx, db.UpdateLectureParams{
			ID:          lecture.ID,
			Title:       opts.Title,
			Description: opts.Description,
		})
		if err != nil {
			log.ErrorContext(ctx, "Failed to update lecture", "error", err)
			return db.Lecture{}, FromDBError(err)
		}

		log.InfoContext(ctx, "Lecture updated successfully")
		return lecture, nil
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return db.Lecture{}, FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err)

	oldPosition := lecture.Position
	lecture, err = qrs.UpdateLectureWithPosition(ctx, db.UpdateLectureWithPositionParams{
		ID:          lecture.ID,
		Title:       opts.Title,
		Description: opts.Description,
		Position:    opts.Position,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to update lecture with position", "error", err)
		return db.Lecture{}, FromDBError(err)
	}

	if oldPosition < opts.Position {
		params := db.DecrementLecturePositionParams{
			SeriesPartID: opts.SeriesPartID,
			Position:     oldPosition,
			Position_2:   opts.Position,
		}
		if err := qrs.DecrementLecturePosition(ctx, params); err != nil {
			log.ErrorContext(ctx, "Failed to decrement lecture position", "error", err)
			return db.Lecture{}, FromDBError(err)
		}
	} else {
		params := db.IncrementLecturePositionParams{
			SeriesPartID: opts.SeriesPartID,
			Position:     opts.Position,
			Position_2:   oldPosition,
		}
		if err := qrs.IncrementLecturePosition(ctx, params); err != nil {
			log.ErrorContext(ctx, "Failed to increment lecture position", "error", err)
			return db.Lecture{}, FromDBError(err)
		}
	}

	log.InfoContext(ctx, "Lecture updated successfully")
	return lecture, nil
}

type DeleteLecturesOptions struct {
	UserID       int32
	SeriesSlug   string
	SeriesPartID int32
	LectureID    int32
}

func (s *Services) DeleteLectures(ctx context.Context, opts DeleteLecturesOptions) *ServiceError {
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
