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

const languageProgressLocation string = "languageProgress"

func (c *Controllers) CreateOrUpdateLanguageProgress(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	slug := ctx.Params("languageSlug")
	log := c.buildLogger(ctx, requestID, languageProgressLocation, "CreateOrUpdateLanguageProgress").With(
		"slug", slug,
	)
	log.InfoContext(userCtx, "Creating or updating language progress...")

	user, err := c.GetUserClaims(ctx)
	if err != nil {
		log.ErrorContext(userCtx, "This route is protected should have not reached here")
		return ctx.Status(fiber.StatusUnauthorized).JSON(exceptions.NewRequestError(exceptions.NewUnauthorizedError()))
	}

	params := dtos.LessonPathParams{LanguageSlug: slug}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	language, languageProgress, serviceErr := c.services.CreateOrUpdateLanguageProgress(
		userCtx,
		services.CreateOrUpdateLanguageProgressOptions{
			RequestID:    requestID,
			UserID:       user.ID,
			LanguageSlug: slug,
		},
	)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(dtos.NewLanguageResponse(c.backendDomain, language.ToLanguageModelWithProgress(languageProgress)))
}

func (c *Controllers) ResetLanguageProgress(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	slug := ctx.Params("languageSlug")
	log := c.buildLogger(ctx, requestID, languageProgressLocation, "ResetLanguageProgress").With(
		"slug", slug,
	)
	log.InfoContext(userCtx, "Resetting language progress...")

	user, err := c.GetUserClaims(ctx)
	if err != nil {
		log.ErrorContext(userCtx, "This route is protected should have not reached here")
		return ctx.Status(fiber.StatusUnauthorized).JSON(exceptions.NewRequestError(exceptions.NewUnauthorizedError()))
	}

	params := dtos.LessonPathParams{LanguageSlug: slug}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	opts := services.DeleteLanguageProgressOptions{
		RequestID:    requestID,
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
	}
	if serviceErr := c.services.DeleteLanguageProgress(userCtx, opts); serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
