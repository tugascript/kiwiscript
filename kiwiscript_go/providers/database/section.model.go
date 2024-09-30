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

package db

import "time"

type SectionModel struct {
	ID               int32
	Title            string
	LanguageSlug     string
	SeriesSlug       string
	Description      string
	Position         int16
	CompletedLessons int16
	TotalLessons     int16
	IsCompleted      bool
	ViewedAt         string
	WatchTimeSeconds int32
	ReadTimeSeconds  int32
	IsPublished      bool
}

type ToSectionModel interface {
	ToSectionModel() *SectionModel
}

type ToSectionModelWithProgress interface {
	ToLanguageModelWithProgress(progress *SectionProgress) *SectionModel
}

func (sp *Section) ToSectionModel() *SectionModel {
	return &SectionModel{
		ID:               sp.ID,
		Title:            sp.Title,
		LanguageSlug:     sp.LanguageSlug,
		SeriesSlug:       sp.SeriesSlug,
		Description:      sp.Description,
		Position:         sp.Position,
		CompletedLessons: 0,
		TotalLessons:     sp.LessonsCount,
		IsCompleted:      false,
		ViewedAt:         "",
		WatchTimeSeconds: sp.WatchTimeSeconds,
		ReadTimeSeconds:  sp.ReadTimeSeconds,
		IsPublished:      sp.IsPublished,
	}
}

func (sp *Section) ToSectionModelWithProgress(progress *SectionProgress) *SectionModel {
	var viewedAt string
	if progress.ViewedAt.Valid {
		viewedAt = progress.ViewedAt.Time.Format(time.RFC3339)
	}

	return &SectionModel{
		ID:               sp.ID,
		Title:            sp.Title,
		LanguageSlug:     sp.LanguageSlug,
		SeriesSlug:       sp.SeriesSlug,
		Description:      sp.Description,
		Position:         sp.Position,
		CompletedLessons: progress.CompletedLessons,
		TotalLessons:     sp.LessonsCount,
		IsCompleted:      progress.CompletedAt.Valid,
		ViewedAt:         viewedAt,
		WatchTimeSeconds: sp.WatchTimeSeconds,
		ReadTimeSeconds:  sp.ReadTimeSeconds,
		IsPublished:      sp.IsPublished,
	}
}

func (sp *FindPublishedSectionBySlugsAndIDWithProgressRow) ToSectionModel() *SectionModel {
	var viewedAt string
	if sp.SectionProgressViewedAt.Valid {
		viewedAt = sp.SectionProgressViewedAt.Time.Format(time.RFC3339)
	}

	return &SectionModel{
		ID:               sp.ID,
		Title:            sp.Title,
		LanguageSlug:     sp.LanguageSlug,
		SeriesSlug:       sp.SeriesSlug,
		Description:      sp.Description,
		Position:         sp.Position,
		CompletedLessons: sp.SectionProgressCompletedLessons.Int16,
		TotalLessons:     sp.LessonsCount,
		IsCompleted:      sp.SectionProgressCompletedAt.Valid,
		ViewedAt:         viewedAt,
		WatchTimeSeconds: sp.WatchTimeSeconds,
		ReadTimeSeconds:  sp.ReadTimeSeconds,
		IsPublished:      sp.IsPublished,
	}
}

func (sp *FindPaginatedPublishedSectionsBySlugsWithProgressRow) ToSectionModel() *SectionModel {
	var viewedAt string
	if sp.SectionProgressViewedAt.Valid {
		viewedAt = sp.SectionProgressViewedAt.Time.Format(time.RFC3339)
	}

	return &SectionModel{
		ID:               sp.ID,
		Title:            sp.Title,
		LanguageSlug:     sp.LanguageSlug,
		SeriesSlug:       sp.SeriesSlug,
		Description:      sp.Description,
		Position:         sp.Position,
		CompletedLessons: sp.SectionProgressCompletedLessons.Int16,
		TotalLessons:     sp.LessonsCount,
		IsCompleted:      sp.SectionProgressCompletedAt.Valid,
		ViewedAt:         viewedAt,
		WatchTimeSeconds: sp.WatchTimeSeconds,
		ReadTimeSeconds:  sp.ReadTimeSeconds,
		IsPublished:      sp.IsPublished,
	}
}

func (sec *FindCurrentSectionRow) ToSectionModel() *SectionModel {
	return &SectionModel{
		ID:               sec.ID,
		Title:            sec.Title,
		LanguageSlug:     sec.LanguageSlug,
		SeriesSlug:       sec.SeriesSlug,
		Description:      sec.Description,
		Position:         sec.Position,
		CompletedLessons: sec.SectionProgressCompletedLessons,
		TotalLessons:     sec.LessonsCount,
		IsCompleted:      sec.SectionProgressCompletedAt.Valid,
		ViewedAt:         sec.SectionProgressViewedAt.Time.Format(time.RFC3339),
		WatchTimeSeconds: sec.WatchTimeSeconds,
		ReadTimeSeconds:  sec.ReadTimeSeconds,
		IsPublished:      sec.IsPublished,
	}
}
