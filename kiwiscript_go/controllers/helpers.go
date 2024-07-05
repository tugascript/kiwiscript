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
