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

type SeriesModel struct {
	ID                int32
	Title             string
	Slug              string
	LanguageSlug      string
	Description       string
	CompletedSections int16
	TotalSections     int16
	CompletedLessons  int16
	TotalLessons      int16
	WatchTime         int32
	ReadTime          int32
	IsPublished       bool
	Author            SeriesAuthor
}

type ToSeriesModel interface {
	ToSeriesModel() *SeriesModel
}

type ToSeriesModelWithAuthor interface {
	ToSeriesModelWithAuthor(authorID int32, firstName, lastName string) *SeriesModel
}

type ToSeriesModelWithProgress interface {
	ToSeriesModelWithProgress(progress *SeriesProgress) *SeriesModel
	ToSeriesModel() *SeriesModel
}

func (s *Series) ToSeriesModelWithAuthor(authorID int32, firstName, lastName string) *SeriesModel {
	return &SeriesModel{
		ID:               s.ID,
		Title:            s.Title,
		Slug:             s.Slug,
		LanguageSlug:     s.LanguageSlug,
		Description:      s.Description,
		TotalSections:    s.SectionsCount,
		WatchTime:        s.WatchTimeSeconds,
		ReadTime:         s.ReadTimeSeconds,
		CompletedLessons: 0,
		TotalLessons:     s.LessonsCount,
		IsPublished:      s.IsPublished,
		Author: SeriesAuthor{
			ID:        authorID,
			FirstName: firstName,
			LastName:  lastName,
		},
	}
}

func (s *FindSeriesBySlugWithAuthorRow) ToSeriesModel() *SeriesModel {
	return &SeriesModel{
		ID:               s.ID,
		Title:            s.Title,
		Slug:             s.Slug,
		LanguageSlug:     s.LanguageSlug,
		Description:      s.Description,
		TotalSections:    s.SectionsCount,
		CompletedLessons: 0,
		TotalLessons:     s.LessonsCount,
		IsPublished:      s.IsPublished,
		WatchTime:        s.WatchTimeSeconds,
		ReadTime:         s.ReadTimeSeconds,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindPublishedSeriesBySlugWithAuthorAndProgressRow) ToSeriesModel() *SeriesModel {
	return &SeriesModel{
		ID:               s.ID,
		Title:            s.Title,
		Slug:             s.Slug,
		LanguageSlug:     s.LanguageSlug,
		Description:      s.Description,
		TotalSections:    s.SectionsCount,
		CompletedLessons: s.SeriesProgressCompletedLessons.Int16,
		TotalLessons:     s.LessonsCount,
		IsPublished:      s.IsPublished,
		WatchTime:        s.WatchTimeSeconds,
		ReadTime:         s.ReadTimeSeconds,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindPaginatedSeriesWithAuthorSortByIDRow) ToSeriesModel() *SeriesModel {
	return &SeriesModel{
		ID:               s.ID,
		Title:            s.Title,
		Slug:             s.Slug,
		LanguageSlug:     s.LanguageSlug,
		Description:      s.Description,
		TotalSections:    s.SectionsCount,
		CompletedLessons: 0,
		TotalLessons:     s.LessonsCount,
		IsPublished:      s.IsPublished,
		WatchTime:        s.WatchTimeSeconds,
		ReadTime:         s.ReadTimeSeconds,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindPaginatedPublishedSeriesWithAuthorSortByIDRow) ToSeriesModel() *SeriesModel {
	return &SeriesModel{
		ID:               s.ID,
		Title:            s.Title,
		Slug:             s.Slug,
		LanguageSlug:     s.LanguageSlug,
		Description:      s.Description,
		TotalSections:    s.SectionsCount,
		CompletedLessons: 0,
		TotalLessons:     s.LessonsCount,
		IsPublished:      s.IsPublished,
		WatchTime:        s.WatchTimeSeconds,
		ReadTime:         s.ReadTimeSeconds,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindFilteredPublishedSeriesWithAuthorSortByIDRow) ToSeriesModel() *SeriesModel {
	return &SeriesModel{
		ID:               s.ID,
		Title:            s.Title,
		Slug:             s.Slug,
		LanguageSlug:     s.LanguageSlug,
		Description:      s.Description,
		TotalSections:    s.SectionsCount,
		CompletedLessons: 0,
		TotalLessons:     s.LessonsCount,
		IsPublished:      s.IsPublished,
		WatchTime:        s.WatchTimeSeconds,
		ReadTime:         s.ReadTimeSeconds,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindPaginatedSeriesWithAuthorSortBySlugRow) ToSeriesModel() *SeriesModel {
	return &SeriesModel{
		ID:               s.ID,
		Title:            s.Title,
		Slug:             s.Slug,
		LanguageSlug:     s.LanguageSlug,
		Description:      s.Description,
		TotalSections:    s.SectionsCount,
		CompletedLessons: 0,
		TotalLessons:     s.LessonsCount,
		IsPublished:      s.IsPublished,
		WatchTime:        s.WatchTimeSeconds,
		ReadTime:         s.ReadTimeSeconds,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindPaginatedPublishedSeriesWithAuthorSortBySlugRow) ToSeriesModel() *SeriesModel {
	return &SeriesModel{
		ID:               s.ID,
		Title:            s.Title,
		Slug:             s.Slug,
		LanguageSlug:     s.LanguageSlug,
		Description:      s.Description,
		TotalSections:    s.SectionsCount,
		CompletedLessons: 0,
		TotalLessons:     s.LessonsCount,
		IsPublished:      s.IsPublished,
		WatchTime:        s.WatchTimeSeconds,
		ReadTime:         s.ReadTimeSeconds,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindFilteredPublishedSeriesWithAuthorSortBySlugRow) ToSeriesModel() *SeriesModel {
	return &SeriesModel{
		ID:               s.ID,
		Title:            s.Title,
		Slug:             s.Slug,
		LanguageSlug:     s.LanguageSlug,
		Description:      s.Description,
		TotalSections:    s.SectionsCount,
		CompletedLessons: 0,
		TotalLessons:     s.LessonsCount,
		IsPublished:      s.IsPublished,
		WatchTime:        s.WatchTimeSeconds,
		ReadTime:         s.ReadTimeSeconds,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindFilteredSeriesWithAuthorSortBySlugRow) ToSeriesModel() *SeriesModel {
	return &SeriesModel{
		ID:               s.ID,
		Title:            s.Title,
		Slug:             s.Slug,
		LanguageSlug:     s.LanguageSlug,
		Description:      s.Description,
		TotalSections:    s.SectionsCount,
		CompletedLessons: 0,
		TotalLessons:     s.LessonsCount,
		IsPublished:      s.IsPublished,
		WatchTime:        s.WatchTimeSeconds,
		ReadTime:         s.ReadTimeSeconds,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindFilteredSeriesWithAuthorSortByIDRow) ToSeriesModel() *SeriesModel {
	return &SeriesModel{
		ID:               s.ID,
		Title:            s.Title,
		Slug:             s.Slug,
		LanguageSlug:     s.LanguageSlug,
		Description:      s.Description,
		TotalSections:    s.SectionsCount,
		CompletedLessons: 0,
		TotalLessons:     s.LessonsCount,
		IsPublished:      s.IsPublished,
		WatchTime:        s.WatchTimeSeconds,
		ReadTime:         s.ReadTimeSeconds,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindPaginatedPublishedSeriesWithAuthorAndProgressSortByIDRow) ToSeriesModel() *SeriesModel {
	return &SeriesModel{
		ID:                s.ID,
		Title:             s.Title,
		Slug:              s.Slug,
		LanguageSlug:      s.LanguageSlug,
		Description:       s.Description,
		CompletedSections: s.SeriesProgressCompletedSections.Int16,
		TotalSections:     s.SectionsCount,
		CompletedLessons:  s.SeriesProgressCompletedLessons.Int16,
		TotalLessons:      s.LessonsCount,
		IsPublished:       s.IsPublished,
		WatchTime:         s.WatchTimeSeconds,
		ReadTime:          s.ReadTimeSeconds,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindPaginatedPublishedSeriesWithAuthorAndProgressSortBySlugRow) ToSeriesModel() *SeriesModel {
	return &SeriesModel{
		ID:                s.ID,
		Title:             s.Title,
		Slug:              s.Slug,
		LanguageSlug:      s.LanguageSlug,
		Description:       s.Description,
		CompletedSections: s.SeriesProgressCompletedSections.Int16,
		TotalSections:     s.SectionsCount,
		CompletedLessons:  s.SeriesProgressCompletedLessons.Int16,
		TotalLessons:      s.LessonsCount,
		IsPublished:       s.IsPublished,
		WatchTime:         s.WatchTimeSeconds,
		ReadTime:          s.ReadTimeSeconds,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindFilteredPublishedSeriesWithAuthorAndProgressSortByIDRow) ToSeriesModel() *SeriesModel {
	return &SeriesModel{
		ID:                s.ID,
		Title:             s.Title,
		Slug:              s.Slug,
		LanguageSlug:      s.LanguageSlug,
		Description:       s.Description,
		CompletedSections: s.SeriesProgressCompletedSections.Int16,
		TotalSections:     s.SectionsCount,
		CompletedLessons:  s.SeriesProgressCompletedLessons.Int16,
		TotalLessons:      s.LessonsCount,
		IsPublished:       s.IsPublished,
		WatchTime:         s.WatchTimeSeconds,
		ReadTime:          s.ReadTimeSeconds,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindFilteredPublishedSeriesWithAuthorAndProgressSortBySlugRow) ToSeriesModel() *SeriesModel {
	return &SeriesModel{
		ID:                s.ID,
		Title:             s.Title,
		Slug:              s.Slug,
		LanguageSlug:      s.LanguageSlug,
		Description:       s.Description,
		CompletedSections: s.SeriesProgressCompletedSections.Int16,
		TotalSections:     s.SectionsCount,
		CompletedLessons:  s.SeriesProgressCompletedLessons.Int16,
		TotalLessons:      s.LessonsCount,
		IsPublished:       s.IsPublished,
		WatchTime:         s.WatchTimeSeconds,
		ReadTime:          s.ReadTimeSeconds,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindPublishedSeriesBySlugsWithAuthorRow) ToSeriesModelWithProgress(progress *SeriesProgress) *SeriesModel {
	return &SeriesModel{
		ID:                s.ID,
		Title:             s.Title,
		Slug:              s.Slug,
		LanguageSlug:      s.LanguageSlug,
		Description:       s.Description,
		CompletedSections: progress.CompletedSections,
		TotalSections:     s.SectionsCount,
		CompletedLessons:  progress.CompletedLessons,
		TotalLessons:      s.LessonsCount,
		IsPublished:       s.IsPublished,
		WatchTime:         s.WatchTimeSeconds,
		ReadTime:          s.ReadTimeSeconds,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}

func (s *FindPublishedSeriesBySlugsWithAuthorRow) ToSeriesModel() *SeriesModel {
	return &SeriesModel{
		ID:               s.ID,
		Title:            s.Title,
		Slug:             s.Slug,
		LanguageSlug:     s.LanguageSlug,
		Description:      s.Description,
		TotalSections:    s.SectionsCount,
		CompletedLessons: 0,
		TotalLessons:     s.LessonsCount,
		IsPublished:      s.IsPublished,
		WatchTime:        s.WatchTimeSeconds,
		ReadTime:         s.ReadTimeSeconds,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
	}
}
