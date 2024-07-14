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

package controllers

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/kiwiscript/kiwiscript_go/paths"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/providers/tokens"
	"github.com/kiwiscript/kiwiscript_go/services"
)

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

func NewAuthResponse(accessToken, refreshToken string, expiresIn int64) AuthResponse {
	return AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		TokenType:    "Bearer",
	}
}

type MessageResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

func NewMessageResponse(message string) MessageResponse {
	return MessageResponse{
		ID:      uuid.NewString(),
		Message: message,
	}
}

type LinkResponse struct {
	Href string `json:"href"`
}

func (l *LinkResponse) ToRef() *LinkResponse {
	if l == nil || l.Href == "" {
		return nil
	}

	return l
}

type SelfLinkResponse struct {
	Self LinkResponse `json:"self"`
}

type PaginatedResponseLinks struct {
	Self LinkResponse  `json:"self"`
	Next *LinkResponse `json:"next,omitempty"`
	Prev *LinkResponse `json:"previous,omitempty"`
}

type PaginatedResponse[T any] struct {
	Count   int64                  `json:"count"`
	Links   PaginatedResponseLinks `json:"_links"`
	Results []T                    `json:"results"`
}

func newPaginatedNavigationURL(frontendDomain, path string, params QueryStr, limit, offset int32) LinkResponse {
	if offset < 0 {
		offset = 0
	}

	var href string
	if params != QueryStr("") {
		href = fmt.Sprintf("https://%s/api/%s?%s&limit=%d&offset=%d", frontendDomain, path, params, limit, offset)
	} else {
		href = fmt.Sprintf("https://%s/api/%s?limit=%d&offset=%d", frontendDomain, path, limit, offset)
	}

	return LinkResponse{href}
}

func NewPaginatedResponse[T any, V any](
	backendDomain,
	path string,
	params FromQueryParams,
	count int64,
	entites []V,
	mapper func(*V) T,
) PaginatedResponse[T] {
	results := make([]T, len(entites))

	for i, entity := range entites {
		results[i] = mapper(&entity)
	}

	offset := params.GetOffset()
	limit := params.GetLimit()
	self := newPaginatedNavigationURL(backendDomain, path, params.ToQueryString(), limit, offset)

	var next LinkResponse
	nextOffset := offset + limit
	if int64(nextOffset) < count {
		next = newPaginatedNavigationURL(backendDomain, path, params.ToQueryString(), limit, nextOffset)
	}

	var prev LinkResponse
	prevOffset := offset - limit
	if prevOffset > 0 {
		prev = newPaginatedNavigationURL(backendDomain, path, params.ToQueryString(), limit, prevOffset)
	}

	return PaginatedResponse[T]{
		Count:   count,
		Links:   PaginatedResponseLinks{Self: self, Next: next.ToRef(), Prev: prev.ToRef()},
		Results: results,
	}
}

type LanguageLinks struct {
	Self   LinkResponse `json:"self"`
	Series LinkResponse `json:"series"`
}

type LanguageResponse struct {
	ID    int32         `json:"id"`
	Name  string        `json:"name"`
	Slug  string        `json:"slug"`
	Icon  string        `json:"icon"`
	Links LanguageLinks `json:"_links"`
}

func (c *Controllers) NewLanguageResponse(language *db.Language) *LanguageResponse {
	return &LanguageResponse{
		ID:   language.ID,
		Name: language.Name,
		Slug: language.Slug,
		Icon: language.Icon,
		Links: LanguageLinks{
			Self: LinkResponse{
				fmt.Sprintf("https://%s%s/%s", c.backendDomain, paths.LanguagePathV1, language.Slug),
			},
			Series: LinkResponse{
				fmt.Sprintf("https://%s%s/%s%s", c.backendDomain, paths.LanguagePathV1, language.Slug, paths.SeriesPath),
			},
		},
	}
}

type SeriesAuthor struct {
	FirstName string           `json:"firstName"`
	LastName  string           `json:"lastName"`
	Links     SelfLinkResponse `json:"_links"`
}

type SeriesTag struct {
	Name  string           `json:"name"`
	Links SelfLinkResponse `json:"_links"`
}

type SeriesLinks struct {
	Self    LinkResponse `json:"self"`
	Author  LinkResponse `json:"author"`
	Parts   LinkResponse `json:"parts"`
	Reviews LinkResponse `json:"reviews"`
}

func (c *Controllers) newSeriesLinks(languageSlug, seriesSlug string, authorID int32) SeriesLinks {
	return SeriesLinks{
		Self: LinkResponse{
			fmt.Sprintf("https://%s%s/%s%s/%s", c.backendDomain, paths.LanguagePathV1, languageSlug, paths.SeriesPath, seriesSlug),
		},
		Author: LinkResponse{
			fmt.Sprintf("https://%s%s/%d", c.backendDomain, paths.UsersPathV1, authorID),
		},
		Parts: LinkResponse{
			fmt.Sprintf("https://%s%s/%s%s/%s%s", c.backendDomain, paths.LanguagePathV1, languageSlug, paths.SeriesPath, seriesSlug, paths.PartsPath),
		},
		Reviews: LinkResponse{
			fmt.Sprintf("https://%s%s/%s%s/%s%s",
				c.backendDomain,
				paths.LanguagePathV1,
				languageSlug,
				paths.SeriesPath,
				seriesSlug,
				paths.ReviewsPath,
			),
		},
	}
}

type SeriesEmbedded struct {
	Author SeriesAuthor `json:"author"`
	Tags   []SeriesTag  `json:"tags"`
}

func (c *Controllers) newSeriesEmbeded(
	authorId int32,
	authorFirstName,
	authorLastName string,
	tags []string,
	languageSlug string,
) SeriesEmbedded {
	seriesTags := make([]SeriesTag, len(tags))

	for i, t := range tags {
		seriesTags[i] = SeriesTag{
			Name: t,
			Links: SelfLinkResponse{
				LinkResponse{
					fmt.Sprintf("https://%s%s/%s%s?tag=%s", c.backendDomain, paths.LanguagePathV1, languageSlug, paths.SeriesPath, t),
				},
			},
		}
	}

	return SeriesEmbedded{
		Author: SeriesAuthor{
			FirstName: authorFirstName,
			LastName:  authorLastName,
			Links: SelfLinkResponse{
				LinkResponse{
					fmt.Sprintf("https://%s%s/%d", c.backendDomain, paths.UsersPathV1, authorId),
				},
			},
		},
		Tags: seriesTags,
	}
}

type SeriesResponse struct {
	ID          int32          `json:"id"`
	Title       string         `json:"title"`
	Slug        string         `json:"slug"`
	Description string         `json:"description"`
	Parts       int16          `json:"parts"`
	Lectures    int16          `json:"lectures"`
	ReviewAvg   int16          `json:"reviewAvg"`
	ReviewCount int32          `json:"reviewCount"`
	IsPublished bool           `json:"isPublished"`
	Tags        []string       `json:"tags"`
	Embedded    SeriesEmbedded `json:"_embedded"`
	Links       SeriesLinks    `json:"_links"`
}

func (c *Controllers) NewSeriesResponse(user *tokens.AccessUserClaims, series *db.Series, tags []db.Tag, languageSlug string) *SeriesResponse {
	strTags := make([]string, len(tags))

	for i, t := range tags {
		strTags[i] = t.Name
	}

	return &SeriesResponse{
		ID:          series.ID,
		Title:       series.Title,
		Slug:        series.Slug,
		Description: series.Description,
		Parts:       series.PartsCount,
		Lectures:    series.LecturesCount,
		ReviewAvg:   series.ReviewAvg,
		ReviewCount: series.ReviewCount,
		IsPublished: series.IsPublished,
		Tags:        strTags,
		Embedded:    c.newSeriesEmbeded(user.ID, user.FirstName, user.LastName, strTags, languageSlug),
		Links:       c.newSeriesLinks(languageSlug, series.Slug, series.AuthorID),
	}
}

func (c *Controllers) NewSeriesFromDto(dto *services.SeriesDto) *SeriesResponse {
	seriesTags := make([]SeriesTag, len(dto.Tags))

	for i, t := range dto.Tags {
		seriesTags[i] = SeriesTag{
			Name: t,
			Links: SelfLinkResponse{
				LinkResponse{
					fmt.Sprintf("https://%s%s/%s%s?tag=%s", c.backendDomain, paths.LanguagePathV1, dto.LanguageSlug, paths.SeriesPath, t),
				},
			},
		}
	}

	return &SeriesResponse{
		ID:          dto.ID,
		Title:       dto.Title,
		Slug:        dto.Slug,
		Description: dto.Description,
		Parts:       dto.Parts,
		Lectures:    dto.Lectures,
		ReviewAvg:   dto.ReviewAvg,
		ReviewCount: dto.ReviewCount,
		IsPublished: dto.IsPublished,
		Tags:        dto.Tags,
	}
}
