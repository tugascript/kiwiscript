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
	"strconv"
)

const lessonProgressLocation string = "lesson_progress"

func (c *Controllers) CreateOrUpdateLessonProgress(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	sectionID := ctx.Params("sectionID")
	lessonID := ctx.Params("lessonID")
	log := c.buildLogger(ctx, requestID, languageProgressLocation, "CreateOrUpdateLessonProgress").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"sectionId", sectionID,
		"lessonId", lessonID,
	)
	log.InfoContext(userCtx, "Create or update lesson progress...")

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil {
		log.ErrorContext(userCtx, "This route is protected should have not reached here")
		return ctx.Status(fiber.StatusUnauthorized).JSON(NewRequestError(services.NewUnauthorizedError()))
	}

	params := dtos.LessonPathParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SectionID:    sectionID,
		LessonID:     lessonID,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	parsedSectionID, err := strconv.Atoi(params.SectionID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "sectionId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SectionID,
			}}))
	}

	parsedLessonID, err := strconv.Atoi(params.LessonID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "lessonId",
				Message: StrFieldErrMessageNumber,
				Value:   params.LessonID,
			}}))
	}

	parsedSectionIDi32 := int32(parsedSectionID)
	parsedLessonIDi32 := int32(parsedLessonID)
	lesson, progress, serviceErr := c.services.CreateOrUpdateLessonProgress(
		userCtx,
		services.CreateOrUpdateLessonProgressOptions{
			RequestID:    requestID,
			UserID:       user.ID,
			LanguageSlug: params.LanguageSlug,
			SeriesSlug:   params.SeriesSlug,
			SectionID:    parsedSectionIDi32,
			LessonID:     parsedLessonIDi32,
		},
	)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(dtos.NewLessonResponse(c.backendDomain, lesson.ToLessonModelWithProgress(progress)))
}

func (c *Controllers) CompleteLessonProgress(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	sectionID := ctx.Params("sectionID")
	lessonID := ctx.Params("lessonID")
	log := c.buildLogger(ctx, requestID, languageProgressLocation, "CompleteLessonProgress").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"sectionId", sectionID,
		"lessonId", lessonID,
	)
	log.InfoContext(userCtx, "Completing lesson progress...")

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil {
		log.ErrorContext(userCtx, "This route is protected should have not reached here")
		return ctx.Status(fiber.StatusUnauthorized).JSON(NewRequestError(services.NewUnauthorizedError()))
	}

	params := dtos.LessonPathParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SectionID:    sectionID,
		LessonID:     lessonID,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	parsedSectionID, err := strconv.Atoi(params.SectionID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "sectionId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SectionID,
			}}))
	}

	parsedLessonID, err := strconv.Atoi(params.LessonID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "lessonId",
				Message: StrFieldErrMessageNumber,
				Value:   params.LessonID,
			}}))
	}

	parsedSectionIDi32 := int32(parsedSectionID)
	parsedLessonIDi32 := int32(parsedLessonID)
	lesson, progress, certificate, serviceErr := c.services.CompleteLessonProgress(
		userCtx,
		services.CompleteLessonProgressOptions{
			RequestID:    requestID,
			UserID:       user.ID,
			LanguageSlug: params.LanguageSlug,
			SeriesSlug:   params.SeriesSlug,
			SectionID:    parsedSectionIDi32,
			LessonID:     parsedLessonIDi32,
		},
	)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	if certificate != nil {
		return ctx.JSON(
			dtos.NewLessonResponseWithCertificate(
				c.backendDomain,
				lesson.ToLessonModelWithProgress(progress),
				certificate,
			),
		)
	}

	return ctx.JSON(dtos.NewLessonResponse(c.backendDomain, lesson.ToLessonModelWithProgress(progress)))
}

func (c *Controllers) ResetLessonProgress(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	sectionID := ctx.Params("sectionID")
	lessonID := ctx.Params("lessonID")
	log := c.buildLogger(ctx, requestID, languageProgressLocation, "CompleteLessonProgress").With(
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"sectionId", sectionID,
		"lessonId", lessonID,
	)
	log.InfoContext(userCtx, "Resetting lesson progress...")

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil {
		log.ErrorContext(userCtx, "This route is protected should have not reached here")
		return ctx.Status(fiber.StatusUnauthorized).JSON(NewRequestError(services.NewUnauthorizedError()))
	}

	params := dtos.LessonPathParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SectionID:    sectionID,
		LessonID:     lessonID,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	parsedSectionID, err := strconv.Atoi(params.SectionID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "sectionId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SectionID,
			}}))
	}

	parsedLessonID, err := strconv.Atoi(params.LessonID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "lessonId",
				Message: StrFieldErrMessageNumber,
				Value:   params.LessonID,
			}}))
	}

	parsedSectionIDi32 := int32(parsedSectionID)
	parsedLessonIDi32 := int32(parsedLessonID)
	opts := services.DeleteLessonProgressOptions{
		RequestID:    requestID,
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SectionID:    parsedSectionIDi32,
		LessonID:     parsedLessonIDi32,
	}
	if serviceErr := c.services.DeleteLessonProgress(userCtx, opts); serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
