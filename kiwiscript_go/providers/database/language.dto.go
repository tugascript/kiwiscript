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

type LanguageDTO struct {
	ID               int32
	Name             string
	Slug             string
	Icon             string
	InProgressSeries int16
	CompletedSeries  int16
	TotalSeries      int16
	IsCurrent        bool
}

type ToLanguageDTO interface {
	ToLanguageDTO() *LanguageDTO
}

type ToLanguageDTOWithProgress interface {
	ToLanguageDTOWithProgress(progress *LanguageProgress) *LanguageDTO
}

func (l *Language) ToLanguageDTO() *LanguageDTO {
	return &LanguageDTO{
		ID:               l.ID,
		Name:             l.Name,
		Slug:             l.Slug,
		Icon:             l.Icon,
		InProgressSeries: 0,
		CompletedSeries:  0,
		TotalSeries:      l.SeriesCount,
		IsCurrent:        false,
	}
}

func (l *Language) ToLanguageDTOWithProgress(progress *LanguageProgress) *LanguageDTO {
	return &LanguageDTO{
		ID:               l.ID,
		Name:             l.Name,
		Slug:             l.Slug,
		Icon:             l.Icon,
		InProgressSeries: progress.InProgressSeries,
		CompletedSeries:  progress.CompletedSeries,
		TotalSeries:      l.SeriesCount,
		IsCurrent:        progress.IsCurrent,
	}
}

func (l *FindPaginatedLanguagesWithLanguageProgressRow) ToLanguageDTO() *LanguageDTO {
	return &LanguageDTO{
		ID:               l.ID,
		Name:             l.Name,
		Slug:             l.Slug,
		Icon:             l.Icon,
		InProgressSeries: l.InProgressSeries.Int16,
		CompletedSeries:  l.CompletedSeries.Int16,
		TotalSeries:      l.SeriesCount,
		IsCurrent:        l.IsCurrent.Bool,
	}
}

func (l *FindFilteredPaginatedLanguagesWithLanguageProgressRow) ToLanguageDTO() *LanguageDTO {
	return &LanguageDTO{
		ID:               l.ID,
		Name:             l.Name,
		Slug:             l.Slug,
		Icon:             l.Icon,
		InProgressSeries: l.InProgressSeries.Int16,
		CompletedSeries:  l.CompletedSeries.Int16,
		TotalSeries:      l.SeriesCount,
		IsCurrent:        l.IsCurrent.Bool,
	}
}

func (l *FindLanguageBySlugWithLanguageProgressRow) ToLanguageDTO() *LanguageDTO {
	return &LanguageDTO{
		ID:               l.ID,
		Name:             l.Name,
		Slug:             l.Slug,
		Icon:             l.Icon,
		InProgressSeries: l.InProgressSeries.Int16,
		CompletedSeries:  l.CompletedSeries.Int16,
		TotalSeries:      l.SeriesCount,
		IsCurrent:        l.IsCurrent.Bool,
	}
}
