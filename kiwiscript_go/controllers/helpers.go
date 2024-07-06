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
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/kiwiscript/kiwiscript_go/services"
)

func (c *Controllers) parseRequestErrorResponse(log *slog.Logger, userCtx context.Context, err error, ctx *fiber.Ctx) error {
	log.WarnContext(userCtx, "Failed to parse request", "error", err)
	return ctx.
		Status(fiber.StatusBadRequest).
		JSON(NewEmptyRequestValidationError(RequestValidationLocationBody))
}

func (c *Controllers) validateErrorResponse(
	log *slog.Logger,
	userCtx context.Context,
	err error,
	ctx *fiber.Ctx,
	location string,
) error {
	log.WarnContext(userCtx, "Failed to validate request", "error", err, "errorType", fmt.Sprintf("%T", err))

	errors, ok := err.(validator.ValidationErrors)
	if ok {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(RequestValidationErrorFromErr(&errors, location))
	}

	return ctx.
		Status(fiber.StatusBadRequest).
		JSON(NewEmptyRequestValidationError(location))
}

func (c *Controllers) validateRequestErrorResponse(log *slog.Logger, userCtx context.Context, err error, ctx *fiber.Ctx) error {
	return c.validateErrorResponse(log, userCtx, err, ctx, RequestValidationLocationBody)
}

func (c *Controllers) validateParamsErrorResponse(log *slog.Logger, userCtx context.Context, err error, ctx *fiber.Ctx) error {
	return c.validateErrorResponse(log, userCtx, err, ctx, RequestValidationLocationParams)
}

func (c *Controllers) validateQueryErrorResponse(log *slog.Logger, userCtx context.Context, err error, ctx *fiber.Ctx) error {
	return c.validateErrorResponse(log, userCtx, err, ctx, RequestValidationLocationQuery)
}

func (c *Controllers) serviceErrorResponse(serviceErr *services.ServiceError, ctx *fiber.Ctx) error {
	return ctx.
		Status(NewRequestErrorStatus(serviceErr.Code)).
		JSON(NewRequestError(serviceErr))
}
