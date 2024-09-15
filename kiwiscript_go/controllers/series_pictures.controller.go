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
	"github.com/kiwiscript/kiwiscript_go/exceptions"
	"github.com/kiwiscript/kiwiscript_go/services"
)

const seriesPicturesLocation string = "seriesPictures"

func (c *Controllers) UploadSeriesPicture(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	log := c.buildLogger(ctx, requestID, seriesPicturesLocation, "UploadSeriesPicture").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
	)
	log.InfoContext(userCtx, "Uploading series picture...")

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
	}

	params := dtos.SeriesPathParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(exceptions.NewRequestValidationError(
				exceptions.RequestValidationLocationBody,
				[]exceptions.FieldError{{
					Param:   "file",
					Message: exceptions.FieldErrMessageRequired,
				}},
			))
	}

	seriesPicture, serviceErr := c.services.UploadSeriesPicture(userCtx, services.UploadSeriesPictureOptions{
		RequestID:    requestID,
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		FileHeader:   file,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	pictureURL, serviceErr := c.services.FindFileURL(userCtx, services.FindFileURLOptions{
		RequestID: requestID,
		UserID:    user.ID,
		FileID:    seriesPicture.ID,
		FileExt:   seriesPicture.Ext,
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
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	log := c.buildLogger(ctx, requestID, seriesPicturesLocation, "GetSeriesPicture").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
	)
	log.InfoContext(userCtx, "Getting series picture...")

	params := dtos.SeriesPathParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
	}
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
		return c.serviceErrorResponse(exceptions.NewNotFoundError(), ctx)
	}

	seriesPicture, serviceErr := c.services.FindSeriesPictureBySeriesID(
		userCtx,
		services.FindSeriesPictureBySeriesIDOptions{
			RequestID: requestID,
			SeriesID:  series.ID,
		},
	)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	pictureURL, serviceErr := c.services.FindFileURL(userCtx, services.FindFileURLOptions{
		RequestID: requestID,
		UserID:    seriesPicture.AuthorID,
		FileID:    seriesPicture.ID,
		FileExt:   seriesPicture.Ext,
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
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	log := c.buildLogger(ctx, requestID, seriesPicturesLocation, "DeleteSeriesPicture").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
	)
	log.InfoContext(userCtx, "Deleting series picture...")

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
	}

	params := dtos.SeriesPathParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	opts := services.DeletePictureOptions{
		RequestID:    requestID,
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
	}
	if serviceErr := c.services.DeleteSeriesPicture(userCtx, opts); serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
