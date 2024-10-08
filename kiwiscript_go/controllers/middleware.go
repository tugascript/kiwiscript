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
	"github.com/kiwiscript/kiwiscript_go/exceptions"
	"github.com/kiwiscript/kiwiscript_go/providers/tokens"
)

func (c *Controllers) AccessClaimsMiddleware(ctx *fiber.Ctx) error {
	authHeader := ctx.Get("Authorization")
	if authHeader == "" {
		return ctx.Next()
	}

	userClaims, err := c.services.ProcessAuthHeader(authHeader)
	if err != nil {
		return ctx.Next()
	}

	ctx.Locals("user", userClaims)
	return ctx.Next()
}

func (c *Controllers) GetUserClaims(ctx *fiber.Ctx) (*tokens.AccessUserClaims, *exceptions.ServiceError) {
	user, ok := ctx.Locals("user").(tokens.AccessUserClaims)

	if !ok || user.ID == 0 {
		return nil, exceptions.NewUnauthorizedError()
	}

	return &user, nil
}

func (c *Controllers) UserMiddleware(ctx *fiber.Ctx) error {
	_, err := c.GetUserClaims(ctx)
	if err != nil {
		return ctx.
			Status(fiber.StatusUnauthorized).
			JSON(exceptions.NewRequestError(err))
	}

	return ctx.Next()
}

func (c *Controllers) AdminUserMiddleware(ctx *fiber.Ctx) error {
	user, err := c.GetUserClaims(ctx)
	if err != nil {
		return ctx.
			Status(fiber.StatusUnauthorized).
			JSON(exceptions.NewRequestError(err))
	}

	if !user.IsAdmin {
		return ctx.
			Status(fiber.StatusForbidden).
			JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
	}

	return ctx.Next()
}

func (c *Controllers) StaffUserMiddleware(ctx *fiber.Ctx) error {
	user, err := c.GetUserClaims(ctx)
	if err != nil {
		return ctx.
			Status(fiber.StatusUnauthorized).
			JSON(exceptions.NewRequestError(err))
	}

	if !user.IsStaff {
		return ctx.
			Status(fiber.StatusForbidden).
			JSON(exceptions.NewRequestError(exceptions.NewForbiddenError()))
	}

	return ctx.Next()
}
