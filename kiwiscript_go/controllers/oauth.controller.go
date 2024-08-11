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
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/kiwiscript/kiwiscript_go/dtos"
	"github.com/kiwiscript/kiwiscript_go/services"
	"github.com/kiwiscript/kiwiscript_go/utils"
	"net/url"
)

func (c *Controllers) generateOAuthAcceptURL(ctx *fiber.Ctx, code, state string) error {
	params := make(url.Values)
	params.Add("code", code)
	params.Add("state", state)
	redirectUrl := fmt.Sprintf("https://%s/auth/callback?%s", c.frontendDomain, params.Encode())
	return ctx.Redirect(redirectUrl, fiber.StatusFound)
}

func (c *Controllers) GitHubSignIn(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.oauth.GitHubSignIn")
	userCtx := ctx.UserContext()
	log.InfoContext(userCtx, "Signing up with GitHub...")

	authUrl, serviceErr := c.services.GetAuthorizationURL(userCtx, utils.ProviderGitHub)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	log.InfoContext(userCtx, "Redirecting the user to login")
	return ctx.Redirect(authUrl, fiber.StatusTemporaryRedirect)
}

func (c *Controllers) GitHubCallback(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.oauth.GitHubCallback")
	userCtx := ctx.UserContext()
	log.InfoContext(userCtx, "Getting GitHub token...")

	queryParams := dtos.OAuthTokenParams{
		Code:  ctx.Query("code"),
		State: ctx.Query("state"),
	}
	if err := c.validate.StructCtx(userCtx, queryParams); err != nil {
		return c.validateQueryErrorResponse(log, userCtx, err, ctx)
	}

	token, serviceErr := c.services.GetOAuthToken(userCtx, services.GetOAuthTokenOptions{
		Provider: utils.ProviderGitHub,
		Code:     queryParams.Code,
		State:    queryParams.State,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	code, state, serviceErr := c.services.ExtOAuthSignIn(userCtx, services.ExtOAuthSignInOptions{
		Provider: utils.ProviderGitHub,
		Token:    token,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return c.generateOAuthAcceptURL(ctx, code, state)
}

func (c *Controllers) GoogleSignIn(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.oauth.GoogleSignIn")
	userCtx := ctx.UserContext()
	log.InfoContext(userCtx, "Signing up with Google...")

	authUrl, serviceErr := c.services.GetAuthorizationURL(userCtx, utils.ProviderGoogle)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	log.InfoContext(userCtx, "Redirecting the user to front-end login")
	return ctx.Redirect(authUrl, fiber.StatusFound)
}

func (c *Controllers) GoogleCallback(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.oauth.GoogleCallback")
	userCtx := ctx.UserContext()
	log.InfoContext(userCtx, "Getting Google token...")

	queryParams := dtos.OAuthTokenParams{
		Code:  ctx.Query("code"),
		State: ctx.Query("state"),
	}
	if err := c.validate.StructCtx(userCtx, queryParams); err != nil {
		return c.validateQueryErrorResponse(log, userCtx, err, ctx)
	}

	token, serviceErr := c.services.GetOAuthToken(userCtx, services.GetOAuthTokenOptions{
		Provider: utils.ProviderGoogle,
		Code:     queryParams.Code,
		State:    queryParams.State,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	code, state, serviceErr := c.services.ExtOAuthSignIn(userCtx, services.ExtOAuthSignInOptions{
		Provider: utils.ProviderGoogle,
		Token:    token,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return c.generateOAuthAcceptURL(ctx, code, state)
}

func (c *Controllers) OAuthToken(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.oauth.OAuthToken")
	userCtx := ctx.UserContext()
	log.InfoContext(userCtx, "Generating oauth token...")

	var body dtos.OAuthTokenBody
	if err := ctx.BodyParser(&body); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.Struct(body); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	authRes, serviceErr := c.services.IntOAuthSignIn(userCtx, services.IntOAuthSignInOptions{
		Code:  body.Code,
		State: body.State,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return c.processAuthResponse(ctx, authRes)
}