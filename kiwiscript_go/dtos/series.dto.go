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
	"github.com/google/uuid"
	"github.com/kiwiscript/kiwiscript_go/paths"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"net/url"
)

// Bodies

type CreateSeriesBody struct {
	Title       string `json:"title" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"required,min=2"`
}

type UpdateSeriesBody struct {
	Title       string `json:"title" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"required,min=2"`
}

// Path Params

type SeriesPathParams struct {
	LanguageSlug string `validate:"required,min=2,max=50,slug"`
	SeriesSlug   string `validate:"required,min=2,max=100,slug"`
}

// Query params

type SeriesQueryParams struct {
	Search string `validate:"omitempty,min=1,max=100"`
	Limit  int32  `validate:"omitempty,gte=1,lte=100"`
	Offset int32  `validate:"omitempty,gte=0"`
	SortBy string `validate:"omitempty,oneof=slug date"`
}

func (p SeriesQueryParams) ToQueryString() string {
	params := make(url.Values)

	if p.Search != "" {
		params.Add("search", p.Search)
	}
	if p.SortBy != "" {
		params.Add("sortBy", p.SortBy)
	}

	return params.Encode()
}
func (p SeriesQueryParams) GetLimit() int32 {
	return p.Limit
}
func (p SeriesQueryParams) GetOffset() int32 {
	return p.Offset
}

type SeriesLinks struct {
	Self     LinkResponse  `json:"self"`
	Author   LinkResponse  `json:"author"`
	Language LinkResponse  `json:"language"`
	Sections LinkResponse  `json:"parts"`
	Picture  *LinkResponse `json:"picture,omitempty"`
}

func newSeriesLinks(backendDomain, languageSlug, seriesSlug string, authorID int32, withPicture bool) SeriesLinks {
	var picture *LinkResponse
	if withPicture {
		picture = &LinkResponse{
			Href: fmt.Sprintf(
				"https://%s/api%s/%s%s/%s%s",
				backendDomain,
				paths.LanguagePathV1,
				languageSlug,
				paths.SeriesPath,
				seriesSlug,
				paths.PicturePath,
			),
		}
	}

	return SeriesLinks{
		Self: LinkResponse{
			fmt.Sprintf(
				"https://%s/api%s/%s%s/%s",
				backendDomain,
				paths.LanguagePathV1,
				languageSlug,
				paths.SeriesPath,
				seriesSlug,
			),
		},
		Language: LinkResponse{
			fmt.Sprintf("https://%s/api%s/%s", backendDomain, paths.LanguagePathV1, languageSlug),
		},
		Author: LinkResponse{
			fmt.Sprintf("https://%s/api%s/%d", backendDomain, paths.UsersPathV1, authorID),
		},
		Sections: LinkResponse{
			fmt.Sprintf(
				"https://%s/api%s/%s%s/%s%s",
				backendDomain,
				paths.LanguagePathV1,
				languageSlug,
				paths.SeriesPath,
				seriesSlug,
				paths.SectionsPath,
			),
		},
		Picture: picture,
	}
}

type SeriesAuthorEmbedded struct {
	ID        int32            `json:"id"`
	FirstName string           `json:"firstName"`
	LastName  string           `json:"lastName"`
	Links     SelfLinkResponse `json:"_links"`
}

type SeriesPictureEmbedded struct {
	ID    uuid.UUID        `json:"ID"`
	EXT   string           `json:"ext"`
	URL   string           `json:"url"`
	Links SelfLinkResponse `json:"_links"`
}

type SeriesEmbedded struct {
	Author  SeriesAuthorEmbedded   `json:"author"`
	Picture *SeriesPictureEmbedded `json:"picture,omitempty"`
}

func newSeriesEmbedded(
	backendDomain string,
	author *db.SeriesAuthor,
	picture *db.SeriesPictureIDAndEXT,
	pictureURL,
	languageSlug,
	seriesSlug string,
) SeriesEmbedded {
	var pictureEmbedded *SeriesPictureEmbedded
	if picture != nil {
		pictureEmbedded = &SeriesPictureEmbedded{
			ID:  picture.ID,
			EXT: picture.EXT,
			URL: pictureURL,
			Links: SelfLinkResponse{
				Self: LinkResponse{
					Href: fmt.Sprintf(
						"https://%s/api%s/%s%s/%s",
						backendDomain,
						paths.LanguagePathV1,
						languageSlug,
						paths.SeriesPath,
						seriesSlug,
					),
				},
			},
		}
	}

	return SeriesEmbedded{
		Author: SeriesAuthorEmbedded{
			ID:        author.ID,
			FirstName: author.FirstName,
			LastName:  author.LastName,
			Links: SelfLinkResponse{
				LinkResponse{
					fmt.Sprintf("https://%s/api%s/%d", backendDomain, paths.UsersPathV1, author.ID),
				},
			},
		},
		Picture: pictureEmbedded,
	}
}

type SeriesResponse struct {
	ID                int32          `json:"id"`
	Title             string         `json:"title"`
	Slug              string         `json:"slug"`
	Description       string         `json:"description"`
	CompletedSections int16          `json:"completedSections"`
	TotalSections     int16          `json:"totalSections"`
	CompletedLessons  int16          `json:"completedLessons"`
	TotalLessons      int16          `json:"totalLessons"`
	WatchTime         int32          `json:"watchTime"`
	ReadTime          int32          `json:"readTime"`
	ViewedAt          string         `json:"viewedAt,omitempty"`
	IsPublished       bool           `json:"isPublished"`
	Embedded          SeriesEmbedded `json:"_embedded"`
	Links             SeriesLinks    `json:"_links"`
}

func NewSeriesResponse(backendDomain string, model *db.SeriesModel, pictureURL string) *SeriesResponse {
	return &SeriesResponse{
		ID:                model.ID,
		Title:             model.Title,
		Slug:              model.Slug,
		Description:       model.Description,
		IsPublished:       model.IsPublished,
		CompletedSections: model.CompletedSections,
		TotalSections:     model.TotalSections,
		CompletedLessons:  model.CompletedLessons,
		TotalLessons:      model.TotalLessons,
		WatchTime:         model.WatchTime,
		ReadTime:          model.ReadTime,
		ViewedAt:          model.ViewedAt,
		Embedded: newSeriesEmbedded(
			backendDomain,
			&model.Author,
			model.Picture,
			pictureURL,
			model.LanguageSlug,
			model.Slug,
		),
		Links: newSeriesLinks(
			backendDomain,
			model.LanguageSlug,
			model.Slug,
			model.Author.ID,
			pictureURL == "",
		),
	}
}
