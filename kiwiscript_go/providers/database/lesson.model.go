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

type LessonModel struct {
	ID               int32
	Title            string
	Position         int16
	WatchTimeSeconds int32
	ReadTimeSeconds  int32
	LanguageSlug     string
	SeriesSlug       string
	SectionID        int32
	ViewedAt         string
	IsPublished      bool
	IsCompleted      bool
}

type ToLessonModel interface {
	ToLessonModel() *LessonModel
}

type ToLessonModelWithProgress interface {
	ToLessonModelWithProgress(progress *LessonProgress) *LessonModel
}

func (l *Lesson) ToLessonModel() *LessonModel {
	return &LessonModel{
		ID:               l.ID,
		Title:            l.Title,
		Position:         l.Position,
		WatchTimeSeconds: l.WatchTimeSeconds,
		ReadTimeSeconds:  l.ReadTimeSeconds,
		LanguageSlug:     l.LanguageSlug,
		SeriesSlug:       l.SeriesSlug,
		SectionID:        l.SectionID,
		IsPublished:      l.IsPublished,
		IsCompleted:      false,
		ViewedAt:         "",
	}
}

func (l *Lesson) ToLessonModelWithProgress(progress *LessonProgress) *LessonModel {
	var viewedAt string
	if progress.ViewedAt.Valid {
		viewedAt = progress.ViewedAt.Time.Format(time.RFC3339)
	}

	return &LessonModel{
		ID:               l.ID,
		Title:            l.Title,
		Position:         l.Position,
		WatchTimeSeconds: l.WatchTimeSeconds,
		ReadTimeSeconds:  l.ReadTimeSeconds,
		LanguageSlug:     l.LanguageSlug,
		SeriesSlug:       l.SeriesSlug,
		SectionID:        l.SectionID,
		ViewedAt:         viewedAt,
		IsPublished:      l.IsPublished,
		IsCompleted:      progress.CompletedAt.Valid,
	}
}

func (l *FindPaginatedPublishedLessonsBySlugsAndSectionIDWithProgressRow) ToLessonModel() *LessonModel {
	var viewedAt string
	if l.LessonProgressViewedAt.Valid {
		viewedAt = l.LessonProgressViewedAt.Time.Format(time.RFC3339)
	}

	return &LessonModel{
		ID:               l.ID,
		Title:            l.Title,
		Position:         l.Position,
		WatchTimeSeconds: l.WatchTimeSeconds,
		ReadTimeSeconds:  l.ReadTimeSeconds,
		LanguageSlug:     l.LanguageSlug,
		SeriesSlug:       l.SeriesSlug,
		SectionID:        l.SectionID,
		IsPublished:      l.IsPublished,
		IsCompleted:      l.LessonProgressCompletedAt.Valid,
		ViewedAt:         viewedAt,
	}
}

func (l *FindPublishedLessonBySlugsAndIDsWithProgressArticleAndVideoRow) ToLessonModel() *LessonModel {
	var viewedAt string
	if l.LessonProgressViewedAt.Valid {
		viewedAt = l.LessonProgressViewedAt.Time.Format(time.RFC3339)
	}

	return &LessonModel{
		ID:               l.ID,
		Title:            l.Title,
		Position:         l.Position,
		WatchTimeSeconds: l.WatchTimeSeconds,
		ReadTimeSeconds:  l.ReadTimeSeconds,
		LanguageSlug:     l.LanguageSlug,
		SeriesSlug:       l.SeriesSlug,
		SectionID:        l.SectionID,
		IsPublished:      l.IsPublished,
		IsCompleted:      l.LessonProgressCompletedAt.Valid,
		ViewedAt:         viewedAt,
	}
}

func (l *FindLessonBySlugsAndIDsWithArticleAndVideoRow) ToLessonModel() *LessonModel {
	return &LessonModel{
		ID:               l.ID,
		Title:            l.Title,
		Position:         l.Position,
		WatchTimeSeconds: l.WatchTimeSeconds,
		ReadTimeSeconds:  l.ReadTimeSeconds,
		LanguageSlug:     l.LanguageSlug,
		SeriesSlug:       l.SeriesSlug,
		SectionID:        l.SectionID,
		IsPublished:      l.IsPublished,
		IsCompleted:      false,
	}
}
