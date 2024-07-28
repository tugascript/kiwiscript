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

type SeriesAuthor struct {
	ID        int32
	FirstName string
	LastName  string
}

type SeriesDTO struct {
	ID                 int32
	Title              string
	Slug               string
	LanguageSlug       string
	Description        string
	InProgressParts    int16
	CompletedParts     int16
	TotalParts         int16
	InProgressLectures int16
	CompletedLectures  int16
	TotalLectures      int16
	IsCurrent          bool
	IsPublished        bool
	Author             SeriesAuthor
}

type ToSeriesDTO interface {
	ToSeriesDTO() *SeriesDTO
}

type ToSeriesDTOWithAuthor interface {
	ToSeriesDTOWithAuthor(authorID int32, firstName, lastName string) *SeriesDTO
}

type ToSeriesDTOWithProgress interface {
	ToSeriesDTOWithProgress(progress *SeriesProgress) *SeriesDTO
	ToSeriesDTO() *SeriesDTO
}

func (s *Series) ToSeriesDTOWithAuthor(authorID int32, firstName, lastName string) *SeriesDTO {
	return &SeriesDTO{
		ID:                 s.ID,
		Title:              s.Title,
		Slug:               s.Slug,
		LanguageSlug:       s.LanguageSlug,
		Description:        s.Description,
		InProgressParts:    0,
		CompletedParts:     0,
		TotalParts:         s.PartsCount,
		InProgressLectures: 0,
		CompletedLectures:  0,
		TotalLectures:      s.LecturesCount,
		IsCurrent:          false,
		IsPublished:        s.IsPublished,
		Author: SeriesAuthor{
			ID:        authorID,
			FirstName: firstName,
			LastName:  lastName,
		},
	}
}

func (s *FindSeriesBySlugWithAuthorRow) ToSeriesDTO() *SeriesDTO {
	return &SeriesDTO{
		ID:                 s.ID,
		Title:              s.Title,
		Slug:               s.Slug,
		LanguageSlug:       s.LanguageSlug,
		Description:        s.Description,
		InProgressParts:    0,
		CompletedParts:     0,
		TotalParts:         s.PartsCount,
		InProgressLectures: 0,
		CompletedLectures:  0,
		TotalLectures:      s.LecturesCount,
		IsCurrent:          false,
		IsPublished:        s.IsPublished,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindPublishedSeriesBySlugWithAuthorAndProgressRow) ToSeriesDTO() *SeriesDTO {
	return &SeriesDTO{
		ID:                 s.ID,
		Title:              s.Title,
		Slug:               s.Slug,
		LanguageSlug:       s.LanguageSlug,
		Description:        s.Description,
		InProgressParts:    s.SeriesProgressInProgressParts.Int16,
		CompletedParts:     s.SeriesProgressCompletedParts.Int16,
		TotalParts:         s.PartsCount,
		InProgressLectures: s.SeriesProgressInProgressLectures.Int16,
		CompletedLectures:  s.SeriesProgressCompletedLectures.Int16,
		TotalLectures:      s.LecturesCount,
		IsCurrent:          s.SeriesProgressIsCurrent.Bool,
		IsPublished:        s.IsPublished,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindPaginatedSeriesWithAuthorSortByIDRow) ToSeriesDTO() *SeriesDTO {
	return &SeriesDTO{
		ID:                 s.ID,
		Title:              s.Title,
		Slug:               s.Slug,
		LanguageSlug:       s.LanguageSlug,
		Description:        s.Description,
		InProgressParts:    0,
		CompletedParts:     0,
		TotalParts:         s.PartsCount,
		InProgressLectures: 0,
		CompletedLectures:  0,
		TotalLectures:      s.LecturesCount,
		IsCurrent:          false,
		IsPublished:        s.IsPublished,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindPaginatedPublishedSeriesWithAuthorSortByIDRow) ToSeriesDTO() *SeriesDTO {
	return &SeriesDTO{
		ID:                 s.ID,
		Title:              s.Title,
		Slug:               s.Slug,
		LanguageSlug:       s.LanguageSlug,
		Description:        s.Description,
		InProgressParts:    0,
		CompletedParts:     0,
		TotalParts:         s.PartsCount,
		InProgressLectures: 0,
		CompletedLectures:  0,
		TotalLectures:      s.LecturesCount,
		IsCurrent:          false,
		IsPublished:        s.IsPublished,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindFilteredPublishedSeriesWithAuthorSortByIDRow) ToSeriesDTO() *SeriesDTO {
	return &SeriesDTO{
		ID:                 s.ID,
		Title:              s.Title,
		Slug:               s.Slug,
		LanguageSlug:       s.LanguageSlug,
		Description:        s.Description,
		InProgressParts:    0,
		CompletedParts:     0,
		TotalParts:         s.PartsCount,
		InProgressLectures: 0,
		CompletedLectures:  0,
		TotalLectures:      s.LecturesCount,
		IsCurrent:          false,
		IsPublished:        s.IsPublished,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindPaginatedSeriesWithAuthorSortBySlugRow) ToSeriesDTO() *SeriesDTO {
	return &SeriesDTO{
		ID:                 s.ID,
		Title:              s.Title,
		Slug:               s.Slug,
		LanguageSlug:       s.LanguageSlug,
		Description:        s.Description,
		InProgressParts:    0,
		CompletedParts:     0,
		TotalParts:         s.PartsCount,
		InProgressLectures: 0,
		CompletedLectures:  0,
		TotalLectures:      s.LecturesCount,
		IsCurrent:          false,
		IsPublished:        s.IsPublished,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindPaginatedPublishedSeriesWithAuthorSortBySlugRow) ToSeriesDTO() *SeriesDTO {
	return &SeriesDTO{
		ID:                 s.ID,
		Title:              s.Title,
		Slug:               s.Slug,
		LanguageSlug:       s.LanguageSlug,
		Description:        s.Description,
		InProgressParts:    0,
		CompletedParts:     0,
		TotalParts:         s.PartsCount,
		InProgressLectures: 0,
		CompletedLectures:  0,
		TotalLectures:      s.LecturesCount,
		IsCurrent:          false,
		IsPublished:        s.IsPublished,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindFilteredPublishedSeriesWithAuthorSortBySlugRow) ToSeriesDTO() *SeriesDTO {
	return &SeriesDTO{
		ID:                 s.ID,
		Title:              s.Title,
		Slug:               s.Slug,
		LanguageSlug:       s.LanguageSlug,
		Description:        s.Description,
		InProgressParts:    0,
		CompletedParts:     0,
		TotalParts:         s.PartsCount,
		InProgressLectures: 0,
		CompletedLectures:  0,
		TotalLectures:      s.LecturesCount,
		IsCurrent:          false,
		IsPublished:        s.IsPublished,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindFilteredSeriesWithAuthorSortBySlugRow) ToSeriesDTO() *SeriesDTO {
	return &SeriesDTO{
		ID:                 s.ID,
		Title:              s.Title,
		Slug:               s.Slug,
		LanguageSlug:       s.LanguageSlug,
		Description:        s.Description,
		InProgressParts:    0,
		CompletedParts:     0,
		TotalParts:         s.PartsCount,
		InProgressLectures: 0,
		CompletedLectures:  0,
		TotalLectures:      s.LecturesCount,
		IsCurrent:          false,
		IsPublished:        s.IsPublished,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindFilteredSeriesWithAuthorSortByIDRow) ToSeriesDTO() *SeriesDTO {
	return &SeriesDTO{
		ID:                 s.ID,
		Title:              s.Title,
		Slug:               s.Slug,
		LanguageSlug:       s.LanguageSlug,
		Description:        s.Description,
		InProgressParts:    0,
		CompletedParts:     0,
		TotalParts:         s.PartsCount,
		InProgressLectures: 0,
		CompletedLectures:  0,
		TotalLectures:      s.LecturesCount,
		IsCurrent:          false,
		IsPublished:        s.IsPublished,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindPaginatedPublishedSeriesWithAuthorAndProgressSortByIDRow) ToSeriesDTO() *SeriesDTO {
	return &SeriesDTO{
		ID:                 s.ID,
		Title:              s.Title,
		Slug:               s.Slug,
		LanguageSlug:       s.LanguageSlug,
		Description:        s.Description,
		InProgressParts:    s.SeriesProgressInProgressParts.Int16,
		CompletedParts:     s.SeriesProgressCompletedParts.Int16,
		TotalParts:         s.PartsCount,
		InProgressLectures: s.SeriesProgressInProgressLectures.Int16,
		CompletedLectures:  s.SeriesProgressCompletedLectures.Int16,
		TotalLectures:      s.LecturesCount,
		IsCurrent:          s.SeriesProgressIsCurrent.Bool,
		IsPublished:        s.IsPublished,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindPaginatedPublishedSeriesWithAuthorAndProgressSortBySlugRow) ToSeriesDTO() *SeriesDTO {
	return &SeriesDTO{
		ID:                 s.ID,
		Title:              s.Title,
		Slug:               s.Slug,
		LanguageSlug:       s.LanguageSlug,
		Description:        s.Description,
		InProgressParts:    s.SeriesProgressInProgressParts.Int16,
		CompletedParts:     s.SeriesProgressCompletedParts.Int16,
		TotalParts:         s.PartsCount,
		InProgressLectures: s.SeriesProgressInProgressLectures.Int16,
		CompletedLectures:  s.SeriesProgressCompletedLectures.Int16,
		TotalLectures:      s.LecturesCount,
		IsCurrent:          s.SeriesProgressIsCurrent.Bool,
		IsPublished:        s.IsPublished,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindFilteredPublishedSeriesWithAuthorAndProgressSortByIDRow) ToSeriesDTO() *SeriesDTO {
	return &SeriesDTO{
		ID:                 s.ID,
		Title:              s.Title,
		Slug:               s.Slug,
		LanguageSlug:       s.LanguageSlug,
		Description:        s.Description,
		InProgressParts:    s.SeriesProgressInProgressParts.Int16,
		CompletedParts:     s.SeriesProgressCompletedParts.Int16,
		TotalParts:         s.PartsCount,
		InProgressLectures: s.SeriesProgressInProgressLectures.Int16,
		CompletedLectures:  s.SeriesProgressCompletedLectures.Int16,
		TotalLectures:      s.LecturesCount,
		IsCurrent:          s.SeriesProgressIsCurrent.Bool,
		IsPublished:        s.IsPublished,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindFilteredPublishedSeriesWithAuthorAndProgressSortBySlugRow) ToSeriesDTO() *SeriesDTO {
	return &SeriesDTO{
		ID:                 s.ID,
		Title:              s.Title,
		Slug:               s.Slug,
		LanguageSlug:       s.LanguageSlug,
		Description:        s.Description,
		InProgressParts:    s.SeriesProgressInProgressParts.Int16,
		CompletedParts:     s.SeriesProgressCompletedParts.Int16,
		TotalParts:         s.PartsCount,
		InProgressLectures: s.SeriesProgressInProgressLectures.Int16,
		CompletedLectures:  s.SeriesProgressCompletedLectures.Int16,
		TotalLectures:      s.LecturesCount,
		IsCurrent:          s.SeriesProgressIsCurrent.Bool,
		IsPublished:        s.IsPublished,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindPublishedSeriesBySlugsWithAuthorRow) ToSeriesDTOWithProgress(progress *SeriesProgress) *SeriesDTO {
	return &SeriesDTO{
		ID:                 s.ID,
		Title:              s.Title,
		Slug:               s.Slug,
		LanguageSlug:       s.LanguageSlug,
		Description:        s.Description,
		InProgressParts:    progress.InProgressParts,
		CompletedParts:     progress.CompletedParts,
		TotalParts:         s.PartsCount,
		InProgressLectures: progress.InProgressLectures,
		CompletedLectures:  progress.CompletedLectures,
		TotalLectures:      s.LecturesCount,
		IsCurrent:          progress.IsCurrent,
		IsPublished:        s.IsPublished,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindPublishedSeriesBySlugsWithAuthorRow) ToSeriesDTO() *SeriesDTO {
	return &SeriesDTO{
		ID:                 s.ID,
		Title:              s.Title,
		Slug:               s.Slug,
		LanguageSlug:       s.LanguageSlug,
		Description:        s.Description,
		InProgressParts:    0,
		CompletedParts:     0,
		TotalParts:         s.PartsCount,
		InProgressLectures: 0,
		CompletedLectures:  0,
		TotalLectures:      s.LecturesCount,
		IsCurrent:          false,
		IsPublished:        s.IsPublished,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}
