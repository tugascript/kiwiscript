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
)

func (c *Controllers) CreateSeries(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.CreateSeries")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	log.InfoContext(userCtx, "Creating series...", "languageSlug", languageSlug)

	user, err := c.GetUserClaims(ctx)
	if err != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(NewRequestError(services.NewForbiddenError()))
	}

	params := LanguageParams{languageSlug}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	var request CreateSeriesRequest
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	series, tags, serviceErr := c.services.CreateSeries(userCtx, services.CreateSeriesOptions{
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		Title:        request.Title,
		Description:  request.Description,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.Status(fiber.StatusCreated).JSON(NewAdminSeriesResponse(series, tags))
}

func (c *Controllers) GetAdminSingleSeries(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.CreateSeries")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	log.InfoContext(userCtx, "Getting series...", "languageSlug", languageSlug, "seriesSlug", seriesSlug)

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

	seriesDto, serviceErr := c.services.FindSeriesBySlugsWithJoins(userCtx, services.FindSeriesBySlugsWithJoinsOptions{
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		IsPublished:  false,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(NewAdminSeriesFromDto(seriesDto))
}
