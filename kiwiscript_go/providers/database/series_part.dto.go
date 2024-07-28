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

type SeriesPartDTO struct {
	ID                 int32
	Title              string
	LanguageSlug       string
	SeriesSlug         string
	Description        string
	Position           int16
	InProgressLectures int16
	CompletedLectures  int16
	TotalLectures      int16
	IsCurrent          bool
	WatchTimeSeconds   int32
	ReadTimeSeconds    int32
	IsPublished        bool
}

type ToSeriesPartDTO interface {
	ToSeriesPartDTO() *SeriesPartDTO
}

type ToSeriesPartDTOWithProgress interface {
	ToSeriesPartDTO() *SeriesPartDTO
	ToSeriesPartDTOWithProgress(progress *SeriesProgress) *SeriesPartDTO
}

func (sp *SeriesPart) ToSeriesPartDTO() *SeriesPartDTO {
	return &SeriesPartDTO{
		ID:                 sp.ID,
		Title:              sp.Title,
		LanguageSlug:       sp.LanguageSlug,
		SeriesSlug:         sp.SeriesSlug,
		Description:        sp.Description,
		Position:           sp.Position,
		InProgressLectures: 0,
		CompletedLectures:  0,
		TotalLectures:      sp.LecturesCount,
		IsCurrent:          false,
		WatchTimeSeconds:   sp.WatchTimeSeconds,
		ReadTimeSeconds:    sp.ReadTimeSeconds,
		IsPublished:        sp.IsPublished,
	}
}

func (sp *SeriesPart) ToSeriesPartDTOWithProgress(progress *SeriesProgress) *SeriesPartDTO {
	return &SeriesPartDTO{
		ID:                 sp.ID,
		Title:              sp.Title,
		LanguageSlug:       sp.LanguageSlug,
		SeriesSlug:         sp.SeriesSlug,
		Description:        sp.Description,
		Position:           sp.Position,
		InProgressLectures: progress.InProgressLectures,
		CompletedLectures:  progress.CompletedLectures,
		TotalLectures:      sp.LecturesCount,
		IsCurrent:          progress.IsCurrent,
		WatchTimeSeconds:   sp.WatchTimeSeconds,
		ReadTimeSeconds:    sp.ReadTimeSeconds,
		IsPublished:        sp.IsPublished,
	}
}
