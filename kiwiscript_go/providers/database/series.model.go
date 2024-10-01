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

import (
	"github.com/google/uuid"
	"time"
)

type SeriesAuthor struct {
	ID        int32
	FirstName string
	LastName  string
}

type SeriesPictureIDAndEXT struct {
	ID  uuid.UUID
	EXT string
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
	ViewedAt          string
	CompletedAt       string
	IsPublished       bool
	Author            SeriesAuthor
	Picture           *SeriesPictureIDAndEXT
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

func (s *Series) ToSeriesModelWithAuthorAndPicture(
	authorID int32,
	firstName,
	lastName string,
	pictureID uuid.UUID,
	pictureEXT string,
) *SeriesModel {
	return &SeriesModel{
		ID:            s.ID,
		Title:         s.Title,
		Slug:          s.Slug,
		LanguageSlug:  s.LanguageSlug,
		Description:   s.Description,
		TotalSections: s.SectionsCount,
		WatchTime:     s.WatchTimeSeconds,
		ReadTime:      s.ReadTimeSeconds,
		TotalLessons:  s.LessonsCount,
		IsPublished:   s.IsPublished,
		Author: SeriesAuthor{
			ID:        authorID,
			FirstName: firstName,
			LastName:  lastName,
		},
		Picture: &SeriesPictureIDAndEXT{
			ID:  pictureID,
			EXT: pictureEXT,
		},
	}
}

func (s *FindSeriesBySlugWithAuthorRow) ToSeriesModel() *SeriesModel {
	var picture *SeriesPictureIDAndEXT
	if s.PictureID.Valid && s.PictureExt.Valid {
		picture = &SeriesPictureIDAndEXT{
			ID:  s.PictureID.Bytes,
			EXT: s.PictureExt.String,
		}
	}

	return &SeriesModel{
		ID:            s.ID,
		Title:         s.Title,
		Slug:          s.Slug,
		LanguageSlug:  s.LanguageSlug,
		Description:   s.Description,
		TotalSections: s.SectionsCount,
		TotalLessons:  s.LessonsCount,
		IsPublished:   s.IsPublished,
		WatchTime:     s.WatchTimeSeconds,
		ReadTime:      s.ReadTimeSeconds,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
		Picture: picture,
	}
}

func (s *FindPublishedSeriesBySlugWithAuthorAndProgressRow) ToSeriesModel() *SeriesModel {
	var picture *SeriesPictureIDAndEXT
	if s.PictureID.Valid && s.PictureExt.Valid {
		picture = &SeriesPictureIDAndEXT{
			ID:  s.PictureID.Bytes,
			EXT: s.PictureExt.String,
		}
	}

	var viewedAt string
	if s.SeriesProgressViewedAt.Valid {
		viewedAt = s.SeriesProgressViewedAt.Time.Format(time.RFC3339)
	}

	var completedAt string
	if s.SeriesProgressCompletedAt.Valid {
		completedAt = s.SeriesProgressCompletedAt.Time.Format(time.RFC3339)
	}

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
		ViewedAt:          viewedAt,
		CompletedAt:       completedAt,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
		Picture: picture,
	}
}

func (s *FindPaginatedSeriesWithAuthorSortByIDRow) ToSeriesModel() *SeriesModel {
	var picture *SeriesPictureIDAndEXT
	if s.PictureID.Valid && s.PictureExt.Valid {
		picture = &SeriesPictureIDAndEXT{
			ID:  s.PictureID.Bytes,
			EXT: s.PictureExt.String,
		}
	}

	return &SeriesModel{
		ID:            s.ID,
		Title:         s.Title,
		Slug:          s.Slug,
		LanguageSlug:  s.LanguageSlug,
		Description:   s.Description,
		TotalSections: s.SectionsCount,
		TotalLessons:  s.LessonsCount,
		IsPublished:   s.IsPublished,
		WatchTime:     s.WatchTimeSeconds,
		ReadTime:      s.ReadTimeSeconds,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
		Picture: picture,
	}
}

func (s *FindPaginatedPublishedSeriesWithAuthorSortByIDRow) ToSeriesModel() *SeriesModel {
	var picture *SeriesPictureIDAndEXT
	if s.PictureID.Valid && s.PictureExt.Valid {
		picture = &SeriesPictureIDAndEXT{
			ID:  s.PictureID.Bytes,
			EXT: s.PictureExt.String,
		}
	}

	return &SeriesModel{
		ID:            s.ID,
		Title:         s.Title,
		Slug:          s.Slug,
		LanguageSlug:  s.LanguageSlug,
		Description:   s.Description,
		TotalSections: s.SectionsCount,
		TotalLessons:  s.LessonsCount,
		IsPublished:   s.IsPublished,
		WatchTime:     s.WatchTimeSeconds,
		ReadTime:      s.ReadTimeSeconds,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
		Picture: picture,
	}
}

func (s *FindFilteredPublishedSeriesWithAuthorSortByIDRow) ToSeriesModel() *SeriesModel {
	var picture *SeriesPictureIDAndEXT
	if s.PictureID.Valid && s.PictureExt.Valid {
		picture = &SeriesPictureIDAndEXT{
			ID:  s.PictureID.Bytes,
			EXT: s.PictureExt.String,
		}
	}

	return &SeriesModel{
		ID:            s.ID,
		Title:         s.Title,
		Slug:          s.Slug,
		LanguageSlug:  s.LanguageSlug,
		Description:   s.Description,
		TotalSections: s.SectionsCount,
		TotalLessons:  s.LessonsCount,
		IsPublished:   s.IsPublished,
		WatchTime:     s.WatchTimeSeconds,
		ReadTime:      s.ReadTimeSeconds,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
		Picture: picture,
	}
}

func (s *FindPaginatedSeriesWithAuthorSortBySlugRow) ToSeriesModel() *SeriesModel {
	var picture *SeriesPictureIDAndEXT
	if s.PictureID.Valid && s.PictureExt.Valid {
		picture = &SeriesPictureIDAndEXT{
			ID:  s.PictureID.Bytes,
			EXT: s.PictureExt.String,
		}
	}

	return &SeriesModel{
		ID:            s.ID,
		Title:         s.Title,
		Slug:          s.Slug,
		LanguageSlug:  s.LanguageSlug,
		Description:   s.Description,
		TotalSections: s.SectionsCount,
		TotalLessons:  s.LessonsCount,
		IsPublished:   s.IsPublished,
		WatchTime:     s.WatchTimeSeconds,
		ReadTime:      s.ReadTimeSeconds,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
		Picture: picture,
	}
}

func (s *FindPaginatedPublishedSeriesWithAuthorSortBySlugRow) ToSeriesModel() *SeriesModel {
	var picture *SeriesPictureIDAndEXT
	if s.PictureID.Valid && s.PictureExt.Valid {
		picture = &SeriesPictureIDAndEXT{
			ID:  s.PictureID.Bytes,
			EXT: s.PictureExt.String,
		}
	}

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
		Picture: picture,
	}
}

func (s *FindFilteredPublishedSeriesWithAuthorSortBySlugRow) ToSeriesModel() *SeriesModel {
	var picture *SeriesPictureIDAndEXT
	if s.PictureID.Valid && s.PictureExt.Valid {
		picture = &SeriesPictureIDAndEXT{
			ID:  s.PictureID.Bytes,
			EXT: s.PictureExt.String,
		}
	}

	return &SeriesModel{
		ID:            s.ID,
		Title:         s.Title,
		Slug:          s.Slug,
		LanguageSlug:  s.LanguageSlug,
		Description:   s.Description,
		TotalSections: s.SectionsCount,
		TotalLessons:  s.LessonsCount,
		IsPublished:   s.IsPublished,
		WatchTime:     s.WatchTimeSeconds,
		ReadTime:      s.ReadTimeSeconds,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
		Picture: picture,
	}
}

func (s *FindFilteredSeriesWithAuthorSortBySlugRow) ToSeriesModel() *SeriesModel {
	var picture *SeriesPictureIDAndEXT
	if s.PictureID.Valid && s.PictureExt.Valid {
		picture = &SeriesPictureIDAndEXT{
			ID:  s.PictureID.Bytes,
			EXT: s.PictureExt.String,
		}
	}

	return &SeriesModel{
		ID:            s.ID,
		Title:         s.Title,
		Slug:          s.Slug,
		LanguageSlug:  s.LanguageSlug,
		Description:   s.Description,
		TotalSections: s.SectionsCount,
		TotalLessons:  s.LessonsCount,
		IsPublished:   s.IsPublished,
		WatchTime:     s.WatchTimeSeconds,
		ReadTime:      s.ReadTimeSeconds,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
		Picture: picture,
	}
}

func (s *FindFilteredSeriesWithAuthorSortByIDRow) ToSeriesModel() *SeriesModel {
	var picture *SeriesPictureIDAndEXT
	if s.PictureID.Valid && s.PictureExt.Valid {
		picture = &SeriesPictureIDAndEXT{
			ID:  s.PictureID.Bytes,
			EXT: s.PictureExt.String,
		}
	}

	return &SeriesModel{
		ID:            s.ID,
		Title:         s.Title,
		Slug:          s.Slug,
		LanguageSlug:  s.LanguageSlug,
		Description:   s.Description,
		TotalSections: s.SectionsCount,
		TotalLessons:  s.LessonsCount,
		IsPublished:   s.IsPublished,
		WatchTime:     s.WatchTimeSeconds,
		ReadTime:      s.ReadTimeSeconds,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
		Picture: picture,
	}
}

func (s *FindPaginatedPublishedSeriesWithAuthorAndProgressSortByIDRow) ToSeriesModel() *SeriesModel {
	var picture *SeriesPictureIDAndEXT
	if s.PictureID.Valid && s.PictureExt.Valid {
		picture = &SeriesPictureIDAndEXT{
			ID:  s.PictureID.Bytes,
			EXT: s.PictureExt.String,
		}
	}

	var viewedAt string
	if s.SeriesProgressViewedAt.Valid {
		viewedAt = s.SeriesProgressViewedAt.Time.Format(time.RFC3339)
	}

	var completedAt string
	if s.SeriesProgressCompletedAt.Valid {
		completedAt = s.SeriesProgressCompletedAt.Time.Format(time.RFC3339)
	}

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
		ViewedAt:          viewedAt,
		CompletedAt:       completedAt,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
		Picture: picture,
	}
}

func (s *FindPaginatedPublishedSeriesWithAuthorAndProgressSortBySlugRow) ToSeriesModel() *SeriesModel {
	var picture *SeriesPictureIDAndEXT
	if s.PictureID.Valid && s.PictureExt.Valid {
		picture = &SeriesPictureIDAndEXT{
			ID:  s.PictureID.Bytes,
			EXT: s.PictureExt.String,
		}
	}

	var viewedAt string
	if s.SeriesProgressViewedAt.Valid {
		viewedAt = s.SeriesProgressViewedAt.Time.Format(time.RFC3339)
	}

	var completedAt string
	if s.SeriesProgressCompletedAt.Valid {
		completedAt = s.SeriesProgressCompletedAt.Time.Format(time.RFC3339)
	}

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
		ViewedAt:          viewedAt,
		CompletedAt:       completedAt,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
		Picture: picture,
	}
}

func (s *FindFilteredPublishedSeriesWithAuthorAndProgressSortByIDRow) ToSeriesModel() *SeriesModel {
	var picture *SeriesPictureIDAndEXT
	if s.PictureID.Valid && s.PictureExt.Valid {
		picture = &SeriesPictureIDAndEXT{
			ID:  s.PictureID.Bytes,
			EXT: s.PictureExt.String,
		}
	}

	var viewedAt string
	if s.SeriesProgressViewedAt.Valid {
		viewedAt = s.SeriesProgressViewedAt.Time.Format(time.RFC3339)
	}

	var completedAt string
	if s.SeriesProgressCompletedAt.Valid {
		completedAt = s.SeriesProgressCompletedAt.Time.Format(time.RFC3339)
	}

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
		ViewedAt:          viewedAt,
		CompletedAt:       completedAt,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
		Picture: picture,
	}
}

func (s *FindFilteredPublishedSeriesWithAuthorAndProgressSortBySlugRow) ToSeriesModel() *SeriesModel {
	var picture *SeriesPictureIDAndEXT
	if s.PictureID.Valid && s.PictureExt.Valid {
		picture = &SeriesPictureIDAndEXT{
			ID:  s.PictureID.Bytes,
			EXT: s.PictureExt.String,
		}
	}

	var viewedAt string
	if s.SeriesProgressViewedAt.Valid {
		viewedAt = s.SeriesProgressViewedAt.Time.Format(time.RFC3339)
	}

	var completedAt string
	if s.SeriesProgressCompletedAt.Valid {
		completedAt = s.SeriesProgressCompletedAt.Time.Format(time.RFC3339)
	}

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
		ViewedAt:          viewedAt,
		CompletedAt:       completedAt,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
		Picture: picture,
	}
}

func (s *FindPublishedSeriesBySlugsWithAuthorRow) ToSeriesModelWithProgress(progress *SeriesProgress) *SeriesModel {
	var picture *SeriesPictureIDAndEXT
	if s.PictureID.Valid && s.PictureExt.Valid {
		picture = &SeriesPictureIDAndEXT{
			ID:  s.PictureID.Bytes,
			EXT: s.PictureExt.String,
		}
	}

	var viewedAt string
	if progress.ViewedAt.Valid {
		viewedAt = progress.ViewedAt.Time.Format(time.RFC3339)
	}

	var completedAt string
	if progress.CompletedAt.Valid {
		completedAt = progress.CompletedAt.Time.Format(time.RFC3339)
	}

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
		ViewedAt:          viewedAt,
		CompletedAt:       completedAt,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
		Picture: picture,
	}
}

func (s *FindPublishedSeriesBySlugsWithAuthorRow) ToSeriesModel() *SeriesModel {
	var picture *SeriesPictureIDAndEXT
	if s.PictureID.Valid && s.PictureExt.Valid {
		picture = &SeriesPictureIDAndEXT{
			ID:  s.PictureID.Bytes,
			EXT: s.PictureExt.String,
		}
	}

	return &SeriesModel{
		ID:            s.ID,
		Title:         s.Title,
		Slug:          s.Slug,
		LanguageSlug:  s.LanguageSlug,
		Description:   s.Description,
		TotalSections: s.SectionsCount,
		TotalLessons:  s.LessonsCount,
		IsPublished:   s.IsPublished,
		WatchTime:     s.WatchTimeSeconds,
		ReadTime:      s.ReadTimeSeconds,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
		Picture: picture,
	}
}

func (s *FindPaginatedPublishedSeriesWithAuthorAndInnerProgressRow) ToSeriesModel() *SeriesModel {
	var picture *SeriesPictureIDAndEXT
	if s.PictureID.Valid && s.PictureExt.Valid {
		picture = &SeriesPictureIDAndEXT{
			ID:  s.PictureID.Bytes,
			EXT: s.PictureExt.String,
		}
	}

	var completedAt string
	if s.SeriesProgressCompletedAt.Valid {
		completedAt = s.SeriesProgressCompletedAt.Time.Format(time.RFC3339)
	}

	return &SeriesModel{
		ID:                s.ID,
		Title:             s.Title,
		Slug:              s.Slug,
		LanguageSlug:      s.LanguageSlug,
		Description:       s.Description,
		CompletedSections: s.SeriesProgressCompletedSections,
		TotalSections:     s.SectionsCount,
		CompletedLessons:  s.SeriesProgressCompletedLessons,
		TotalLessons:      s.LessonsCount,
		IsPublished:       s.IsPublished,
		WatchTime:         s.WatchTimeSeconds,
		ReadTime:          s.ReadTimeSeconds,
		ViewedAt:          s.SeriesProgressViewedAt.Time.Format(time.RFC3339),
		CompletedAt:       completedAt,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
		Picture: picture,
	}
}

func (s *FindPaginatedDiscoverySeriesWithAuthorRow) ToSeriesModel() *SeriesModel {
	var picture *SeriesPictureIDAndEXT
	if s.PictureID.Valid && s.PictureExt.Valid {
		picture = &SeriesPictureIDAndEXT{
			ID:  s.PictureID.Bytes,
			EXT: s.PictureExt.String,
		}
	}

	return &SeriesModel{
		ID:            s.ID,
		Title:         s.Title,
		Slug:          s.Slug,
		LanguageSlug:  s.LanguageSlug,
		Description:   s.Description,
		TotalSections: s.SectionsCount,
		TotalLessons:  s.LessonsCount,
		IsPublished:   s.IsPublished,
		WatchTime:     s.WatchTimeSeconds,
		ReadTime:      s.ReadTimeSeconds,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
		Picture: picture,
	}
}

func (s *FindFilteredDiscoverySeriesWithAuthorRow) ToSeriesModel() *SeriesModel {
	var picture *SeriesPictureIDAndEXT
	if s.PictureID.Valid && s.PictureExt.Valid {
		picture = &SeriesPictureIDAndEXT{
			ID:  s.PictureID.Bytes,
			EXT: s.PictureExt.String,
		}
	}

	return &SeriesModel{
		ID:            s.ID,
		Title:         s.Title,
		Slug:          s.Slug,
		LanguageSlug:  s.LanguageSlug,
		Description:   s.Description,
		TotalSections: s.SectionsCount,
		TotalLessons:  s.LessonsCount,
		IsPublished:   s.IsPublished,
		WatchTime:     s.WatchTimeSeconds,
		ReadTime:      s.ReadTimeSeconds,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
		Picture: picture,
	}
}

func (s *FindPaginatedDiscoverySeriesWithAuthorAndProgressRow) ToSeriesModel() *SeriesModel {
	var picture *SeriesPictureIDAndEXT
	if s.PictureID.Valid && s.PictureExt.Valid {
		picture = &SeriesPictureIDAndEXT{
			ID:  s.PictureID.Bytes,
			EXT: s.PictureExt.String,
		}
	}

	var viewedAt string
	if s.SeriesProgressViewedAt.Valid {
		viewedAt = s.SeriesProgressViewedAt.Time.Format(time.RFC3339)
	}

	var completedAt string
	if s.SeriesProgressCompletedAt.Valid {
		completedAt = s.SeriesProgressCompletedAt.Time.Format(time.RFC3339)
	}

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
		ViewedAt:          viewedAt,
		CompletedAt:       completedAt,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
		Picture: picture,
	}
}

func (s *FindFilteredDiscoverySeriesWithAuthorAndProgressRow) ToSeriesModel() *SeriesModel {
	var picture *SeriesPictureIDAndEXT
	if s.PictureID.Valid && s.PictureExt.Valid {
		picture = &SeriesPictureIDAndEXT{
			ID:  s.PictureID.Bytes,
			EXT: s.PictureExt.String,
		}
	}

	var viewedAt string
	if s.SeriesProgressViewedAt.Valid {
		viewedAt = s.SeriesProgressViewedAt.Time.Format(time.RFC3339)
	}

	var completedAt string
	if s.SeriesProgressCompletedAt.Valid {
		completedAt = s.SeriesProgressCompletedAt.Time.Format(time.RFC3339)
	}

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
		ViewedAt:          viewedAt,
		CompletedAt:       completedAt,
		Author: SeriesAuthor{
			ID:        s.AuthorID,
			FirstName: s.AuthorFirstName,
			LastName:  s.AuthorLastName,
		},
		Picture: picture,
	}
}
