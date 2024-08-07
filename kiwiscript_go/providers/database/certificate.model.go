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

type CertificateLanguage struct {
	ID   int32
	Name string
	Slug string
}

type CertificateModel struct {
	ID          uuid.UUID
	FirstName   string
	LastName    string
	SeriesTitle string
	SeriesSlug  string
	Lessons     int16
	WatchTime   int32
	ReadTime    int32
	CompletedAt string
	Language    CertificateLanguage
}

type ToCertificateModel interface {
	ToCertificateModel() *CertificateModel
}

type ToCertificateModelWithAuthor interface {
	ToCertificateModelWithAuthor(firstName, lastName string) *CertificateModel
}

func (c *FindCertificateByIDWithUserAndLanguageRow) ToCertificateModel() *CertificateModel {
	var completedAt string
	if c.CompletedAt.Valid {
		completedAt = c.CompletedAt.Time.Format(time.RFC3339)
	}

	return &CertificateModel{
		ID:          c.ID,
		FirstName:   c.AuthorFirstName,
		LastName:    c.AuthorLastName,
		SeriesTitle: c.SeriesTitle,
		SeriesSlug:  c.SeriesSlug,
		Lessons:     c.Lessons,
		WatchTime:   c.WatchTimeSeconds,
		ReadTime:    c.ReadTimeSeconds,
		CompletedAt: completedAt,
		Language: CertificateLanguage{
			ID:   c.LanguageID,
			Name: c.LanguageName,
			Slug: c.LanguageSlug,
		},
	}
}

func (c *FindPaginatedCertificatesByUserIDRow) ToCertificateModelWithAuthor(
	firstName,
	lastName string,
) *CertificateModel {
	var completedAt string
	if c.CompletedAt.Valid {
		completedAt = c.CompletedAt.Time.Format(time.RFC3339)
	}

	return &CertificateModel{
		ID:          c.ID,
		FirstName:   firstName,
		LastName:    lastName,
		SeriesTitle: c.SeriesTitle,
		SeriesSlug:  c.SeriesSlug,
		Lessons:     c.Lessons,
		WatchTime:   c.WatchTimeSeconds,
		ReadTime:    c.ReadTimeSeconds,
		CompletedAt: completedAt,
		Language: CertificateLanguage{
			ID:   c.LanguageID,
			Name: c.LanguageName,
			Slug: c.LanguageSlug,
		},
	}
}
