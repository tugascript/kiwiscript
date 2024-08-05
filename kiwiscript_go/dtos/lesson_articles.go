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

type LessonArticleBody struct {
	Content string `json:"content"  validate:"required,min=1,markdown"`
}

// Responses

type LessonArticleLinks struct {
	Self   LinkResponse `json:"self"`
	Lesson LinkResponse `json:"lesson"`
}

func newLessonArticleLinks(
	backendDomain,
	languageSlug,
	seriesSlug string,
	sectionID,
	lessonID int32,
) LessonArticleLinks {
	return LessonArticleLinks{
		Self: LinkResponse{
			Href: fmt.Sprintf(
				"https://%s%s/%s%s/%s%s/%d%s/%d%s",
				backendDomain,
				paths.LanguagePathV1,
				languageSlug,
				paths.SeriesPath,
				seriesSlug,
				paths.SectionsPath,
				sectionID,
				paths.LessonsPath,
				lessonID,
				paths.ArticlePath,
			),
		},
		Lesson: LinkResponse{
			Href: fmt.Sprintf(
				"https://%s%s/%s%s/%s%s/%d%s/%d",
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

type LessonArticleResponse struct {
	ID       int32              `json:"id"`
	Content  string             `json:"content"`
	ReadTime int32              `json:"readTime"`
	Links    LessonArticleLinks `json:"_links"`
}

func NewLessonArticleResponse(
	backendDomain,
	languageSlug,
	seriesSlug string,
	sectionID int32,
	article *db.LessonArticle,
) *LessonArticleResponse {
	return &LessonArticleResponse{
		ID:       article.ID,
		Content:  article.Content,
		ReadTime: article.ReadTimeSeconds,
		Links: newLessonArticleLinks(
			backendDomain,
			languageSlug,
			seriesSlug,
			sectionID,
			article.LessonID,
		),
	}
}
