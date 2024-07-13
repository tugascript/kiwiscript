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
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
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

type PaginatedResponse[T any] struct {
	Count    int64  `json:"count"`
	Next     string `json:"next,omitempty"`
	Previous string `json:"previous,omitempty"`
	Results  []T    `json:"results"`
}

func newPaginatedNavigationURL(frontendDomain string, limit, offset int32) string {
	if offset < 0 {
		offset = 0
	}

	return fmt.Sprintf("https://%s?limit=%d&offset=%d", frontendDomain, limit, offset)
}

func NewPaginatedResponse[T any, V any](
	frontendDomain string,
	limit,
	offset int32,
	count int64,
	entites []V,
	mapper func(V) T,
) PaginatedResponse[T] {
	results := make([]T, len(entites))

	for i, entity := range entites {
		results[i] = mapper(entity)
	}

	var next, prev string
	nextOffset := offset + limit
	prevOffset := offset - limit
	if int64(nextOffset) < count {
		next = newPaginatedNavigationURL(frontendDomain, limit, nextOffset)
	}
	if prevOffset > 0 {
		prev = newPaginatedNavigationURL(frontendDomain, limit, prevOffset)
	}

	return PaginatedResponse[T]{
		Count:    count,
		Next:     next,
		Previous: prev,
		Results:  results,
	}
}

type LanguageResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Icon string `json:"icon"`
}

func NewLanguageResponse(language db.Language) LanguageResponse {
	return LanguageResponse{
		ID:   language.ID,
		Name: language.Name,
		Slug: language.Slug,
		Icon: language.Icon,
	}
}

type AdminSeriesResponse struct {
	ID          int32    `json:"id"`
	Title       string   `json:"title"`
	Slug        string   `json:"slug"`
	Description string   `json:"description"`
	Parts       int16    `json:"parts"`
	Lectures    int16    `json:"lectures"`
	ReviewAvg   int16    `json:"reviewAvg"`
	ReviewCount int32    `json:"reviewCount"`
	IsPublished bool     `json:"isPublished"`
	Tags        []string `json:"tags"`
}

func NewAdminSeriesResponse(series *db.Series, tags []db.Tag) *AdminSeriesResponse {
	strTags := make([]string, len(tags))

	for i, t := range tags {
		strTags[i] = t.Name
	}

	return &AdminSeriesResponse{
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
	}
}

func NewAdminSeriesFromDto(dtos *services.SeriesDto) *AdminSeriesResponse {
	return &AdminSeriesResponse{
		ID:          dtos.ID,
		Title:       dtos.Title,
		Slug:        dtos.Slug,
		Description: dtos.Description,
		Parts:       dtos.Parts,
		Lectures:    dtos.Lectures,
		ReviewAvg:   dtos.ReviewAvg,
		ReviewCount: dtos.ReviewCount,
		IsPublished: dtos.IsPublished,
		Tags:        dtos.Tags,
	}
}
