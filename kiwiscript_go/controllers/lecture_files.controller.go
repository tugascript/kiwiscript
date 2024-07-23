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
	"github.com/kiwiscript/kiwiscript_go/services"
	"strconv"
)

func (c *Controllers) UploadLectureFile(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.lecture_files.UploadLectureFile")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	seriesPartID := ctx.Params("seriesPartID")
	lectureID := ctx.Params("lectureID")
	name := ctx.FormValue("name")
	log.InfoContext(
		userCtx,
		"Uploading lecture file...",
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"seriesPartID", seriesPartID,
		"lectureID", lectureID,
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

	params := LectureParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SeriesPartID: seriesPartID,
		LectureID:    lectureID,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	parsedSeriesPartID, err := strconv.Atoi(params.SeriesPartID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "seriesPartId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SeriesPartID,
			}}))
	}

	parsedLectureID, err := strconv.Atoi(params.LectureID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "lecturesId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SeriesPartID,
			}}))
	}

	request := LectureFileRequest{name}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	seriesPartIDi32 := int32(parsedSeriesPartID)
	lectureIDi32 := int32(parsedLectureID)
	lectureFile, serviceErr := c.services.UploadLectureFile(userCtx, services.UploadLectureFileOptions{
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SeriesPartID: seriesPartIDi32,
		LectureID:    lectureIDi32,
		FileName:     request.Name,
		FileHeader:   file,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	fileUrl, serviceErr := c.services.FindFileURL(userCtx, services.FindFileURLOptions{
		UserID:  user.ID,
		FileID:  lectureFile.File,
		FileExt: lectureFile.Ext,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.Status(fiber.StatusCreated).JSON(
		c.NewLectureFileResponse(
			lectureFile,
			fileUrl,
			params.LanguageSlug,
			params.SeriesSlug,
			seriesPartIDi32,
		),
	)
}

func (c *Controllers) GetLectureFiles(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.lecture_files.GetLectureFiles")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	seriesPartID := ctx.Params("seriesPartID")
	lectureID := ctx.Params("lectureID")
	log.InfoContext(
		userCtx,
		"Getting lecture files...",
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"seriesPartID", seriesPartID,
		"lectureID", lectureID,
	)

	params := LectureParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SeriesPartID: seriesPartID,
		LectureID:    lectureID,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	parsedSeriesPartID, err := strconv.Atoi(params.SeriesPartID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "seriesPartId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SeriesPartID,
			}}))
	}

	parsedLectureID, err := strconv.Atoi(params.LectureID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "lecturesId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SeriesPartID,
			}}))
	}

	seriesPartIDi32 := int32(parsedSeriesPartID)
	lectureIDi32 := int32(parsedLectureID)
	isPublished := false
	if user, serviceErr := c.GetUserClaims(ctx); serviceErr != nil || !user.IsStaff {
		isPublished = true
	}

	lectureFiles, serviceErr := c.services.FindLectureFiles(userCtx, services.FindLectureFilesOptions{
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SeriesPartID: seriesPartIDi32,
		LectureID:    lectureIDi32,
		IsPublished:  isPublished,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	filesLen := len(lectureFiles)
	if filesLen == 0 {
		return ctx.JSON(make([]*LectureFileResponse, 0))
	}

	optsList := make([]services.FindFileURLOptions, 0, filesLen)
	for _, lectureFile := range lectureFiles {
		optsList = append(optsList, services.FindFileURLOptions{
			UserID:  lectureFile.AuthorID,
			FileID:  lectureFile.File,
			FileExt: lectureFile.Ext,
		})
	}

	fileUrls, serviceErr := c.services.FindFileURLs(userCtx, optsList)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	responses := make([]*LectureFileResponse, 0, len(lectureFiles))
	for _, lectureFile := range lectureFiles {
		url, ok := fileUrls.Get(lectureFile.File)
		if !ok {
			log.WarnContext(userCtx, "Could not find URL for lecture file", "fileId", lectureFile.File)
			continue
		}
		responses = append(
			responses,
			c.NewLectureFileResponse(&lectureFile, url, params.LanguageSlug, params.SeriesSlug, seriesPartIDi32),
		)
	}

	return ctx.JSON(responses)
}

func (c *Controllers) GetLectureFile(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.lecture_files.GetLectureFile")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	seriesPartID := ctx.Params("seriesPartID")
	lectureID := ctx.Params("lectureID")
	fileID := ctx.Params("fileID")
	log.InfoContext(
		userCtx,
		"Deleting lecture file...",
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"seriesPartID", seriesPartID,
		"lectureID", lectureID,
		"fileID", fileID,
	)

	params := LectureFileParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SeriesPartID: seriesPartID,
		LectureID:    lectureID,
		FileID:       fileID,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	parsedSeriesPartID, err := strconv.Atoi(params.SeriesPartID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "seriesPartId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SeriesPartID,
			}}))
	}

	parsedLectureID, err := strconv.Atoi(params.LectureID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "lecturesId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SeriesPartID,
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

	seriesPartIDi32 := int32(parsedSeriesPartID)
	lectureIDi32 := int32(parsedLectureID)
	isPublished := false
	if user, serviceErr := c.GetUserClaims(ctx); serviceErr != nil || !user.IsStaff {
		isPublished = true
	}

	lectureFile, serviceErr := c.services.FindLectureFile(userCtx, services.FindLectureFileOptions{
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SeriesPartID: seriesPartIDi32,
		LectureID:    lectureIDi32,
		File:         parsedFileID,
		IsPublished:  isPublished,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	fileUrl, serviceErr := c.services.FindFileURL(userCtx, services.FindFileURLOptions{
		UserID:  lectureFile.AuthorID,
		FileID:  lectureFile.File,
		FileExt: lectureFile.Ext,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(
		c.NewLectureFileResponse(lectureFile, fileUrl, params.LanguageSlug, params.SeriesSlug, seriesPartIDi32),
	)
}

func (c *Controllers) DeleteLectureFile(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.lecture_files.DeleteLectureFile")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	seriesPartID := ctx.Params("seriesPartID")
	lectureID := ctx.Params("lectureID")
	fileID := ctx.Params("fileID")
	log.InfoContext(
		userCtx,
		"Deleting lecture file...",
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"seriesPartID", seriesPartID,
		"lectureID", lectureID,
		"fileID", fileID,
	)

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(NewRequestError(services.NewForbiddenError()))
	}

	params := LectureFileParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SeriesPartID: seriesPartID,
		LectureID:    lectureID,
		FileID:       fileID,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	parsedSeriesPartID, err := strconv.Atoi(params.SeriesPartID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "seriesPartId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SeriesPartID,
			}}))
	}

	parsedLectureID, err := strconv.Atoi(params.LectureID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "lecturesId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SeriesPartID,
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

	seriesPartIDi32 := int32(parsedSeriesPartID)
	lectureIDi32 := int32(parsedLectureID)
	opts := services.DeleteLectureFileOptions{
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SeriesPartID: seriesPartIDi32,
		LectureID:    lectureIDi32,
		File:         parsedFileID,
	}
	if serviceErr := c.services.DeleteLectureFile(userCtx, opts); serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

func (c *Controllers) UpdateLectureFile(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.lecture_files.UpdateLectureFile")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	seriesPartID := ctx.Params("seriesPartID")
	lectureID := ctx.Params("lectureID")
	fileID := ctx.Params("fileID")
	name := ctx.FormValue("name")
	log.InfoContext(
		userCtx,
		"Updating lecture file...",
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"seriesPartID", seriesPartID,
		"lectureID", lectureID,
		"fileID", fileID,
		"name", name,
	)

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(NewRequestError(services.NewForbiddenError()))
	}

	params := LectureFileParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SeriesPartID: seriesPartID,
		LectureID:    lectureID,
		FileID:       fileID,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	parsedSeriesPartID, err := strconv.Atoi(params.SeriesPartID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "seriesPartId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SeriesPartID,
			}}))
	}

	parsedLectureID, err := strconv.Atoi(params.LectureID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "lecturesId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SeriesPartID,
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

	request := LectureFileRequest{name}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	seriesPartIDi32 := int32(parsedSeriesPartID)
	lectureIDi32 := int32(parsedLectureID)
	lectureFile, serviceErr := c.services.UpdateLectureFile(userCtx, services.UpdateLectureFileOptions{
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SeriesPartID: seriesPartIDi32,
		LectureID:    lectureIDi32,
		File:         parsedFileID,
		Name:         request.Name,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	fileUrl, serviceErr := c.services.FindFileURL(userCtx, services.FindFileURLOptions{
		UserID:  user.ID,
		FileID:  lectureFile.File,
		FileExt: lectureFile.Ext,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(
		c.NewLectureFileResponse(
			lectureFile,
			fileUrl,
			params.LanguageSlug,
			params.SeriesSlug,
			seriesPartIDi32,
		),
	)
}
