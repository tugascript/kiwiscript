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
	"github.com/gofiber/fiber/v2"
	"github.com/kiwiscript/kiwiscript_go/services"
	"strconv"
)

func (c *Controllers) CreateLectureVideo(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.CreateLectureVideo")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	seriesPartID := ctx.Params("seriesPartID")
	lectureID := ctx.Params("lectureID")
	log.InfoContext(
		userCtx,
		"Creating lecture video...",
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

	var request LectureVideoRequest
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	seriesPartIDi32 := int32(parsedSeriesPartID)
	lectureIDi32 := int32(parsedLectureID)
	video, serviceErr := c.services.CreateLectureVideo(userCtx, services.CreateLectureVideoOptions{
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SeriesPartID: seriesPartIDi32,
		LectureID:    lectureIDi32,
		Url:          request.URL,
		WatchTime:    request.WatchTime,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.
		Status(fiber.StatusCreated).
		JSON(
			c.NewLectureVideoResponse(
				video,
				params.LanguageSlug,
				params.SeriesSlug,
				seriesPartIDi32,
			),
		)
}

func (c *Controllers) GetLectureVideo(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.CreateLectureActicle")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	seriesPartID := ctx.Params("seriesPartID")
	lectureID := ctx.Params("lectureID")
	log.InfoContext(
		userCtx,
		"Creating lecture video...",
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
	isPublished := false
	if user, serviceErr := c.GetUserClaims(ctx); serviceErr != nil || !user.IsStaff {
		isPublished = true
	}

	video, serviceErr := c.services.FindLectureVideo(userCtx, services.FindLectureVideoOptions{
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SeriesPartID: seriesPartIDi32,
		LectureID:    lectureIDi32,
		IsPublished:  isPublished,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.
		Status(fiber.StatusOK).
		JSON(
			c.NewLectureVideoResponse(
				video,
				params.LanguageSlug,
				params.SeriesSlug,
				seriesPartIDi32,
			),
		)
}

func (c *Controllers) UpdateLectureVideo(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.UpdateLectureVideo")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	seriesPartID := ctx.Params("seriesPartID")
	lectureID := ctx.Params("lectureID")
	log.InfoContext(
		userCtx,
		"Updating lecture video...",
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

	var request LectureVideoRequest
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	seriesPartIDi32 := int32(parsedSeriesPartID)
	lectureIDi32 := int32(parsedLectureID)
	video, serviceErr := c.services.UpdateLectureVideo(userCtx, services.UpdateLectureVideoOptions{
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SeriesPartID: seriesPartIDi32,
		LectureID:    lectureIDi32,
		Url:          request.URL,
		WatchTime:    request.WatchTime,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.
		JSON(
			c.NewLectureVideoResponse(
				video,
				params.LanguageSlug,
				params.SeriesSlug,
				seriesPartIDi32,
			),
		)
}

func (c *Controllers) DeleteLectureVideo(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.DeleteLectureVideo")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	seriesPartID := ctx.Params("seriesPartID")
	lectureID := ctx.Params("lectureID")
	log.InfoContext(
		userCtx,
		"Delete lecture video...",
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
	opts := services.DeleteLectureVideoOptions{
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SeriesPartID: seriesPartIDi32,
		LectureID:    lectureIDi32,
	}
	if serviceErr := c.services.DeleteLectureVideo(userCtx, opts); serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
