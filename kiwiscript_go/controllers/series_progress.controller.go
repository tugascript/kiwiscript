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

const seriesProgressLocation string = "series_progress"

func (c *Controllers) CreateOrUpdateSeriesProgress(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	log := c.buildLogger(ctx, requestID, seriesProgressLocation, "CreateOrUpdateSeriesProgress").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
	)
	log.InfoContext(userCtx, "Creating or updating series progress...")

	user, err := c.GetUserClaims(ctx)
	if err != nil {
		log.ErrorContext(userCtx, "This route is protected should have not reached here")
		return ctx.Status(fiber.StatusUnauthorized).JSON(exceptions.NewRequestError(exceptions.NewUnauthorizedError()))
	}

	if user.IsStaff || user.IsAdmin {
		log.WarnContext(userCtx, "Staff or admin user cannot create or update series progress")
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
	}

	params := dtos.SeriesPathParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	series, seriesProgress, created, serviceErr := c.services.CreateOrUpdateSeriesProgress(
		userCtx,
		services.CreateOrUpdateSeriesProgressOptions{
			RequestID:    requestID,
			UserID:       user.ID,
			LanguageSlug: params.LanguageSlug,
			SeriesSlug:   params.SeriesSlug,
		},
	)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	if series.PictureID.Valid && series.PictureExt.Valid {
		fileUrl, serviceErr := c.services.FindFileURL(userCtx, services.FindFileURLOptions{
			UserID:  series.AuthorID,
			FileID:  series.PictureID.Bytes,
			FileExt: series.PictureExt.String,
		})
		if serviceErr != nil {
			return c.serviceErrorResponse(serviceErr, ctx)
		}

		return ctx.JSON(dtos.NewSeriesResponse(
			c.backendDomain,
			series.ToSeriesModelWithProgress(seriesProgress),
			fileUrl,
		))
	}

	if created {
		return ctx.Status(fiber.StatusCreated).JSON(dtos.NewSeriesResponse(
			c.backendDomain,
			series.ToSeriesModelWithProgress(seriesProgress),
			"",
		))
	}

	return ctx.JSON(dtos.NewSeriesResponse(
		c.backendDomain,
		series.ToSeriesModelWithProgress(seriesProgress),
		"",
	))
}

func (c *Controllers) ResetSeriesProgress(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	log := c.buildLogger(ctx, requestID, seriesProgressLocation, "ResetSeriesProgress").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
	)
	log.InfoContext(userCtx, "Resetting series progress...")

	user, err := c.GetUserClaims(ctx)
	if err != nil {
		log.ErrorContext(userCtx, "This route is protected should have not reached here")
		return ctx.Status(fiber.StatusUnauthorized).JSON(exceptions.NewRequestError(exceptions.NewUnauthorizedError()))
	}

	if user.IsStaff || user.IsAdmin {
		log.WarnContext(userCtx, "Staff or admin user is trying to reset series progress")
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
	}

	params := dtos.SeriesPathParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	opts := services.DeleteSeriesProgressOptions{
		RequestID:    requestID,
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
	}
	if serviceErr := c.services.DeleteSeriesProgress(userCtx, opts); serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
