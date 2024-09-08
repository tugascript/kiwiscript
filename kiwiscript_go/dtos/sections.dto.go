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

package dtos

import (
	"fmt"
	"github.com/kiwiscript/kiwiscript_go/paths"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
)

// Path Params

type SectionPathParams struct {
	LanguageSlug string `validate:"required,min=2,max=50,slug"`
	SeriesSlug   string `validate:"required,min=2,max=100,slug"`
	SectionID    string `validate:"required,number,min=1"`
}

// Bodies

type CreateSectionBody struct {
	Title       string `json:"title" validate:"required,min=2,max=250"`
	Description string `json:"description" validate:"required,min=2"`
}

type UpdateSectionBody struct {
	Title       string `json:"title" validate:"required,min=2,max=250"`
	Description string `json:"description" validate:"required,min=2"`
	Position    int16  `json:"position" validate:"required,gte=1"`
}

type SectionLinks struct {
	Self     LinkResponse `json:"self"`
	Series   LinkResponse `json:"series"`
	Lectures LinkResponse `json:"lectures"`
}

func newSectionLinks(backendDomain, languageSlug, seriesSlug string, sectionID int32) SectionLinks {
	return SectionLinks{
		Self: LinkResponse{
			Href: fmt.Sprintf(
				"https://%s/api%s/%s%s/%s%s/%d",
				backendDomain,
				paths.LanguagePathV1,
				languageSlug,
				paths.SeriesPath,
				seriesSlug,
				paths.SectionsPath,
				sectionID,
			),
		},
		Series: LinkResponse{
			Href: fmt.Sprintf(
				"https://%s/api%s/%s%s/%s",
				backendDomain,
				paths.LanguagePathV1,
				languageSlug,
				paths.SeriesPath,
				seriesSlug,
			),
		},
		Lectures: LinkResponse{
			Href: fmt.Sprintf(
				"https://%s/api%s/%s%s/%s%s/%d%s",
				backendDomain,
				paths.LanguagePathV1,
				languageSlug,
				paths.SeriesPath,
				seriesSlug,
				paths.SectionsPath,
				sectionID,
				paths.LessonsPath,
			),
		},
	}
}

type SectionLesson struct {
	ID               int32
	Title            string
	Position         int16
	WatchTimeSeconds int32
	ReadTimeSeconds  int32
	Links            SelfLinkResponse `json:"_links"`
}

type SectionEmbedded struct {
	Lessons []SectionLesson `json:"lessons"`
}

func newSectionEmbedded(backendDomain string, lessons []db.Lesson) *SectionEmbedded {
	if lessons == nil {
		return nil
	}

	sectionLessons := make([]SectionLesson, 0, len(lessons))
	for _, l := range lessons {
		sectionLessons = append(sectionLessons, SectionLesson{
			ID:               l.ID,
			Title:            l.Title,
			Position:         l.Position,
			WatchTimeSeconds: l.WatchTimeSeconds,
			ReadTimeSeconds:  l.ReadTimeSeconds,
			Links: SelfLinkResponse{
				Self: LinkResponse{
					Href: fmt.Sprintf(
						"https://%s/api%s/%s%s/%s%s/%d%s/%d",
						backendDomain,
						paths.LanguagePathV1,
						l.LanguageSlug,
						paths.SeriesPath,
						l.SeriesSlug,
						paths.SectionsPath,
						l.SectionID,
						paths.LessonsPath,
						l.ID,
					),
				},
			},
		})
	}

	return &SectionEmbedded{Lessons: sectionLessons}
}

type SectionResponse struct {
	ID               int32            `json:"id"`
	Title            string           `json:"title"`
	Description      string           `json:"description"`
	Position         int16            `json:"position"`
	CompletedLessons int16            `json:"completedLessons"`
	TotalLessons     int16            `json:"totalLessons"`
	IsCompleted      bool             `json:"isCompleted"`
	ReadTime         int32            `json:"readTime"`
	WatchTime        int32            `json:"watchTime"`
	IsPublished      bool             `json:"isPublished"`
	ViewedAt         string           `json:"viewedAt,omitempty"`
	Links            SectionLinks     `json:"_links"`
	Embedded         *SectionEmbedded `json:"_embedded,omitempty"`
}

func NewSectionResponse(backendDomain string, section *db.SectionModel, lessons []db.Lesson) *SectionResponse {
	return &SectionResponse{
		ID:               section.ID,
		Title:            section.Title,
		Description:      section.Description,
		Position:         section.Position,
		CompletedLessons: section.CompletedLessons,
		TotalLessons:     section.TotalLessons,
		IsCompleted:      section.IsCompleted,
		ReadTime:         section.ReadTimeSeconds,
		WatchTime:        section.WatchTimeSeconds,
		IsPublished:      section.IsPublished,
		ViewedAt:         section.ViewedAt,
		Links:            newSectionLinks(backendDomain, section.LanguageSlug, section.SeriesSlug, section.ID),
		Embedded:         newSectionEmbedded(backendDomain, lessons),
	}
}
