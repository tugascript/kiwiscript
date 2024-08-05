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
	"github.com/kiwiscript/kiwiscript_go/dtos"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/kiwiscript/kiwiscript_go/paths"
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

	params := dtos.LanguagePathParams{LanguageSlug: languageSlug}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	var request dtos.CreateSeriesBody
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	series, serviceErr := c.services.CreateSeries(userCtx, services.CreateSeriesOptions{
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		Title:        request.Title,
		Description:  request.Description,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.Status(fiber.StatusCreated).JSON(
		dtos.NewSeriesResponse(
			c.backendDomain,
			series.ToSeriesModelWithAuthor(user.ID, user.FirstName, user.LastName),
		),
	)
}

func (c *Controllers) GetSingleSeries(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.GetSingleSeries")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	log.InfoContext(userCtx, "Getting series...", "languageSlug", languageSlug, "seriesSlug", seriesSlug)

	params := dtos.SeriesPathParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	if user, err := c.GetUserClaims(ctx); err == nil {
		if user.IsStaff {
			series, serviceErr := c.services.FindSeriesBySlugs(userCtx, services.FindSeriesBySlugsOptions{
				SeriesSlug:   params.SeriesSlug,
				LanguageSlug: params.LanguageSlug,
			})
			if serviceErr != nil {
				return c.serviceErrorResponse(serviceErr, ctx)
			}

			user, serviceErr := c.services.FindUserByID(userCtx, series.AuthorID)
			if serviceErr != nil {
				return c.serviceErrorResponse(serviceErr, ctx)
			}

			return ctx.JSON(
				dtos.NewSeriesResponse(
					c.backendDomain,
					series.ToSeriesModelWithAuthor(user.ID, user.FirstName, user.LastName),
				),
			)
		}

		seriesModel, serviceErr := c.services.FindPublishedSeriesBySlugsWithProgress(
			userCtx,
			services.FindSeriesBySlugsWithProgressOptions{
				UserID:       user.ID,
				LanguageSlug: params.LanguageSlug,
				SeriesSlug:   params.SeriesSlug,
			},
		)
		if serviceErr != nil {
			return c.serviceErrorResponse(serviceErr, ctx)
		}

		return ctx.JSON(dtos.NewSeriesResponse(c.backendDomain, seriesModel))
	}

	series, serviceErr := c.services.FindPublishedSeriesBySlugsWithAuthor(userCtx, services.FindSeriesBySlugsOptions{
		SeriesSlug:   params.SeriesSlug,
		LanguageSlug: params.LanguageSlug,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(dtos.NewSeriesResponse(c.backendDomain, series.ToSeriesModel()))
}

func (c *Controllers) GetPaginatedSeries(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.GetAdminPaginatedSeries")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	log.InfoContext(userCtx, "Getting series...", "languageSlug", languageSlug)

	params := dtos.LanguagePathParams{LanguageSlug: languageSlug}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	queryParams := dtos.SeriesQueryParams{
		Search: ctx.Query("search"),
		Offset: int32(ctx.QueryInt("offset", dtos.OffsetDefault)),
		Limit:  int32(ctx.QueryInt("limit", dtos.LimitDefault)),
		SortBy: ctx.Query("sortBy", "date"),
	}
	if err := c.validate.StructCtx(userCtx, queryParams); err != nil {
		return c.validateQueryErrorResponse(log, userCtx, err, ctx)
	}

	sortBySlug := utils.Lowered(queryParams.SortBy) == "slug"
	paginationPath := fmt.Sprintf("%s/%s%s", paths.LanguagePathV1, languageSlug, paths.SeriesPath)

	var seriesModels []db.SeriesModel
	var count int64
	var serviceErr *services.ServiceError

	if user, err := c.GetUserClaims(ctx); err == nil {
		if user.IsStaff {
			if queryParams.Search != "" {
				seriesModels, count, serviceErr = c.services.FindFilteredSeries(
					userCtx,
					services.FindFilteredSeriesOptions{
						Search:       queryParams.Search,
						LanguageSlug: params.LanguageSlug,
						Offset:       queryParams.Offset,
						Limit:        queryParams.Limit,
						SortBySlug:   sortBySlug,
					},
				)
			} else {
				seriesModels, count, serviceErr = c.services.FindPaginatedSeries(
					userCtx,
					services.FindPaginatedSeriesOptions{
						LanguageSlug: params.LanguageSlug,
						Offset:       queryParams.Offset,
						Limit:        queryParams.Limit,
						SortBySlug:   sortBySlug,
					},
				)
			}
			if serviceErr != nil {
				return c.serviceErrorResponse(serviceErr, ctx)
			}

			return ctx.JSON(
				dtos.NewPaginatedResponse(
					c.backendDomain,
					paginationPath,
					&queryParams,
					count,
					seriesModels,
					func(dto *db.SeriesModel) *dtos.SeriesResponse {
						return dtos.NewSeriesResponse(c.backendDomain, dto)
					},
				),
			)
		}

		if queryParams.Search != "" {
			seriesModels, count, serviceErr = c.services.FindFilteredPublishedSeriesWithProgress(
				userCtx,
				services.FindFilteredSeriesOptions{
					Search:       queryParams.Search,
					LanguageSlug: params.LanguageSlug,
					Offset:       queryParams.Offset,
					Limit:        queryParams.Limit,
					SortBySlug:   sortBySlug,
				},
			)
		} else {
			seriesModels, count, serviceErr = c.services.FindPaginatedPublishedSeriesWithProgress(
				userCtx,
				services.FindPaginatedSeriesOptions{
					LanguageSlug: params.LanguageSlug,
					Offset:       queryParams.Offset,
					Limit:        queryParams.Limit,
					SortBySlug:   sortBySlug,
				},
			)
		}
		if serviceErr != nil {
			return c.serviceErrorResponse(serviceErr, ctx)
		}

		return ctx.JSON(
			dtos.NewPaginatedResponse(
				c.backendDomain,
				paginationPath,
				&queryParams,
				count,
				seriesModels,
				func(dto *db.SeriesModel) *dtos.SeriesResponse {
					return dtos.NewSeriesResponse(c.backendDomain, dto)
				},
			),
		)
	}

	if queryParams.Search != "" {
		seriesModels, count, serviceErr = c.services.FindFilteredPublishedSeries(
			userCtx,
			services.FindFilteredSeriesOptions{
				Search:       queryParams.Search,
				LanguageSlug: params.LanguageSlug,
				Offset:       queryParams.Offset,
				Limit:        queryParams.Limit,
				SortBySlug:   sortBySlug,
			},
		)
	} else {
		seriesModels, count, serviceErr = c.services.FindPaginatedPublishedSeries(
			userCtx,
			services.FindPaginatedSeriesOptions{
				LanguageSlug: params.LanguageSlug,
				Offset:       queryParams.Offset,
				Limit:        queryParams.Limit,
				SortBySlug:   sortBySlug,
			},
		)
	}
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(
		dtos.NewPaginatedResponse(
			c.backendDomain,
			paginationPath,
			&queryParams,
			count,
			seriesModels,
			func(dto *db.SeriesModel) *dtos.SeriesResponse {
				return dtos.NewSeriesResponse(c.backendDomain, dto)
			},
		),
	)
}

func (c *Controllers) UpdateSeries(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.UpdateSeries")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	log.InfoContext(userCtx, "Updating series...", "languageSlug", languageSlug, "seriesSlug", seriesSlug)

	user, err := c.GetUserClaims(ctx)
	if err != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(NewRequestError(services.NewForbiddenError()))
	}

	params := dtos.SeriesPathParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	var request dtos.UpdateSeriesBody
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	series, serviceErr := c.services.UpdateSeries(userCtx, services.UpdateSeriesOptions{
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		Title:        request.Title,
		Description:  request.Description,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(dtos.NewSeriesResponse(c.backendDomain, series.ToSeriesModelWithAuthor(user.ID, user.FirstName, user.LastName)))
}

func (c *Controllers) UpdateSeriesIsPublished(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.UpdateIsPublished")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	log.InfoContext(userCtx, "Updating series published status...", "languageSlug", languageSlug, "seriesSlug", seriesSlug)

	user, err := c.GetUserClaims(ctx)
	if err != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(NewRequestError(services.NewForbiddenError()))
	}

	params := dtos.SeriesPathParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	var request dtos.UpdateIsPublishedBody
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	series, serviceErr := c.services.UpdateSeriesIsPublished(userCtx, services.UpdateSeriesIsPublishedOptions{
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		IsPublished:  request.IsPublished,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(dtos.NewSeriesResponse(c.backendDomain, series.ToSeriesModelWithAuthor(user.ID, user.FirstName, user.LastName)))
}

func (c *Controllers) DeleteSeries(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.DeleteSeries")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	log.InfoContext(userCtx, "Deleting series...", "languageSlug", languageSlug, "seriesSlug", seriesSlug)

	user, err := c.GetUserClaims(ctx)
	if err != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(NewRequestError(services.NewForbiddenError()))
	}

	params := dtos.SeriesPathParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	opts := services.DeleteSeriesOptions{
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
	}
	if serviceErr := c.services.DeleteSeries(userCtx, opts); serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
