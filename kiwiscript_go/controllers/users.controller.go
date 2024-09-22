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
	"github.com/kiwiscript/kiwiscript_go/utils"
	"strconv"
)

const usersLocation string = "users"

func (c *Controllers) GetUser(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	userID := ctx.Params("userID")
	log := c.buildLogger(ctx, requestID, usersLocation, "GetUser").With(
		"userID", userID,
	)
	log.InfoContext(userCtx, "Getting user...")

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
	currentUser, serviceErr := c.GetUserClaims(ctx)
	if serviceErr == nil && currentUser.ID == userIDi32 && !currentUser.IsStaff {
		user, serviceErr := c.services.FindUserByID(userCtx, services.FindUserByIDOptions{
			RequestID: requestID,
			ID:        userIDi32,
		})
		if serviceErr != nil {
			return c.serviceErrorResponse(serviceErr, ctx)
		}

		return ctx.JSON(dtos.NewUserResponse(c.backendDomain, user.ToUserModel()))
	}

	user, serviceErr := c.services.FindStaffUserWithProfileAndPicture(userCtx, services.FindUserByIDOptions{
		RequestID: requestID,
		ID:        userIDi32,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	if user.PictureID.Valid && user.PictureExt.Valid {
		pictureUrl, pictureUrlErr := c.services.FindFileURL(userCtx, services.FindFileURLOptions{
			RequestID: requestID,
			UserID:    user.ID,
			FileID:    user.PictureID.Bytes,
			FileExt:   user.PictureExt.String,
		})
		if pictureUrlErr != nil {
			return c.serviceErrorResponse(pictureUrlErr, ctx)
		}

		return ctx.JSON(
			dtos.NewUserResponseWithEmbedded(
				c.backendDomain,
				user.ToUserModel(),
				user.ToUserProfileModel(),
				user.ToUserPictureModel(pictureUrl),
			),
		)
	}

	return ctx.JSON(
		dtos.NewUserResponseWithEmbedded(
			c.backendDomain,
			user.ToUserModel(),
			user.ToUserProfileModel(),
			nil,
		),
	)
}

func (c *Controllers) GetCurrentAccount(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, usersLocation, "GetCurrentAccount")
	log.InfoContext(userCtx, "Getting me...")

	currentUser, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil {
		return c.serviceErrorResponse(exceptions.NewUnauthorizedError(), ctx)
	}

	user, serviceErr := c.services.FindUserByID(userCtx, services.FindUserByIDOptions{
		RequestID: requestID,
		ID:        currentUser.ID,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	if user.IsStaff {
		profile, profileErr := c.services.FindUserProfile(userCtx, services.FindUserProfileOptions{
			RequestID: requestID,
			UserID:    user.ID,
		})
		if profileErr != nil && profileErr.Code != exceptions.CodeNotFound {
			return c.serviceErrorResponse(profileErr, ctx)
		}

		picture, pictureErr := c.services.FindUserPicture(userCtx, services.FindUserPictureOptions{
			RequestID: requestID,
			UserID:    user.ID,
		})
		if pictureErr != nil && pictureErr.Code != exceptions.CodeNotFound {
			return c.serviceErrorResponse(pictureErr, ctx)
		}

		if picture != nil {
			pictureUrl, pictureUrlErr := c.services.FindFileURL(userCtx, services.FindFileURLOptions{
				RequestID: requestID,
				UserID:    user.ID,
				FileID:    picture.ID,
				FileExt:   picture.Ext,
			})
			if pictureUrlErr != nil {
				return c.serviceErrorResponse(pictureUrlErr, ctx)
			}

			if profile != nil {
				return ctx.JSON(
					dtos.NewUserResponseWithEmbedded(
						c.backendDomain,
						user.ToUserModel(),
						profile.ToUserProfileModel(),
						picture.ToUserPictureModel(pictureUrl),
					),
				)
			}

			return ctx.JSON(
				dtos.NewUserResponseWithEmbedded(
					c.backendDomain,
					user.ToUserModel(),
					nil,
					picture.ToUserPictureModel(pictureUrl),
				),
			)
		}

		if profile != nil {
			return ctx.JSON(
				dtos.NewUserResponseWithEmbedded(
					c.backendDomain,
					user.ToUserModel(),
					profile.ToUserProfileModel(),
					nil,
				),
			)
		}
	}

	return ctx.JSON(dtos.NewUserResponse(c.backendDomain, user.ToUserModel()))
}

func (c *Controllers) UpdateCurrentAccount(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, usersLocation, "UpdateCurrentAccount")
	log.InfoContext(userCtx, "Updating me...")

	currentUser, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil {
		return c.serviceErrorResponse(exceptions.NewUnauthorizedError(), ctx)
	}

	var body dtos.UpdateUserBody
	if err := ctx.BodyParser(&body); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, body); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	user, serviceErr := c.services.UpdateUser(userCtx, services.UpdateUserOptions{
		RequestID: requestID,
		ID:        currentUser.ID,
		FirstName: utils.Capitalized(body.FirstName),
		LastName:  utils.Capitalized(body.LastName),
		Location:  utils.Uppercased(body.Location),
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(dtos.NewUserResponse(c.backendDomain, user.ToUserModel()))
}

func (c *Controllers) DeleteCurrentAccount(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, usersLocation, "DeleteCurrentAccount")
	log.InfoContext(userCtx, "Deleting me...")

	currentUser, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil {
		return c.serviceErrorResponse(exceptions.NewUnauthorizedError(), ctx)
	}

	if currentUser.IsAdmin || currentUser.IsStaff {
		log.WarnContext(userCtx, "Staff and admin users cannot delete their accounts")
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
	}

	var body dtos.DeleteUserBody
	if err := ctx.BodyParser(&body); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, body); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	serviceErr = c.services.DeleteUser(userCtx, services.DeleteUserOptions{
		RequestID: requestID,
		ID:        currentUser.ID,
		Password:  body.Password,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
