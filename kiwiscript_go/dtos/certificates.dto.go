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

type CertificatesPathParams struct {
	CertificateID string `validate:"required,uuid"`
}

type CertificateLinks struct {
	Self     LinkResponse `json:"self"`
	Series   LinkResponse `json:"series"`
	Language LinkResponse `json:"language"`
}

func newCertificatesLink(backendDomain, languageSlug, seriesSlug string, certificateID uuid.UUID) CertificateLinks {
	return CertificateLinks{
		Self: LinkResponse{
			Href: fmt.Sprintf(
				"https://%s/api%s/%s",
				backendDomain,
				paths.CertificatesV1,
				certificateID.String(),
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
		Language: LinkResponse{
			Href: fmt.Sprintf("https://%s/api%s/%s", backendDomain, paths.LanguagePathV1, languageSlug),
		},
	}
}

type CertificateLanguageEmbedded struct {
	ID   int32            `json:"id"`
	Name string           `json:"name"`
	Slug string           `json:"slug"`
	Link SelfLinkResponse `json:"_link"`
}

func newCertificateLanguageEmbedded(
	backendDomain string,
	language *db.CertificateLanguage,
) CertificateLanguageEmbedded {
	return CertificateLanguageEmbedded{
		ID:   language.ID,
		Name: language.Name,
		Slug: language.Slug,
		Link: SelfLinkResponse{
			Self: LinkResponse{
				Href: fmt.Sprintf("https://%s/api%s/%s", backendDomain, paths.LanguagePathV1, language.Slug),
			},
		},
	}
}

type CertificateEmbedded struct {
	Language CertificateLanguageEmbedded `json:"language"`
}

type CertificateResponse struct {
	ID          uuid.UUID           `json:"id"`
	FirstName   string              `json:"firstName"`
	LastName    string              `json:"lastName"`
	SeriesTitle string              `json:"seriesTitle"`
	Lessons     int16               `json:"lessons"`
	WatchTime   int32               `json:"watchTime"`
	ReadTime    int32               `json:"readTime"`
	CompletedAt string              `json:"completedAt,omitempty"`
	Links       CertificateLinks    `json:"_links"`
	Embedded    CertificateEmbedded `json:"_embedded"`
}

func NewCertificateResponse(backendDomain string, certificate *db.CertificateModel) *CertificateResponse {
	return &CertificateResponse{
		ID:          certificate.ID,
		FirstName:   certificate.FirstName,
		LastName:    certificate.LastName,
		SeriesTitle: certificate.SeriesTitle,
		Lessons:     certificate.Lessons,
		WatchTime:   certificate.WatchTime,
		ReadTime:    certificate.ReadTime,
		CompletedAt: certificate.CompletedAt,
		Links: newCertificatesLink(
			backendDomain,
			certificate.Language.Slug,
			certificate.SeriesSlug,
			certificate.ID,
		),
		Embedded: CertificateEmbedded{
			Language: newCertificateLanguageEmbedded(backendDomain, &certificate.Language),
		},
	}
}
