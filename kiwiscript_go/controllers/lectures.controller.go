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
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/kiwiscript/kiwiscript_go/paths"
	"github.com/kiwiscript/kiwiscript_go/services"
)

func (c *Controllers) CreateLecture(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.CreateLecture")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	seriesPartID := ctx.Params("seriesPartID")
	log.InfoContext(
		userCtx,
		"Creating lecture...",
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"seriesPartID", seriesPartID,
	)

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(NewRequestError(services.NewForbiddenError()))
	}

	params := SeriesPartParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SeriesPartID: seriesPartID,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	parsedSeriesPartID, err := strconv.Atoi(params.SeriesPartID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "seriesPartId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SeriesPartID,
			}}))
	}

	var request CreateLectureRequest
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	seriesPartIDi32 := int32(parsedSeriesPartID)
	lecture, serviceErr := c.services.CreateLectures(userCtx, services.CreateLecturesOptions{
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SeriesPartID: seriesPartIDi32,
		Title:        request.Title,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.
		Status(fiber.StatusCreated).
		JSON(
			c.NewLectureResponse(
				lecture,
				nil, nil,
				nil, nil,
			),
		)
}

func (c *Controllers) findLectureFiles(
	userCtx context.Context,
	params *LectureParams,
	seriesPartID,
	lectureID int32,
) ([]db.LectureFile, *services.FileURLsContainer, *services.ServiceError) {
	lectureFiles, serviceErr := c.services.FindLectureFiles(userCtx, services.FindLectureFilesOptions{
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SeriesPartID: seriesPartID,
		LectureID:    lectureID,
		IsPublished:  false,
	})
	if serviceErr != nil {
		return nil, nil, serviceErr
	}

	filesLen := len(lectureFiles)
	var fileUrls *services.FileURLsContainer
	if filesLen > 0 {
		optsList := make([]services.FindFileURLOptions, 0, filesLen)
		for _, lectureFile := range lectureFiles {
			optsList = append(optsList, services.FindFileURLOptions{
				UserID:  lectureFile.AuthorID,
				FileID:  lectureFile.File,
				FileExt: lectureFile.Ext,
			})
		}
		var serviceErr *services.ServiceError
		fileUrls, serviceErr = c.services.FindFileURLs(userCtx, optsList)
		if serviceErr != nil {
			return nil, nil, serviceErr
		}
	}

	return lectureFiles, fileUrls, nil
}

func (c *Controllers) GetLecture(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.GetLecture")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	seriesPartID := ctx.Params("seriesPartID")
	lectureID := ctx.Params("lectureID")
	log.InfoContext(
		userCtx,
		"Getting lecture...",
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"seriesPartID", seriesPartID,
		"lectureID", lectureID,
	)

	params := LectureParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SeriesPartID: seriesPartID,
		LectureID:    lectureID,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	parsedSeriesPartID, err := strconv.Atoi(params.SeriesPartID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "seriesPartId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SeriesPartID,
			}}))
	}

	parsedLectureID, err := strconv.Atoi(params.LectureID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "lecturesId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SeriesPartID,
			}}))
	}

	seriesPartIDi32 := int32(parsedSeriesPartID)
	lectureIDi32 := int32(parsedLectureID)
	lecture, serviceErr := c.services.FindLectureWithArticleAndVideo(userCtx, services.FindLectureOptions{
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SeriesPartID: seriesPartIDi32,
		LectureID:    lectureIDi32,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	if user, serviceErr := c.GetUserClaims(ctx); (serviceErr != nil || !user.IsStaff) && !lecture.IsPublished {
		return c.serviceErrorResponse(services.NewNotFoundError(), ctx)
	}

	files, fileUrls, serviceErr := c.findLectureFiles(userCtx, &params, seriesPartIDi32, lectureIDi32)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(c.NewLectureResponseFromJoinedRow(lecture, files, fileUrls))
}

func (c *Controllers) GetLectures(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.GetLectures")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	seriesPartID := ctx.Params("seriesPartID")
	log.InfoContext(
		userCtx,
		"Creating lecture...",
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"seriesPartID", seriesPartID,
	)

	params := SeriesPartParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SeriesPartID: seriesPartID,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	parsedSeriesPartID, err := strconv.Atoi(params.SeriesPartID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "seriesPartId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SeriesPartID,
			}}))
	}

	queryParams := LecturesQueryParams{
		IsPublished: ctx.QueryBool("isPublished", false),
		Offset:      int32(ctx.QueryInt("offset", OffsetDefault)),
		Limit:       int32(ctx.QueryInt("limit", LimitDefault)),
	}
	if err := c.validate.StructCtx(userCtx, queryParams); err != nil {
		return c.validateQueryErrorResponse(log, userCtx, err, ctx)
	}

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		queryParams.IsPublished = true
	}

	seriesPartIDi32 := int32(parsedSeriesPartID)
	lectures, count, serviceErr := c.services.FindPaginatedLectures(userCtx, services.FindPaginatedLecturesOptions{
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SeriesPartID: seriesPartIDi32,
		IsPublished:  queryParams.IsPublished,
		Offset:       queryParams.Offset,
		Limit:        queryParams.Limit,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(
		NewPaginatedResponse(
			c.backendDomain,
			fmt.Sprintf(
				"https://%s%s/%s%s/%s%s/%d%s",
				c.backendDomain,
				paths.LanguagePathV1,
				languageSlug,
				paths.SeriesPath,
				seriesSlug,
				paths.PartsPath,
				seriesPartIDi32,
				paths.LecturesPath,
			),
			&queryParams,
			count,
			lectures,
			func(l *db.FindPaginatedPublishedLecturesBySeriesPartIDWithArticleAndVideoRow) *LectureResponse {
				return c.NewLectureResponseFromJoinedRow(l, nil, nil)
			},
		),
	)
}

func (c *Controllers) UpdateLecture(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.UpdateLecture")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	seriesPartID := ctx.Params("seriesPartID")
	lectureID := ctx.Params("lectureID")
	log.InfoContext(
		userCtx,
		"Updating lecture...",
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"seriesPartID", seriesPartID,
		"lectureID", lectureID,
	)

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(NewRequestError(services.NewForbiddenError()))
	}

	params := LectureParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SeriesPartID: seriesPartID,
		LectureID:    lectureID,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	parsedSeriesPartID, err := strconv.Atoi(params.SeriesPartID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "seriesPartId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SeriesPartID,
			}}))
	}

	parsedLectureID, err := strconv.Atoi(params.LectureID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "lecturesId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SeriesPartID,
			}}))
	}

	var request UpdateLectureRequest
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	seriesPartIDi32 := int32(parsedSeriesPartID)
	lectureIDi32 := int32(parsedLectureID)
	lecture, serviceErr := c.services.UpdateLecture(userCtx, services.UpdateLectureOptions{
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SeriesPartID: seriesPartIDi32,
		LectureID:    lectureIDi32,
		Title:        request.Title,
		Position:     request.Position,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	article, serviceErr := c.services.FindLectureArticleByLectureID(userCtx, lecture.ID)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	video, serviceErr := c.services.FindLectureVideoByLectureID(userCtx, lecture.ID)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	files, fileUrls, serviceErr := c.findLectureFiles(userCtx, &params, seriesPartIDi32, lectureIDi32)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(
		c.NewLectureResponse(
			lecture,
			article,
			video,
			files,
			fileUrls,
		),
	)
}

func (c *Controllers) UpdateLectureIsPublished(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.UpdateLecture")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	seriesPartID := ctx.Params("seriesPartID")
	lectureID := ctx.Params("lectureID")
	log.InfoContext(
		userCtx,
		"Updating lecture...",
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"seriesPartID", seriesPartID,
		"lectureID", lectureID,
	)

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(NewRequestError(services.NewForbiddenError()))
	}

	params := LectureParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SeriesPartID: seriesPartID,
		LectureID:    lectureID,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	parsedSeriesPartID, err := strconv.Atoi(params.SeriesPartID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "seriesPartId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SeriesPartID,
			}}))
	}

	parsedLectureID, err := strconv.Atoi(params.LectureID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "lecturesId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SeriesPartID,
			}}))
	}

	var request UpdateIsPublishedRequest
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	seriesPartIDi32 := int32(parsedSeriesPartID)
	lectureIDi32 := int32(parsedLectureID)
	lecture, serviceErr := c.services.UpdateLectureIsPublished(userCtx, services.UpdateLectureIsPublishedOptions{
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SeriesPartID: seriesPartIDi32,
		LectureID:    lectureIDi32,
		IsPublished:  request.IsPublished,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	article, serviceErr := c.services.FindLectureArticleByLectureID(userCtx, lecture.ID)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	video, serviceErr := c.services.FindLectureVideoByLectureID(userCtx, lecture.ID)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	files, fileUrls, serviceErr := c.findLectureFiles(userCtx, &params, seriesPartIDi32, lectureIDi32)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(
		c.NewLectureResponse(
			lecture,
			article,
			video,
			files,
			fileUrls,
		),
	)
}

func (c *Controllers) DeleteLecture(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.DeleteLecture")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	seriesPartID := ctx.Params("seriesPartID")
	lectureID := ctx.Params("lectureID")
	log.InfoContext(
		userCtx,
		"Updating lecture...",
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"seriesPartID", seriesPartID,
		"lectureID", lectureID,
	)

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(NewRequestError(services.NewForbiddenError()))
	}

	params := LectureParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SeriesPartID: seriesPartID,
		LectureID:    lectureID,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	parsedSeriesPartID, err := strconv.Atoi(params.SeriesPartID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "seriesPartId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SeriesPartID,
			}}))
	}

	parsedLectureID, err := strconv.Atoi(params.LectureID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "lecturesId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SeriesPartID,
			}}))
	}

	seriesPartIDi32 := int32(parsedSeriesPartID)
	lectureIDi32 := int32(parsedLectureID)
	opts := services.DeleteLectureOptions{
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SeriesPartID: seriesPartIDi32,
		LectureID:    lectureIDi32,
	}
	if serviceErr := c.services.DeleteLecture(userCtx, opts); serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
