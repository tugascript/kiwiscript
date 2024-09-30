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
	"github.com/kiwiscript/kiwiscript_go/exceptions"
	"github.com/kiwiscript/kiwiscript_go/paths"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/kiwiscript/kiwiscript_go/services"
)

const sectionsLocation string = "sections"

func (c *Controllers) CreateSection(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	log := c.buildLogger(ctx, requestID, sectionsLocation, "CreateSection").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
	)
	log.InfoContext(userCtx, "Creating section...")

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

	var request dtos.CreateSectionBody
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	section, serviceErr := c.services.CreateSection(userCtx, services.CreateSectionOptions{
		RequestID:    requestID,
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		Title:        request.Title,
		Description:  request.Description,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.Status(fiber.StatusCreated).JSON(
		dtos.NewSectionResponse(c.backendDomain, section.ToSectionModel()),
	)
}

func (c *Controllers) GetSection(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	sectionID := ctx.Params("sectionID")
	log := c.buildLogger(ctx, requestID, sectionsLocation, "GetSection").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"sectionId", sectionID,
	)
	log.InfoContext(userCtx, "Getting section...")

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
			JSON(exceptions.NewRequestValidationError(exceptions.RequestValidationLocationParams, []exceptions.FieldError{{
				Param:   "sectionId",
				Message: exceptions.StrFieldErrMessageNumber,
				Value:   params.SectionID,
			}}))
	}

	parsedSectionIDi32 := int32(parsedSectionID)
	if user, serviceErr := c.GetUserClaims(ctx); serviceErr == nil {
		if user.IsStaff {
			section, serviceErr := c.services.FindSectionBySlugsAndID(userCtx, services.FindSectionBySlugsAndIDOptions{
				RequestID:    requestID,
				LanguageSlug: params.LanguageSlug,
				SeriesSlug:   params.SeriesSlug,
				SectionID:    parsedSectionIDi32,
			})
			if serviceErr != nil {
				return c.serviceErrorResponse(serviceErr, ctx)
			}

			return ctx.JSON(dtos.NewSectionResponse(c.backendDomain, section.ToSectionModel()))
		}

		section, serviceErr := c.services.FindPublishedSectionBySlugsAndIDWithProgress(
			userCtx,
			services.FindPublishedSectionBySlugsAndIDWithProgressOptions{
				RequestID:    requestID,
				UserID:       user.ID,
				LanguageSlug: params.LanguageSlug,
				SeriesSlug:   params.SeriesSlug,
				SectionID:    parsedSectionIDi32,
			},
		)
		if serviceErr != nil {
			return c.serviceErrorResponse(serviceErr, ctx)
		}

		return ctx.JSON(dtos.NewSectionResponse(c.backendDomain, section.ToSectionModel()))
	}

	section, serviceErr := c.services.FindPublishedSectionBySlugsAndID(userCtx, services.FindSectionBySlugsAndIDOptions{
		RequestID:    requestID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SectionID:    int32(parsedSectionID),
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(dtos.NewSectionResponse(c.backendDomain, section.ToSectionModel()))
}

func (c *Controllers) GetSections(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	log := c.buildLogger(ctx, requestID, sectionsLocation, "GetSections").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
	)
	log.InfoContext(userCtx, "Getting sections...")

	params := dtos.SeriesPathParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
	}
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

	if user, serviceErr := c.GetUserClaims(ctx); serviceErr == nil {
		if user.IsStaff {
			sections, count, serviceErr := c.services.FindPaginatedSectionsBySlugs(
				userCtx,
				services.FindPaginatedSectionsBySlugsOptions{
					RequestID:    requestID,
					LanguageSlug: params.LanguageSlug,
					SeriesSlug:   params.SeriesSlug,
					Offset:       queryParams.Offset,
					Limit:        queryParams.Limit,
				},
			)
			if serviceErr != nil {
				return c.serviceErrorResponse(serviceErr, ctx)
			}

			return ctx.JSON(
				dtos.NewPaginatedResponse(
					c.backendDomain,
					fmt.Sprintf(
						"%s/%s%s/%s%s",
						paths.LanguagePathV1,
						languageSlug,
						paths.SeriesPath,
						seriesSlug,
						paths.SectionsPath,
					),
					queryParams,
					count,
					sections,
					func(s *db.Section) *dtos.SectionResponse {
						return dtos.NewSectionResponse(c.backendDomain, s.ToSectionModel())
					},
				),
			)
		}

		sections, count, serviceErr := c.services.FindPaginatedPublishedSectionsBySlugsWithProgress(
			userCtx,
			services.FindSectionBySlugsAndIDWithProgressOptions{
				RequestID:    requestID,
				UserID:       user.ID,
				LanguageSlug: params.LanguageSlug,
				SeriesSlug:   params.SeriesSlug,
				Limit:        queryParams.Limit,
				Offset:       queryParams.Offset,
			},
		)
		if serviceErr != nil {
			return c.serviceErrorResponse(serviceErr, ctx)
		}

		return ctx.JSON(
			dtos.NewPaginatedResponse(
				c.backendDomain,
				fmt.Sprintf(
					"%s/%s%s/%s%s",
					paths.LanguagePathV1,
					languageSlug,
					paths.SeriesPath,
					seriesSlug,
					paths.SectionsPath,
				),
				queryParams,
				count,
				sections,
				func(s *db.FindPaginatedPublishedSectionsBySlugsWithProgressRow) *dtos.SectionResponse {
					return dtos.NewSectionResponse(c.backendDomain, s.ToSectionModel())
				},
			),
		)
	}

	sections, count, serviceErr := c.services.FindPaginatedPublishedSectionsBySlugs(
		userCtx,
		services.FindPaginatedSectionsBySlugsOptions{
			RequestID:    requestID,
			LanguageSlug: params.LanguageSlug,
			SeriesSlug:   params.SeriesSlug,
			Offset:       queryParams.Offset,
			Limit:        queryParams.Limit,
		},
	)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(
		dtos.NewPaginatedResponse(
			c.backendDomain,
			fmt.Sprintf(
				"%s/%s%s/%s%s",
				paths.LanguagePathV1,
				languageSlug,
				paths.SeriesPath,
				seriesSlug,
				paths.SectionsPath,
			),
			queryParams,
			count,
			sections,
			func(s *db.Section) *dtos.SectionResponse {
				return dtos.NewSectionResponse(c.backendDomain, s.ToSectionModel())
			},
		),
	)
}

func (c *Controllers) UpdateSection(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	sectionID := ctx.Params("sectionID")
	log := c.buildLogger(ctx, requestID, sectionsLocation, "UpdateSection").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"sectionId", sectionID,
	)
	log.InfoContext(userCtx, "Updating series part...")

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
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
			JSON(exceptions.NewRequestValidationError(exceptions.RequestValidationLocationParams, []exceptions.FieldError{{
				Param:   "sectionId",
				Message: exceptions.StrFieldErrMessageNumber,
				Value:   params.SectionID,
			}}))
	}

	var request dtos.UpdateSectionBody
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	section, serviceErr := c.services.UpdateSection(userCtx, services.UpdateSectionOptions{
		RequestID:    requestID,
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SectionID:    int32(parsedSectionID),
		Title:        request.Title,
		Description:  request.Description,
		Position:     request.Position,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(dtos.NewSectionResponse(c.backendDomain, section.ToSectionModel()))
}

func (c *Controllers) UpdateSectionIsPublished(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	sectionID := ctx.Params("sectionID")
	log := c.buildLogger(ctx, requestID, sectionsLocation, "UpdateSectionIsPublished").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"sectionId", sectionID,
	)
	log.InfoContext(userCtx, "Updating section is published...")

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
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
			JSON(exceptions.NewRequestValidationError(exceptions.RequestValidationLocationParams, []exceptions.FieldError{{
				Param:   "sectionId",
				Message: exceptions.StrFieldErrMessageNumber,
				Value:   params.SectionID,
			}}))
	}

	var request dtos.UpdateIsPublishedBody
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	section, serviceErr := c.services.UpdateSectionIsPublished(userCtx, services.UpdateSectionIsPublishedOptions{
		RequestID:    requestID,
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SectionID:    int32(parsedSectionID),
		IsPublished:  request.IsPublished,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(dtos.NewSectionResponse(c.backendDomain, section.ToSectionModel()))
}

func (c *Controllers) DeleteSection(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	sectionID := ctx.Params("sectionID")
	log := c.buildLogger(ctx, requestID, sectionsLocation, "DeleteSection").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"sectionID", sectionID,
	)
	log.InfoContext(userCtx, "Updating section...")

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
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
			JSON(exceptions.NewRequestValidationError(exceptions.RequestValidationLocationParams, []exceptions.FieldError{{
				Param:   "sectionId",
				Message: exceptions.StrFieldErrMessageNumber,
				Value:   params.SectionID,
			}}))
	}

	serviceErr = c.services.DeleteSection(userCtx, services.DeleteSectionOptions{
		RequestID:    requestID,
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SectionID:    int32(parsedSectionID),
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

func (c *Controllers) GetCurrentSection(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	log := c.buildLogger(ctx, requestID, sectionsLocation, "DeleteSection").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
	)
	log.InfoContext(userCtx, "Finding current section")

	user, err := c.GetUserClaims(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(exceptions.NewRequestError(exceptions.NewUnauthorizedError()))
	}
	if user.IsStaff || user.IsAdmin {
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
	}

	params := dtos.SeriesPathParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	section, serviceErr := c.services.FindCurrentSection(userCtx, services.FindCurrentSectionOptions{
		RequestID:    requestID,
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(dtos.NewSectionResponse(c.backendDomain, section.ToSectionModel()))
}
