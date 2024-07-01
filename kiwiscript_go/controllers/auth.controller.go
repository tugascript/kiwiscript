package controllers

import (
	"unicode"

	"github.com/gofiber/fiber/v2"
	"github.com/kiwiscript/kiwiscript_go/services"
	"github.com/kiwiscript/kiwiscript_go/utils"
)

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

func (c *Controllers) processAuthResponse(ctx *fiber.Ctx, authRes services.AuthResponse) error {
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
		JSON(NewAuthResponse(authRes.AccessToken, authRes.RefreshToken, authRes.ExpiresIn))
}

func (c *Controllers) SignUp(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.auth.SignUp")
	userCtx := ctx.UserContext()
	var request SignUpRequest

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
				{Param: "password", Error: err.Message},
			}))
	}
	if request.Password1 != request.Password2 {
		errMsg := "Passwords do not match"
		log.WarnContext(userCtx, errMsg)
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationBody, []FieldError{
				{Param: "password", Error: errMsg},
			}))
	}

	opts := services.SignUpOptions{
		Email:     utils.Lowered(request.Email),
		FirstName: utils.Capitalized(request.FirstName),
		LastName:  utils.Capitalized(request.LastName),
		Location:  utils.Uppercased(request.Location),
		BirthDate: request.BirthDate,
		Password:  request.Password1,
	}
	if serviceErr := c.services.SignUp(userCtx, opts); serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.
		Status(fiber.StatusCreated).
		JSON(NewMessageResponse("Confirmation email has been sent"))
}

func (c *Controllers) SignIn(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.auth.SignIn")
	userCtx := ctx.UserContext()
	var request SignInRequest

	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	opts := services.SignInOptions{
		Email:    utils.Lowered(request.Email),
		Password: request.Password,
	}
	if serviceErr := c.services.SignIn(userCtx, opts); serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.
		Status(fiber.StatusOK).
		JSON(NewMessageResponse("Sign in successful"))
}

func (c *Controllers) ConfirmSignIn(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.auth.ConfirmSignIn")
	userCtx := ctx.UserContext()
	var request ConfirmSignInRequest

	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.Struct(request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	authRes, serviceErr := c.services.TwoFactor(ctx.UserContext(), services.TwoFactorOptions{
		Email: request.Email,
		Code:  request.Code,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return c.processAuthResponse(ctx, authRes)
}

func (c *Controllers) SignOut(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.auth.SignOut")
	userCtx := ctx.UserContext()
	refreshToken := ctx.Cookies(c.refreshCookieName)

	log.Info("SignOut", "refreshToken", refreshToken)
	if refreshToken == "" {
		var request SignOutRequest

		if err := ctx.BodyParser(&request); err != nil {
			return c.parseRequestErrorResponse(log, userCtx, err, ctx)
		}
		if err := c.validate.Struct(request); err != nil {
			return c.validateRequestErrorResponse(log, userCtx, err, ctx)
		}

		refreshToken = request.RefreshToken
	}
	if serviceErr := c.services.SignOut(ctx.UserContext(), refreshToken); serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

func (c *Controllers) Refresh(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.auth.Refresh")
	userCtx := ctx.UserContext()
	refreshToken := ctx.Cookies(c.refreshCookieName)

	if refreshToken == "" {
		var request RefreshRequest

		if err := ctx.BodyParser(&request); err != nil {
			return c.parseRequestErrorResponse(log, userCtx, err, ctx)
		}
		if err := c.validate.Struct(request); err != nil {
			return c.validateRequestErrorResponse(log, userCtx, err, ctx)
		}

		refreshToken = request.RefreshToken
	}

	authRes, serviceErr := c.services.Refresh(ctx.UserContext(), refreshToken)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return c.processAuthResponse(ctx, authRes)
}

func (c *Controllers) ConfirmEmail(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.auth.ConfirmEmail")
	userCtx := ctx.UserContext()
	var request ConfirmRequest

	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.Struct(request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	authRes, serviceErr := c.services.ConfirmEmail(ctx.UserContext(), request.ConfirmationToken)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return c.processAuthResponse(ctx, authRes)
}

func (c *Controllers) ForgotPassword(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.auth.ForgotPassword")
	userCtx := ctx.UserContext()
	var request ForgotPasswordRequest

	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.Struct(request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}
	if serviceErr := c.services.ForgotPassword(ctx.UserContext(), request.Email); serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.
		Status(fiber.StatusOK).
		JSON(NewMessageResponse("If the email exists, a password reset email has been sent"))
}

func (c *Controllers) ResetPassword(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.auth.ForgotPassword")
	userCtx := ctx.UserContext()
	var request ResetPasswordRequest

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
	if request.Password1 != request.Password2 {
		errMsg := "Passwords do not match"
		log.WarnContext(userCtx, errMsg)
		return c.serviceErrorResponse(services.NewValidationError(errMsg), ctx)
	}

	opts := services.ResetPasswordOptions{
		ResetToken:  request.ResetToken,
		NewPassword: request.Password1,
	}
	if serviceErr := c.services.ResetPassword(userCtx, opts); serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.
		Status(fiber.StatusOK).
		JSON(NewMessageResponse("Password reseted successfully"))
}

func (c *Controllers) UpdatePassword(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.auth.UpdatePassword")
	userCtx := ctx.UserContext()

	userClaims, err := c.GetUserClaims(ctx)
	if err != nil {
		log.ErrorContext(userCtx, "This should be a protected route", "error", err)
		return c.serviceErrorResponse(err, ctx)
	}

	var request UpdatePasswordRequest
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.Struct(request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := passwordValidator(request.Password1); err != nil {
		return c.serviceErrorResponse(err, ctx)
	}
	if request.Password1 != request.Password2 {
		errMsg := "Passwords do not match"
		log.WarnContext(userCtx, errMsg)
		return c.serviceErrorResponse(services.NewValidationError(errMsg), ctx)
	}

	authRes, serviceErr := c.services.UpdatePassword(ctx.UserContext(), services.UpdatePasswordOptions{
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
	log := c.log.WithGroup("controllers.auth.UpdatePassword")
	userCtx := ctx.UserContext()

	userClaims, err := c.GetUserClaims(ctx)
	if err != nil {
		log.ErrorContext(userCtx, "This should be a protected route", "error", err)
		return c.serviceErrorResponse(err, ctx)
	}

	var request UpdateEmailRequest
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.Struct(request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	authRes, serviceErr := c.services.UpdateEmail(ctx.UserContext(), services.UpdateEmailOptions{
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
