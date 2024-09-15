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
)

// Bodies

type LessonVideoBody struct {
	URL       string `json:"url" validate:"required,url"`
	WatchTime int32  `json:"watchTime" validate:"required,number,min=1"`
}

// Responses

type LessonVideoLinks struct {
	Self   LinkResponse `json:"self"`
	Lesson LinkResponse `json:"lesson"`
}

func newLessonVideoLinks(
	backendDomain,
	languageSlug,
	seriesSlug string,
	sectionID,
	lessonID int32,
) LessonVideoLinks {
	return LessonVideoLinks{
		Self: LinkResponse{
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
				paths.VideoPath,
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

type LessonVideoResponse struct {
	ID        int32            `json:"id"`
	URL       string           `json:"url"`
	WatchTime int32            `json:"watchTime"`
	Links     LessonVideoLinks `json:"_links"`
}

func NewLessonVideoResponse(
	backendDomain,
	languageSlug,
	seriesSlug string,
	sectionID int32,
	video *db.LessonVideo,
) *LessonVideoResponse {
	return &LessonVideoResponse{
		ID:        video.ID,
		URL:       video.Url,
		WatchTime: video.WatchTimeSeconds,
		Links: newLessonVideoLinks(
			backendDomain,
			languageSlug,
			seriesSlug,
			sectionID,
			video.LessonID,
		),
	}
}
