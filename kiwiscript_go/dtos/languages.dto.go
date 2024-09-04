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
	"net/url"
)

// Bodies

type LanguageBody struct {
	Name string `json:"name" validate:"required,min=2,max=50,extalphanum"`
	Icon string `json:"icon" validate:"required,svg"`
}

// Path Params

type LanguagePathParams struct {
	LanguageSlug string `validate:"required,min=2,max=50,slug"`
}

// Query Params

type GetLanguagesQueryParams struct {
	Limit  int32  `validate:"omitempty,gte=1,lte=100"`
	Offset int32  `validate:"omitempty,gte=0"`
	Search string `validate:"omitempty,min=1,max=50,extalphanum"`
}

func (p GetLanguagesQueryParams) ToQueryString() string {
	params := make(url.Values)

	if p.Search != "" {
		params.Add("search", p.Search)
	}

	return params.Encode()
}
func (p GetLanguagesQueryParams) GetLimit() int32 {
	return p.Limit
}
func (p GetLanguagesQueryParams) GetOffset() int32 {
	return p.Offset
}

// Responses

type LanguageLinks struct {
	Self   LinkResponse `json:"self"`
	Series LinkResponse `json:"series"`
}

func newLanguageLinks(backendDomain, languageSlug string) LanguageLinks {
	return LanguageLinks{
		Self: LinkResponse{
			fmt.Sprintf("https://%s/api%s/%s", backendDomain, paths.LanguagePathV1, languageSlug),
		},
		Series: LinkResponse{
			fmt.Sprintf(
				"https://%s/api%s/%s%s",
				backendDomain,
				paths.LanguagePathV1,
				languageSlug,
				paths.SeriesPath,
			),
		},
	}
}

type LanguageResponse struct {
	ID              int32         `json:"id"`
	Name            string        `json:"name"`
	Slug            string        `json:"slug"`
	Icon            string        `json:"icon"`
	CompletedSeries int16         `json:"completedSeries"`
	TotalSeries     int16         `json:"totalSeries"`
	ViewedAt        string        `json:"viewedAt,omitempty"`
	Links           LanguageLinks `json:"_links"`
}

func NewLanguageResponse(backendDomain string, model *db.LanguageModel) *LanguageResponse {
	return &LanguageResponse{
		ID:              model.ID,
		Name:            model.Name,
		Slug:            model.Slug,
		Icon:            model.Icon,
		CompletedSeries: model.CompletedSeries,
		TotalSeries:     model.TotalSeries,
		ViewedAt:        model.ViewedAt,
		Links:           newLanguageLinks(backendDomain, model.Slug),
	}
}
