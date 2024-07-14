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
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/kiwiscript/kiwiscript_go/paths"
	"github.com/kiwiscript/kiwiscript_go/services"
	"github.com/kiwiscript/kiwiscript_go/utils"
)

func (c *Controllers) CreateLanguage(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.languages.CreateLanguage")
	userCtx := ctx.UserContext()
	var request LanguageRequest
	user, err := c.GetUserClaims(ctx)

	if err != nil || !user.IsAdmin {
		return ctx.Status(fiber.StatusForbidden).JSON(NewRequestError(services.NewForbiddenError()))
	}
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	language, serviceErr := c.services.CreateLanguage(userCtx, services.CreateLanguageOptions{
		UserID: user.ID,
		Name:   utils.CapitalizedFirst(request.Name),
		Icon:   strings.TrimSpace(request.Icon),
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.Status(fiber.StatusCreated).JSON(c.NewLanguageResponse(language))
}

func (c *Controllers) GetLanguage(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.languages.GetLanguage")
	userCtx := ctx.UserContext()
	slug := ctx.Params("languageSlug")
	log.InfoContext(userCtx, "get language", "slug", slug)
	params := LanguageParams{slug}

	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	language, serviceErr := c.services.FindLanguageBySlug(userCtx, slug)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(c.NewLanguageResponse(language))
}

func (c *Controllers) GetLanguages(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.languages.GetLanguages")
	userCtx := ctx.UserContext()
	log.InfoContext(userCtx, "get languages")
	queryParams := GetLanguagesQueryParams{
		Offset: int32(ctx.QueryInt("offset", OffsetDefault)),
		Limit:  int32(ctx.QueryInt("limit", LimitDefault)),
		Search: ctx.Query("search"),
	}

	if err := c.validate.StructCtx(userCtx, queryParams); err != nil {
		return c.validateQueryErrorResponse(log, userCtx, err, ctx)
	}

	languages, count, serviceErr := c.services.FindPaginatedLanguages(userCtx, services.FindPaginatedLanguagesOptions{
		Search: queryParams.Search,
		Offset: queryParams.Offset,
		Limit:  queryParams.Limit,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(NewPaginatedResponse(
		c.backendDomain,
		paths.LanguagePathV1,
		&queryParams,
		count,
		languages,
		c.NewLanguageResponse,
	))
}

func (c *Controllers) UpdateLanguage(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.languages.UpdateLanguage")
	userCtx := ctx.UserContext()
	slug := ctx.Params("languageSlug")
	log.InfoContext(userCtx, "update language", "slug", slug)

	user, err := c.GetUserClaims(ctx)
	if err != nil || !user.IsAdmin {
		return ctx.Status(fiber.StatusForbidden).JSON(NewRequestError(services.NewForbiddenError()))
	}

	params := LanguageParams{slug}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	var request LanguageRequest
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	language, serviceErr := c.services.UpdateLanguage(userCtx, services.UpdateLanguageOptions{
		Slug: slug,
		Name: utils.CapitalizedFirst(request.Name),
		Icon: strings.TrimSpace(request.Icon),
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(c.NewLanguageResponse(language))
}

func (c *Controllers) DeleteLanguage(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.languages.DeleteLanguage")
	userCtx := ctx.UserContext()
	slug := ctx.Params("languageSlug")
	log.InfoContext(userCtx, "delete language", "name", slug)

	user, err := c.GetUserClaims(ctx)
	if err != nil || !user.IsAdmin {
		return ctx.Status(fiber.StatusForbidden).JSON(NewRequestError(services.NewForbiddenError()))
	}

	params := LanguageParams{slug}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	serviceErr := c.services.DeleteLanguage(userCtx, slug)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
