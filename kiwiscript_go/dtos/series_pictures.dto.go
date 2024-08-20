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
)

// Response

type SeriesPictureLinks struct {
	Self   LinkResponse `json:"self"`
	Series LinkResponse `json:"series"`
}

func newSeriesPictureLinks(
	backendDomain,
	languageSlug,
	seriesSlug string,
) SeriesPictureLinks {
	return SeriesPictureLinks{
		Self: LinkResponse{
			Href: fmt.Sprintf(
				"https://%s%s/%s%s/%s%s",
				backendDomain,
				paths.LanguagePathV1,
				languageSlug,
				paths.SeriesPath,
				seriesSlug,
				paths.PicturePath,
			),
		},
		Series: LinkResponse{
			Href: fmt.Sprintf(
				"https://%s%s/%s%s/%s",
				backendDomain,
				paths.LanguagePathV1,
				languageSlug,
				paths.SeriesPath,
				seriesSlug,
			),
		},
	}
}

type SeriesPictureResponse struct {
	ID    uuid.UUID          `json:"id"`
	EXT   string             `json:"ext"`
	URL   string             `json:"url"`
	Links SeriesPictureLinks `json:"_links"`
}

func NewSeriesPictureResponse(
	backendDomain,
	languageSlug,
	seriesSlug string,
	picture *db.SeriesPictureModel,
) *SeriesPictureResponse {
	return &SeriesPictureResponse{
		ID:    picture.ID,
		EXT:   picture.EXT,
		URL:   picture.URL,
		Links: newSeriesPictureLinks(backendDomain, languageSlug, seriesSlug),
	}
}
