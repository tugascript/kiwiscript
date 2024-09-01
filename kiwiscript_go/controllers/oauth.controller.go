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
	"github.com/kiwiscript/kiwiscript_go/exceptions"
	"github.com/kiwiscript/kiwiscript_go/services"
	"github.com/kiwiscript/kiwiscript_go/utils"
	"net/url"
	"strconv"
)

const oauthLocation string = "oauth"

func (c *Controllers) generateOAuthAcceptURL(ctx *fiber.Ctx, response *services.OAuthResponse) error {
	params := make(url.Values)
	params.Add("code", response.Code)
	params.Add("accessToken", response.AccessToken)
	params.Add("tokenType", "Bearer")
	params.Add("expiresIn", strconv.FormatInt(response.ExpiresIn, 10))
	redirectUrl := fmt.Sprintf("https://%s/auth/callback?%s", c.frontendDomain, params.Encode())
	return ctx.Redirect(redirectUrl, fiber.StatusFound)
}

func (c *Controllers) GitHubSignIn(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, oauthLocation, "GitHubSignIn")
	log.InfoContext(userCtx, "Signing up with GitHub...")

	authUrl, serviceErr := c.services.GetAuthorizationURL(userCtx, services.GetAuthorizationURLOptions{
		RequestID: requestID,
		Provider:  utils.ProviderGitHub,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	log.InfoContext(userCtx, "Redirecting the user to login")
	return ctx.Redirect(authUrl, fiber.StatusTemporaryRedirect)
}

func (c *Controllers) GitHubCallback(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, oauthLocation, "GitHubCallback")
	log.InfoContext(userCtx, "Getting GitHub token...")

	queryParams := dtos.OAuthTokenParams{
		Code:  ctx.Query("code"),
		State: ctx.Query("state"),
	}
	if err := c.validate.StructCtx(userCtx, queryParams); err != nil {
		return c.validateQueryErrorResponse(log, userCtx, err, ctx)
	}

	token, serviceErr := c.services.GetOAuthToken(userCtx, services.GetOAuthTokenOptions{
		RequestID: requestID,
		Provider:  utils.ProviderGitHub,
		Code:      queryParams.Code,
		State:     queryParams.State,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	response, serviceErr := c.services.ExtOAuthSignIn(userCtx, services.ExtOAuthSignInOptions{
		RequestID: requestID,
		Provider:  utils.ProviderGitHub,
		Token:     token,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return c.generateOAuthAcceptURL(ctx, response)
}

func (c *Controllers) GoogleSignIn(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, oauthLocation, "GoogleSignIn")
	log.InfoContext(userCtx, "Signing up with Google...")

	authUrl, serviceErr := c.services.GetAuthorizationURL(userCtx, services.GetAuthorizationURLOptions{
		RequestID: requestID,
		Provider:  utils.ProviderGitHub,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	log.InfoContext(userCtx, "Redirecting the user to front-end login")
	return ctx.Redirect(authUrl, fiber.StatusFound)
}

func (c *Controllers) GoogleCallback(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, oauthLocation, "GoogleCallback")
	log.InfoContext(userCtx, "Getting Google token...")

	queryParams := dtos.OAuthTokenParams{
		Code:  ctx.Query("code"),
		State: ctx.Query("state"),
	}
	if err := c.validate.StructCtx(userCtx, queryParams); err != nil {
		return c.validateQueryErrorResponse(log, userCtx, err, ctx)
	}

	token, serviceErr := c.services.GetOAuthToken(userCtx, services.GetOAuthTokenOptions{
		RequestID: requestID,
		Provider:  utils.ProviderGoogle,
		Code:      queryParams.Code,
		State:     queryParams.State,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	response, serviceErr := c.services.ExtOAuthSignIn(userCtx, services.ExtOAuthSignInOptions{
		RequestID: requestID,
		Provider:  utils.ProviderGoogle,
		Token:     token,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return c.generateOAuthAcceptURL(ctx, response)
}

func (c *Controllers) OAuthToken(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, oauthLocation, "OAuthToken")
	log.InfoContext(userCtx, "Generating oauth token...")

	userClaims, serviceErr := c.services.ProcessOAuthHeader(userCtx, ctx.Get("Authorization"))
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	var body dtos.OAuthTokenBody
	if err := ctx.BodyParser(&body); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.Struct(body); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	redirectURI := fmt.Sprintf("https://%s/auth/callback", c.frontendDomain)
	if body.RedirectURI != redirectURI {
		return c.serviceErrorResponse(exceptions.NewUnauthorizedError(), ctx)
	}

	authRes, serviceErr := c.services.OAuthToken(userCtx, services.IntOAuthSignInOptions{
		RequestID:   requestID,
		UserID:      userClaims.ID,
		UserVersion: userClaims.Version,
		Code:        body.Code,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return c.processAuthResponse(ctx, authRes)
}
