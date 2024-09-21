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

const userPicturesLocation string = "user_pictures"

func (c *Controllers) UploadUserPicture(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, userPicturesLocation, "UploadUserPicture")
	log.InfoContext(userCtx, "Uploading user picture...")

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil {
		log.ErrorContext(userCtx, "User is not logged in, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewUnauthorizedError()))
	}
	if !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(exceptions.NewRequestValidationError(
				exceptions.RequestValidationLocationBody,
				[]exceptions.FieldError{{
					Param:   "file",
					Message: exceptions.FieldErrMessageRequired,
				}},
			))
	}

	picture, serviceErr := c.services.UploadUserPicture(userCtx, services.UploadUserPictureOptions{
		RequestID:  requestID,
		UserID:     user.ID,
		FileHeader: file,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	pictureURL, serviceErr := c.services.FindFileURL(userCtx, services.FindFileURLOptions{
		RequestID: requestID,
		UserID:    user.ID,
		FileID:    picture.ID,
		FileExt:   picture.Ext,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.Status(fiber.StatusCreated).JSON(
		dtos.NewUserPictureResponse(c.backendDomain, picture.ToUserPictureModel(pictureURL)),
	)
}

func (c *Controllers) GetUserPicture(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	userID := ctx.Params("userID")
	log := c.buildLogger(ctx, requestID, userPicturesLocation, "GetUserPicture").With(
		"userId", userID,
	)
	log.InfoContext(userCtx, "Getting user picture...")

	params := dtos.UserPathParams{UserID: userID}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	parsedUserID, err := strconv.Atoi(params.UserID)
	if err != nil || parsedUserID <= 0 {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(exceptions.NewRequestValidationError(exceptions.RequestValidationLocationParams, []exceptions.FieldError{{
				Param:   "sectionId",
				Message: exceptions.StrFieldErrMessageNumber,
				Value:   params.UserID,
			}}))
	}

	userIDi32 := int32(parsedUserID)
	picture, serviceErr := c.services.FindUserPicture(userCtx, services.FindUserPictureOptions{
		RequestID: requestID,
		UserID:    userIDi32,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	pictureURL, serviceErr := c.services.FindFileURL(userCtx, services.FindFileURLOptions{
		RequestID: requestID,
		UserID:    userIDi32,
		FileID:    picture.ID,
		FileExt:   picture.Ext,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(
		dtos.NewUserPictureResponse(c.backendDomain, picture.ToUserPictureModel(pictureURL)),
	)
}

func (c *Controllers) GetCurrentAccountPicture(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, userPicturesLocation, "GetCurrentAccountPicture")
	log.InfoContext(userCtx, "Getting current user picture...")

	user, err := c.GetUserClaims(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(exceptions.NewRequestError(exceptions.NewUnauthorizedError()))
	}
	if !user.IsStaff {
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
	}

	picture, serviceErr := c.services.FindUserPicture(userCtx, services.FindUserPictureOptions{
		RequestID: requestID,
		UserID:    user.ID,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	pictureURL, serviceErr := c.services.FindFileURL(userCtx, services.FindFileURLOptions{
		RequestID: requestID,
		UserID:    user.ID,
		FileID:    picture.ID,
		FileExt:   picture.Ext,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(
		dtos.NewUserPictureResponse(c.backendDomain, picture.ToUserPictureModel(pictureURL)),
	)
}

func (c *Controllers) DeleteUserPicture(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, userPicturesLocation, "DeleteUserPicture")
	log.InfoContext(userCtx, "Deleting user picture...")

	user, err := c.GetUserClaims(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(exceptions.NewRequestError(exceptions.NewUnauthorizedError()))
	}
	if !user.IsStaff {
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
	}

	delOpts := services.DeleteUserPictureOptions{
		RequestID: requestID,
		UserID:    user.ID,
	}
	if serviceErr := c.services.DeleteUserPicture(userCtx, delOpts); serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
