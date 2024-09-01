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
	"github.com/kiwiscript/kiwiscript_go/exceptions"
	cc "github.com/kiwiscript/kiwiscript_go/providers/cache"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/providers/oauth"
	"github.com/kiwiscript/kiwiscript_go/providers/tokens"
	"github.com/kiwiscript/kiwiscript_go/utils"
	"log/slog"
	"strings"
)

const oauthLocation string = "oauth"

type GetAuthorizationURLOptions struct {
	RequestID string
	Provider  string
}

func (s *Services) GetAuthorizationURL(ctx context.Context, opts GetAuthorizationURLOptions) (string, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, oauthLocation, "GetAuthorizationURL")
	log.InfoContext(ctx, "Getting authorization url")

	var url, state string
	var err error
	switch opts.Provider {
	case utils.ProviderGitHub:
		url, state, err = s.oauthProviders.GetGitHubAuthorizationURL(ctx, opts.RequestID)
	case utils.ProviderGoogle:
		url, state, err = s.oauthProviders.GetGoogleAuthorizationURL(ctx, opts.RequestID)
	default:
		log.ErrorContext(ctx, "Authorization url must be for 'github' or 'google'")
		return "", exceptions.NewServerError()
	}
	if err != nil {
		log.ErrorContext(ctx, "Failed to generate state", "error", err)
		return "", exceptions.NewServerError()
	}

	stateOpts := cc.AddOAuthStateOptions{
		RequestID: opts.RequestID,
		State:     state,
		Provider:  opts.Provider,
	}
	if err := s.cache.AddOAuthState(ctx, stateOpts); err != nil {
		log.ErrorContext(ctx, "Failed to cache state", "error", err)
		return "", exceptions.NewServerError()
	}

	return url, nil
}

type GetOAuthTokenOptions struct {
	RequestID string
	Provider  string
	Code      string
	State     string
}

func (s *Services) GetOAuthToken(ctx context.Context, opts GetOAuthTokenOptions) (string, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, oauthLocation, "GetOAuthToken")
	log.InfoContext(ctx, "Getting oauth token")

	ok, err := s.cache.VerifyOAuthState(ctx, cc.VerifyOAuthStateOptions{
		RequestID: opts.RequestID,
		State:     opts.State,
		Provider:  opts.Provider,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to verify oauth state", "error", err)
		return "", exceptions.NewServerError()
	}
	if !ok {
		log.WarnContext(ctx, "OAuth state is invalid")
		return "", exceptions.NewUnauthorizedError()
	}

	var token string
	switch opts.Provider {
	case utils.ProviderGitHub:
		token, err = s.oauthProviders.GetGitHubAccessToken(ctx, oauth.GetGitHubAccessTokenOptions{
			RequestID: opts.RequestID,
			Code:      opts.Code,
		})
	case utils.ProviderGoogle:
		token, err = s.oauthProviders.GetGoogleAccessToken(ctx, oauth.GetGoogleAccessTokenOptions{
			RequestID: opts.RequestID,
			Code:      opts.Code,
		})
	default:
		log.ErrorContext(ctx, "Provider must be 'github' or 'google'")
		return "", exceptions.NewServerError()
	}

	if err != nil {
		log.WarnContext(ctx, "Failed to get oauth access token", "error", err)
		return "", exceptions.NewUnauthorizedError()
	}

	return token, nil
}

func (s *Services) generateEmailCode(
	ctx context.Context,
	log *slog.Logger,
	requestId,
	email string,
) (string, *exceptions.ServiceError) {
	log.InfoContext(ctx, "Generating email code...")

	code := utils.Base62UUID()
	oauthEmailOpts := cc.AddOAuthEmailOptions{
		RequestID:       requestId,
		Code:            code,
		Email:           email,
		DurationSeconds: s.jwt.GetOAuthTtl(),
	}
	if err := s.cache.AddOAuthEmail(ctx, oauthEmailOpts); err != nil {
		log.ErrorContext(ctx, "Failed to cache code", "error", err)
		return "", exceptions.NewServerError()
	}

	return code, nil
}

func (s *Services) generateOAuthResponse(
	ctx context.Context,
	log *slog.Logger,
	user *db.User,
	requestID string,
) (*OAuthResponse, *exceptions.ServiceError) {
	code, serviceErr := s.generateEmailCode(ctx, log, requestID, user.Email)
	if serviceErr != nil {
		return nil, serviceErr
	}

	log.InfoContext(ctx, "Generating oauth access token")
	accessToken, err := s.jwt.CreateOAuthToken(user)
	if err != nil {
		log.ErrorContext(ctx, "Failed to create OAuth access token", "error", err)
		return nil, exceptions.NewServerError()
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
	RequestID string
	Provider  string
	Token     string
}

func (s *Services) ExtOAuthSignIn(ctx context.Context, opts ExtOAuthSignInOptions) (*OAuthResponse, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, oauthLocation, "ExtOAuthSignIn")
	log.InfoContext(ctx, "Generating internal code and state...")

	var toUserData oauth.ToUserData
	var status int
	var err error
	switch opts.Provider {
	case utils.ProviderGitHub:
		toUserData, status, err = s.oauthProviders.GetGitHubUserData(ctx, oauth.GetGitHubUserDataOptions{
			RequestID: opts.RequestID,
			Token:     opts.Token,
		})
	case utils.ProviderGoogle:
		toUserData, status, err = s.oauthProviders.GetGoogleUserData(ctx, oauth.GetGoogleUserDataOptions{
			RequestID: opts.RequestID,
			Token:     opts.Token,
		})
	default:
		log.ErrorContext(ctx, "Provider must be 'github' or 'google'")
		return nil, exceptions.NewServerError()
	}

	if err != nil {
		if status > 0 && status < 500 {
			log.WarnContext(ctx, "User data got non 200 status code", "error", err, "status", status)
			return nil, exceptions.NewUnauthorizedError()
		}

		log.ErrorContext(ctx, "Failed to fetch userData data", "error", err)
		return nil, exceptions.NewServerError()
	}

	userData := toUserData.ToUserData()
	user, serviceErr := s.FindUserByEmail(ctx, FindUserByEmailOptions{
		RequestID: opts.RequestID,
		Email:     userData.Email,
	})
	if serviceErr != nil {
		if serviceErr.Code != exceptions.CodeNotFound {
			log.ErrorContext(ctx, "Failed to find user by email", "error", serviceErr)
			return nil, exceptions.NewServerError()
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
			return nil, exceptions.NewServerError()
		}

		return s.generateOAuthResponse(ctx, log, user, opts.RequestID)
	}

	findProvPrms := db.FindAuthProviderByEmailAndProviderParams{
		Email:    user.Email,
		Provider: opts.Provider,
	}
	if _, err := s.database.FindAuthProviderByEmailAndProvider(ctx, findProvPrms); err != nil {
		serviceErr := exceptions.FromDBError(err)
		if serviceErr.Code != exceptions.CodeNotFound {
			log.ErrorContext(ctx, "Failed to find auth provider", "error", err)
			return nil, serviceErr
		}

		createProvPrms := db.CreateAuthProviderParams{
			Email:    userData.Email,
			Provider: opts.Provider,
		}
		if err := s.database.CreateAuthProvider(ctx, createProvPrms); err != nil {
			log.ErrorContext(ctx, "Failed to create auth provider", "error", err)
			return nil, exceptions.FromDBError(err)
		}
	}

	return s.generateOAuthResponse(ctx, log, user, opts.RequestID)
}

func (s *Services) ProcessOAuthHeader(ctx context.Context, authHeader string) (*tokens.OAuthUserClaims, *exceptions.ServiceError) {
	log := s.log.WithGroup("services.oauth.ProcessOAuthHeader")
	log.InfoContext(ctx, "Processing OAuth authentication header...")

	if authHeader == "" {
		log.WarnContext(ctx, "OAuth authentication header is empty")
		return nil, exceptions.NewUnauthorizedError()
	}

	authHeaderSlice := strings.Split(authHeader, " ")
	if len(authHeaderSlice) != 2 {
		log.WarnContext(ctx, "OAuth authentication header is invalid", "authHeader", authHeader)
		return nil, exceptions.NewUnauthorizedError()
	}

	tokenType, accessToken := authHeaderSlice[0], authHeaderSlice[1]
	if strings.ToLower(tokenType) != "bearer" {
		log.WarnContext(ctx, "OAuth token type is not Bearer", "tokenType", tokenType)
		return nil, exceptions.NewUnauthorizedError()
	}

	userClaims, err := s.jwt.VerifyOAuthToken(accessToken)
	if err != nil {
		log.WarnContext(ctx, "Failed to verify OAuth access token", "error", err)
		return nil, exceptions.NewUnauthorizedError()
	}

	return &userClaims, nil
}

type IntOAuthSignInOptions struct {
	RequestID   string
	UserID      int32
	UserVersion int16
	Code        string
}

func (s *Services) OAuthToken(ctx context.Context, opts IntOAuthSignInOptions) (*AuthResponse, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, oauthLocation, "OAuthToken").With(
		"tokenUserId", opts.UserID,
		"tokenUserVersion", opts.UserVersion,
	)
	log.InfoContext(ctx, "Sign in the user with a local token...")

	email, err := s.cache.GetOAuthEmail(ctx, cc.GetOAuthEmailOptions{
		RequestID: opts.RequestID,
		Code:      opts.Code,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to fetch email by code")
		return nil, exceptions.NewServerError()
	}

	if email == "" {
		log.WarnContext(ctx, "Email does not exist in cache")
		return nil, exceptions.NewUnauthorizedError()
	}

	user, serviceErr := s.FindUserByEmail(ctx, FindUserByEmailOptions{
		RequestID: opts.RequestID,
		Email:     email,
	})
	if serviceErr != nil {
		if serviceErr.Code == exceptions.CodeNotFound {
			return nil, exceptions.NewUnauthorizedError()
		}

		return nil, serviceErr
	}

	if user.ID != opts.UserID || user.Version != opts.UserVersion {
		log.WarnContext(
			ctx,
			"User id or version do not match the token's user id or version",
			"userId", user.ID,
			"userVersion", user.Version,
		)
		return nil, exceptions.NewUnauthorizedError()
	}

	return s.generateAuthResponse(ctx, log, "User OAuth signed in successfully", user)
}
