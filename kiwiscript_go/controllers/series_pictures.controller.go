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
)

func (c *Controllers) UploadSeriesPicture(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series_pictures.UploadLessonFile")
	userCtx := ctx.UserContext()
	params := dtos.SeriesPathParams{
		LanguageSlug: ctx.Params("languageSlug"),
		SeriesSlug:   ctx.Params("seriesSlug"),
	}
	log.InfoContext(userCtx, "Uploading series picture...",
		"languageSlug", params.LanguageSlug,
		"seriesSlug", params.SeriesSlug,
	)

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(NewRequestError(services.NewForbiddenError()))
	}

	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationBody, []FieldError{{
				Param:   "file",
				Message: FieldErrMessageRequired,
			}}))
	}

	seriesPicture, serviceErr := c.services.UploadSeriesPicture(userCtx, services.UploadSeriesPictureOptions{
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		FileHeader:   file,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	pictureURL, serviceErr := c.services.FindFileURL(userCtx, services.FindFileURLOptions{
		UserID:  user.ID,
		FileID:  seriesPicture.ID,
		FileExt: seriesPicture.Ext,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.Status(fiber.StatusCreated).JSON(dtos.NewSeriesPictureResponse(
		c.backendDomain,
		params.LanguageSlug,
		params.SeriesSlug,
		seriesPicture.ToSeriesPictureModel(pictureURL),
	))
}

func (c *Controllers) GetSeriesPicture(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series_pictures.GetSeriesPicture")
	userCtx := ctx.UserContext()
	params := dtos.SeriesPathParams{
		LanguageSlug: ctx.Params("languageSlug"),
		SeriesSlug:   ctx.Params("seriesSlug"),
	}
	log.InfoContext(userCtx, "Getting series picture...",
		"languageSlug", params.LanguageSlug,
		"seriesSlug", params.SeriesSlug,
	)

	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	series, serviceErr := c.services.FindSeriesBySlugs(userCtx, services.FindSeriesBySlugsOptions{
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	if user, serviceErr := c.GetUserClaims(ctx); serviceErr != nil || (!user.IsStaff && !series.IsPublished) {
		return c.serviceErrorResponse(services.NewNotFoundError(), ctx)
	}

	seriesPicture, serviceErr := c.services.FindSeriesPictureBySeriesID(userCtx, series.ID)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	pictureURL, serviceErr := c.services.FindFileURL(userCtx, services.FindFileURLOptions{
		UserID:  seriesPicture.AuthorID,
		FileID:  seriesPicture.ID,
		FileExt: seriesPicture.Ext,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(dtos.NewSeriesPictureResponse(
		c.backendDomain,
		params.LanguageSlug,
		params.SeriesSlug,
		seriesPicture.ToSeriesPictureModel(pictureURL),
	))
}

func (c *Controllers) DeleteSeriesPicture(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series_pictures.DeleteSeriesPicture")
	userCtx := ctx.UserContext()
	params := dtos.SeriesPathParams{
		LanguageSlug: ctx.Params("languageSlug"),
		SeriesSlug:   ctx.Params("seriesSlug"),
	}
	log.InfoContext(userCtx, "Deleting series picture...",
		"languageSlug", params.LanguageSlug,
		"seriesSlug", params.SeriesSlug,
	)

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(NewRequestError(services.NewForbiddenError()))
	}

	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	opts := services.DeletePictureOptions{
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
	}
	if serviceErr := c.services.DeleteSeriesPicture(userCtx, opts); serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
