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
	"github.com/kiwiscript/kiwiscript_go/paths"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/providers/tokens"
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

func newPaginatedNavigationURL(frontendDomain, path string, params QueryStr, limit, offset int32) LinkResponse {
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
	entites []V,
	mapper func(*V) T,
) PaginatedResponse[T] {
	results := make([]T, len(entites))

	for i, entity := range entites {
		results[i] = mapper(&entity)
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

type LanguageLinks struct {
	Self   LinkResponse `json:"self"`
	Series LinkResponse `json:"series"`
}

type LanguageResponse struct {
	ID    int32         `json:"id"`
	Name  string        `json:"name"`
	Slug  string        `json:"slug"`
	Icon  string        `json:"icon"`
	Links LanguageLinks `json:"_links"`
}

func (c *Controllers) NewLanguageResponse(language *db.Language) *LanguageResponse {
	return &LanguageResponse{
		ID:   language.ID,
		Name: language.Name,
		Slug: language.Slug,
		Icon: language.Icon,
		Links: LanguageLinks{
			Self: LinkResponse{
				fmt.Sprintf("https://%s%s/%s", c.backendDomain, paths.LanguagePathV1, language.Slug),
			},
			Series: LinkResponse{
				fmt.Sprintf("https://%s%s/%s%s", c.backendDomain, paths.LanguagePathV1, language.Slug, paths.SeriesPath),
			},
		},
	}
}

type SeriesAuthor struct {
	FirstName string           `json:"firstName"`
	LastName  string           `json:"lastName"`
	Links     SelfLinkResponse `json:"_links"`
}

type SeriesTag struct {
	Name  string           `json:"name"`
	Links SelfLinkResponse `json:"_links"`
}

type SeriesLinks struct {
	Self     LinkResponse `json:"self"`
	Author   LinkResponse `json:"author"`
	Language LinkResponse `json:"language"`
	Parts    LinkResponse `json:"parts"`
	Reviews  LinkResponse `json:"reviews"`
}

func (c *Controllers) newSeriesLinks(languageSlug, seriesSlug string, authorID int32) SeriesLinks {
	return SeriesLinks{
		Self: LinkResponse{
			fmt.Sprintf("https://%s%s/%s%s/%s", c.backendDomain, paths.LanguagePathV1, languageSlug, paths.SeriesPath, seriesSlug),
		},
		Language: LinkResponse{
			fmt.Sprintf("https://%s%s/%s", c.backendDomain, paths.LanguagePathV1, languageSlug),
		},
		Author: LinkResponse{
			fmt.Sprintf("https://%s%s/%d", c.backendDomain, paths.UsersPathV1, authorID),
		},
		Parts: LinkResponse{
			fmt.Sprintf("https://%s%s/%s%s/%s%s", c.backendDomain, paths.LanguagePathV1, languageSlug, paths.SeriesPath, seriesSlug, paths.PartsPath),
		},
		Reviews: LinkResponse{
			fmt.Sprintf("https://%s%s/%s%s/%s%s",
				c.backendDomain,
				paths.LanguagePathV1,
				languageSlug,
				paths.SeriesPath,
				seriesSlug,
				paths.ReviewsPath,
			),
		},
	}
}

type SeriesEmbedded struct {
	Author SeriesAuthor `json:"author"`
	Tags   []SeriesTag  `json:"tags"`
}

func (c *Controllers) newSeriesEmbeded(
	authorId int32,
	authorFirstName,
	authorLastName string,
	tags []string,
	languageSlug string,
) SeriesEmbedded {
	seriesTags := make([]SeriesTag, len(tags))

	for i, t := range tags {
		seriesTags[i] = SeriesTag{
			Name: t,
			Links: SelfLinkResponse{
				LinkResponse{
					fmt.Sprintf("https://%s%s/%s%s?tag=%s", c.backendDomain, paths.LanguagePathV1, languageSlug, paths.SeriesPath, t),
				},
			},
		}
	}

	return SeriesEmbedded{
		Author: SeriesAuthor{
			FirstName: authorFirstName,
			LastName:  authorLastName,
			Links: SelfLinkResponse{
				LinkResponse{
					fmt.Sprintf("https://%s%s/%d", c.backendDomain, paths.UsersPathV1, authorId),
				},
			},
		},
		Tags: seriesTags,
	}
}

type SeriesResponse struct {
	ID          int32          `json:"id"`
	Title       string         `json:"title"`
	Slug        string         `json:"slug"`
	Description string         `json:"description"`
	Parts       int16          `json:"parts"`
	Lectures    int16          `json:"lectures"`
	ReviewAvg   int16          `json:"reviewAvg"`
	ReviewCount int32          `json:"reviewCount"`
	IsPublished bool           `json:"isPublished"`
	Tags        []string       `json:"tags"`
	Embedded    SeriesEmbedded `json:"_embedded"`
	Links       SeriesLinks    `json:"_links"`
}

func (c *Controllers) NewSeriesResponse(user *tokens.AccessUserClaims, series *db.Series, tags []db.Tag) *SeriesResponse {
	strTags := make([]string, len(tags))

	for i, t := range tags {
		strTags[i] = t.Name
	}

	return &SeriesResponse{
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
		Embedded:    c.newSeriesEmbeded(user.ID, user.FirstName, user.LastName, strTags, series.LanguageSlug),
		Links:       c.newSeriesLinks(series.LanguageSlug, series.Slug, series.AuthorID),
	}
}

func (c *Controllers) NewSeriesFromDto(dto *services.SeriesDto, languageSlug string) *SeriesResponse {
	seriesTags := make([]SeriesTag, len(dto.Tags))

	for i, t := range dto.Tags {
		seriesTags[i] = SeriesTag{
			Name: t,
			Links: SelfLinkResponse{
				LinkResponse{
					fmt.Sprintf(
						"https://%s%s/%s%s?tag=%s",
						c.backendDomain,
						paths.LanguagePathV1,
						languageSlug,
						paths.SeriesPath,
						t,
					),
				},
			},
		}
	}

	return &SeriesResponse{
		ID:          dto.ID,
		Title:       dto.Title,
		Slug:        dto.Slug,
		Description: dto.Description,
		Parts:       dto.Parts,
		Lectures:    dto.Lectures,
		ReviewAvg:   dto.ReviewAvg,
		ReviewCount: dto.ReviewCount,
		IsPublished: dto.IsPublished,
		Tags:        dto.Tags,
	}
}

type SeriesPartLinks struct {
	Self     LinkResponse `json:"self"`
	Series   LinkResponse `json:"series"`
	Lectures LinkResponse `json:"lectures"`
}

func (c *Controllers) newSeriesPartLinks(languageSlug, seriesSlug string, partID int32) SeriesPartLinks {
	return SeriesPartLinks{
		Self: LinkResponse{
			fmt.Sprintf("https://%s%s/%s%s/%s%s/%d", c.backendDomain, paths.LanguagePathV1, languageSlug, paths.SeriesPath, seriesSlug, paths.PartsPath, partID),
		},
		Series: LinkResponse{
			fmt.Sprintf("https://%s%s/%s%s/%s", c.backendDomain, paths.LanguagePathV1, languageSlug, paths.SeriesPath, seriesSlug),
		},
		Lectures: LinkResponse{
			fmt.Sprintf("https://%s%s/%s%s/%s%s/%d%s", c.backendDomain, paths.LanguagePathV1, languageSlug, paths.SeriesPath, seriesSlug, paths.PartsPath, partID, paths.LecturesPath),
		},
	}
}

type SeriesPartLecture struct {
	ID               int32            `json:"id"`
	Title            string           `json:"title"`
	WatchTimeSeconds int32            `json:"watchTimeSeconds"`
	ReadTimeSeconds  int32            `json:"readTimeSeconds"`
	IsPublished      bool             `json:"isPublished"`
	Links            SelfLinkResponse `json:"_links"`
}

type SeriesPartEmbedded struct {
	Lectures []SeriesPartLecture `json:"lectures"`
}

func (c *Controllers) newSeriesPartEmbedded(languageSlug, seriesSlug string, partID int32, lectures []db.Lecture) SeriesPartEmbedded {
	seriesPartLectures := make([]SeriesPartLecture, len(lectures))

	for i, l := range lectures {
		seriesPartLectures[i] = SeriesPartLecture{
			ID:               l.ID,
			Title:            l.Title,
			WatchTimeSeconds: l.WatchTimeSeconds,
			ReadTimeSeconds:  l.ReadTimeSeconds,
			IsPublished:      l.IsPublished,
			Links: SelfLinkResponse{
				LinkResponse{
					fmt.Sprintf(
						"https://%s%s/%s%s/%s%s/%d%s/%d",
						c.backendDomain,
						paths.LanguagePathV1,
						languageSlug,
						paths.SeriesPath,
						seriesSlug,
						paths.PartsPath,
						partID,
						paths.LecturesPath,
						l.ID,
					),
				},
			},
		}
	}

	return SeriesPartEmbedded{Lectures: seriesPartLectures}
}

func (c *Controllers) newSeriesPartEmbeddedFromDto(
	languageSlug,
	seriesSlug string,
	partID int32,
	lectures []services.SeriesPartLecture,
) SeriesPartEmbedded {
	seriesPartLectures := make([]SeriesPartLecture, len(lectures))

	for i, l := range lectures {
		seriesPartLectures[i] = SeriesPartLecture{
			ID:               l.ID,
			Title:            l.Title,
			WatchTimeSeconds: l.WatchTimeSeconds,
			ReadTimeSeconds:  l.ReadTimeSeconds,
			IsPublished:      l.IsPublished,
			Links: SelfLinkResponse{
				LinkResponse{
					fmt.Sprintf(
						"https://%s%s/%s%s/%s%s/%d%s/%d",
						c.backendDomain,
						paths.LanguagePathV1,
						languageSlug,
						paths.SeriesPath,
						seriesSlug,
						paths.PartsPath,
						partID,
						paths.LecturesPath,
						l.ID,
					),
				},
			},
		}
	}

	return SeriesPartEmbedded{Lectures: seriesPartLectures}
}

type SeriesPartResponse struct {
	ID          int32              `json:"id"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Position    int16              `json:"position"`
	Lectures    int16              `json:"lectures"`
	ReadTime    int32              `json:"readTime"`
	WatchTime   int32              `json:"watchTime"`
	IsPublished bool               `json:"isPublished"`
	Embedded    SeriesPartEmbedded `json:"_embedded"`
	Links       SeriesPartLinks    `json:"_links"`
}

func (c *Controllers) NewSeriesPartResponse(part *db.SeriesPart, lectures []db.Lecture) *SeriesPartResponse {
	return &SeriesPartResponse{
		ID:          part.ID,
		Title:       part.Title,
		Description: part.Description,
		Position:    part.Position,
		Lectures:    part.LecturesCount,
		ReadTime:    part.ReadTimeSeconds,
		WatchTime:   part.WatchTimeSeconds,
		IsPublished: part.IsPublished,
		Links:       c.newSeriesPartLinks(part.LanguageSlug, part.SeriesSlug, part.ID),
		Embedded:    c.newSeriesPartEmbedded(part.LanguageSlug, part.SeriesSlug, part.ID, lectures),
	}
}

func (c *Controllers) NewSeriesPartResponseFromDTO(dto *services.SeriesPartDto, languageSlug, seriesSlug string) *SeriesPartResponse {
	return &SeriesPartResponse{
		ID:          dto.ID,
		Title:       dto.Title,
		Description: dto.Description,
		Position:    dto.Position,
		Lectures:    dto.LecturesCount,
		ReadTime:    dto.ReadTimeSeconds,
		WatchTime:   dto.WatchTimeSeconds,
		IsPublished: dto.IsPublished,
		Links:       c.newSeriesPartLinks(languageSlug, seriesSlug, dto.ID),
		Embedded:    c.newSeriesPartEmbeddedFromDto(languageSlug, seriesSlug, dto.ID, dto.Lectures),
	}
}

type LectureLinks struct {
	Self       LinkResponse  `json:"self"`
	SeriesPart LinkResponse  `json:"seriesPart"`
	Article    *LinkResponse `json:"article,omitempty"`
	Video      *LinkResponse `json:"video,omitempty"`
}

func (c *Controllers) newLectureLinks(languageSlug, seriesSlug string, partID, lectureID, articleID, videoID int32) LectureLinks {
	var articleLink *LinkResponse = nil
	if articleID > 0 {
		articleLink = &LinkResponse{
			fmt.Sprintf("https://%s%s/%s%s/%s%s/%d%s/%d", c.backendDomain, paths.LanguagePathV1, languageSlug, paths.SeriesPath, seriesSlug, paths.PartsPath, partID, paths.LecturesPath, articleID),
		}
	}

	var videoLink *LinkResponse = nil
	if videoID > 0 {
		videoLink = &LinkResponse{
			fmt.Sprintf("https://%s%s/%s%s/%s%s/%d%s/%d", c.backendDomain, paths.LanguagePathV1, languageSlug, paths.SeriesPath, seriesSlug, paths.PartsPath, partID, paths.LecturesPath, videoID),
		}
	}

	return LectureLinks{
		Self: LinkResponse{
			fmt.Sprintf("https://%s%s/%s%s/%s%s/%d%s/%d", c.backendDomain, paths.LanguagePathV1, languageSlug, paths.SeriesPath, seriesSlug, paths.PartsPath, partID, paths.LecturesPath, lectureID),
		},
		SeriesPart: LinkResponse{
			fmt.Sprintf("https://%s%s/%s%s/%s%s/%d", c.backendDomain, paths.LanguagePathV1, languageSlug, paths.SeriesPath, seriesSlug, paths.PartsPath, partID),
		},
		Article: articleLink,
		Video:   videoLink,
	}
}

type EmbeddedLectureArticle struct {
	ID          int32            `json:"id"`
	ReadingTime int32            `json:"readingTime"`
	Content     string           `json:"content"`
	Links       SelfLinkResponse `json:"_links"`
}

func (c *Controllers) newLectureArticleResponse(
	articleID,
	readingTime,
	lectureID int32,
	content string,
	seriesPartID int32,
	languageSlug,
	seriesSlug string,
) *EmbeddedLectureArticle {
	return &EmbeddedLectureArticle{
		ID:          articleID,
		ReadingTime: readingTime,
		Content:     content,
		Links: SelfLinkResponse{
			LinkResponse{
				fmt.Sprintf(
					"https://%s%s/%s%s/%s%s/%d%s/%d%s",
					c.backendDomain,
					paths.LanguagePathV1,
					languageSlug,
					paths.SeriesPath,
					seriesSlug,
					paths.PartsPath,
					seriesPartID,
					paths.LecturesPath,
					lectureID,
					paths.ArticlePath,
				),
			},
		},
	}
}

type EmbeddedLectureVideo struct {
	ID        int32            `json:"id"`
	WatchTime int32            `json:"watchTime"`
	Url       string           `json:"url"`
	Links     SelfLinkResponse `json:"_links"`
}

func (c *Controllers) newLectureVideoResponse(
	videoID,
	watchTime int32,
	lectureID int32,
	url string,
	seriesPartID int32,
	languageSlug,
	seriesSlug string,
) *EmbeddedLectureVideo {
	return &EmbeddedLectureVideo{
		ID:        videoID,
		WatchTime: watchTime,
		Url:       url,
		Links: SelfLinkResponse{
			LinkResponse{
				fmt.Sprintf(
					"https://%s%s/%s%s/%s%s/%d%s/%d%s",
					c.backendDomain,
					paths.LanguagePathV1,
					languageSlug,
					paths.SeriesPath,
					seriesSlug,
					paths.PartsPath,
					seriesPartID,
					paths.LecturesPath,
					lectureID,
					paths.VideoPath,
				),
			},
		},
	}
}

type EmbeddedLectureFile struct {
	ID    uuid.UUID        `json:"id"`
	Name  string           `json:"name"`
	Url   string           `json:"url"`
	Links SelfLinkResponse `json:"_links"`
}

func (c *Controllers) newEmbeddedLectureFiles(
	files []db.LectureFile,
	fileUrlsContainer *services.FileURLsContainer,
	languageSlug,
	seriesSlug string,
	seriesPartID int32,
) []EmbeddedLectureFile {
	filesLen := len(files)
	if filesLen == 0 || fileUrlsContainer == nil {
		return []EmbeddedLectureFile{}
	}

	embeddedFiles := make([]EmbeddedLectureFile, 0, filesLen)
	for _, f := range files {
		url, ok := fileUrlsContainer.Get(f.File)
		if !ok {
			continue
		}

		embeddedFiles = append(embeddedFiles, EmbeddedLectureFile{
			ID:   f.File,
			Name: f.Filename,
			Url:  url,
			Links: SelfLinkResponse{
				LinkResponse{
					fmt.Sprintf(
						"https://%s%s/%s%s/%s%s/%d%s/%d%s/%s",
						c.backendDomain,
						paths.LanguagePathV1,
						languageSlug,
						paths.SeriesPath,
						seriesSlug,
						paths.PartsPath,
						seriesPartID,
						paths.LecturesPath,
						f.LectureID,
						paths.FilesPath,
						f.File.String(),
					),
				},
			},
		})
	}

	return embeddedFiles
}

type LectureEmbedded struct {
	Article *EmbeddedLectureArticle `json:"article,omitempty"`
	Video   *EmbeddedLectureVideo   `json:"video,omitempty"`
	Files   []EmbeddedLectureFile   `json:"files,omitempty"`
}

func (c *Controllers) newLectureEmbedded(
	article *EmbeddedLectureArticle,
	video *EmbeddedLectureVideo,
	files []EmbeddedLectureFile,
) *LectureEmbedded {
	if article == nil && video == nil && files == nil {
		return nil
	}

	return &LectureEmbedded{
		Article: article,
		Video:   video,
		Files:   files,
	}
}

type LectureResponse struct {
	ID       int32            `json:"id"`
	Title    string           `json:"title"`
	Position int16            `json:"position"`
	Comments int32            `json:"comments"`
	Embedded *LectureEmbedded `json:"_embedded,omitempty"`
	Links    LectureLinks     `json:"_links"`
}

func (c *Controllers) NewLectureResponse(
	lecture *db.Lecture,
	article *db.LectureArticle,
	video *db.LectureVideo,
	files []db.LectureFile,
	fileUrlsContainer *services.FileURLsContainer,
) *LectureResponse {
	var articleID, videoID int32
	var lecArt *EmbeddedLectureArticle
	var lecVid *EmbeddedLectureVideo
	var lecFs []EmbeddedLectureFile

	if article != nil {
		articleID = article.ID
		lecArt = c.newLectureArticleResponse(
			article.ID,
			article.ReadTimeSeconds,
			article.LectureID,
			article.Content,
			lecture.SeriesPartID,
			lecture.LanguageSlug,
			lecture.SeriesSlug,
		)
	}
	if video != nil {
		videoID = video.ID
		lecVid = c.newLectureVideoResponse(
			video.ID,
			video.WatchTimeSeconds,
			video.LectureID,
			video.Url,
			lecture.SeriesPartID,
			lecture.LanguageSlug,
			lecture.SeriesSlug,
		)
	}
	if files != nil {
		lecFs = c.newEmbeddedLectureFiles(
			files,
			fileUrlsContainer,
			lecture.LanguageSlug,
			lecture.SeriesSlug,
			lecture.SeriesPartID,
		)
	}

	return &LectureResponse{
		ID:       lecture.ID,
		Title:    lecture.Title,
		Position: lecture.Position,
		Comments: lecture.CommentsCount,
		Embedded: c.newLectureEmbedded(lecArt, lecVid, lecFs),
		Links: c.newLectureLinks(
			lecture.LanguageSlug,
			lecture.SeriesSlug,
			lecture.SeriesPartID,
			lecture.ID,
			articleID,
			videoID,
		),
	}
}

func (c *Controllers) NewLectureResponseFromJoinedRow(
	lecture *db.FindPaginatedPublishedLecturesBySeriesPartIDWithArticleAndVideoRow,
	files []db.LectureFile,
	fileUrlsContainer *services.FileURLsContainer,
) *LectureResponse {
	var articleID, videoID int32
	var lecArt *EmbeddedLectureArticle
	var lecVid *EmbeddedLectureVideo
	var lecFs []EmbeddedLectureFile

	if lecture.ArticleID.Valid {
		articleID = lecture.ArticleID.Int32
		lecArt = c.newLectureArticleResponse(
			articleID,
			lecture.ReadTimeSeconds,
			lecture.ID,
			lecture.ArticleContent.String,
			lecture.SeriesPartID,
			lecture.LanguageSlug,
			lecture.SeriesSlug,
		)
	}

	if lecture.VideoID.Valid {
		videoID = lecture.VideoID.Int32
		lecVid = c.newLectureVideoResponse(
			videoID,
			lecture.WatchTimeSeconds,
			lecture.ID,
			lecture.VideoUrl.String,
			lecture.SeriesPartID,
			lecture.LanguageSlug,
			lecture.SeriesSlug,
		)
	}

	if files != nil && fileUrlsContainer != nil {
		lecFs = c.newEmbeddedLectureFiles(
			files,
			fileUrlsContainer,
			lecture.LanguageSlug,
			lecture.SeriesSlug,
			lecture.SeriesPartID,
		)
	}

	return &LectureResponse{
		ID:       lecture.ID,
		Title:    lecture.Title,
		Position: lecture.Position,
		Embedded: c.newLectureEmbedded(lecArt, lecVid, lecFs),
		Links: c.newLectureLinks(
			lecture.LanguageSlug,
			lecture.SeriesSlug,
			lecture.SeriesPartID,
			lecture.ID,
			articleID,
			videoID,
		),
	}
}

type LectureArticleLinks struct {
	Self    LinkResponse `json:"self"`
	Lecture LinkResponse `json:"lecture"`
}

func (c *Controllers) newLectureArticleLinks(languageSlug, seriesSlug string, partID, lectureID int32) LectureArticleLinks {
	return LectureArticleLinks{
		Self: LinkResponse{
			fmt.Sprintf(
				"https://%s%s/%s%s/%s%s/%d%s/%d%s",
				c.backendDomain,
				paths.LanguagePathV1,
				languageSlug,
				paths.SeriesPath,
				seriesSlug,
				paths.PartsPath,
				partID,
				paths.LecturesPath,
				lectureID,
				paths.ArticlePath,
			),
		},
		Lecture: LinkResponse{
			fmt.Sprintf(
				"https://%s%s/%s%s/%s%s/%d%s/%d",
				c.backendDomain,
				paths.LanguagePathV1,
				languageSlug,
				paths.SeriesPath,
				seriesSlug,
				paths.PartsPath,
				partID,
				paths.LecturesPath,
				lectureID,
			),
		},
	}
}

type LectureArticleResponse struct {
	ID          int32               `json:"id"`
	ReadingTime int32               `json:"readingTime"`
	Content     string              `json:"content"`
	Links       LectureArticleLinks `json:"_links"`
}

func (c *Controllers) NewLectureArticleResponse(
	article *db.LectureArticle,
	languageSlug,
	seriesSlug string,
	seriesPartID int32,
) *LectureArticleResponse {
	return &LectureArticleResponse{
		ID:          article.ID,
		ReadingTime: article.ReadTimeSeconds,
		Content:     article.Content,
		Links: c.newLectureArticleLinks(
			languageSlug,
			seriesSlug,
			seriesPartID,
			article.LectureID,
		),
	}
}

type LectureVideoLinks struct {
	Self    LinkResponse `json:"self"`
	Lecture LinkResponse `json:"lecture"`
}

func (c *Controllers) newLectureVideoLinks(languageSlug, seriesSlug string, partID, lectureID int32) LectureVideoLinks {
	return LectureVideoLinks{
		Self: LinkResponse{
			fmt.Sprintf(
				"https://%s%s/%s%s/%s%s/%d%s/%d%s",
				c.backendDomain,
				paths.LanguagePathV1,
				languageSlug,
				paths.SeriesPath,
				seriesSlug,
				paths.PartsPath,
				partID,
				paths.LecturesPath,
				lectureID,
				paths.VideoPath,
			),
		},
		Lecture: LinkResponse{
			fmt.Sprintf(
				"https://%s%s/%s%s/%s%s/%d%s/%d",
				c.backendDomain,
				paths.LanguagePathV1,
				languageSlug,
				paths.SeriesPath,
				seriesSlug,
				paths.PartsPath,
				partID,
				paths.LecturesPath,
				lectureID,
			),
		},
	}
}

type LectureVideoResponse struct {
	ID        int32             `json:"id"`
	WatchTime int32             `json:"watchTime"`
	Url       string            `json:"url"`
	Links     LectureVideoLinks `json:"_links"`
}

func (c *Controllers) NewLectureVideoResponse(
	video *db.LectureVideo,
	languageSlug,
	seriesSlug string,
	seriesPartID int32,
) *LectureVideoResponse {
	return &LectureVideoResponse{
		ID:        video.ID,
		WatchTime: video.WatchTimeSeconds,
		Url:       video.Url,
		Links: c.newLectureVideoLinks(
			languageSlug,
			seriesSlug,
			seriesPartID,
			video.LectureID,
		),
	}
}

type LectureFileLinks struct {
	Self         LinkResponse `json:"self"`
	LectureFiles LinkResponse `json:"lectureFiles"`
}

func (c *Controllers) newLectureFileLinks(
	languageSlug,
	seriesSlug string,
	partID,
	lectureID int32,
	fileID uuid.UUID,
) LectureFileLinks {
	return LectureFileLinks{
		Self: LinkResponse{
			fmt.Sprintf(
				"https://%s%s/%s%s/%s%s/%d%s/%d%s/%s",
				c.backendDomain,
				paths.LanguagePathV1,
				languageSlug,
				paths.SeriesPath,
				seriesSlug,
				paths.PartsPath,
				partID,
				paths.LecturesPath,
				lectureID,
				paths.FilesPath,
				fileID.String(),
			),
		},
		LectureFiles: LinkResponse{
			fmt.Sprintf(
				"https://%s%s/%s%s/%s%s/%d%s/%d%s",
				c.backendDomain,
				paths.LanguagePathV1,
				languageSlug,
				paths.SeriesPath,
				seriesSlug,
				paths.PartsPath,
				partID,
				paths.LecturesPath,
				lectureID,
				paths.FilesPath,
			),
		},
	}
}

type LectureFileResponse struct {
	ID    int32            `json:"id"`
	Name  string           `json:"name"`
	Ext   string           `json:"ext"`
	URL   string           `json:"url"`
	Links LectureFileLinks `json:"_links"`
}

func (c *Controllers) NewLectureFileResponse(
	file *db.LectureFile,
	fileUrl,
	languageSlug,
	seriesSlug string,
	partID int32,
) *LectureFileResponse {
	return &LectureFileResponse{
		ID:   file.ID,
		Name: file.Filename,
		Ext:  file.Ext,
		URL:  fileUrl,
		Links: c.newLectureFileLinks(
			languageSlug,
			seriesSlug,
			partID,
			file.LectureID,
			file.File,
		),
	}
}
