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
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "sectionId",
				Message: StrFieldErrMessageNumber,
				Value:   params.UserID,
			}}))
	}

	user, serviceErr := c.services.FindUserByID(userCtx, services.FindUserByIDOptions{
		RequestID: requestID,
		ID:        int32(parsedUserID),
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	currentUser, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil {
		if !user.IsStaff {
			return c.serviceErrorResponse(services.NewNotFoundError(), ctx)
		}

		return ctx.JSON(dtos.NewUserResponse(c.backendDomain, user.ToUserModel()))
	}

	if user.ID != currentUser.ID || !user.IsStaff {
		return c.serviceErrorResponse(services.NewNotFoundError(), ctx)
	}
	return ctx.JSON(dtos.NewUserResponse(c.backendDomain, user.ToUserModel()))
}

func (c *Controllers) GetMe(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, usersLocation, "GetMe")
	log.InfoContext(userCtx, "Getting me...")

	currentUser, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil {
		return c.serviceErrorResponse(services.NewUnauthorizedError(), ctx)
	}

	user, serviceErr := c.services.FindUserByID(userCtx, services.FindUserByIDOptions{
		RequestID: requestID,
		ID:        currentUser.ID,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
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
		return c.serviceErrorResponse(services.NewUnauthorizedError(), ctx)
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
		return c.serviceErrorResponse(services.NewUnauthorizedError(), ctx)
	}

	if currentUser.IsAdmin || currentUser.IsStaff {
		log.WarnContext(userCtx, "Staff and admin users cannot delete their accounts")
		return ctx.Status(fiber.StatusForbidden).JSON(NewRequestError(services.NewForbiddenError()))
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
