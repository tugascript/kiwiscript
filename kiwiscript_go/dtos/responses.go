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

import "fmt"

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

func newPaginatedNavigationURL(frontendDomain, path string, params string, limit, offset int32) LinkResponse {
	if offset < 0 {
		offset = 0
	}

	var href string
	if params != "" {
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
	entities []V,
	mapper func(*V) T,
) PaginatedResponse[T] {
	results := make([]T, 0, len(entities))

	for _, entity := range entities {
		results = append(results, mapper(&entity))
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
