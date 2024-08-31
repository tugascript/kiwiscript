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

package services

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	cc "github.com/kiwiscript/kiwiscript_go/providers/cache"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/providers/email"
	"github.com/kiwiscript/kiwiscript_go/providers/tokens"
	"github.com/kiwiscript/kiwiscript_go/utils"
	"log/slog"
	"strings"
)

const authLocation string = "auth"

type SignUpOptions struct {
	RequestID string
	Email     string
	FirstName string
	LastName  string
	Location  string
	Password  string
}

func (s *Services) sendConfirmationEmail(
	ctx context.Context,
	log *slog.Logger,
	requestID string,
	user *db.User,
) *ServiceError {
	log.InfoContext(ctx, "Sending confirmation email")
	confirmationToken, err := s.jwt.CreateEmailToken(tokens.EmailTokenConfirmation, user)
	if err != nil {
		log.ErrorContext(ctx, "Failed to create confirmation token", "error", err)
		return NewServerError()
	}

	go func() {
		opts := email.ConfirmationEmailOptions{
			RequestID:         requestID,
			Email:             user.Email,
			FirstName:         user.FirstName,
			LastName:          user.LastName,
			ConfirmationToken: confirmationToken,
		}
		if err := s.mail.SendConfirmationEmail(ctx, opts); err != nil {
			log.WarnContext(ctx, "Failed to send confirmation email", "error", err)
		}
	}()

	return nil
}

func (s *Services) SignUp(ctx context.Context, opts SignUpOptions) *ServiceError {
	log := s.buildLogger(opts.RequestID, authLocation, "SignUp").With(
		"firstName", opts.FirstName,
		"lastName", opts.LastName,
	)
	log.InfoContext(ctx, "Signing up...")

	password, err := utils.HashPassword(opts.Password)
	if err != nil {
		log.ErrorContext(ctx, "Failed to hash password", "error", err)
		return NewServerError()
	}

	prms := db.FindAuthProviderByEmailAndProviderParams{
		Email:    opts.Email,
		Provider: utils.ProviderEmail,
	}
	if _, err := s.database.FindAuthProviderByEmailAndProvider(ctx, prms); err == nil {
		log.WarnContext(ctx, "Email already in use")
		return NewConflictError("Email already exists")
	}

	user, serviceErr := s.CreateUser(ctx, CreateUserOptions{
		RequestID: opts.RequestID,
		Email:     opts.Email,
		FirstName: opts.FirstName,
		LastName:  opts.LastName,
		Location:  opts.Location,
		Password:  password,
		Provider:  utils.ProviderEmail,
	})
	if serviceErr != nil {
		log.WarnContext(ctx, "Failed to create user", "error", serviceErr)
		return serviceErr
	}

	if err := s.sendConfirmationEmail(ctx, log, opts.RequestID, user); err != nil {
		return err
	}
	log.InfoContext(ctx, "Sign up successfully")
	return nil
}

type AuthResponse struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

func (s *Services) generateAuthResponse(
	ctx context.Context,
	log *slog.Logger,
	successMsg string,
	user *db.User,
) (*AuthResponse, *ServiceError) {
	accessToken, err := s.jwt.CreateAccessToken(user)
	if err != nil {
		log.ErrorContext(ctx, "Failed to create access token", "error", err)
		return nil, NewServerError()
	}

	refreshToken, err := s.jwt.CreateRefreshToken(user)
	if err != nil {
		log.ErrorContext(ctx, "Failed to create refresh token", "error", err)
		return nil, NewServerError()
	}

	response := AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.jwt.GetAccessTtl(),
	}
	log.InfoContext(ctx, successMsg)
	return &response, nil
}

type ConfirmEmailOptions struct {
	RequestID string
	Token     string
}

func (s *Services) ConfirmEmail(ctx context.Context, opts ConfirmEmailOptions) (*AuthResponse, *ServiceError) {
	log := s.buildLogger(opts.RequestID, authLocation, "ConfirmEmail")
	log.InfoContext(ctx, "Confirming email...")

	tokenType, claims, err := s.jwt.VerifyEmailToken(opts.Token)
	if err != nil {
		log.WarnContext(ctx, "Invalid token", "error", err)
		return nil, NewUnauthorizedError()
	}

	if tokenType != tokens.EmailTokenConfirmation {
		log.WarnContext(ctx, "Invalid token type")
		return nil, NewUnauthorizedError()
	}

	user, serviceErr := s.FindUserByID(ctx, FindUserByIDOptions{
		RequestID: opts.RequestID,
		ID:        claims.ID,
	})
	if serviceErr != nil {
		log.WarnContext(ctx, "User not found", "error", serviceErr)
		return nil, NewUnauthorizedError()
	}
	if user.IsConfirmed {
		log.WarnContext(ctx, "User already confirmed")
		return nil, NewUnauthorizedError()
	}
	if user.Version != claims.Version {
		log.WarnContext(ctx, "Invalid token version")
		return nil, NewUnauthorizedError()
	}

	user, serviceErr = s.ConfirmUser(ctx, ConfirmUserOptions{
		RequestID: opts.RequestID,
		ID:        user.ID,
	})
	if serviceErr != nil {
		log.ErrorContext(ctx, "Failed to confirm user", "error", serviceErr)
		return nil, serviceErr
	}

	return s.generateAuthResponse(ctx, log, "Confirmed email successfully", user)
}

type SignInOptions struct {
	RequestID string
	Email     string
	Password  string
}

func (s *Services) SignIn(ctx context.Context, opts SignInOptions) *ServiceError {
	log := s.buildLogger(opts.RequestID, authLocation, "SignIn")
	log.InfoContext(ctx, "Signing in...")

	prms := db.FindAuthProviderByEmailAndProviderParams{
		Email:    opts.Email,
		Provider: utils.ProviderEmail,
	}
	if _, err := s.database.FindAuthProviderByEmailAndProvider(ctx, prms); err != nil {
		log.WarnContext(ctx, "Failed to find auth provider", "error", err)
		return NewUnauthorizedError()
	}

	user, serviceErr := s.FindUserByEmail(ctx, FindUserByEmailOptions{
		RequestID: opts.RequestID,
		Email:     opts.Email,
	})
	if serviceErr != nil {
		log.WarnContext(ctx, "Failed to find user", "error", serviceErr)
		return NewUnauthorizedError()
	}

	if !utils.VerifyPassword(opts.Password, user.Password.String) {
		log.WarnContext(ctx, "Invalid password")
		return NewUnauthorizedError()
	}
	if !user.IsConfirmed {
		log.WarnContext(ctx, "User still not confirmed, sending confirmation email")

		if err := s.sendConfirmationEmail(ctx, log, opts.RequestID, user); err != nil {
			return err
		}

		return NewValidationError("User not confirmed")
	}

	code, err := s.cache.AddTwoFactorCode(ctx, cc.AddTwoFactorCodeOptions{
		RequestID: opts.RequestID,
		UserID:    user.ID,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to generate two factor code", "error", serviceErr)
		return NewServerError()
	}

	go func() {
		opts := email.CodeEmailOptions{
			RequestID: opts.RequestID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Code:      code,
		}
		if err := s.mail.SendCodeEmail(ctx, opts); err != nil {
			log.ErrorContext(ctx, "Failed to send two factor email", "error", err)
		}
	}()

	log.InfoContext(ctx, "Sign in successful")
	return nil
}

type TwoFactorOptions struct {
	RequestID string
	Email     string
	Code      string
}

func (s *Services) TwoFactor(ctx context.Context, opts TwoFactorOptions) (*AuthResponse, *ServiceError) {
	log := s.buildLogger(opts.RequestID, authLocation, "TwoFactor")
	log.InfoContext(ctx, "Confirming two factor code...")

	user, serviceErr := s.FindUserByEmail(ctx, FindUserByEmailOptions{
		RequestID: opts.RequestID,
		Email:     opts.Email,
	})
	if serviceErr != nil {
		log.WarnContext(ctx, "Failed to find user", "error", serviceErr)
		return nil, NewUnauthorizedError()
	}

	verified, err := s.cache.VerifyTwoFactorCode(ctx, cc.VerifyTwoFactorCodeOptions{
		RequestID: opts.RequestID,
		UserID:    user.ID,
		Code:      opts.Code,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to verify two factor code", "error", err)
		return nil, NewServerError()
	}
	if !verified {
		log.WarnContext(ctx, "Invalid two factor code")
		return nil, NewUnauthorizedError()
	}
	if !user.IsConfirmed {
		errMsg := "User not confirmed"
		log.WarnContext(ctx, errMsg)
		return nil, NewValidationError(errMsg)
	}

	return s.generateAuthResponse(ctx, log, "Confirmed two factor successfully", user)
}

type RefreshOptions struct {
	RequestID string
	Token     string
}

func (s *Services) Refresh(ctx context.Context, opts RefreshOptions) (*AuthResponse, *ServiceError) {
	log := s.buildLogger(opts.RequestID, authLocation, "Refresh")
	log.InfoContext(ctx, "Refreshing access token...")

	claims, id, _, err := s.jwt.VerifyRefreshToken(opts.Token)
	if err != nil {
		log.WarnContext(ctx, "Invalid token", "error", err)
		return nil, NewUnauthorizedError()
	}

	isBl, err := s.cache.IsBlacklisted(ctx, cc.IsBlacklistedOptions{
		RequestID: "",
		ID:        id,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to check black list", "error", err)
		return nil, NewServerError()
	}
	if isBl {
		log.WarnContext(ctx, "Token black listed")
		return nil, NewUnauthorizedError()
	}

	user, serviceErr := s.FindUserByID(ctx, FindUserByIDOptions{
		RequestID: opts.RequestID,
		ID:        claims.ID,
	})
	if serviceErr != nil {
		log.WarnContext(ctx, "Failed to find user", "error", err)
		return nil, NewUnauthorizedError()
	}
	if user.Version != claims.Version {
		log.WarnContext(ctx, "Invalid token version")
		return nil, NewUnauthorizedError()
	}

	return s.generateAuthResponse(ctx, log, "Refreshed token successfully", user)
}

type SignOutOptions struct {
	RequestID string
	Token     string
}

func (s *Services) SignOut(ctx context.Context, opts SignOutOptions) *ServiceError {
	log := s.buildLogger(opts.RequestID, authLocation, "SignOut")
	log.InfoContext(ctx, "Signing out...")

	_, id, exp, err := s.jwt.VerifyRefreshToken(opts.Token)
	if err != nil {
		log.WarnContext(ctx, "Invalid token", "error", err)
		return NewUnauthorizedError()
	}

	err = s.cache.AddBlacklist(ctx, cc.AddBlacklistOptions{
		ID:  id,
		Exp: exp,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to add black list", "error", err)
		return NewServerError()
	}

	log.InfoContext(ctx, "Signed out successfully")
	return nil
}

type UpdatePasswordOptions struct {
	RequestID   string
	UserID      int32
	UserVersion int16
	OldPassword string
	NewPassword string
}

func (s *Services) UpdatePassword(ctx context.Context, opts UpdatePasswordOptions) (*AuthResponse, *ServiceError) {
	log := s.buildLogger(opts.RequestID, authLocation, "UpdatePassword").With(
		"userId", opts.UserID,
		"userVersion", opts.UserVersion,
	)
	log.InfoContext(ctx, "Updating password...")

	user, serviceErr := s.FindUserByID(ctx, FindUserByIDOptions{
		RequestID: opts.RequestID,
		ID:        opts.UserID,
	})
	if serviceErr != nil {
		log.WarnContext(ctx, "Failed to find user", "error", serviceErr)
		return nil, NewUnauthorizedError()
	}
	if user.Version != opts.UserVersion {
		log.WarnContext(ctx, "Invalid user version")
		return nil, NewUnauthorizedError()
	}

	authProviderParams := db.FindAuthProviderByEmailAndProviderParams{
		Email:    user.Email,
		Provider: utils.ProviderEmail,
	}
	if _, err := s.database.FindAuthProviderByEmailAndProvider(ctx, authProviderParams); err != nil {
		if serviceErr := FromDBError(err); serviceErr.Code != CodeNotFound {
			log.ErrorContext(ctx, "Failed to find auth provider", "error", err)
			return nil, serviceErr
		}

		qrs, txn, err := s.database.BeginTx(ctx)
		if err != nil {
			log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
			return nil, FromDBError(err)
		}
		defer s.database.FinalizeTx(ctx, txn, err, serviceErr)

		err = qrs.CreateAuthProvider(ctx, db.CreateAuthProviderParams{
			Email:    user.Email,
			Provider: utils.ProviderEmail,
		})
		if err != nil {
			log.ErrorContext(ctx, "Failed to create auth provider", "error", err)
			return nil, FromDBError(err)
		}

		passwordHash, err := utils.HashPassword(opts.NewPassword)
		if err != nil {
			log.ErrorContext(ctx, "Failed to hash new password", "error", err)
			return nil, NewServerError()
		}

		var password pgtype.Text
		if err = password.Scan(passwordHash); err != nil {
			log.ErrorContext(ctx, "Failed to scan password", "error", err)
			return nil, NewServerError()
		}

		*user, err = qrs.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
			ID:       opts.UserID,
			Password: password,
		})
		if err != nil {
			log.ErrorContext(ctx, "Failed to update password", "error", err)
			return nil, FromDBError(err)
		}

		return s.generateAuthResponse(ctx, log, "Updated password successfully", user)
	}

	if !utils.VerifyPassword(opts.OldPassword, user.Password.String) {
		errMsg := "Old password is incorrect"
		log.WarnContext(ctx, errMsg)
		return nil, NewValidationError(errMsg)
	}

	password, err := utils.HashPassword(opts.NewPassword)
	if err != nil {
		log.ErrorContext(ctx, "Failed to hash new password", "error", err)
		return nil, NewServerError()
	}

	user, serviceErr = s.UpdateUserPassword(ctx, UpdateUserPasswordOptions{
		RequestID: opts.RequestID,
		ID:        opts.UserID,
		Password:  password,
	})
	if serviceErr != nil {
		log.ErrorContext(ctx, "Failed to update password", "error", serviceErr)
		return nil, serviceErr
	}

	return s.generateAuthResponse(ctx, log, "Updated password successfully", user)
}

type ForgotPasswordOptions struct {
	RequestID string
	Email     string
}

func (s *Services) ForgotPassword(ctx context.Context, opts ForgotPasswordOptions) *ServiceError {
	log := s.buildLogger(opts.RequestID, authLocation, "ForgotPassword")
	log.InfoContext(ctx, "Reset password")

	user, serviceErr := s.FindUserByEmail(ctx, FindUserByEmailOptions{
		RequestID: opts.RequestID,
		Email:     opts.Email,
	})
	if serviceErr != nil {
		if serviceErr.Code == CodeNotFound {
			log.InfoContext(ctx, "User not found, skip reset password")
			return nil
		}

		log.WarnContext(ctx, "Failed to find user", "error", serviceErr)
		return serviceErr
	}

	emailToken, err := s.jwt.CreateEmailToken(tokens.EmailTokenReset, user)
	if err != nil {
		log.ErrorContext(ctx, "Failed to create email token", "error", err)
		return NewServerError()
	}

	go func() {
		opts := email.ResetEmailOptions{
			RequestID:  opts.RequestID,
			Email:      user.Email,
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			ResetToken: emailToken,
		}
		if err := s.mail.SendResetEmail(ctx, opts); err != nil {
			log.ErrorContext(ctx, "Failed to send two factor email", "error", err)
		}
	}()

	log.InfoContext(ctx, "Reset password successfully")
	return nil
}

type ResetPasswordOptions struct {
	RequestID   string
	ResetToken  string
	NewPassword string
}

func (s *Services) ResetPassword(ctx context.Context, opts ResetPasswordOptions) *ServiceError {
	log := s.buildLogger(opts.RequestID, authLocation, "ResetPassword")
	log.InfoContext(ctx, "Resetting password...")

	tokenType, claims, err := s.jwt.VerifyEmailToken(opts.ResetToken)
	if err != nil {
		log.WarnContext(ctx, "Invalid token", "error", err)
		return NewUnauthorizedError()
	}

	if tokenType != tokens.EmailTokenReset {
		log.WarnContext(ctx, "Invalid token type")
		return NewUnauthorizedError()
	}

	user, serviceErr := s.FindUserByID(ctx, FindUserByIDOptions{
		RequestID: opts.RequestID,
		ID:        claims.ID,
	})
	if serviceErr != nil {
		log.WarnContext(ctx, "Failed to find user", "error", serviceErr)
		return NewUnauthorizedError()
	}
	if user.Version != claims.Version {
		log.WarnContext(ctx, "Invalid token version")
		return NewUnauthorizedError()
	}

	passwordHash, err := utils.HashPassword(opts.NewPassword)
	if err != nil {
		log.ErrorContext(ctx, "Failed to hash new password", "error", err)
		return NewServerError()
	}

	// TODO: fix, create email oauth provider if it does not exist
	authProviderParams := db.FindAuthProviderByEmailAndProviderParams{
		Email:    user.Email,
		Provider: utils.ProviderEmail,
	}
	if _, err := s.database.FindAuthProviderByEmailAndProvider(ctx, authProviderParams); err != nil {
		if serviceErr := FromDBError(err); serviceErr.Code != CodeNotFound {
			log.ErrorContext(ctx, "Failed to find auth provider", "error", err)
			return serviceErr
		}

		qrs, txn, err := s.database.BeginTx(ctx)
		if err != nil {
			log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
			return FromDBError(err)
		}
		defer s.database.FinalizeTx(ctx, txn, err, serviceErr)

		authProvParams := db.CreateAuthProviderParams{
			Email:    user.Email,
			Provider: utils.ProviderEmail,
		}
		if err := qrs.CreateAuthProvider(ctx, authProvParams); err != nil {
			log.ErrorContext(ctx, "Failed to create auth provider", "error", err)
			return FromDBError(err)
		}

		var password pgtype.Text
		if err := password.Scan(opts.NewPassword); err != nil || opts.NewPassword == "" {
			log.WarnContext(ctx, "Password is invalid")
			return NewValidationError("'password' is invalid")
		}

		passParams := db.UpdateUserPasswordParams{
			Password: password,
			ID:       user.ID,
		}
		if _, err := qrs.UpdateUserPassword(ctx, passParams); err != nil {
			log.ErrorContext(ctx, "Failed to update user password")
			return FromDBError(err)
		}

		log.InfoContext(ctx, "Reset password successful")
		return nil
	}

	updatePassOpts := UpdateUserPasswordOptions{
		RequestID: opts.RequestID,
		ID:        user.ID,
		Password:  passwordHash,
	}
	if _, serviceErr = s.UpdateUserPassword(ctx, updatePassOpts); serviceErr != nil {
		log.ErrorContext(ctx, "Failed to update password", "error", serviceErr)
		return serviceErr
	}

	log.InfoContext(ctx, "Reset password successful")
	return nil
}

type UpdateEmailOptions struct {
	RequestID   string
	UserID      int32
	UserVersion int16
	NewEmail    string
	Password    string
}

func (s *Services) UpdateEmail(ctx context.Context, opts UpdateEmailOptions) (*AuthResponse, *ServiceError) {
	log := s.buildLogger(opts.RequestID, authLocation, "UpdateEmail").With(
		"userId", opts.UserID,
		"userVersion", opts.UserVersion,
	)
	log.InfoContext(ctx, "Updating email...")

	user, serviceErr := s.FindUserByID(ctx, FindUserByIDOptions{
		RequestID: opts.RequestID,
		ID:        opts.UserID,
	})
	if serviceErr != nil {
		log.WarnContext(ctx, "Failed to find user", "error", serviceErr)
		return nil, NewUnauthorizedError()
	}
	if user.Version != opts.UserVersion {
		log.WarnContext(ctx, "Invalid user version")
		return nil, NewUnauthorizedError()
	}

	authProvParams := db.FindAuthProviderByEmailAndProviderParams{
		Email:    user.Email,
		Provider: utils.ProviderEmail,
	}
	if _, err := s.database.FindAuthProviderByEmailAndProvider(ctx, authProvParams); err != nil {
		serviceErr = FromDBError(err)

		if serviceErr.Code == CodeNotFound {
			log.WarnContext(ctx, "Email auth provider not found", "error", err)
			return nil, NewUnauthorizedError()
		}

		log.ErrorContext(ctx, "Failed to find auth provider", "error", err)
		return nil, serviceErr
	}
	if !utils.VerifyPassword(opts.Password, user.Password.String) {
		errMsg := "Invalid password"
		log.WarnContext(ctx, errMsg)
		return nil, NewValidationError(errMsg)
	}

	newAuthProvParams := db.FindAuthProviderByEmailAndProviderParams{
		Email:    opts.NewEmail,
		Provider: utils.ProviderEmail,
	}
	if _, err := s.database.FindAuthProviderByEmailAndProvider(ctx, newAuthProvParams); err == nil {
		log.WarnContext(ctx, "Email already exists")
		return nil, NewValidationError("Email already in use")
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return nil, FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err, serviceErr)

	*user, err = qrs.UpdateUserEmail(ctx, db.UpdateUserEmailParams{
		ID:    user.ID,
		Email: opts.NewEmail,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to update email", "error", err)
		return nil, FromDBError(err)
	}
	if err = qrs.DeleteProviderByEmailAndNotProvider(ctx, db.DeleteProviderByEmailAndNotProviderParams{
		Email:    user.Email,
		Provider: utils.ProviderEmail,
	}); err != nil {
		log.ErrorContext(ctx, "Failed to delete auth provider", "error", err)
		return nil, FromDBError(err)
	}

	return s.generateAuthResponse(ctx, log, "Update email successfully", user)
}

func (s *Services) ProcessAuthHeader(authHeader string) (tokens.AccessUserClaims, *ServiceError) {
	authHeaderSlice := strings.Split(authHeader, " ")
	var userClaims tokens.AccessUserClaims

	if len(authHeaderSlice) != 2 {
		return userClaims, NewUnauthorizedError()
	}
	if strings.ToLower(authHeaderSlice[0]) != "bearer" {
		return userClaims, NewUnauthorizedError()
	}

	userClaims, err := s.jwt.VerifyAccessToken(authHeaderSlice[1])
	if err != nil {
		return userClaims, NewUnauthorizedError()
	}
	return userClaims, nil
}
