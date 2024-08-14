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
	"github.com/kiwiscript/kiwiscript_go/utils"
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

func (s *Services) generateEmailCodeAndState(ctx context.Context, email string) (string, string, *ServiceError) {
	log := s.log.WithGroup("services.oauth.generateEmailCodeAndState")
	log.InfoContext(ctx, "Generating email code and state...")

	state, err := oauth.GenerateState()
	if err != nil {
		log.ErrorContext(ctx, "Failed to generate state", "error", err)
		return "", "", NewServerError()
	}

	code := utils.Base62UUID()
	if err := s.cache.AddOAuthEmail(code, email); err != nil {
		log.ErrorContext(ctx, "Failed to cache code", "error", err)
		return "", "", NewServerError()
	}
	if err := s.cache.AddOAuthState(state, utils.ProviderEmail); err != nil {
		log.ErrorContext(ctx, "Failed to cache state", "error", err)
		return "", "", NewServerError()
	}

	return code, state, nil
}

type ExtOAuthSignInOptions struct {
	Provider string
	Token    string
}

func (s *Services) ExtOAuthSignIn(ctx context.Context, opts ExtOAuthSignInOptions) (string, string, *ServiceError) {
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
		return "", "", NewServerError()
	}

	if err != nil {
		if status > 0 && status < 500 {
			log.WarnContext(ctx, "User data got non 200 status code", "error", err, "status", status)
			return "", "", NewUnauthorizedError()
		}

		log.ErrorContext(ctx, "Failed to fetch userData data", "error", err)
		return "", "", NewServerError()
	}

	userData := toUserData.ToUserData()
	if _, serviceErr := s.FindUserByEmail(ctx, userData.Email); serviceErr != nil {
		if serviceErr.Code != CodeNotFound {
			log.ErrorContext(ctx, "Failed to find user by email", "error", serviceErr)
			return "", "", NewServerError()
		}

		userOpts := CreateUserOptions{
			FirstName: userData.FirstName,
			LastName:  userData.LastName,
			Location:  userData.Location,
			Email:     userData.Email,
			Provider:  opts.Provider,
			Password:  "",
		}
		if _, serviceErr := s.CreateUser(ctx, userOpts); serviceErr != nil {
			log.ErrorContext(ctx, "Failed to create user", "error", serviceErr)
			return "", "", NewServerError()
		}

		return s.generateEmailCodeAndState(ctx, userData.Email)
	}

	findProvPrms := db.FindAuthProviderByEmailAndProviderParams{
		Email:    userData.Email,
		Provider: opts.Provider,
	}
	if _, err := s.database.FindAuthProviderByEmailAndProvider(ctx, findProvPrms); err != nil {
		serviceErr := FromDBError(err)
		if serviceErr.Code != CodeNotFound {
			log.ErrorContext(ctx, "Failed to find auth provider", "error", err)
			return "", "", serviceErr
		}

		createProvPrms := db.CreateAuthProviderParams{
			Email:    userData.Email,
			Provider: opts.Provider,
		}
		if err := s.database.CreateAuthProvider(ctx, createProvPrms); err != nil {
			log.ErrorContext(ctx, "Failed to create auth provider", "error", err)
			return "", "", FromDBError(err)
		}
	}

	return s.generateEmailCodeAndState(ctx, userData.Email)
}

type IntOAuthSignInOptions struct {
	Code  string
	State string
}

func (s *Services) IntOAuthSignIn(ctx context.Context, opts IntOAuthSignInOptions) (*AuthResponse, *ServiceError) {
	log := s.log.WithGroup("services.oauth.IntOAuthSignIn")
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

	ok, err := s.cache.VerifyOAuthState(opts.State, utils.ProviderEmail)
	if err != nil {
		log.ErrorContext(ctx, "Failed to fetch email state")
		return nil, NewServerError()
	}

	if !ok {
		log.WarnContext(ctx, "Email context is invalid")
		return nil, NewUnauthorizedError()
	}

	user, serviceErr := s.FindUserByEmail(ctx, email)
	if serviceErr != nil {
		if serviceErr.Code == CodeNotFound {
			return nil, NewUnauthorizedError()
		}

		return nil, serviceErr
	}

	return s.generateAuthResponse(ctx, log, "User OAuth signed in successfully", user)
}
