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
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kiwiscript/kiwiscript_go/paths"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/services"
	"time"
)

// Path params

type LessonPathParams struct {
	LanguageSlug string `validate:"required,min=2,max=50,slug"`
	SeriesSlug   string `validate:"required,min=2,max=100,slug"`
	SectionID    string `validate:"required,number,min=1"`
	LessonID     string `validate:"required,number,min=1"`
}

// Bodies

type CreateLessonBody struct {
	Title string `json:"title" validate:"required,min=2,max=250"`
}

type UpdateLessonBody struct {
	Title    string `json:"title" validate:"required,min=2,max=250"`
	Position int16  `json:"position" validate:"required,gte=1"`
}

// Responses

type LessonLinks struct {
	Self        LinkResponse  `json:"self"`
	Section     LinkResponse  `json:"section"`
	Article     *LinkResponse `json:"article,omitempty"`
	Video       *LinkResponse `json:"video,omitempty"`
	Certificate *LinkResponse `json:"certificate,omitempty"`
}

func newLessonLinks(
	backendDomain,
	languageSlug,
	seriesSlug string,
	sectionID,
	lessonID,
	readTimeSeconds,
	watchTimeSeconds int32,
	certificateID *uuid.UUID,
) LessonLinks {
	var article, video, certificate *LinkResponse

	if readTimeSeconds > 0 {
		article = &LinkResponse{
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
		}
	}
	if watchTimeSeconds > 0 {
		video = &LinkResponse{
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
				paths.VideoPath,
			),
		}
	}
	if certificateID != nil {
		certificate = &LinkResponse{
			Href: fmt.Sprintf("https://%s%s/%s", backendDomain, paths.CertificatesV1, certificateID.String()),
		}
	}

	return LessonLinks{
		Self: LinkResponse{
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
		Section: LinkResponse{
			Href: fmt.Sprintf(
				"https://%s%s/%s%s/%s%s/%d",
				backendDomain,
				paths.LanguagePathV1,
				languageSlug,
				paths.SeriesPath,
				seriesSlug,
				paths.SectionsPath,
				sectionID,
			),
		},
		Article:     article,
		Video:       video,
		Certificate: certificate,
	}
}

type LessonArticleEmbedded struct {
	ID      int32            `json:"id"`
	Content string           `json:"content"`
	Links   SelfLinkResponse `json:"_links"`
}

func newLessonArticleEmbedded(
	backendDomain string,
	lesson *db.LessonModel,
	articleID pgtype.Int4,
	content string,
) *LessonArticleEmbedded {
	if !articleID.Valid {
		return nil
	}

	return &LessonArticleEmbedded{
		ID:      articleID.Int32,
		Content: content,
		Links: SelfLinkResponse{
			Self: LinkResponse{
				Href: fmt.Sprintf(
					"https://%s%s/%s%s/%s%s/%d%s/%d%s",
					backendDomain,
					paths.LanguagePathV1,
					lesson.LanguageSlug,
					paths.SeriesPath,
					lesson.SeriesSlug,
					paths.SectionsPath,
					lesson.SectionID,
					paths.LessonsPath,
					lesson.ID,
					paths.ArticlePath,
				),
			},
		},
	}
}

type LessonVideoEmbedded struct {
	ID    int32            `json:"id"`
	URL   string           `json:"url"`
	Links SelfLinkResponse `json:"_links"`
}

func newLessonVideoEmbedded(
	backendDomain string,
	lesson *db.LessonModel,
	videoID pgtype.Int4,
	url string,
) *LessonVideoEmbedded {
	if !videoID.Valid {
		return nil
	}

	return &LessonVideoEmbedded{
		ID:  videoID.Int32,
		URL: url,
		Links: SelfLinkResponse{
			Self: LinkResponse{
				Href: fmt.Sprintf(
					"https://%s%s/%s%s/%s%s/%d%s/%d%s",
					backendDomain,
					paths.LanguagePathV1,
					lesson.LanguageSlug,
					paths.SeriesPath,
					lesson.SeriesSlug,
					paths.SectionsPath,
					lesson.SectionID,
					paths.LessonsPath,
					lesson.ID,
					paths.VideoPath,
				),
			},
		},
	}
}

type LessonFileEmbedded struct {
	ID    uuid.UUID        `json:"id"`
	Name  string           `json:"name"`
	Ext   string           `json:"ext"`
	URL   string           `json:"url"`
	Links SelfLinkResponse `json:"_links"`
}

func newFileEmbedded(
	backendDomain string,
	lesson *db.LessonModel,
	file *db.LessonFile,
	url string,
) LessonFileEmbedded {
	return LessonFileEmbedded{
		ID:   file.ID,
		Name: file.Name,
		Ext:  file.Ext,
		URL:  url,
		Links: SelfLinkResponse{
			Self: LinkResponse{
				Href: fmt.Sprintf(
					"https://%s%s/%s%s/%s%s/%d%s/%d%s/%s",
					backendDomain,
					paths.LanguagePathV1,
					lesson.LanguageSlug,
					paths.SeriesPath,
					lesson.SeriesSlug,
					paths.SectionsPath,
					lesson.SectionID,
					paths.LessonsPath,
					lesson.ID,
					paths.FilesPath,
					file.ID.String(),
				),
			},
		},
	}
}

type LessonCertificateEmbedded struct {
	ID          uuid.UUID        `json:"id"`
	SeriesTitle string           `json:"seriesTitle"`
	Lessons     int16            `json:"lessons"`
	WatchTime   int32            `json:"watchTime"`
	ReadTime    int32            `json:"readTime"`
	CompletedAt string           `json:"completedAt,omitempty"`
	Links       SelfLinkResponse `json:"_links"`
}

func newCertificateEmbedded(
	backendDomain string,
	certificate *db.Certificate,
) *LessonCertificateEmbedded {
	var completedAt string
	if certificate.CompletedAt.Valid {
		completedAt = certificate.CompletedAt.Time.Format(time.RFC3339)
	}

	return &LessonCertificateEmbedded{
		ID:          certificate.ID,
		SeriesTitle: certificate.SeriesTitle,
		Lessons:     certificate.Lessons,
		WatchTime:   certificate.WatchTimeSeconds,
		ReadTime:    certificate.ReadTimeSeconds,
		CompletedAt: completedAt,
		Links: SelfLinkResponse{
			Self: LinkResponse{
				Href: fmt.Sprintf(
					"https://%s%s/%s",
					backendDomain,
					paths.CertificatesV1,
					certificate.ID.String(),
				),
			},
		},
	}
}

type LessonEmbedded struct {
	Article     *LessonArticleEmbedded     `json:"article,omitempty"`
	Video       *LessonVideoEmbedded       `json:"video,omitempty"`
	Files       []LessonFileEmbedded       `json:"files,omitempty"`
	Certificate *LessonCertificateEmbedded `json:"certificate,omitempty"`
}

func newLessonEmbedded(
	article *LessonArticleEmbedded,
	video *LessonVideoEmbedded,
	files []LessonFileEmbedded,
) *LessonEmbedded {
	if article == nil && video == nil && files == nil {
		return nil
	}

	return &LessonEmbedded{
		Article: article,
		Video:   video,
		Files:   files,
	}
}

func newLessonEmbeddedCertificate(certificate *LessonCertificateEmbedded) *LessonEmbedded {
	return &LessonEmbedded{
		Certificate: certificate,
	}
}

type LessonResponse struct {
	ID          int32           `json:"id"`
	Title       string          `json:"title"`
	Position    int16           `json:"position"`
	IsCompleted bool            `json:"isCompleted"`
	IsPublished bool            `json:"isPublished"`
	WatchTime   int32           `json:"watchTime"`
	ReadTime    int32           `json:"readTime"`
	Embedded    *LessonEmbedded `json:"_embedded,omitempty"`
	Links       LessonLinks     `json:"_links"`
}

func NewLessonResponse(backendDomain string, lesson *db.LessonModel) *LessonResponse {
	return &LessonResponse{
		ID:          lesson.ID,
		Title:       lesson.Title,
		Position:    lesson.Position,
		IsCompleted: lesson.IsCompleted,
		IsPublished: lesson.IsPublished,
		WatchTime:   lesson.WatchTimeSeconds,
		ReadTime:    lesson.ReadTimeSeconds,
		Links: newLessonLinks(
			backendDomain,
			lesson.LanguageSlug,
			lesson.SeriesSlug,
			lesson.SectionID,
			lesson.ID,
			lesson.ReadTimeSeconds,
			lesson.WatchTimeSeconds,
			nil,
		),
	}
}

func NewLessonResponseWithEmbeddedOptions(
	backendDomain string,
	lesson *db.LessonModel,
	articleID,
	videoID pgtype.Int4,
	articleContent,
	videoURL string,
	files []db.LessonFile,
	fileURLs *services.FileURLsContainer,
) *LessonResponse {
	var embeddedFiles []LessonFileEmbedded
	filesLen := len(files)
	if filesLen > 0 && fileURLs != nil {
		embeddedFiles = make([]LessonFileEmbedded, 0, filesLen)

		for _, f := range files {
			if url, ok := fileURLs.Get(f.ID); ok {
				embeddedFiles = append(embeddedFiles, newFileEmbedded(backendDomain, lesson, &f, url))
			}
		}
	}

	return &LessonResponse{
		ID:          lesson.ID,
		Title:       lesson.Title,
		Position:    lesson.Position,
		IsCompleted: lesson.IsCompleted,
		IsPublished: lesson.IsPublished,
		Links: newLessonLinks(
			backendDomain,
			lesson.LanguageSlug,
			lesson.SeriesSlug,
			lesson.SectionID,
			lesson.ID,
			lesson.ReadTimeSeconds,
			lesson.WatchTimeSeconds,
			nil,
		),
		Embedded: newLessonEmbedded(
			newLessonArticleEmbedded(backendDomain, lesson, articleID, articleContent),
			newLessonVideoEmbedded(backendDomain, lesson, videoID, videoURL),
			embeddedFiles,
		),
	}
}

func NewLessonResponseWithCertificate(
	backendDomain string,
	lesson *db.LessonModel,
	certificate *db.Certificate,
) *LessonResponse {
	return &LessonResponse{
		ID:          lesson.ID,
		Title:       lesson.Title,
		Position:    lesson.Position,
		IsCompleted: lesson.IsCompleted,
		IsPublished: lesson.IsPublished,
		WatchTime:   lesson.WatchTimeSeconds,
		ReadTime:    lesson.ReadTimeSeconds,
		Links: newLessonLinks(
			backendDomain,
			lesson.LanguageSlug,
			lesson.SeriesSlug,
			lesson.SectionID,
			lesson.ID,
			lesson.ReadTimeSeconds,
			lesson.WatchTimeSeconds,
			&certificate.ID,
		),
		Embedded: newLessonEmbeddedCertificate(newCertificateEmbedded(backendDomain, certificate)),
	}
}
