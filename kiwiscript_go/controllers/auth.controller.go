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
	"github.com/kiwiscript/kiwiscript_go/dtos"
	"unicode"

	"github.com/gofiber/fiber/v2"
	"github.com/kiwiscript/kiwiscript_go/services"
	"github.com/kiwiscript/kiwiscript_go/utils"
)

const authLocation string = "auth"

type passwordValidity struct {
	hasLowercase bool
	hasUppercase bool
	hasNumber    bool
	hasSymbol    bool
}

func passwordValidator(password string) *services.ServiceError {
	validity := passwordValidity{}

	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			validity.hasLowercase = true
		case unicode.IsUpper(char):
			validity.hasUppercase = true
		case unicode.IsNumber(char):
			validity.hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			validity.hasSymbol = true
		}
	}

	if !validity.hasLowercase || !validity.hasUppercase || !validity.hasNumber || !validity.hasSymbol {
		return services.NewValidationError("Password must contain at least one lowercase letter, one uppercase letter, one number, and one symbol")
	}

	return nil
}

func (c *Controllers) processAuthResponse(ctx *fiber.Ctx, authRes *services.AuthResponse) error {
	ctx.Cookie(&fiber.Cookie{
		Name:     c.refreshCookieName,
		Value:    authRes.RefreshToken,
		Path:     "/api/auth",
		HTTPOnly: true,
		SameSite: "None",
		Secure:   true,
	})
	return ctx.
		Status(fiber.StatusOK).
		JSON(dtos.NewAuthResponse(authRes.AccessToken, authRes.RefreshToken, authRes.ExpiresIn))
}

func (c *Controllers) SignUp(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, authLocation, "SignUp")
	log.InfoContext(userCtx, "Signing up...")

	var request dtos.SignUpBody
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	if err := passwordValidator(request.Password1); err != nil {
		log.WarnContext(userCtx, "Failed to validate password", "error", err)
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationBody, []FieldError{
				{Param: "password", Message: err.Message},
			}))
	}

	opts := services.SignUpOptions{
		RequestID: requestID,
		Email:     utils.Lowered(request.Email),
		FirstName: utils.Capitalized(request.FirstName),
		LastName:  utils.Capitalized(request.LastName),
		Location:  utils.Uppercased(request.Location),
		Password:  request.Password1,
	}
	if serviceErr := c.services.SignUp(userCtx, opts); serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.
		Status(fiber.StatusOK).
		JSON(dtos.NewMessageResponse("Confirmation email has been sent"))
}

func (c *Controllers) SignIn(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, authLocation, "SignIn")
	log.InfoContext(userCtx, "Signing in...")

	var request dtos.SignInBody
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	opts := services.SignInOptions{
		RequestID: requestID,
		Email:     utils.Lowered(request.Email),
		Password:  request.Password,
	}
	if serviceErr := c.services.SignIn(userCtx, opts); serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.
		Status(fiber.StatusOK).
		JSON(dtos.NewMessageResponse("Confirmation code has been sent to your email"))
}

func (c *Controllers) ConfirmSignIn(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, authLocation, "ConfirmSignIn")
	log.InfoContext(userCtx, "Confirming sign in...")

	var request dtos.ConfirmSignInBody
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.Struct(request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	authRes, serviceErr := c.services.TwoFactor(userCtx, services.TwoFactorOptions{
		RequestID: requestID,
		Email:     request.Email,
		Code:      request.Code,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return c.processAuthResponse(ctx, authRes)
}

func (c *Controllers) SignOut(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, authLocation, "SignOut")
	log.InfoContext(userCtx, "Signing out...")

	refreshToken := ctx.Cookies(c.refreshCookieName)
	if refreshToken == "" {
		var request dtos.SignOutBody

		if err := ctx.BodyParser(&request); err != nil {
			return c.parseRequestErrorResponse(log, userCtx, err, ctx)
		}
		if err := c.validate.Struct(request); err != nil {
			return c.validateRequestErrorResponse(log, userCtx, err, ctx)
		}

		refreshToken = request.RefreshToken
	}

	opts := services.SignOutOptions{
		RequestID: requestID,
		Token:     refreshToken,
	}
	if serviceErr := c.services.SignOut(userCtx, opts); serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

func (c *Controllers) Refresh(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, authLocation, "Refresh")
	log.InfoContext(userCtx, "Refreshing access token...")

	refreshToken := ctx.Cookies(c.refreshCookieName)
	if refreshToken == "" {
		var request dtos.RefreshBody

		if err := ctx.BodyParser(&request); err != nil {
			return c.parseRequestErrorResponse(log, userCtx, err, ctx)
		}
		if err := c.validate.Struct(request); err != nil {
			return c.validateRequestErrorResponse(log, userCtx, err, ctx)
		}

		refreshToken = request.RefreshToken
	}

	authRes, serviceErr := c.services.Refresh(userCtx, services.RefreshOptions{
		RequestID: requestID,
		Token:     refreshToken,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return c.processAuthResponse(ctx, authRes)
}

func (c *Controllers) ConfirmEmail(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, authLocation, "ConfirmEmail")
	log.InfoContext(userCtx, "Confirming user email...")

	var request dtos.ConfirmBody
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.Struct(request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	authRes, serviceErr := c.services.ConfirmEmail(userCtx, services.ConfirmEmailOptions{
		RequestID: requestID,
		Token:     request.ConfirmationToken,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return c.processAuthResponse(ctx, authRes)
}

func (c *Controllers) ForgotPassword(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, authLocation, "ForgotPassword")
	log.InfoContext(userCtx, "Sending forgot password email...")

	var request dtos.ForgotPasswordBody
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.Struct(request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	opts := services.ForgotPasswordOptions{
		RequestID: requestID,
		Email:     utils.Lowered(request.Email),
	}
	if serviceErr := c.services.ForgotPassword(userCtx, opts); serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.
		Status(fiber.StatusOK).
		JSON(dtos.NewMessageResponse("If the email exists, a password reset email has been sent"))
}

func (c *Controllers) ResetPassword(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, authLocation, "ResetPassword")
	log.InfoContext(userCtx, "Resetting user password...")

	var request dtos.ResetPasswordBody
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.Struct(request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	if serviceErr := passwordValidator(request.Password1); serviceErr != nil {
		log.WarnContext(userCtx, "Failed to validate password", "error", serviceErr)
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	opts := services.ResetPasswordOptions{
		RequestID:   requestID,
		ResetToken:  request.ResetToken,
		NewPassword: request.Password1,
	}
	if serviceErr := c.services.ResetPassword(userCtx, opts); serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.
		Status(fiber.StatusOK).
		JSON(dtos.NewMessageResponse("Password reset successfully"))
}

func (c *Controllers) UpdatePassword(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, authLocation, "UpdatePassword")
	log.InfoContext(userCtx, "Updating password...")

	userClaims, err := c.GetUserClaims(ctx)
	if err != nil {
		log.ErrorContext(userCtx, "This should be a protected route", "error", err)
		return c.serviceErrorResponse(err, ctx)
	}

	var request dtos.UpdatePasswordBody
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.Struct(request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := passwordValidator(request.Password1); err != nil {
		return c.serviceErrorResponse(err, ctx)
	}

	authRes, serviceErr := c.services.UpdatePassword(userCtx, services.UpdatePasswordOptions{
		RequestID:   requestID,
		UserID:      userClaims.ID,
		UserVersion: userClaims.Version,
		OldPassword: request.OldPassword,
		NewPassword: request.Password1,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return c.processAuthResponse(ctx, authRes)
}

func (c *Controllers) UpdateEmail(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, authLocation, "UpdateEmail")
	log.InfoContext(userCtx, "Updating email...")

	userClaims, err := c.GetUserClaims(ctx)
	if err != nil {
		log.ErrorContext(userCtx, "This should be a protected route", "error", err)
		return c.serviceErrorResponse(err, ctx)
	}

	var request dtos.UpdateEmailBody
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.Struct(request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	authRes, serviceErr := c.services.UpdateEmail(userCtx, services.UpdateEmailOptions{
		RequestID:   requestID,
		UserID:      userClaims.ID,
		UserVersion: userClaims.Version,
		NewEmail:    request.Email,
		Password:    request.Password,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return c.processAuthResponse(ctx, authRes)
}
