package controllers

import (
	"unicode"

	"github.com/gofiber/fiber/v2"
	"github.com/kiwiscript/kiwiscript_go/services"
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
	var request SignUpRequest
	serviceErr := services.NewValidationError("Invalid request")

	if err := ctx.BodyParser(&request); err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestError(serviceErr))
	}
	if err := c.validate.Struct(request); err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestError(serviceErr))
	}
	if err := passwordValidator(request.Password1); err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestError(err))
	}
	if request.Password1 != request.Password2 {
		serviceErr = services.NewValidationError("Password does not match")
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestError(serviceErr))
	}

	serviceErr = c.services.SignUp(ctx.UserContext(), services.SignUpOptions{
		Email:     request.Email,
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Location:  request.Location,
		BirthDate: request.BirthDate,
		Password:  request.Password1,
	})
	if serviceErr != nil {
		return ctx.
			Status(NewRequestErrorStatus(serviceErr.Code)).
			JSON(NewRequestError(serviceErr))
	}

	return ctx.
		Status(fiber.StatusCreated).
		JSON(NewMessageResponse("Confirmation email has been sent"))
}

func (c *Controllers) SignIn(ctx *fiber.Ctx) error {
	var request SignInRequest
	serviceErr := services.NewValidationError("Invalid request")

	if err := ctx.BodyParser(&request); err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestError(serviceErr))
	}
	if err := c.validate.Struct(request); err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestError(serviceErr))
	}

	serviceErr = c.services.SignIn(ctx.UserContext(), services.SignInOptions{
		Email:    request.Email,
		Password: request.Password,
	})
	if serviceErr != nil {
		return ctx.
			Status(NewRequestErrorStatus(serviceErr.Code)).
			JSON(NewRequestError(serviceErr))
	}

	return ctx.
		Status(fiber.StatusOK).
		JSON(NewMessageResponse("Sign in successful"))
}

func (c *Controllers) ConfirmSignIn(ctx *fiber.Ctx) error {
	var request ConfirmSignInRequest
	serviceErr := services.NewValidationError("Invalid request")

	if err := ctx.BodyParser(&request); err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestError(serviceErr))
	}
	if err := c.validate.Struct(request); err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestError(serviceErr))
	}

	authRes, serviceErr := c.services.TwoFactor(ctx.UserContext(), services.TwoFactorOptions{
		Email: request.Email,
		Code:  request.Code,
	})
	if serviceErr != nil {
		return ctx.
			Status(NewRequestErrorStatus(serviceErr.Code)).
			JSON(NewRequestError(serviceErr))
	}

	return c.processAuthResponse(ctx, authRes)
}

func (c *Controllers) SignOut(ctx *fiber.Ctx) error {
	refreshToken := ctx.Cookies(c.refreshCookieName)

	if refreshToken == "" {
		var request SignOutRequest
		serviceErr := services.NewValidationError("Invalid request")

		if err := ctx.BodyParser(&request); err != nil {
			return ctx.
				Status(fiber.StatusBadRequest).
				JSON(NewRequestError(serviceErr))
		}
		if err := c.validate.Struct(request); err != nil {
			return ctx.
				Status(fiber.StatusBadRequest).
				JSON(NewRequestError(serviceErr))
		}

		refreshToken = request.RefreshToken
	}
	if serviceErr := c.services.SignOut(ctx.UserContext(), refreshToken); serviceErr != nil {
		return ctx.
			Status(NewRequestErrorStatus(serviceErr.Code)).
			JSON(NewRequestError(serviceErr))
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

func (c *Controllers) Refresh(ctx *fiber.Ctx) error {
	refreshToken := ctx.Cookies(c.refreshCookieName)

	if refreshToken == "" {
		var request RefreshRequest
		serviceErr := services.NewValidationError("Invalid request")

		if err := ctx.BodyParser(&request); err != nil {
			return ctx.
				Status(fiber.StatusBadRequest).
				JSON(NewRequestError(serviceErr))
		}
		if err := c.validate.Struct(request); err != nil {
			return ctx.
				Status(fiber.StatusBadRequest).
				JSON(NewRequestError(serviceErr))
		}

		refreshToken = request.RefreshToken
	}

	authRes, serviceErr := c.services.Refresh(ctx.UserContext(), refreshToken)
	if serviceErr != nil {
		return ctx.
			Status(NewRequestErrorStatus(serviceErr.Code)).
			JSON(NewRequestError(serviceErr))
	}

	return c.processAuthResponse(ctx, authRes)
}

func (c *Controllers) ConfirmEmail(ctx *fiber.Ctx) error {
	var request ConfirmRequest
	serviceErr := services.NewValidationError("Invalid request")

	if err := ctx.BodyParser(&request); err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestError(serviceErr))
	}
	if err := c.validate.Struct(request); err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestError(serviceErr))
	}

	authRes, serviceErr := c.services.ConfirmEmail(ctx.UserContext(), request.ConfirmationToken)
	if serviceErr != nil {
		return ctx.
			Status(NewRequestErrorStatus(serviceErr.Code)).
			JSON(NewRequestError(serviceErr))
	}

	return c.processAuthResponse(ctx, authRes)
}

func (c *Controllers) ForgotPassword(ctx *fiber.Ctx) error {
	var request ForgotPasswordRequest
	serviceErr := services.NewValidationError("Invalid request")

	if err := ctx.BodyParser(&request); err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestError(serviceErr))
	}
	if err := c.validate.Struct(request); err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestError(serviceErr))
	}
	if serviceErr = c.services.ForgotPassword(ctx.UserContext(), request.Email); serviceErr != nil {
		return ctx.
			Status(NewRequestErrorStatus(serviceErr.Code)).
			JSON(NewRequestError(serviceErr))
	}

	return ctx.
		Status(fiber.StatusOK).
		JSON(NewMessageResponse("If the email exists, a password reset email has been sent"))
}

func (c *Controllers) ResetPassword(ctx *fiber.Ctx) error {
	var request ResetPasswordRequest
	serviceErr := services.NewValidationError("Invalid request")

	if err := ctx.BodyParser(&request); err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestError(serviceErr))
	}
	if err := c.validate.Struct(request); err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestError(serviceErr))
	}
	if serviceErr = passwordValidator(request.Password1); serviceErr != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestError(serviceErr))
	}
	if request.Password1 != request.Password2 {
		serviceErr = services.NewValidationError("Password does not match")
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestError(serviceErr))
	}
	if serviceErr = c.services.ResetPassword(ctx.UserContext(), services.ResetPasswordOptions{
		ResetToken:  request.ResetToken,
		NewPassword: request.Password1,
	}); serviceErr != nil {
		return ctx.
			Status(NewRequestErrorStatus(serviceErr.Code)).
			JSON(NewRequestError(serviceErr))
	}

	return ctx.
		Status(fiber.StatusOK).
		JSON(NewMessageResponse("Password reset successful"))
}

func (c *Controllers) UpdatePassword(ctx *fiber.Ctx) error {
	userClaims, err := c.GetUserClaims(ctx)
	if err != nil {
		return ctx.
			Status(NewRequestErrorStatus(err.Code)).
			JSON(NewRequestError(err))
	}

	var request UpdatePasswordRequest
	serviceErr := services.NewValidationError("Invalid request")

	if err := ctx.BodyParser(&request); err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestError(serviceErr))
	}
	if err := c.validate.Struct(request); err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestError(serviceErr))
	}
	if serviceErr = passwordValidator(request.Password1); serviceErr != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestError(serviceErr))
	}
	if request.Password1 != request.Password2 {
		serviceErr = services.NewValidationError("Password does not match")
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestError(serviceErr))
	}

	authRes, serviceErr := c.services.UpdatePassword(ctx.UserContext(), services.UpdatePasswordOptions{
		UserID:      userClaims.ID,
		OldPassword: request.OldPassword,
		NewPassword: request.Password1,
	})
	if serviceErr != nil {
		return ctx.
			Status(NewRequestErrorStatus(serviceErr.Code)).
			JSON(NewRequestError(serviceErr))
	}

	return c.processAuthResponse(ctx, authRes)
}

func (c *Controllers) UpdateEmail(ctx *fiber.Ctx) error {
	userClaims, err := c.GetUserClaims(ctx)
	if err != nil {
		return ctx.
			Status(NewRequestErrorStatus(err.Code)).
			JSON(NewRequestError(err))
	}

	var request UpdateEmailRequest
	serviceErr := services.NewValidationError("Invalid request")

	if err := ctx.BodyParser(&request); err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestError(serviceErr))
	}
	if err := c.validate.Struct(request); err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestError(serviceErr))
	}

	authRes, serviceErr := c.services.UpdateEmail(ctx.UserContext(), services.UpdateEmailOptions{
		UserID:      userClaims.ID,
		UserVersion: userClaims.Version,
		NewEmail:    request.Email,
		Password:    request.Password,
	})
	if serviceErr != nil {
		return ctx.
			Status(NewRequestErrorStatus(serviceErr.Code)).
			JSON(NewRequestError(serviceErr))
	}

	return c.processAuthResponse(ctx, authRes)
}
