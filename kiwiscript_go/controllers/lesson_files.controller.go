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
	"github.com/google/uuid"
	"github.com/kiwiscript/kiwiscript_go/dtos"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/services"
	"strconv"
)

func (c *Controllers) UploadLessonFile(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.lesson_files.UploadLessonFile")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	sectionID := ctx.Params("sectionID")
	lessonID := ctx.Params("lessonID")
	name := ctx.FormValue("name")
	log.InfoContext(
		userCtx,
		"Uploading lesson file...",
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"sectionID", sectionID,
		"lessonID", lessonID,
		"name", name,
	)

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(NewRequestError(services.NewForbiddenError()))
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationBody, []FieldError{{
				Param:   "file",
				Message: FieldErrMessageRequired,
			}}))
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
				Param:   "lessonsId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SectionID,
			}}))
	}

	request := dtos.LessonFileBody{Name: name}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	sectionIDi32 := int32(parsedSectionID)
	lessonIDi32 := int32(parsedLessonID)
	lessonFile, serviceErr := c.services.UploadLessonFile(userCtx, services.UploadLessonFileOptions{
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SectionID:    sectionIDi32,
		LessonID:     lessonIDi32,
		Name:         request.Name,
		FileHeader:   file,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	fileUrl, serviceErr := c.services.FindFileURL(userCtx, services.FindFileURLOptions{
		UserID:  user.ID,
		FileID:  lessonFile.ID,
		FileExt: lessonFile.Ext,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.Status(fiber.StatusCreated).JSON(
		dtos.NewLessonFileResponse(
			c.backendDomain,
			params.LanguageSlug,
			params.SeriesSlug,
			sectionIDi32,
			lessonFile.ToLessonFileModel(fileUrl),
		),
	)
}

func (c *Controllers) GetLessonFiles(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.lesson_files.GetLessonFiles")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	sectionID := ctx.Params("sectionID")
	lessonID := ctx.Params("lessonID")
	log.InfoContext(
		userCtx,
		"Getting lesson files...",
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"sectionID", sectionID,
		"lessonID", lessonID,
	)

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
				Param:   "lessonsId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SectionID,
			}}))
	}

	sectionIDi32 := int32(parsedSectionID)
	lessonIDi32 := int32(parsedLessonID)
	isPublished := false
	if user, serviceErr := c.GetUserClaims(ctx); serviceErr != nil || !user.IsStaff {
		isPublished = true
	}

	lessonFiles, serviceErr := c.services.FindLessonFiles(userCtx, services.FindLessonFilesOptions{
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SectionID:    sectionIDi32,
		LessonID:     lessonIDi32,
		IsPublished:  isPublished,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	filesLen := len(lessonFiles)
	if filesLen == 0 {
		return ctx.JSON(make([]db.LessonFileModel, 0))
	}

	optsList := make([]services.FindFileURLOptions, 0, filesLen)
	for _, lessonFile := range lessonFiles {
		optsList = append(optsList, services.FindFileURLOptions{
			UserID:  lessonFile.AuthorID,
			FileID:  lessonFile.ID,
			FileExt: lessonFile.Ext,
		})
	}

	fileUrls, serviceErr := c.services.FindFileURLs(userCtx, optsList)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	responses := make([]dtos.LessonFileResponse, 0, len(lessonFiles))
	for _, lessonFile := range lessonFiles {
		url, ok := fileUrls.Get(lessonFile.ID)
		if !ok {
			log.WarnContext(userCtx, "Could not find URL for lesson file", "fileId", lessonFile.ID)
			continue
		}
		responses = append(
			responses,
			*dtos.NewLessonFileResponse(
				c.backendDomain,
				params.LanguageSlug,
				params.SeriesSlug,
				sectionIDi32,
				lessonFile.ToLessonFileModel(url),
			),
		)
	}

	return ctx.JSON(responses)
}

func (c *Controllers) GetLessonFile(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.lesson_files.GetLessonFile")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	sectionID := ctx.Params("sectionID")
	lessonID := ctx.Params("lessonID")
	fileID := ctx.Params("fileID")
	log.InfoContext(
		userCtx,
		"Deleting lesson file...",
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"sectionID", sectionID,
		"lessonID", lessonID,
		"fileID", fileID,
	)

	params := dtos.LessonFilePathParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SectionID:    sectionID,
		LessonID:     lessonID,
		FileID:       fileID,
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
				Param:   "lessonsId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SectionID,
			}}))
	}

	parsedFileID, err := uuid.FromBytes([]byte(params.FileID))
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "fileId",
				Message: StrFieldErrMessageUUID,
				Value:   params.FileID,
			}}))
	}

	sectionIDi32 := int32(parsedSectionID)
	lessonIDi32 := int32(parsedLessonID)
	isPublished := false
	if user, serviceErr := c.GetUserClaims(ctx); serviceErr != nil || !user.IsStaff {
		isPublished = true
	}

	lessonFile, serviceErr := c.services.FindLessonFile(userCtx, services.FindLessonFileOptions{
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SectionID:    sectionIDi32,
		LessonID:     lessonIDi32,
		File:         parsedFileID,
		IsPublished:  isPublished,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	fileUrl, serviceErr := c.services.FindFileURL(userCtx, services.FindFileURLOptions{
		UserID:  lessonFile.AuthorID,
		FileID:  lessonFile.ID,
		FileExt: lessonFile.Ext,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(
		dtos.NewLessonFileResponse(
			c.backendDomain,
			params.LanguageSlug,
			params.SeriesSlug,
			sectionIDi32,
			lessonFile.ToLessonFileModel(fileUrl),
		),
	)
}

func (c *Controllers) DeleteLessonFile(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.lesson_files.DeleteLessonFile")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	sectionID := ctx.Params("sectionID")
	lessonID := ctx.Params("lessonID")
	fileID := ctx.Params("fileID")
	log.InfoContext(
		userCtx,
		"Deleting lesson file...",
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"sectionID", sectionID,
		"lessonID", lessonID,
		"fileID", fileID,
	)

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(NewRequestError(services.NewForbiddenError()))
	}

	params := dtos.LessonFilePathParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SectionID:    sectionID,
		LessonID:     lessonID,
		FileID:       fileID,
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
				Param:   "lessonsId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SectionID,
			}}))
	}

	parsedFileID, err := uuid.FromBytes([]byte(params.FileID))
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "fileId",
				Message: StrFieldErrMessageUUID,
				Value:   params.FileID,
			}}))
	}

	sectionIDi32 := int32(parsedSectionID)
	lessonIDi32 := int32(parsedLessonID)
	opts := services.DeleteLessonFileOptions{
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SectionID:    sectionIDi32,
		LessonID:     lessonIDi32,
		File:         parsedFileID,
	}
	if serviceErr := c.services.DeleteLessonFile(userCtx, opts); serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

func (c *Controllers) UpdateLessonFile(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.lesson_files.UpdateLessonFile")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	sectionID := ctx.Params("sectionID")
	lessonID := ctx.Params("lessonID")
	fileID := ctx.Params("fileID")
	name := ctx.FormValue("name")
	log.InfoContext(
		userCtx,
		"Updating lesson file...",
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"sectionID", sectionID,
		"lessonID", lessonID,
		"fileID", fileID,
		"name", name,
	)

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(NewRequestError(services.NewForbiddenError()))
	}

	params := dtos.LessonFilePathParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SectionID:    sectionID,
		LessonID:     lessonID,
		FileID:       fileID,
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
				Param:   "lessonsId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SectionID,
			}}))
	}

	parsedFileID, err := uuid.FromBytes([]byte(params.FileID))
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "fileId",
				Message: StrFieldErrMessageUUID,
				Value:   params.FileID,
			}}))
	}

	request := dtos.LessonFileBody{Name: name}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	sectionIDi32 := int32(parsedSectionID)
	lessonIDi32 := int32(parsedLessonID)
	lessonFile, serviceErr := c.services.UpdateLessonFile(userCtx, services.UpdateLessonFileOptions{
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SectionID:    sectionIDi32,
		LessonID:     lessonIDi32,
		File:         parsedFileID,
		Name:         request.Name,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	fileUrl, serviceErr := c.services.FindFileURL(userCtx, services.FindFileURLOptions{
		UserID:  user.ID,
		FileID:  lessonFile.ID,
		FileExt: lessonFile.Ext,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(
		dtos.NewLessonFileResponse(
			c.backendDomain,
			params.LanguageSlug,
			params.SeriesSlug,
			sectionIDi32,
			lessonFile.ToLessonFileModel(fileUrl),
		),
	)
}
