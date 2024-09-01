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
	"strconv"
)

const lessonArticlesLocation string = "lesson_articles"

func (c *Controllers) CreateLessonArticle(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	sectionID := ctx.Params("sectionID")
	lectureID := ctx.Params("lectureID")
	log := c.buildLogger(ctx, requestID, lessonArticlesLocation, "CreateLessonArticle").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"sectionID", sectionID,
		"lectureID", lectureID,
	)
	log.InfoContext(userCtx, "Creating lecture article...")

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
	}

	params := dtos.LessonPathParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SectionID:    sectionID,
		LessonID:     lectureID,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	parsedSectionID, err := strconv.Atoi(params.SectionID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(exceptions.NewRequestValidationError(
				exceptions.RequestValidationLocationParams,
				[]exceptions.FieldError{{
					Param:   "sectionId",
					Message: exceptions.StrFieldErrMessageNumber,
					Value:   params.SectionID,
				}},
			))
	}

	parsedLessonID, err := strconv.Atoi(params.LessonID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(exceptions.NewRequestValidationError(
				exceptions.RequestValidationLocationParams,
				[]exceptions.FieldError{{
					Param:   "lecturesId",
					Message: exceptions.StrFieldErrMessageNumber,
					Value:   params.SectionID,
				}},
			))
	}

	var request dtos.LessonArticleBody
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	sectionIDi32 := int32(parsedSectionID)
	lectureIDi32 := int32(parsedLessonID)
	article, serviceErr := c.services.CreateLessonArticle(userCtx, services.CreateLessonArticleOptions{
		RequestID:    requestID,
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SectionID:    sectionIDi32,
		LessonID:     lectureIDi32,
		Content:      request.Content,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.
		Status(fiber.StatusCreated).
		JSON(
			dtos.NewLessonArticleResponse(
				c.backendDomain,
				params.LanguageSlug,
				params.SeriesSlug,
				sectionIDi32,
				article,
			),
		)
}

func (c *Controllers) GetLessonArticle(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	sectionID := ctx.Params("sectionID")
	lectureID := ctx.Params("lectureID")
	log := c.buildLogger(ctx, requestID, lessonArticlesLocation, "GetLessonArticle").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"sectionId", sectionID,
		"lectureId", lectureID,
	)
	log.InfoContext(userCtx, "Creating lecture article...")

	params := dtos.LessonPathParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SectionID:    sectionID,
		LessonID:     lectureID,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	parsedSectionID, err := strconv.Atoi(params.SectionID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(exceptions.NewRequestValidationError(
				exceptions.RequestValidationLocationParams,
				[]exceptions.FieldError{{
					Param:   "sectionId",
					Message: exceptions.StrFieldErrMessageNumber,
					Value:   params.SectionID,
				}},
			))
	}

	parsedLessonID, err := strconv.Atoi(params.LessonID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(exceptions.NewRequestValidationError(
				exceptions.RequestValidationLocationParams,
				[]exceptions.FieldError{{
					Param:   "lecturesId",
					Message: exceptions.StrFieldErrMessageNumber,
					Value:   params.SectionID,
				}},
			))
	}

	sectionIDi32 := int32(parsedSectionID)
	lectureIDi32 := int32(parsedLessonID)
	isPublished := false
	if user, serviceErr := c.GetUserClaims(ctx); serviceErr != nil || !user.IsStaff {
		isPublished = true
	}

	article, serviceErr := c.services.FindLessonArticle(userCtx, services.FindLessonArticleOptions{
		RequestID:    requestID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SectionID:    sectionIDi32,
		LessonID:     lectureIDi32,
		IsPublished:  isPublished,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.
		Status(fiber.StatusOK).
		JSON(
			dtos.NewLessonArticleResponse(
				c.backendDomain,
				params.LanguageSlug,
				params.SeriesSlug,
				sectionIDi32,
				article,
			),
		)
}

func (c *Controllers) UpdateLessonArticle(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	sectionID := ctx.Params("sectionID")
	lectureID := ctx.Params("lectureID")
	log := c.buildLogger(ctx, requestID, lessonArticlesLocation, "UpdateLessonArticle").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"sectionID", sectionID,
		"lectureID", lectureID,
	)
	log.InfoContext(userCtx, "Updating lecture article...")

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
	}

	params := dtos.LessonPathParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SectionID:    sectionID,
		LessonID:     lectureID,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	parsedSectionID, err := strconv.Atoi(params.SectionID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(exceptions.NewRequestValidationError(
				exceptions.RequestValidationLocationParams,
				[]exceptions.FieldError{{
					Param:   "sectionId",
					Message: exceptions.StrFieldErrMessageNumber,
					Value:   params.SectionID,
				}},
			))
	}

	parsedLessonID, err := strconv.Atoi(params.LessonID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(exceptions.NewRequestValidationError(
				exceptions.RequestValidationLocationParams,
				[]exceptions.FieldError{{
					Param:   "lecturesId",
					Message: exceptions.StrFieldErrMessageNumber,
					Value:   params.SectionID,
				}},
			))
	}

	var request dtos.LessonArticleBody
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	sectionIDi32 := int32(parsedSectionID)
	lectureIDi32 := int32(parsedLessonID)
	article, serviceErr := c.services.UpdateLessonArticle(userCtx, services.UpdateLessonArticleOptions{
		RequestID:    requestID,
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SectionID:    sectionIDi32,
		LessonID:     lectureIDi32,
		Content:      request.Content,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.
		JSON(
			dtos.NewLessonArticleResponse(
				c.backendDomain,
				params.LanguageSlug,
				params.SeriesSlug,
				sectionIDi32,
				article,
			),
		)
}

func (c *Controllers) DeleteLessonArticle(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	sectionID := ctx.Params("sectionID")
	lectureID := ctx.Params("lectureID")
	log := c.buildLogger(ctx, requestID, lessonArticlesLocation, "DeleteLessonArticle").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"sectionID", sectionID,
		"lectureID", lectureID,
	)
	log.InfoContext(userCtx, "Delete lecture article...")

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
	}

	params := dtos.LessonPathParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SectionID:    sectionID,
		LessonID:     lectureID,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	parsedSectionID, err := strconv.Atoi(params.SectionID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(exceptions.NewRequestValidationError(
				exceptions.RequestValidationLocationParams,
				[]exceptions.FieldError{{
					Param:   "sectionId",
					Message: exceptions.StrFieldErrMessageNumber,
					Value:   params.SectionID,
				}},
			))
	}

	parsedLessonID, err := strconv.Atoi(params.LessonID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(exceptions.NewRequestValidationError(
				exceptions.RequestValidationLocationParams,
				[]exceptions.FieldError{{
					Param:   "lecturesId",
					Message: exceptions.StrFieldErrMessageNumber,
					Value:   params.SectionID,
				}},
			))
	}

	sectionIDi32 := int32(parsedSectionID)
	lectureIDi32 := int32(parsedLessonID)
	opts := services.DeleteLessonArticleOptions{
		RequestID:    requestID,
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SectionID:    sectionIDi32,
		LessonID:     lectureIDi32,
	}
	if serviceErr := c.services.DeleteLessonArticle(userCtx, opts); serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
