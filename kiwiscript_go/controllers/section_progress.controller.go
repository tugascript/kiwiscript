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
	"github.com/kiwiscript/kiwiscript_go/dtos"
	"github.com/kiwiscript/kiwiscript_go/services"
	"strconv"
)

func (c *Controllers) CreateOrUpdateSectionProgress(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.GetSingleSeries")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	sectionID := ctx.Params("sectionID")
	log.InfoContext(
		userCtx, "Getting series part...",
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"sectionID", sectionID,
	)

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil {
		log.ErrorContext(userCtx, "This route is protected should have not reached here")
		return ctx.Status(fiber.StatusUnauthorized).JSON(NewRequestError(services.NewUnauthorizedError()))
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
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "sectionId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SectionID,
			}}))
	}

	parsedSectionIDi32 := int32(parsedSectionID)
	section, progress, serviceErr := c.services.CreateOrUpdateSectionProgress(
		userCtx,
		services.CreateOrUpdateSectionProgressOptions{
			UserID:       user.ID,
			LanguageSlug: params.LanguageSlug,
			SeriesSlug:   params.SeriesSlug,
			SectionID:    parsedSectionIDi32,
		},
	)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(dtos.NewSectionResponse(c.backendDomain, section.ToSectionModelWithProgress(progress), nil))
}

func (c *Controllers) ResetSectionProgress(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.ResetSectionProgress")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	sectionID := ctx.Params("sectionID")
	log.InfoContext(
		userCtx, "Resetting series part progress",
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"sectionID", sectionID,
	)

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil {
		log.ErrorContext(userCtx, "This route is protected should have not reached here")
		return ctx.Status(fiber.StatusUnauthorized).JSON(NewRequestError(services.NewUnauthorizedError()))
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
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "sectionId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SectionID,
			}}))
	}

	parsedSectionIDi32 := int32(parsedSectionID)
	opts := services.DeleteSectionProgressOptions{
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SectionID:    parsedSectionIDi32,
	}
	if serviceErr := c.services.DeleteSectionProgress(userCtx, opts); serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
