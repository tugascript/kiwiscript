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

type LanguageModel struct {
	ID              int32
	Name            string
	Slug            string
	Icon            string
	CompletedSeries int16
	TotalSeries     int16
	ViewedAt        string
}

type ToLanguageModel interface {
	ToLanguageModel() *LanguageModel
}

type ToLanguageModelWithProgress interface {
	ToLanguageModelWithProgress(progress *LanguageProgress) *LanguageModel
}

func (l *Language) ToLanguageModel() *LanguageModel {
	return &LanguageModel{
		ID:              l.ID,
		Name:            l.Name,
		Slug:            l.Slug,
		Icon:            l.Icon,
		CompletedSeries: 0,
		TotalSeries:     l.SeriesCount,
		ViewedAt:        "",
	}
}

func (l *Language) ToLanguageModelWithProgress(progress *LanguageProgress) *LanguageModel {
	var viewedAt string
	if progress.ViewedAt.Valid {
		viewedAt = progress.ViewedAt.Time.Format(time.RFC3339)
	}

	return &LanguageModel{
		ID:              l.ID,
		Name:            l.Name,
		Slug:            l.Slug,
		Icon:            l.Icon,
		CompletedSeries: progress.CompletedSeries,
		TotalSeries:     l.SeriesCount,
		ViewedAt:        viewedAt,
	}
}

func (l *FindPaginatedLanguagesWithLanguageProgressRow) ToLanguageModel() *LanguageModel {
	var viewedAt string
	if l.LanguageProgressViewedAt.Valid {
		viewedAt = l.LanguageProgressViewedAt.Time.Format(time.RFC3339)
	}

	return &LanguageModel{
		ID:              l.ID,
		Name:            l.Name,
		Slug:            l.Slug,
		Icon:            l.Icon,
		CompletedSeries: l.LanguageProgressCompletedSeries.Int16,
		TotalSeries:     l.SeriesCount,
		ViewedAt:        viewedAt,
	}
}

func (l *FindFilteredPaginatedLanguagesWithLanguageProgressRow) ToLanguageModel() *LanguageModel {
	var viewedAt string
	if l.LanguageProgressViewedAt.Valid {
		viewedAt = l.LanguageProgressViewedAt.Time.Format(time.RFC3339)
	}

	return &LanguageModel{
		ID:              l.ID,
		Name:            l.Name,
		Slug:            l.Slug,
		Icon:            l.Icon,
		CompletedSeries: l.LanguageProgressCompletedSeries.Int16,
		TotalSeries:     l.SeriesCount,
		ViewedAt:        viewedAt,
	}
}

func (l *FindLanguageBySlugWithLanguageProgressRow) ToLanguageModel() *LanguageModel {
	var viewedAt string
	if l.ViewedAt.Valid {
		viewedAt = l.ViewedAt.Time.Format(time.RFC3339)
	}

	return &LanguageModel{
		ID:              l.ID,
		Name:            l.Name,
		Slug:            l.Slug,
		Icon:            l.Icon,
		CompletedSeries: l.CompletedSeries.Int16,
		TotalSeries:     l.SeriesCount,
		ViewedAt:        viewedAt,
	}
}
