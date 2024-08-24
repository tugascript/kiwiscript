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
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/providers/oauth"
	"github.com/kiwiscript/kiwiscript_go/providers/tokens"
	"github.com/kiwiscript/kiwiscript_go/utils"
	"log/slog"
	"strings"
)

func (s *Services) GetAuthorizationURL(ctx context.Context, provider string) (string, *ServiceError) {
	log := s.log.WithGroup("services.oauth.GetAuthorizationURL").With("provider", provider)
	log.InfoContext(ctx, "Getting authorization url")

	var url, state string
	var err error
	switch provider {
	case utils.ProviderGitHub:
		url, state, err = s.oauthProviders.GetGitHubAuthorizationURL(ctx)
	case utils.ProviderGoogle:
		url, state, err = s.oauthProviders.GetGoogleAuthorizationURL(ctx)
	default:
		log.ErrorContext(ctx, "Authorization url must be for 'github' or 'google'")
		return "", NewServerError()
	}
	if err != nil {
		log.ErrorContext(ctx, "Failed to generate state", "error", err)
		return "", NewServerError()
	}

	if err := s.cache.AddOAuthState(state, provider); err != nil {
		log.ErrorContext(ctx, "Failed to cache state", "error", err)
		return "", NewServerError()
	}

	return url, nil
}

type GetOAuthTokenOptions struct {
	Provider string
	Code     string
	State    string
}

func (s *Services) GetOAuthToken(ctx context.Context, opts GetOAuthTokenOptions) (string, *ServiceError) {
	log := s.log.WithGroup("services.oauth.GetOAuthToken")
	log.InfoContext(ctx, "Getting oauth token")

	ok, err := s.cache.VerifyOAuthState(opts.State, opts.Provider)
	if err != nil {
		log.ErrorContext(ctx, "Failed to verify oauth state", "error", err)
		return "", NewServerError()
	}
	if !ok {
		log.WarnContext(ctx, "OAuth state is invalid")
		return "", NewUnauthorizedError()
	}

	var token string
	switch opts.Provider {
	case utils.ProviderGitHub:
		token, err = s.oauthProviders.GetGitHubAccessToken(ctx, opts.Code)
	case utils.ProviderGoogle:
		token, err = s.oauthProviders.GetGoogleAccessToken(ctx, opts.Code)
	default:
		log.ErrorContext(ctx, "Provider must be 'github' or 'google'")
		return "", NewServerError()
	}

	return token, nil
}

func (s *Services) generateEmailCode(ctx context.Context, log *slog.Logger, email string) (string, *ServiceError) {
	log.InfoContext(ctx, "Generating email code...")

	code := utils.Base62UUID()
	if err := s.cache.AddOAuthEmail(code, email); err != nil {
		log.ErrorContext(ctx, "Failed to cache code", "error", err)
		return "", NewServerError()
	}

	return code, nil
}

func (s *Services) generateOAuthResponse(
	ctx context.Context,
	log *slog.Logger,
	user *db.User,
) (*OAuthResponse, *ServiceError) {
	code, serviceErr := s.generateEmailCode(ctx, log, user.Email)
	if serviceErr != nil {
		return nil, serviceErr
	}

	log.InfoContext(ctx, "Generating oauth access token")
	accessToken, err := s.jwt.CreateOAuthToken(user)
	if err != nil {
		log.ErrorContext(ctx, "Failed to create OAuth access token", "error", err)
		return nil, NewServerError()
	}

	response := OAuthResponse{
		AccessToken: accessToken,
		Code:        code,
		ExpiresIn:   s.jwt.GetOAuthTtl(),
	}
	return &response, nil
}

type OAuthResponse struct {
	AccessToken string
	Code        string
	ExpiresIn   int64
}

type ExtOAuthSignInOptions struct {
	Provider string
	Token    string
}

func (s *Services) ExtOAuthSignIn(ctx context.Context, opts ExtOAuthSignInOptions) (*OAuthResponse, *ServiceError) {
	log := s.log.WithGroup("services.oauth.ExtOAuthSignIn")
	log.InfoContext(ctx, "Generating internal code and state...")

	var toUserData oauth.ToUserData
	var status int
	var err error
	switch opts.Provider {
	case utils.ProviderGitHub:
		toUserData, status, err = s.oauthProviders.GetGitHubUserData(ctx, opts.Token)
	case utils.ProviderGoogle:
		toUserData, status, err = s.oauthProviders.GetGoogleUserData(ctx, opts.Token)
	default:
		log.ErrorContext(ctx, "Provider must be 'github' or 'google'")
		return nil, NewServerError()
	}

	if err != nil {
		if status > 0 && status < 500 {
			log.WarnContext(ctx, "User data got non 200 status code", "error", err, "status", status)
			return nil, NewUnauthorizedError()
		}

		log.ErrorContext(ctx, "Failed to fetch userData data", "error", err)
		return nil, NewServerError()
	}

	userData := toUserData.ToUserData()
	user, serviceErr := s.FindUserByEmail(ctx, userData.Email)
	if serviceErr != nil {
		if serviceErr.Code != CodeNotFound {
			log.ErrorContext(ctx, "Failed to find user by email", "error", serviceErr)
			return nil, NewServerError()
		}

		user, serviceErr := s.CreateUser(ctx, CreateUserOptions{
			FirstName: userData.FirstName,
			LastName:  userData.LastName,
			Location:  userData.Location,
			Email:     userData.Email,
			Provider:  opts.Provider,
			Password:  "",
		})
		if serviceErr != nil {
			log.ErrorContext(ctx, "Failed to create user", "error", serviceErr)
			return nil, NewServerError()
		}

		return s.generateOAuthResponse(ctx, log, user)
	}

	findProvPrms := db.FindAuthProviderByEmailAndProviderParams{
		Email:    user.Email,
		Provider: opts.Provider,
	}
	if _, err := s.database.FindAuthProviderByEmailAndProvider(ctx, findProvPrms); err != nil {
		serviceErr := FromDBError(err)
		if serviceErr.Code != CodeNotFound {
			log.ErrorContext(ctx, "Failed to find auth provider", "error", err)
			return nil, serviceErr
		}

		createProvPrms := db.CreateAuthProviderParams{
			Email:    userData.Email,
			Provider: opts.Provider,
		}
		if err := s.database.CreateAuthProvider(ctx, createProvPrms); err != nil {
			log.ErrorContext(ctx, "Failed to create auth provider", "error", err)
			return nil, FromDBError(err)
		}
	}

	return s.generateOAuthResponse(ctx, log, user)
}

func (s *Services) ProcessOAuthHeader(ctx context.Context, authHeader string) (*tokens.OAuthUserClaims, *ServiceError) {
	log := s.log.WithGroup("services.oauth.ProcessOAuthHeader")
	log.InfoContext(ctx, "Processing OAuth authentication header...")

	if authHeader == "" {
		log.WarnContext(ctx, "OAuth authentication header is empty")
		return nil, NewUnauthorizedError()
	}

	authHeaderSlice := strings.Split(authHeader, " ")
	if len(authHeaderSlice) != 2 {
		log.WarnContext(ctx, "OAuth authentication header is invalid", "authHeader", authHeader)
		return nil, NewUnauthorizedError()
	}

	tokenType, accessToken := authHeaderSlice[0], authHeaderSlice[1]
	if strings.ToLower(tokenType) != "bearer" {
		log.WarnContext(ctx, "OAuth token type is not Bearer", "tokenType", tokenType)
		return nil, NewUnauthorizedError()
	}

	userClaims, err := s.jwt.VerifyOAuthToken(accessToken)
	if err != nil {
		log.WarnContext(ctx, "Failed to verify OAuth access token", "error", err)
		return nil, NewUnauthorizedError()
	}

	return &userClaims, nil
}

type IntOAuthSignInOptions struct {
	UserID      int32
	UserVersion int16
	Code        string
}

func (s *Services) OAuthToken(ctx context.Context, opts IntOAuthSignInOptions) (*AuthResponse, *ServiceError) {
	log := s.log.WithGroup("services.oauth.OAuthToken")
	log.InfoContext(ctx, "Sign in the user with a local token...")

	email, err := s.cache.GetOAuthEmail(opts.Code)
	if err != nil {
		log.ErrorContext(ctx, "Failed to fetch email by code")
		return nil, NewServerError()
	}

	if email == "" {
		log.WarnContext(ctx, "Email does not exist in cache")
		return nil, NewUnauthorizedError()
	}

	user, serviceErr := s.FindUserByEmail(ctx, email)
	if serviceErr != nil {
		if serviceErr.Code == CodeNotFound {
			return nil, NewUnauthorizedError()
		}

		return nil, serviceErr
	}

	if user.ID != opts.UserID || user.Version != opts.UserVersion {
		log.WarnContext(
			ctx,
			"User id or version do not match the token's user id or version",
			"userId", user.ID,
			"userVersion", user.Version,
			"tokenUserId", opts.UserID,
			"tokenUserVersion", opts.UserVersion,
		)
		return nil, NewUnauthorizedError()
	}

	return s.generateAuthResponse(ctx, log, "User OAuth signed in successfully", user)
}
