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
	"github.com/kiwiscript/kiwiscript_go/dtos"
	"github.com/kiwiscript/kiwiscript_go/exceptions"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/kiwiscript/kiwiscript_go/paths"
	"github.com/kiwiscript/kiwiscript_go/services"
	"github.com/kiwiscript/kiwiscript_go/utils"
)

const languagesLocation string = "languages"

func (c *Controllers) CreateLanguage(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, languagesLocation, "CreateLanguage")
	log.InfoContext(userCtx, "Creating language...")

	user, err := c.GetUserClaims(ctx)
	if err != nil || !user.IsAdmin {
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
	}

	var request dtos.LanguageBody
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

	return ctx.Status(fiber.StatusCreated).JSON(dtos.NewLanguageResponse(c.backendDomain, language.ToLanguageModel()))
}

func (c *Controllers) GetLanguage(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	slug := ctx.Params("languageSlug")
	log := c.buildLogger(ctx, requestID, languagesLocation, "GetLanguage").With("slug", slug)
	log.InfoContext(userCtx, "Getting language...")

	params := dtos.LanguagePathParams{LanguageSlug: slug}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	if user, err := c.GetUserClaims(ctx); err == nil {
		language, serviceErr := c.services.FindLanguageWithProgressBySlug(
			userCtx,
			services.FindLanguageWithProgressBySlugOptions{
				RequestID:    requestID,
				UserID:       user.ID,
				LanguageSlug: slug,
			},
		)
		if serviceErr != nil {
			return c.serviceErrorResponse(serviceErr, ctx)
		}

		return ctx.JSON(dtos.NewLanguageResponse(c.backendDomain, language.ToLanguageModel()))
	}

	language, serviceErr := c.services.FindLanguageBySlug(userCtx, slug)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(dtos.NewLanguageResponse(c.backendDomain, language.ToLanguageModel()))
}

func (c *Controllers) GetLanguages(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, languagesLocation, "GetLanguages")
	log.InfoContext(userCtx, "Getting languages...")

	queryParams := dtos.GetLanguagesQueryParams{
		Offset: int32(ctx.QueryInt("offset", dtos.OffsetDefault)),
		Limit:  int32(ctx.QueryInt("limit", dtos.LimitDefault)),
		Search: ctx.Query("search"),
	}
	if err := c.validate.StructCtx(userCtx, queryParams); err != nil {
		return c.validateQueryErrorResponse(log, userCtx, err, ctx)
	}

	if user, err := c.GetUserClaims(ctx); err == nil {
		if queryParams.Search != "" {
			languages, count, serviceErr := c.services.FindFilteredPaginatedLanguagesWithProgress(
				userCtx,
				services.FindFilteredPaginatedLanguagesWithProgressOptions{
					RequestID: requestID,
					UserID:    user.ID,
					Search:    queryParams.Search,
					Offset:    queryParams.Offset,
					Limit:     queryParams.Limit,
				},
			)
			if serviceErr != nil {
				return c.serviceErrorResponse(serviceErr, ctx)
			}

			return ctx.JSON(dtos.NewPaginatedResponse(
				c.backendDomain,
				paths.LanguagePathV1,
				&queryParams,
				count,
				languages,
				func(l *db.FindFilteredPaginatedLanguagesWithLanguageProgressRow) *dtos.LanguageResponse {
					return dtos.NewLanguageResponse(c.backendDomain, l.ToLanguageModel())
				},
			))
		}

		languages, count, serviceErr := c.services.FindPaginatedLanguagesWithProgress(
			userCtx,
			services.FindPaginatedLanguagesWithProgressOptions{
				RequestID: requestID,
				UserID:    user.ID,
				Offset:    queryParams.Offset,
				Limit:     queryParams.Limit,
			},
		)
		if serviceErr != nil {
			return c.serviceErrorResponse(serviceErr, ctx)
		}
		return ctx.JSON(dtos.NewPaginatedResponse(
			c.backendDomain,
			paths.LanguagePathV1,
			&queryParams,
			count,
			languages,
			func(l *db.FindPaginatedLanguagesWithLanguageProgressRow) *dtos.LanguageResponse {
				return dtos.NewLanguageResponse(c.backendDomain, l.ToLanguageModel())
			},
		))
	}

	languages, count, serviceErr := c.services.FindPaginatedLanguages(userCtx, services.FindPaginatedLanguagesOptions{
		RequestID: requestID,
		Search:    queryParams.Search,
		Offset:    queryParams.Offset,
		Limit:     queryParams.Limit,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(dtos.NewPaginatedResponse(
		c.backendDomain,
		paths.LanguagePathV1,
		&queryParams,
		count,
		languages,
		func(l *db.Language) *dtos.LanguageResponse {
			return dtos.NewLanguageResponse(c.backendDomain, l.ToLanguageModel())
		},
	))
}

func (c *Controllers) UpdateLanguage(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	slug := ctx.Params("languageSlug")
	log := c.buildLogger(ctx, requestID, languagesLocation, "UpdateLanguage").With(
		"slug", slug,
	)
	log.InfoContext(userCtx, "Updating languages...")

	user, err := c.GetUserClaims(ctx)
	if err != nil || !user.IsAdmin {
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
	}

	params := dtos.LanguagePathParams{LanguageSlug: slug}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	var request dtos.LanguageBody
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	language, serviceErr := c.services.UpdateLanguage(userCtx, services.UpdateLanguageOptions{
		RequestID: requestID,
		Slug:      slug,
		Name:      utils.CapitalizedFirst(request.Name),
		Icon:      strings.TrimSpace(request.Icon),
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(dtos.NewLanguageResponse(c.backendDomain, language.ToLanguageModel()))
}

func (c *Controllers) DeleteLanguage(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	slug := ctx.Params("languageSlug")
	log := c.buildLogger(ctx, requestID, languagesLocation, "DeleteLanguage").With("slug", slug)
	log.InfoContext(userCtx, "Deleting language...")

	user, err := c.GetUserClaims(ctx)
	if err != nil || !user.IsAdmin {
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
	}

	params := dtos.LanguagePathParams{LanguageSlug: slug}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	serviceErr := c.services.DeleteLanguage(userCtx, services.DeleteLanguageOptions{
		RequestID: requestID,
		Slug:      params.LanguageSlug,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

func (c *Controllers) GetViewedLanguages(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, languagesLocation, "GetViewedLanguages")
	log.InfoContext(userCtx, "Getting viewed languages...")

	user, err := c.GetUserClaims(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(exceptions.NewRequestError(exceptions.NewUnauthorizedError()))
	}
	if user.IsStaff || user.IsAdmin {
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
	}

	queryParams := dtos.PaginationQueryParams{
		Offset: int32(ctx.QueryInt("offset", dtos.OffsetDefault)),
		Limit:  int32(ctx.QueryInt("limit", dtos.LimitDefault)),
	}
	if err := c.validate.StructCtx(userCtx, queryParams); err != nil {
		return c.validateQueryErrorResponse(log, userCtx, err, ctx)
	}

	languages, count, serviceErr := c.services.FindPaginatedViewedLanguagesWithProgress(
		userCtx,
		services.FindPaginatedLanguagesWithProgressOptions{
			RequestID: requestID,
			UserID:    user.ID,
			Offset:    queryParams.Offset,
			Limit:     queryParams.Limit,
		},
	)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(dtos.NewPaginatedResponse(
		c.backendDomain,
		paths.LanguagePathV1,
		&queryParams,
		count,
		languages,
		func(l *db.FindPaginatedLanguagesWithInnerProgressRow) *dtos.LanguageResponse {
			return dtos.NewLanguageResponse(c.backendDomain, l.ToLanguageModel())
		},
	))
}
