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
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kiwiscript/kiwiscript_go/dtos"
	"github.com/kiwiscript/kiwiscript_go/exceptions"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/kiwiscript/kiwiscript_go/paths"
	"github.com/kiwiscript/kiwiscript_go/services"
)

const lessonLocation string = "lesson"

func (c *Controllers) CreateLesson(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	sectionID := ctx.Params("sectionID")
	log := c.buildLogger(ctx, requestID, lessonLocation, "CreateLesson").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"sectionId", sectionID,
	)
	log.InfoContext(userCtx, "Creating lesson...")

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
	}

	params := dtos.SectionPathParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SectionID:    sectionID,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	parsedSectionID, err := strconv.Atoi(params.SectionID)
	if err != nil || parsedSectionID <= 0 {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(exceptions.NewRequestValidationError(exceptions.RequestValidationLocationParams, []exceptions.FieldError{{
				Param:   "sectionId",
				Message: exceptions.StrFieldErrMessageNumber,
				Value:   params.SectionID,
			}}))
	}

	var request dtos.CreateLessonBody
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	sectionIDi32 := int32(parsedSectionID)
	lesson, serviceErr := c.services.CreateLesson(userCtx, services.CreateLessonOptions{
		RequestID:    requestID,
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SectionID:    sectionIDi32,
		Title:        request.Title,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.
		Status(fiber.StatusCreated).
		JSON(dtos.NewLessonResponse(c.backendDomain, lesson.ToLessonModel()))
}

func (c *Controllers) findLessonFiles(
	userCtx context.Context,
	lessonID int32,
) ([]db.LessonFile, *services.FileURLsContainer, *exceptions.ServiceError) {
	lessonFiles, serviceErr := c.services.FindLessonFilesWithNoCheck(userCtx, lessonID)
	if serviceErr != nil {
		return nil, nil, serviceErr
	}

	filesLen := len(lessonFiles)
	var fileUrls *services.FileURLsContainer
	if filesLen > 0 {
		optsList := make([]services.FindFileURLOptions, 0, filesLen)
		for _, lessonFile := range lessonFiles {
			optsList = append(optsList, services.FindFileURLOptions{
				UserID:  lessonFile.AuthorID,
				FileID:  lessonFile.ID,
				FileExt: lessonFile.Ext,
			})
		}
		var serviceErr *exceptions.ServiceError
		fileUrls, serviceErr = c.services.FindFileURLs(userCtx, optsList)
		if serviceErr != nil {
			return nil, nil, serviceErr
		}
	}

	return lessonFiles, fileUrls, nil
}

func (c *Controllers) GetLesson(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	sectionID := ctx.Params("sectionID")
	lessonID := ctx.Params("lessonID")
	log := c.buildLogger(ctx, requestID, lessonLocation, "GetLesson").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"sectionID", sectionID,
		"lessonID", lessonID,
	)
	log.InfoContext(userCtx, "Getting lesson...")

	params := dtos.LessonPathParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SectionID:    sectionID,
		LessonID:     lessonID,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	parsedSectionID, err := strconv.Atoi(params.SectionID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(exceptions.NewRequestValidationError(exceptions.RequestValidationLocationParams, []exceptions.FieldError{{
				Param:   "sectionId",
				Message: exceptions.StrFieldErrMessageNumber,
				Value:   params.SectionID,
			}}))
	}

	parsedLessonID, err := strconv.Atoi(params.LessonID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(exceptions.NewRequestValidationError(exceptions.RequestValidationLocationParams, []exceptions.FieldError{{
				Param:   "lessonsId",
				Message: exceptions.StrFieldErrMessageNumber,
				Value:   params.SectionID,
			}}))
	}

	sectionIDi32 := int32(parsedSectionID)
	lessonIDi32 := int32(parsedLessonID)
	if user, serviceErr := c.GetUserClaims(ctx); serviceErr == nil {
		if user.IsStaff {
			lesson, serviceErr := c.services.FindLessonWithArticleAndVideo(userCtx, services.FindLessonOptions{
				RequestID:    requestID,
				LanguageSlug: params.LanguageSlug,
				SeriesSlug:   params.SeriesSlug,
				SectionID:    sectionIDi32,
				LessonID:     lessonIDi32,
			})
			if serviceErr != nil {
				return c.serviceErrorResponse(serviceErr, ctx)
			}

			files, fileUrls, serviceErr := c.findLessonFiles(userCtx, lessonIDi32)
			if serviceErr != nil {
				return c.serviceErrorResponse(serviceErr, ctx)
			}

			return ctx.JSON(
				dtos.NewLessonResponseWithEmbeddedOptions(
					c.backendDomain,
					lesson.ToLessonModel(),
					lesson.LessonActicleID,
					lesson.LessonVideoID,
					lesson.LessonArticleContent.String,
					lesson.LessonVideoUrl.String,
					files,
					fileUrls,
				),
			)
		}

		lesson, serviceErr := c.services.FindPublishedLessonWithProgressArticleAndVideo(
			userCtx,
			services.FindLessonWithProgressOptions{
				RequestID:    requestID,
				UserID:       user.ID,
				LanguageSlug: params.LanguageSlug,
				SeriesSlug:   params.SeriesSlug,
				SectionID:    sectionIDi32,
				LessonID:     lessonIDi32,
			},
		)
		if serviceErr != nil {
			return c.serviceErrorResponse(serviceErr, ctx)
		}

		files, fileUrls, serviceErr := c.findLessonFiles(userCtx, lessonIDi32)
		if serviceErr != nil {
			return c.serviceErrorResponse(serviceErr, ctx)
		}

		return ctx.JSON(
			dtos.NewLessonResponseWithEmbeddedOptions(
				c.backendDomain,
				lesson.ToLessonModel(),
				lesson.LessonActicleID,
				lesson.LessonVideoID,
				lesson.LessonArticleContent.String,
				lesson.LessonVideoUrl.String,
				files,
				fileUrls,
			),
		)
	}

	lesson, serviceErr := c.services.FindPublishedLessonWithArticleAndVideo(userCtx, services.FindLessonOptions{
		RequestID:    requestID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SectionID:    sectionIDi32,
		LessonID:     lessonIDi32,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	files, fileUrls, serviceErr := c.findLessonFiles(userCtx, lessonIDi32)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(
		dtos.NewLessonResponseWithEmbeddedOptions(
			c.backendDomain,
			lesson.ToLessonModel(),
			lesson.LessonActicleID,
			lesson.LessonVideoID,
			lesson.LessonArticleContent.String,
			lesson.LessonVideoUrl.String,
			files,
			fileUrls,
		),
	)
}

func (c *Controllers) GetLessons(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	sectionID := ctx.Params("sectionID")
	log := c.buildLogger(ctx, requestID, lessonLocation, "GetLessons").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"sectionId", sectionID,
	)
	log.InfoContext(userCtx, "Creating lesson...")

	params := dtos.SectionPathParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SectionID:    sectionID,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	parsedSectionID, err := strconv.Atoi(params.SectionID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(exceptions.NewRequestValidationError(exceptions.RequestValidationLocationParams, []exceptions.FieldError{{
				Param:   "sectionId",
				Message: exceptions.StrFieldErrMessageNumber,
				Value:   params.SectionID,
			}}))
	}

	queryParams := dtos.PaginationQueryParams{
		Offset: int32(ctx.QueryInt("offset", dtos.OffsetDefault)),
		Limit:  int32(ctx.QueryInt("limit", dtos.LimitDefault)),
	}
	if err := c.validate.StructCtx(userCtx, queryParams); err != nil {
		return c.validateQueryErrorResponse(log, userCtx, err, ctx)
	}

	sectionIDi32 := int32(parsedSectionID)
	if user, serviceErr := c.GetUserClaims(ctx); serviceErr == nil {
		if user.IsStaff {
			lessons, count, serviceErr := c.services.FindPaginatedLessons(userCtx, services.FindPaginatedLessonsOptions{
				RequestID:    requestID,
				LanguageSlug: params.LanguageSlug,
				SeriesSlug:   params.SeriesSlug,
				SectionID:    sectionIDi32,
				Offset:       queryParams.Offset,
				Limit:        queryParams.Limit,
			})
			if serviceErr != nil {
				return c.serviceErrorResponse(serviceErr, ctx)
			}

			return ctx.JSON(
				dtos.NewPaginatedResponse(
					c.backendDomain,
					fmt.Sprintf(
						"%s/%s%s/%s%s/%d%s",
						paths.LanguagePathV1,
						languageSlug,
						paths.SeriesPath,
						seriesSlug,
						paths.SectionsPath,
						sectionIDi32,
						paths.LessonsPath,
					),
					&queryParams,
					count,
					lessons,
					func(l *db.Lesson) *dtos.LessonResponse {
						return dtos.NewLessonResponse(c.backendDomain, l.ToLessonModel())
					},
				),
			)
		}

		lessons, count, serviceErr := c.services.FindPaginatedPublishedLessonsWithProgress(
			userCtx,
			services.FindPaginatedPublishedLessonsWithProgressOptions{
				RequestID:    requestID,
				UserID:       user.ID,
				LanguageSlug: params.LanguageSlug,
				SeriesSlug:   params.SeriesSlug,
				SectionID:    sectionIDi32,
				Offset:       queryParams.Offset,
				Limit:        queryParams.Limit,
			},
		)
		if serviceErr != nil {
			return c.serviceErrorResponse(serviceErr, ctx)
		}

		return ctx.JSON(
			dtos.NewPaginatedResponse(
				c.backendDomain,
				fmt.Sprintf(
					"%s/%s%s/%s%s/%d%s",
					paths.LanguagePathV1,
					languageSlug,
					paths.SeriesPath,
					seriesSlug,
					paths.SectionsPath,
					sectionIDi32,
					paths.LessonsPath,
				),
				&queryParams,
				count,
				lessons,
				func(l *db.FindPaginatedPublishedLessonsBySlugsAndSectionIDWithProgressRow) *dtos.LessonResponse {
					return dtos.NewLessonResponse(c.backendDomain, l.ToLessonModel())
				},
			),
		)
	}

	lessons, count, serviceErr := c.services.FindPaginatedPublishedLessons(
		userCtx,
		services.FindPaginatedLessonsOptions{
			RequestID:    requestID,
			LanguageSlug: params.LanguageSlug,
			SeriesSlug:   params.SeriesSlug,
			SectionID:    sectionIDi32,
			Offset:       queryParams.Offset,
			Limit:        queryParams.Limit,
		},
	)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(
		dtos.NewPaginatedResponse(
			c.backendDomain,
			fmt.Sprintf(
				"%s/%s%s/%s%s/%d%s",
				paths.LanguagePathV1,
				languageSlug,
				paths.SeriesPath,
				seriesSlug,
				paths.SectionsPath,
				sectionIDi32,
				paths.LessonsPath,
			),
			&queryParams,
			count,
			lessons,
			func(l *db.Lesson) *dtos.LessonResponse {
				return dtos.NewLessonResponse(c.backendDomain, l.ToLessonModel())
			},
		),
	)
}

func (c *Controllers) UpdateLesson(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	sectionID := ctx.Params("sectionID")
	lessonID := ctx.Params("lessonID")
	log := c.buildLogger(ctx, requestID, lessonLocation, "UpdateLesson").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"sectionId", sectionID,
		"lessonID", lessonID,
	)
	log.InfoContext(userCtx, "Updating lesson...")

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
	}

	params := dtos.LessonPathParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SectionID:    sectionID,
		LessonID:     lessonID,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	parsedSectionID, err := strconv.Atoi(params.SectionID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(exceptions.NewRequestValidationError(exceptions.RequestValidationLocationParams, []exceptions.FieldError{{
				Param:   "sectionId",
				Message: exceptions.StrFieldErrMessageNumber,
				Value:   params.SectionID,
			}}))
	}

	parsedLessonID, err := strconv.Atoi(params.LessonID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(exceptions.NewRequestValidationError(exceptions.RequestValidationLocationParams, []exceptions.FieldError{{
				Param:   "lessonsId",
				Message: exceptions.StrFieldErrMessageNumber,
				Value:   params.SectionID,
			}}))
	}

	var request dtos.UpdateLessonBody
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	sectionIDi32 := int32(parsedSectionID)
	lessonIDi32 := int32(parsedLessonID)
	lesson, serviceErr := c.services.UpdateLesson(userCtx, services.UpdateLessonOptions{
		RequestID:    requestID,
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SectionID:    sectionIDi32,
		LessonID:     lessonIDi32,
		Title:        request.Title,
		Position:     request.Position,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	var article *db.LessonArticle
	var articleID pgtype.Int4
	var articleContent string
	if lesson.ReadTimeSeconds > 0 {
		article, serviceErr = c.services.FindLessonArticleByLessonID(
			userCtx,
			services.FindLessonArticleByLessonIDOptions{
				RequestID: requestID,
				LessonID:  lessonIDi32,
			},
		)
		if serviceErr != nil {
			return c.serviceErrorResponse(serviceErr, ctx)
		}

		articleID = pgtype.Int4{
			Int32: article.ID,
			Valid: true,
		}
		articleContent = article.Content
	}

	var video *db.LessonVideo
	var videoID pgtype.Int4
	var videoURL string
	if lesson.WatchTimeSeconds > 0 {
		video, serviceErr = c.services.FindLessonVideoByLessonID(userCtx, services.FindLessonVideoByLessonIDOptions{
			RequestID: requestID,
			LessonID:  lessonIDi32,
		})
		if serviceErr != nil {
			return c.serviceErrorResponse(serviceErr, ctx)
		}

		videoID = pgtype.Int4{
			Int32: video.ID,
			Valid: true,
		}
		videoURL = video.Url
	}

	files, fileUrls, serviceErr := c.findLessonFiles(userCtx, lessonIDi32)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(
		dtos.NewLessonResponseWithEmbeddedOptions(
			c.backendDomain,
			lesson.ToLessonModel(),
			articleID,
			videoID,
			articleContent,
			videoURL,
			files,
			fileUrls,
		),
	)
}

func (c *Controllers) UpdateLessonIsPublished(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	sectionID := ctx.Params("sectionID")
	lessonID := ctx.Params("lessonID")
	log := c.buildLogger(ctx, requestID, lessonLocation, "UpdateLessonIsPublished").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"sectionId", sectionID,
		"lessonID", lessonID,
	)
	log.InfoContext(userCtx, "Updating lesson...")

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
	}

	params := dtos.LessonPathParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SectionID:    sectionID,
		LessonID:     lessonID,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	parsedSectionID, err := strconv.Atoi(params.SectionID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(exceptions.NewRequestValidationError(exceptions.RequestValidationLocationParams, []exceptions.FieldError{{
				Param:   "sectionId",
				Message: exceptions.StrFieldErrMessageNumber,
				Value:   params.SectionID,
			}}))
	}

	parsedLessonID, err := strconv.Atoi(params.LessonID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(exceptions.NewRequestValidationError(exceptions.RequestValidationLocationParams, []exceptions.FieldError{{
				Param:   "lessonsId",
				Message: exceptions.StrFieldErrMessageNumber,
				Value:   params.SectionID,
			}}))
	}

	var request dtos.UpdateIsPublishedBody
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	sectionIDi32 := int32(parsedSectionID)
	lessonIDi32 := int32(parsedLessonID)
	lesson, serviceErr := c.services.UpdateLessonIsPublished(userCtx, services.UpdateLessonIsPublishedOptions{
		RequestID:    requestID,
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SectionID:    sectionIDi32,
		LessonID:     lessonIDi32,
		IsPublished:  request.IsPublished,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	var article *db.LessonArticle
	var articleID pgtype.Int4
	var articleContent string
	if lesson.ReadTimeSeconds > 0 {
		article, serviceErr = c.services.FindLessonArticleByLessonID(
			userCtx,
			services.FindLessonArticleByLessonIDOptions{
				RequestID: requestID,
				LessonID:  lessonIDi32,
			},
		)
		if serviceErr != nil {
			return c.serviceErrorResponse(serviceErr, ctx)
		}

		articleID = pgtype.Int4{
			Int32: article.ID,
			Valid: true,
		}
		articleContent = article.Content
	}

	var video *db.LessonVideo
	var videoID pgtype.Int4
	var videoURL string
	if lesson.WatchTimeSeconds > 0 {
		video, serviceErr = c.services.FindLessonVideoByLessonID(userCtx, services.FindLessonVideoByLessonIDOptions{
			RequestID: requestID,
			LessonID:  lessonIDi32,
		})
		if serviceErr != nil {
			return c.serviceErrorResponse(serviceErr, ctx)
		}

		videoID = pgtype.Int4{
			Int32: video.ID,
			Valid: true,
		}
		videoURL = video.Url
	}

	files, fileUrls, serviceErr := c.findLessonFiles(userCtx, lessonIDi32)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(
		dtos.NewLessonResponseWithEmbeddedOptions(
			c.backendDomain,
			lesson.ToLessonModel(),
			articleID,
			videoID,
			articleContent,
			videoURL,
			files,
			fileUrls,
		),
	)
}

func (c *Controllers) DeleteLesson(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	sectionID := ctx.Params("sectionID")
	lessonID := ctx.Params("lessonID")
	log := c.buildLogger(ctx, requestID, lessonLocation, "DeleteLesson").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"sectionId", sectionID,
		"lessonID", lessonID,
	)
	log.InfoContext(userCtx, "Deleting lesson...")

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
	}

	params := dtos.LessonPathParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SectionID:    sectionID,
		LessonID:     lessonID,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	parsedSectionID, err := strconv.Atoi(params.SectionID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(exceptions.NewRequestValidationError(exceptions.RequestValidationLocationParams, []exceptions.FieldError{{
				Param:   "sectionId",
				Message: exceptions.StrFieldErrMessageNumber,
				Value:   params.SectionID,
			}}))
	}

	parsedLessonID, err := strconv.Atoi(params.LessonID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(exceptions.NewRequestValidationError(exceptions.RequestValidationLocationParams, []exceptions.FieldError{{
				Param:   "lessonsId",
				Message: exceptions.StrFieldErrMessageNumber,
				Value:   params.SectionID,
			}}))
	}

	sectionIDi32 := int32(parsedSectionID)
	lessonIDi32 := int32(parsedLessonID)
	opts := services.DeleteLessonOptions{
		RequestID:    requestID,
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SectionID:    sectionIDi32,
		LessonID:     lessonIDi32,
	}
	if serviceErr := c.services.DeleteLesson(userCtx, opts); serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

func (c *Controllers) GetCurrentLesson(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	sectionID := ctx.Params("sectionID")
	log := c.buildLogger(ctx, requestID, lessonLocation, "GetCurrentLesson").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"sectionId", sectionID,
	)
	log.InfoContext(userCtx, "Getting current lesson...")

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(exceptions.NewRequestError(exceptions.NewUnauthorizedError()))
	}
	if user.IsStaff || user.IsAdmin {
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
	}

	params := dtos.SectionPathParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SectionID:    sectionID,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	parsedSectionID, err := strconv.Atoi(params.SectionID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(exceptions.NewRequestValidationError(exceptions.RequestValidationLocationParams, []exceptions.FieldError{{
				Param:   "sectionId",
				Message: exceptions.StrFieldErrMessageNumber,
				Value:   params.SectionID,
			}}))
	}

	lesson, serviceErr := c.services.FindCurrentLesson(userCtx, services.FindCurrentLessonOptions{
		RequestID:    requestID,
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SectionID:    int32(parsedSectionID),
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	files, fileUrls, serviceErr := c.findLessonFiles(userCtx, lesson.ID)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(
		dtos.NewLessonResponseWithEmbeddedOptions(
			c.backendDomain,
			lesson.ToLessonModel(),
			lesson.LessonActicleID,
			lesson.LessonVideoID,
			lesson.LessonArticleContent.String,
			lesson.LessonVideoUrl.String,
			files,
			fileUrls,
		),
	)
}
