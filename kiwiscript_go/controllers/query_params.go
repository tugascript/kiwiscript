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

import "fmt"

const (
	OffsetDefault int = 0
	LimitDefault  int = 25
)

type QueryStr string

func (s *QueryStr) checkNotEmpty() bool {
	if s != nil && *s != "" {
		*s += "&"
		return true
	}

	return false
}

func (s *QueryStr) add(key, value string) {
	queryStrParam := QueryStr(fmt.Sprintf("%s=%s", key, value))

	if s.checkNotEmpty() {
		*s += queryStrParam
	} else {
		*s = queryStrParam
	}
}

type FromQueryParams interface {
	ToQueryString() QueryStr
	GetLimit() int32
	GetOffset() int32
}

type GetLanguagesQueryParams struct {
	Limit  int32  `validate:"omitempty,gte=1,lte=100"`
	Offset int32  `validate:"omitempty,gte=0"`
	Search string `validate:"omitempty,min=1,max=50,extalphanum"`
}

func (p GetLanguagesQueryParams) ToQueryString() QueryStr {
	var queryStr QueryStr = ""

	if p.Search != "" {
		queryStr.add("search", p.Search)
	}

	return queryStr
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

func (p PaginationQueryParams) ToQueryString() QueryStr {
	return ""
}
func (p PaginationQueryParams) GetLimit() int32 {
	return p.Limit
}
func (p PaginationQueryParams) GetOffset() int32 {
	return p.Offset
}

type SeriesQueryParams struct {
	IsPublished bool   `validate:"omitempty"`
	AuthorID    int32  `validate:"omitempty,gte=1"`
	Search      string `validate:"omitempty,min=1,max=100"`
	Tag         string `validate:"omitempty,min=2,max=50,slug"`
	Limit       int32  `validate:"omitempty,gte=1,lte=100"`
	Offset      int32  `validate:"omitempty,gte=0"`
	SortBy      string `validate:"omitempty,oneof=slug id"`
	Order       string `validate:"omitempty,oneof=ASC DESC"`
}

func (p SeriesQueryParams) ToQueryString() QueryStr {
	var queryStr QueryStr = ""

	if p.IsPublished {
		queryStr.add("isPublished", "true")
	}
	if p.Search != "" {
		queryStr.add("search", p.Search)
	}
	if p.Tag != "" {
		queryStr.add("tag", p.Tag)
	}
	if p.SortBy != "" {
		queryStr.add("sortBy", p.SortBy)
	}
	if p.Order != "" {
		queryStr.add("order", p.Order)
	}

	return queryStr
}
func (p SeriesQueryParams) GetLimit() int32 {
	return p.Limit
}
func (p SeriesQueryParams) GetOffset() int32 {
	return p.Offset
}

type SeriesPartsQueryParams struct {
	IsPublished       bool  `validate:"omitempty"`
	PublishedLectures bool  `validate:"omitempty"`
	Limit             int32 `validate:"omitempty,gte=1,lte=100"`
	Offset            int32 `validate:"omitempty,gte=0"`
}

func (p SeriesPartsQueryParams) ToQueryString() QueryStr {
	var queryStr QueryStr = ""

	if p.IsPublished {
		queryStr.add("isPublished", "true")
	}
	if p.PublishedLectures {
		queryStr.add("publishedLectures", "true")
	}

	return queryStr
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

func (p *LecturesQueryParams) ToQueryString() QueryStr {
	var queryStr QueryStr = ""

	if p.IsPublished {
		queryStr.add("isPublished", "true")
	}

	return queryStr
}
func (p *LecturesQueryParams) GetLimit() int32 {
	return p.Limit
}
func (p *LecturesQueryParams) GetOffset() int32 {
	return p.Offset
}
