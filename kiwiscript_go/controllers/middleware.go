package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kiwiscript/kiwiscript_go/providers/tokens"
	"github.com/kiwiscript/kiwiscript_go/services"
)

func (c *Controllers) AccessClaimsMiddleware(ctx *fiber.Ctx) error {
	authHeader := ctx.Get("Authorization")
	if authHeader == "" {
		return ctx.
			Status(fiber.StatusUnauthorized).
			JSON(NewRequestError(services.NewUnauthorizedError()))
	}

	userClaims, err := c.services.ProcessAuthHeader(authHeader)
	if err != nil {
		return ctx.
			Status(fiber.StatusUnauthorized).
			JSON(NewRequestError(err))
	}

	ctx.Locals("user", userClaims)
	return ctx.Next()
}

func (c *Controllers) GetUserClaims(ctx *fiber.Ctx) (tokens.AccessUserClaims, *services.ServiceError) {
	user, ok := ctx.Locals("user").(tokens.AccessUserClaims)

	if !ok || user.ID == 0 {
		return user, services.NewUnauthorizedError()
	}

	return user, nil
}

func (c *Controllers) AdminUserMiddleware(ctx *fiber.Ctx) error {
	user, err := c.GetUserClaims(ctx)
	if err != nil {
		return ctx.
			Status(fiber.StatusUnauthorized).
			JSON(NewRequestError(err))
	}

	if !user.IsAdmin {
		return ctx.
			Status(fiber.StatusForbidden).
			JSON(NewRequestError(services.NewForbiddenError()))
	}

	return ctx.Next()
}
