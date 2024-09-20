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

const userProfilesLocation string = "user_profiles"

func (c *Controllers) CreateUserProfile(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, userProfilesLocation, "CreateUserProfile")
	log.InfoContext(userCtx, "Creating user profile...")

	user, err := c.GetUserClaims(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(exceptions.NewRequestError(exceptions.NewUnauthorizedError()))
	}
	if !user.IsStaff {
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
	}

	var body dtos.UserProfileBody
	if err := ctx.BodyParser(&body); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, body); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	profile, serviceErr := c.services.CreateUserProfile(userCtx, services.UserProfileOptions{
		RequestID: requestID,
		UserID:    user.ID,
		Bio:       body.Bio,
		GitHub:    body.GitHub,
		LinkedIn:  body.LinkedIn,
		Website:   body.Website,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.Status(fiber.StatusCreated).JSON(
		dtos.NewUserProfileResponse(c.backendDomain, profile.ToUserProfileModel()),
	)
}

func (c *Controllers) UpdateUserProfile(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, userProfilesLocation, "UpdateUserProfile")
	log.InfoContext(userCtx, "Updating user profile...")

	user, err := c.GetUserClaims(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(exceptions.NewRequestError(exceptions.NewUnauthorizedError()))
	}
	if !user.IsStaff {
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
	}

	var body dtos.UserProfileBody
	if err := ctx.BodyParser(&body); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, body); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	profile, serviceErr := c.services.UpdateUserProfile(userCtx, services.UserProfileOptions{
		RequestID: requestID,
		UserID:    user.ID,
		Bio:       body.Bio,
		GitHub:    body.GitHub,
		LinkedIn:  body.LinkedIn,
		Website:   body.Website,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(dtos.NewUserProfileResponse(c.backendDomain, profile.ToUserProfileModel()))
}

func (c *Controllers) DeleteUserProfile(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, userProfilesLocation, "UpdateUserProfile")
	log.InfoContext(userCtx, "Updating user profile...")

	user, err := c.GetUserClaims(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(exceptions.NewRequestError(exceptions.NewUnauthorizedError()))
	}
	if !user.IsStaff {
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
	}

	delOpts := services.DeleteUserProfileOptions{
		RequestID: requestID,
		UserID:    user.ID,
	}
	if serviceErr := c.services.DeleteUserProfile(userCtx, delOpts); serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

func (c *Controllers) GetUserProfile(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	userID := ctx.Params("userID")
	log := c.buildLogger(ctx, requestID, userProfilesLocation, "GetUserProfile").With(
		"userId", userID,
	)
	log.InfoContext(userCtx, "Getting user profile...")

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

	profile, serviceErr := c.services.FindUserProfile(userCtx, services.FindUserProfileOptions{
		RequestID: requestID,
		UserID:    int32(parsedUserID),
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(dtos.NewUserProfileResponse(c.backendDomain, profile.ToUserProfileModel()))
}

func (c *Controllers) GetMyProfile(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, userProfilesLocation, "GetMyProfile")
	log.InfoContext(userCtx, "Getting current user profile...")

	user, err := c.GetUserClaims(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(exceptions.NewRequestError(exceptions.NewUnauthorizedError()))
	}
	if !user.IsStaff {
		return ctx.Status(fiber.StatusForbidden).JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
	}

	profile, serviceErr := c.services.FindUserProfile(userCtx, services.FindUserProfileOptions{
		RequestID: requestID,
		UserID:    user.ID,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(dtos.NewUserProfileResponse(c.backendDomain, profile.ToUserProfileModel()))
}
