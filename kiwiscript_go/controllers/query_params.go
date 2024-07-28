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

import "net/url"

const (
	OffsetDefault int = 0
	LimitDefault  int = 25
)

type FromQueryParams interface {
	ToQueryString() string
	GetLimit() int32
	GetOffset() int32
}

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

type PaginationQueryParams struct {
	Limit  int32 `validate:"omitempty,gte=1,lte=100"`
	Offset int32 `validate:"omitempty,gte=0"`
}

func (p PaginationQueryParams) ToQueryString() string {
	return ""
}
func (p PaginationQueryParams) GetLimit() int32 {
	return p.Limit
}
func (p PaginationQueryParams) GetOffset() int32 {
	return p.Offset
}

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

type SeriesPartsQueryParams struct {
	Limit  int32 `validate:"omitempty,gte=1,lte=100"`
	Offset int32 `validate:"omitempty,gte=0"`
}

func (p SeriesPartsQueryParams) ToQueryString() string {
	return ""
}
func (p SeriesPartsQueryParams) GetLimit() int32 {
	return p.Limit
}
func (p SeriesPartsQueryParams) GetOffset() int32 {
	return p.Offset
}

type LecturesQueryParams struct {
	IsPublished bool  `validate:"omitempty"`
	Limit       int32 `validate:"omitempty,gte=1,lte=100"`
	Offset      int32 `validate:"omitempty,gte=0"`
}

func (p *LecturesQueryParams) ToQueryString() string {
	return ""
}
func (p *LecturesQueryParams) GetLimit() int32 {
	return p.Limit
}
func (p *LecturesQueryParams) GetOffset() int32 {
	return p.Offset
}
