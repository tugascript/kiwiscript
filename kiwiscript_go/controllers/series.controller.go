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
	"context"
	"fmt"
	"github.com/kiwiscript/kiwiscript_go/dtos"
	"github.com/kiwiscript/kiwiscript_go/exceptions"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/kiwiscript/kiwiscript_go/paths"
	"github.com/kiwiscript/kiwiscript_go/services"
)

const seriesLocation string = "series"

// TODO: add discovery endpoint

func (c *Controllers) CreateSeries(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	log := c.buildLogger(ctx, requestID, sectionsLocation, "CreateSeries").With(
		"languageSlug", languageSlug,
	)
	log.InfoContext(userCtx, "Creating series...")

	user, err := c.GetUserClaims(ctx)
	if err != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
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
		RequestID:    requestID,
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
			"",
		),
	)
}

func (c *Controllers) GetSingleSeries(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	log := c.buildLogger(ctx, requestID, seriesLocation, "GetSingleSeries").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
	)
	log.InfoContext(userCtx, "Getting single series...")

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
				RequestID:    requestID,
				SeriesSlug:   params.SeriesSlug,
				LanguageSlug: params.LanguageSlug,
			})
			if serviceErr != nil {
				return c.serviceErrorResponse(serviceErr, ctx)
			}

			user, serviceErr := c.services.FindUserByID(userCtx, services.FindUserByIDOptions{
				RequestID: requestID,
				ID:        series.AuthorID,
			})
			if serviceErr != nil {
				return c.serviceErrorResponse(serviceErr, ctx)
			}

			picture, serviceErr := c.services.FindSeriesPictureBySeriesID(
				userCtx,
				services.FindSeriesPictureBySeriesIDOptions{
					RequestID: requestID,
					SeriesID:  series.ID,
				},
			)
			if serviceErr != nil {
				if serviceErr.Code != exceptions.CodeNotFound {
					return c.serviceErrorResponse(serviceErr, ctx)
				}

				return ctx.JSON(
					dtos.NewSeriesResponse(
						c.backendDomain,
						series.ToSeriesModelWithAuthor(user.ID, user.FirstName, user.LastName),
						"",
					),
				)
			}

			fileUrl, serviceErr := c.services.FindFileURL(userCtx, services.FindFileURLOptions{
				RequestID: requestID,
				UserID:    picture.AuthorID,
				FileID:    picture.ID,
				FileExt:   picture.Ext,
			})
			if serviceErr != nil {
				return c.serviceErrorResponse(serviceErr, ctx)
			}

			return ctx.JSON(
				dtos.NewSeriesResponse(
					c.backendDomain,
					series.ToSeriesModelWithAuthorAndPicture(
						user.ID,
						user.FirstName,
						user.LastName,
						picture.ID,
						picture.Ext,
					),
					fileUrl,
				),
			)
		}

		series, serviceErr := c.services.FindPublishedSeriesBySlugsWithProgress(
			userCtx,
			services.FindSeriesBySlugsWithProgressOptions{
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
				RequestID: requestID,
				UserID:    series.AuthorID,
				FileID:    series.PictureID.Bytes,
				FileExt:   series.PictureExt.String,
			})
			if serviceErr != nil {
				return c.serviceErrorResponse(serviceErr, ctx)
			}

			return ctx.JSON(dtos.NewSeriesResponse(c.backendDomain, series.ToSeriesModel(), fileUrl))
		}

		return ctx.JSON(dtos.NewSeriesResponse(c.backendDomain, series.ToSeriesModel(), ""))
	}

	series, serviceErr := c.services.FindPublishedSeriesBySlugsWithAuthor(userCtx, services.FindSeriesBySlugsOptions{
		RequestID:    requestID,
		SeriesSlug:   params.SeriesSlug,
		LanguageSlug: params.LanguageSlug,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	if series.PictureID.Valid && series.PictureExt.Valid {
		fileUrl, serviceErr := c.services.FindFileURL(userCtx, services.FindFileURLOptions{
			RequestID: requestID,
			UserID:    series.AuthorID,
			FileID:    series.PictureID.Bytes,
			FileExt:   series.PictureExt.String,
		})
		if serviceErr != nil {
			return c.serviceErrorResponse(serviceErr, ctx)
		}

		return ctx.JSON(dtos.NewSeriesResponse(c.backendDomain, series.ToSeriesModel(), fileUrl))
	}

	return ctx.JSON(dtos.NewSeriesResponse(c.backendDomain, series.ToSeriesModel(), ""))
}

func (c *Controllers) findSeriesPictureURLs(
	userCtx context.Context,
	models []db.SeriesModel,
) (*services.FileURLsContainer, *exceptions.ServiceError) {
	optsList := make([]services.FindFileURLOptions, 0)

	for _, m := range models {
		if m.Picture != nil {
			optsList = append(optsList, services.FindFileURLOptions{
				UserID:  m.ID,
				FileID:  m.Picture.ID,
				FileExt: m.Picture.EXT,
			})
		}
	}
	if len(optsList) == 0 {
		return nil, nil
	}

	fileURLs, serviceErr := c.services.FindFileURLs(userCtx, optsList)
	if serviceErr != nil {
		return nil, serviceErr
	}

	return fileURLs, nil
}

func mapSeriesResponse(
	backendDomain string,
	model *db.SeriesModel,
	fileURLs *services.FileURLsContainer,
) *dtos.SeriesResponse {
	if model.Picture != nil {
		if url, ok := fileURLs.Get(model.Picture.ID); ok {
			return dtos.NewSeriesResponse(backendDomain, model, url)
		}
	}

	return dtos.NewSeriesResponse(backendDomain, model, "")
}

func (c *Controllers) GetPaginatedSeries(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	log := c.buildLogger(ctx, requestID, seriesLocation, "GetPaginatedSeries").With(
		"languageSlug", languageSlug,
	)
	log.InfoContext(userCtx, "Getting paginated series...")

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
	var serviceErr *exceptions.ServiceError

	if user, err := c.GetUserClaims(ctx); err == nil {
		if user.IsStaff {
			if queryParams.Search != "" {
				seriesModels, count, serviceErr = c.services.FindFilteredSeries(
					userCtx,
					services.FindFilteredSeriesOptions{
						RequestID:    requestID,
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
						RequestID:    requestID,
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

			fileURLs, serviceErr := c.findSeriesPictureURLs(userCtx, seriesModels)
			if serviceErr != nil {
				return c.serviceErrorResponse(serviceErr, ctx)
			}

			if fileURLs == nil {
				return ctx.JSON(
					dtos.NewPaginatedResponse(
						c.backendDomain,
						paginationPath,
						&queryParams,
						count,
						seriesModels,
						func(dto *db.SeriesModel) *dtos.SeriesResponse {
							return dtos.NewSeriesResponse(c.backendDomain, dto, "")
						},
					),
				)
			}

			return ctx.JSON(
				dtos.NewPaginatedResponse(
					c.backendDomain,
					paginationPath,
					&queryParams,
					count,
					seriesModels,
					func(model *db.SeriesModel) *dtos.SeriesResponse {
						return mapSeriesResponse(c.backendDomain, model, fileURLs)
					},
				),
			)
		}

		if queryParams.Search != "" {
			seriesModels, count, serviceErr = c.services.FindFilteredPublishedSeriesWithProgress(
				userCtx,
				services.FindFilteredPublishedSeriesWithProgressOptions{
					RequestID:    requestID,
					UserID:       user.ID,
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
				services.FindPaginatedSeriesWithProgressOptions{
					RequestID:    requestID,
					UserID:       user.ID,
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

		fileURLs, serviceErr := c.findSeriesPictureURLs(userCtx, seriesModels)
		if serviceErr != nil {
			return c.serviceErrorResponse(serviceErr, ctx)
		}

		if fileURLs == nil {
			return ctx.JSON(
				dtos.NewPaginatedResponse(
					c.backendDomain,
					paginationPath,
					&queryParams,
					count,
					seriesModels,
					func(dto *db.SeriesModel) *dtos.SeriesResponse {
						return dtos.NewSeriesResponse(c.backendDomain, dto, "")
					},
				),
			)
		}

		return ctx.JSON(
			dtos.NewPaginatedResponse(
				c.backendDomain,
				paginationPath,
				&queryParams,
				count,
				seriesModels,
				func(model *db.SeriesModel) *dtos.SeriesResponse {
					return mapSeriesResponse(c.backendDomain, model, fileURLs)
				},
			),
		)
	}

	if queryParams.Search != "" {
		seriesModels, count, serviceErr = c.services.FindFilteredPublishedSeries(
			userCtx,
			services.FindFilteredSeriesOptions{
				RequestID:    requestID,
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

	fileURLs, serviceErr := c.findSeriesPictureURLs(userCtx, seriesModels)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	if fileURLs == nil {
		return ctx.JSON(
			dtos.NewPaginatedResponse(
				c.backendDomain,
				paginationPath,
				&queryParams,
				count,
				seriesModels,
				func(dto *db.SeriesModel) *dtos.SeriesResponse {
					return dtos.NewSeriesResponse(c.backendDomain, dto, "")
				},
			),
		)
	}

	return ctx.JSON(
		dtos.NewPaginatedResponse(
			c.backendDomain,
			paginationPath,
			&queryParams,
			count,
			seriesModels,
			func(model *db.SeriesModel) *dtos.SeriesResponse {
				return mapSeriesResponse(c.backendDomain, model, fileURLs)
			},
		),
	)
}

func (c *Controllers) UpdateSeries(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	log := c.buildLogger(ctx, requestID, seriesLocation, "UpdateSeries").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
	)
	log.InfoContext(userCtx, "Updating series...")

	user, err := c.GetUserClaims(ctx)
	if err != nil || !user.IsStaff {
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

	picture, serviceErr := c.services.FindSeriesPictureBySeriesID(userCtx, services.FindSeriesPictureBySeriesIDOptions{
		RequestID: requestID,
		SeriesID:  series.ID,
	})
	if serviceErr != nil {
		if serviceErr.Code != exceptions.CodeNotFound {
			return c.serviceErrorResponse(serviceErr, ctx)
		}

		return ctx.JSON(
			dtos.NewSeriesResponse(
				c.backendDomain,
				series.ToSeriesModelWithAuthor(user.ID, user.FirstName, user.LastName),
				"",
			),
		)
	}

	fileUrl, serviceErr := c.services.FindFileURL(userCtx, services.FindFileURLOptions{
		UserID:  picture.AuthorID,
		FileID:  picture.ID,
		FileExt: picture.Ext,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(
		dtos.NewSeriesResponse(
			c.backendDomain,
			series.ToSeriesModelWithAuthorAndPicture(
				user.ID,
				user.FirstName,
				user.LastName,
				picture.ID,
				picture.Ext,
			),
			fileUrl,
		),
	)
}

func (c *Controllers) UpdateSeriesIsPublished(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	log := c.buildLogger(ctx, requestID, seriesLocation, "UpdateSeriesIsPublished").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
	)
	log.InfoContext(userCtx, "Updating series published status...")

	user, err := c.GetUserClaims(ctx)
	if err != nil || !user.IsStaff {
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

	picture, serviceErr := c.services.FindSeriesPictureBySeriesID(userCtx, services.FindSeriesPictureBySeriesIDOptions{
		RequestID: requestID,
		SeriesID:  series.ID,
	})
	if serviceErr != nil {
		if serviceErr.Code != exceptions.CodeNotFound {
			return c.serviceErrorResponse(serviceErr, ctx)
		}

		return ctx.JSON(
			dtos.NewSeriesResponse(
				c.backendDomain,
				series.ToSeriesModelWithAuthor(user.ID, user.FirstName, user.LastName),
				"",
			),
		)
	}

	fileUrl, serviceErr := c.services.FindFileURL(userCtx, services.FindFileURLOptions{
		UserID:  picture.AuthorID,
		FileID:  picture.ID,
		FileExt: picture.Ext,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(
		dtos.NewSeriesResponse(
			c.backendDomain,
			series.ToSeriesModelWithAuthorAndPicture(
				user.ID,
				user.FirstName,
				user.LastName,
				picture.ID,
				picture.Ext,
			),
			fileUrl,
		),
	)
}

func (c *Controllers) DeleteSeries(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	log := c.buildLogger(ctx, requestID, seriesLocation, "DeleteSeries").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
	)
	log.InfoContext(userCtx, "Deleting series...")

	user, err := c.GetUserClaims(ctx)
	if err != nil || !user.IsStaff {
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

	opts := services.DeleteSeriesOptions{
		RequestID:    requestID,
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
	}
	if serviceErr := c.services.DeleteSeries(userCtx, opts); serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

func (c *Controllers) GetPaginatedViewedSeries(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	log := c.buildLogger(ctx, requestID, seriesLocation, "GetPaginatedViewedSeries").With(
		"languageSlug", languageSlug,
	)
	log.InfoContext(userCtx, "Getting paginated viewed series...")

	user, err := c.GetUserClaims(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(exceptions.NewRequestError(exceptions.NewUnauthorizedError()))
	}
	if user.IsStaff || user.IsAdmin {
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
	}

	params := dtos.LanguagePathParams{LanguageSlug: languageSlug}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	queryParams := dtos.PaginationQueryParams{
		Offset: int32(ctx.QueryInt("offset", dtos.OffsetDefault)),
		Limit:  int32(ctx.QueryInt("limit", dtos.LimitDefault)),
	}
	if err := c.validate.StructCtx(userCtx, queryParams); err != nil {
		return c.validateQueryErrorResponse(log, userCtx, err, ctx)
	}

	seriesModels, count, serviceErr := c.services.FindPaginatedViewedSeriesWithProgress(
		userCtx,
		services.FindPaginatedViewedSeriesWithProgressOptions{
			RequestID:    requestID,
			UserID:       user.ID,
			LanguageSlug: params.LanguageSlug,
			Offset:       queryParams.Offset,
			Limit:        queryParams.Limit,
		},
	)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	fileURLs, serviceErr := c.findSeriesPictureURLs(userCtx, seriesModels)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	paginationPath := fmt.Sprintf("%s/%s%s%s", paths.LanguagePathV1, languageSlug, paths.SeriesPath, paths.ProgressPath)
	if fileURLs == nil {
		return ctx.JSON(
			dtos.NewPaginatedResponse(
				c.backendDomain,
				paginationPath,
				&queryParams,
				count,
				seriesModels,
				func(dto *db.SeriesModel) *dtos.SeriesResponse {
					return dtos.NewSeriesResponse(c.backendDomain, dto, "")
				},
			),
		)
	}

	return ctx.JSON(
		dtos.NewPaginatedResponse(
			c.backendDomain,
			paginationPath,
			&queryParams,
			count,
			seriesModels,
			func(model *db.SeriesModel) *dtos.SeriesResponse {
				return mapSeriesResponse(c.backendDomain, model, fileURLs)
			},
		),
	)
}

func (c *Controllers) GetDiscoverySeries(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, seriesLocation, "GetDiscoverySeries")
	log.InfoContext(userCtx, "Getting paginated published series for discovery...")

	queryParams := dtos.DiscoverySeriesQueryParams{
		Search: ctx.Query("search"),
		Offset: int32(ctx.QueryInt("offset", dtos.OffsetDefault)),
		Limit:  int32(ctx.QueryInt("limit", dtos.LimitDefault)),
	}
	if err := c.validate.StructCtx(userCtx, queryParams); err != nil {
		return c.validateQueryErrorResponse(log, userCtx, err, ctx)
	}

	var seriesModels []db.SeriesModel
	var count int64
	var serviceErr *exceptions.ServiceError

	if user, err := c.GetUserClaims(ctx); err == nil && !user.IsStaff {
		if queryParams.Search != "" {
			seriesModels, count, serviceErr = c.services.FindFilteredDiscoverySeriesWithProgress(
				userCtx,
				services.FindFilteredDiscoverySeriesWithProgressOptions{
					RequestID: requestID,
					UserID:    user.ID,
					Search:    queryParams.Search,
					Offset:    queryParams.Offset,
					Limit:     queryParams.Limit,
				},
			)
		} else {
			seriesModels, count, serviceErr = c.services.FindPaginatedDiscoverySeriesWithProgress(
				userCtx,
				services.FindPaginatedDiscoverySeriesWithProgressOptions{
					RequestID: requestID,
					UserID:    user.ID,
					Offset:    queryParams.Offset,
					Limit:     queryParams.Limit,
				},
			)
		}
	} else {
		if queryParams.Search != "" {
			seriesModels, count, serviceErr = c.services.FindFilteredDiscoverySeries(
				userCtx,
				services.FindFilteredDiscoverySeriesOptions{
					RequestID: requestID,
					Search:    queryParams.Search,
					Offset:    queryParams.Offset,
					Limit:     queryParams.Limit,
				},
			)
		} else {
			seriesModels, count, serviceErr = c.services.FindPaginatedDiscoverySeries(
				userCtx,
				services.FindPaginatedDiscoverySeriesOptions{
					RequestID: requestID,
					Offset:    queryParams.Offset,
					Limit:     queryParams.Limit,
				},
			)
		}
	}
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	fileURLs, serviceErr := c.findSeriesPictureURLs(userCtx, seriesModels)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	if fileURLs == nil {
		return ctx.JSON(
			dtos.NewPaginatedResponse(
				c.backendDomain,
				paths.DiscoverV1,
				&queryParams,
				count,
				seriesModels,
				func(dto *db.SeriesModel) *dtos.SeriesResponse {
					return dtos.NewSeriesResponse(c.backendDomain, dto, "")
				},
			),
		)
	}

	return ctx.JSON(
		dtos.NewPaginatedResponse(
			c.backendDomain,
			paths.DiscoverV1,
			&queryParams,
			count,
			seriesModels,
			func(model *db.SeriesModel) *dtos.SeriesResponse {
				return mapSeriesResponse(c.backendDomain, model, fileURLs)
			},
		),
	)
}
