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

// Path params

type LessonFilePathParams struct {
	LanguageSlug string `validate:"required,min=2,max=50,slug"`
	SeriesSlug   string `validate:"required,min=2,max=100,slug"`
	SectionID    string `validate:"required,number,min=1"`
	LessonID     string `validate:"required,number,min=1"`
	FileID       string `validate:"required,uuid"`
}

// Bodies

type LessonFileBody struct {
	Name string `validate:"required,min=2,max=250,extalphanum"`
}

// Response

type LessonFileLinks struct {
	Self        LinkResponse `json:"self"`
	LessonFiles LinkResponse `json:"lessonFiles"`
	Lesson      LinkResponse `json:"lesson"`
}

func newLessonFileLinks(
	backendDomain,
	languageSlug,
	seriesSlug string,
	sectionID,
	lessonID int32,
	fileID uuid.UUID,
) LessonFileLinks {
	return LessonFileLinks{
		Self: LinkResponse{
			Href: fmt.Sprintf(
				"https://%s/api%s/%s%s/%s%s/%d%s/%d%s/%s",
				backendDomain,
				paths.LanguagePathV1,
				languageSlug,
				paths.SeriesPath,
				seriesSlug,
				paths.SectionsPath,
				sectionID,
				paths.LessonsPath,
				lessonID,
				paths.FilesPath,
				fileID.String(),
			),
		},
		LessonFiles: LinkResponse{
			Href: fmt.Sprintf(
				"https://%s/api%s/%s%s/%s%s/%d%s/%d%s",
				backendDomain,
				paths.LanguagePathV1,
				languageSlug,
				paths.SeriesPath,
				seriesSlug,
				paths.SectionsPath,
				sectionID,
				paths.LessonsPath,
				lessonID,
				paths.FilesPath,
			),
		},
		Lesson: LinkResponse{
			Href: fmt.Sprintf(
				"https://%s/api%s/%s%s/%s%s/%d%s/%d",
				backendDomain,
				paths.LanguagePathV1,
				languageSlug,
				paths.SeriesPath,
				seriesSlug,
				paths.SectionsPath,
				sectionID,
				paths.LessonsPath,
				lessonID,
			),
		},
	}
}

type LessonFileResponse struct {
	ID    uuid.UUID       `json:"id"`
	Name  string          `json:"name"`
	Ext   string          `json:"ext"`
	URL   string          `json:"url"`
	Links LessonFileLinks `json:"_links"`
}

func NewLessonFileResponse(
	backendDomain,
	languageSlug,
	seriesSlug string,
	sectionID int32,
	file *db.LessonFileModel,
) *LessonFileResponse {
	return &LessonFileResponse{
		ID:   file.ID,
		Name: file.Name,
		Ext:  file.Ext,
		URL:  file.URL,
		Links: newLessonFileLinks(
			backendDomain,
			languageSlug,
			seriesSlug,
			sectionID,
			file.LessonID,
			file.ID,
		),
	}
}
