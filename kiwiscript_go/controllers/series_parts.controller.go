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
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/kiwiscript/kiwiscript_go/paths"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/services"
)

func (c *Controllers) CreateSeriesPart(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.CreateSeriesPart")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	log.InfoContext(userCtx, "Creating series part...", "languageSlug", languageSlug, "seriesSlug", seriesSlug)

	user, err := c.GetUserClaims(ctx)
	if err != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(NewRequestError(services.NewForbiddenError()))
	}

	params := SeriesParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	var request CreateSeriesPartRequest
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	seriesPart, serviceErr := c.services.CreateSeriesPart(userCtx, services.CreateSeriesPartOptions{
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		Title:        request.Title,
		Description:  request.Description,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.Status(fiber.StatusCreated).JSON(
		c.NewSeriesPartResponse(
			seriesPart,
			make([]db.Lecture, 0),
			params.LanguageSlug,
			params.SeriesSlug,
		),
	)
}

func (c *Controllers) GetSeriesPart(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.GetSingleSeries")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	seriesPartID := ctx.Params("seriesPartID")
	log.InfoContext(userCtx, "Getting series part...", "languageSlug", languageSlug, "seriesSlug", seriesSlug, "seriesPartID", seriesPartID)

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

	isPublished := false
	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		isPublished = true
	}

	seriesPart, serviceErr := c.services.FindSeriesPartBySlugAndID(userCtx, services.FindSeriesPartBySlugsAndIDOptions{
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		IsPublished:  isPublished,
		SeriesPartID: int32(parsedSeriesPartID),
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(c.NewSeriesPartResponseFromDTO(seriesPart, params.LanguageSlug, params.SeriesSlug))
}

func (c *Controllers) GetSeriesParts(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.GetSingleSeries")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	log.InfoContext(userCtx, "Getting series parts...", "languageSlug", languageSlug, "seriesSlug", seriesSlug)

	params := SeriesParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	queryParams := SeriesPartsQueryParams{
		IsPublished:       ctx.QueryBool("isPublished", false),
		PublishedLectures: ctx.QueryBool("publishedLectures", false),
		Offset:            int32(ctx.QueryInt("offset", OffsetDefault)),
		Limit:             int32(ctx.QueryInt("limit", LimitDefault)),
	}
	if err := c.validate.StructCtx(userCtx, queryParams); err != nil {
		return c.validateQueryErrorResponse(log, userCtx, err, ctx)
	}

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		queryParams.IsPublished = true
	}

	seriesParts, count, serviceErr := c.services.FindPaginatedSeriesPartsBySlugsAndID(userCtx, services.FindSeriesPartsBySlugsAndIDOptions{
		LanguageSlug:      params.LanguageSlug,
		SeriesSlug:        params.SeriesSlug,
		IsPublished:       queryParams.IsPublished,
		PublishedLectures: queryParams.PublishedLectures,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(
		NewPaginatedResponse(
			c.backendDomain,
			fmt.Sprintf(
				"%s/%s%s/%s%s",
				paths.LanguagePathV1,
				languageSlug,
				paths.SeriesPath,
				seriesSlug,
				paths.PartsPath,
			),
			queryParams,
			count,
			seriesParts,
			func(dto *services.SeriesPartDto) *SeriesPartResponse {
				return c.NewSeriesPartResponseFromDTO(dto, languageSlug, seriesSlug)
			},
		),
	)
}

func (c *Controllers) UpdateSeriesPart(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.UpdateSingleSeries")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	seriesPartID := ctx.Params("seriesPartID")
	log.InfoContext(
		userCtx,
		"Updating series part...",
		"languageSlug",
		languageSlug,
		"seriesSlug",
		seriesSlug,
		"seriesPartID",
		seriesPartID,
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

	var request UpdateSeriesPartRequest
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	seriesPart, serviceErr := c.services.UpdateSeriesPart(userCtx, services.UpdateSeriesPartOptions{
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SeriesPartID: int32(parsedSeriesPartID),
		Title:        request.Title,
		Description:  request.Description,
		Position:     request.Position,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	lectures, serviceErr := c.services.FindLecturesBySeriesPartID(userCtx, seriesPart.ID)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.
		JSON(
			c.NewSeriesPartResponse(
				seriesPart,
				lectures,
				params.LanguageSlug,
				params.SeriesSlug,
			),
		)
}

func (c *Controllers) UpdateSeriesPartIsPublished(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.UpdateSingleSeries")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	seriesPartID := ctx.Params("seriesPartID")
	log.InfoContext(
		userCtx,
		"Updating series part...",
		"languageSlug",
		languageSlug,
		"seriesSlug",
		seriesSlug,
		"seriesPartID",
		seriesPartID,
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

	var request UpdateIsPublishedRequest
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	seriesPart, serviceErr := c.services.UpdateSeriesPartIsPublished(userCtx, services.UpdateSeriesPartIsPublishedOptions{
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SeriesPartID: int32(parsedSeriesPartID),
		IsPublished:  request.IsPublished,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	lectures, serviceErr := c.services.FindLecturesBySeriesPartID(userCtx, seriesPart.ID)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.
		JSON(
			c.NewSeriesPartResponse(
				seriesPart,
				lectures,
				params.LanguageSlug,
				params.SeriesSlug,
			),
		)
}

func (c *Controllers) DeleteSeriesPart(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.UpdateSingleSeries")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	seriesPartID := ctx.Params("seriesPartID")
	log.InfoContext(
		userCtx,
		"Updating series part...",
		"languageSlug",
		languageSlug,
		"seriesSlug",
		seriesSlug,
		"seriesPartID",
		seriesPartID,
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

	serviceErr = c.services.DeleteSeriesPart(userCtx, services.DeleteSeriesPartOptions{
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SeriesPartID: int32(parsedSeriesPartID),
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
