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

package tests

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/h2non/gock"
	"github.com/kiwiscript/kiwiscript_go/dtos"
	cc "github.com/kiwiscript/kiwiscript_go/providers/cache"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/providers/oauth"
	"github.com/kiwiscript/kiwiscript_go/utils"
	"net/http"
	"net/url"
	"testing"
)

const baseExtAuthPath = "/api/auth/ext"

func TestOAuthGet(t *testing.T) {
	userCleanUp(t)

	testCases := []TestRequestCase[string]{
		{
			Name: "GET github should return URL and 307 TEMPORARY REDIRECT",
			ReqFn: func(t *testing.T) (string, string) {
				return "", ""
			},
			ExpStatus: fiber.StatusTemporaryRedirect,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				defer gock.OffAll()
				location := resp.Header.Get("Location")
				AssertNotEmpty(t, location)
				AssertStringContains(t, location, "github.com/login/oauth/authorize")
				AssertStringContains(t, location, "kiwiscript.com")
			},
			Path: baseExtAuthPath + "/github",
		},
		{
			Name: "GET google should return URL and 307 TEMPORARY REDIRECT",
			ReqFn: func(t *testing.T) (string, string) {
				return "", ""
			},
			ExpStatus: fiber.StatusTemporaryRedirect,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				defer gock.OffAll()
				location := resp.Header.Get("Location")
				AssertNotEmpty(t, location)
				AssertStringContains(t, location, "accounts.google.com/o/oauth2/auth")
				AssertStringContains(t, location, "kiwiscript.com")
			},
			Path: baseExtAuthPath + "/google",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodGet, tc.Path, tc)
		})
	}

	userCleanUp(t)
}

func TestOAuthCallback(t *testing.T) {
	defer gock.OffAll()
	userCleanUp(t)

	var state, code string
	beforeEach := func(provider string) {
		ctx := context.Background()

		var err error
		state, err = oauth.GenerateState()
		if err != nil {
			t.Fatalf("Error generating state: %v", err)
		}

		testCache := GetTestCache(t)
		stateOpts := cc.AddOAuthStateOptions{
			RequestID: uuid.NewString(),
			State:     state,
			Provider:  provider,
		}
		if err := testCache.AddOAuthState(ctx, stateOpts); err != nil {
			t.Fatalf("Error adding state to cache: %v", err)
		}

		code = utils.Base62UUID()
	}

	addParams := func() string {
		params := make(url.Values)
		params.Add("code", code)
		params.Add("state", state)
		return params.Encode()
	}

	testCases := []TestRequestCase[string]{
		{
			Name: "GET github callback should return 302 FOUND and redirect code",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(utils.ProviderGitHub)
				gock.New("https://github.com").
					Post("/login/oauth/access_token").
					Reply(http.StatusOK).
					JSON(map[string]interface{}{
						"access_token":  "123",
						"token_type":    "Bearer",
						"expires_in":    3600,
						"refresh_token": "456",
					})

				gock.New("https://api.github.com").
					Get("/user").
					Reply(http.StatusOK).
					JSON(map[string]interface{}{
						"name":     "John Doe",
						"location": "nz",
						"email":    "john.doe@gmail.com",
					})

				return "", ""
			},
			ExpStatus: fiber.StatusFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				location := resp.Header.Get("Location")
				AssertNotEmpty(t, location)
				AssertStringContains(t, location, "kiwiscript.com")
				AssertStringContains(t, location, "access_token")
				AssertStringContains(t, location, "token_type")
				AssertStringContains(t, location, "expires_in")
				AssertEqual(t, gock.IsDone(), true)
				defer gock.OffAll()
			},
			PathFn: func() string {
				return baseExtAuthPath + "/github/callback?" + addParams()
			},
		},
		{
			Name: "GET google callback should return 302 FOUND and redirect code",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(utils.ProviderGoogle)
				gock.New("https://oauth2.googleapis.com").
					Post("/token").
					Reply(http.StatusOK).
					JSON(map[string]interface{}{
						"access_token":  "123",
						"token_type":    "Bearer",
						"expires_in":    3600,
						"refresh_token": "456",
					})

				gock.New("https://www.googleapis.com").
					Get("/oauth2/v3/userinfo").
					Reply(http.StatusOK).
					JSON(map[string]interface{}{
						"name":        "John Doe",
						"locale":      "en_NZ",
						"email":       "john.doe@gmail.com",
						"given_name":  "John",
						"family_name": "Doe",
					})

				return "", ""
			},
			ExpStatus: fiber.StatusFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				location := resp.Header.Get("Location")
				AssertNotEmpty(t, location)
				AssertStringContains(t, location, "kiwiscript.com")
				AssertStringContains(t, location, "access_token")
				AssertStringContains(t, location, "token_type")
				AssertStringContains(t, location, "expires_in")
				AssertEqual(t, gock.IsDone(), true)
				defer gock.OffAll()
			},
			PathFn: func() string {
				return baseExtAuthPath + "/google/callback?" + addParams()
			},
		},
		{
			Name: "GET github callback with invalid state should return 401 UNAUTHORIZED",
			ReqFn: func(t *testing.T) (string, string) {
				return "", ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertUnauthorizedResponse(t, resp)
				AssertEqual(t, gock.IsDone(), true)
				defer gock.OffAll()
			},
			PathFn: func() string {
				fakeCode := utils.Base62UUID()
				fakeState, err := oauth.GenerateState()
				if err != nil {
					t.Fatalf("Error generating state: %v", err)
				}
				return baseExtAuthPath + "/github/callback?" + "code=" + fakeCode + "&state=" + fakeState
			},
		},
		{
			Name: "GET google callback should return 401 UNAUTHORIZED if it fails to get the token",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(utils.ProviderGoogle)
				gock.New("https://oauth2.googleapis.com").
					Post("/token").
					Reply(http.StatusBadRequest).
					JSON(map[string]interface{}{
						"code": 400,
						"msg":  "Bad Request",
					})

				return "", ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertUnauthorizedResponse(t, resp)
				AssertEqual(t, gock.IsDone(), true)
				defer gock.OffAll()
			},
			PathFn: func() string {
				return baseExtAuthPath + "/google/callback?" + addParams()
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCaseWithPathFn(t, http.MethodGet, tc)
		})
	}

	userCleanUp(t)
}

func TestOAuthToken(t *testing.T) {
	userCleanUp(t)
	generateUserAndToken := func(t *testing.T) (*db.User, string) {
		testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
		tks := GetTestTokens(t)
		temporaryToken, err := tks.CreateOAuthToken(testUser)
		if err != nil {
			t.Fatalf("Error creating temporary token: %v", err)
		}
		return testUser, temporaryToken
	}

	testCases := []TestRequestCase[dtos.OAuthTokenBody]{
		{
			Name: "POST /api/auth/ext/token should return 200 with access token",
			ReqFn: func(t *testing.T) (dtos.OAuthTokenBody, string) {
				testUser, oauthToken := generateUserAndToken(t)
				code := utils.Base62UUID()
				ctx := context.Background()
				opts := cc.AddOAuthEmailOptions{
					RequestID:       uuid.NewString(),
					Code:            code,
					Email:           testUser.Email,
					DurationSeconds: 300,
				}
				if err := GetTestCache(t).AddOAuthEmail(ctx, opts); err != nil {
					t.Fatalf("Error adding email to cache: %v", err)
				}

				body := dtos.OAuthTokenBody{
					Code:        code,
					RedirectURI: "https://kiwiscript.com/auth/callback",
				}
				return body, oauthToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ dtos.OAuthTokenBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.AuthResponse{})
				AssertNotEmpty(t, resBody.AccessToken)
				AssertNotEmpty(t, resBody.RefreshToken)
				AssertEqual(t, resBody.TokenType, "Bearer")
				AssertGreaterThan(t, resBody.ExpiresIn, int64(0))
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if the code is invalid or expired",
			ReqFn: func(t *testing.T) (dtos.OAuthTokenBody, string) {
				_, oauthToken := generateUserAndToken(t)
				body := dtos.OAuthTokenBody{
					Code:        utils.Base62UUID(),
					RedirectURI: "https://kiwiscript.com/auth/callback",
				}
				return body, oauthToken
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ dtos.OAuthTokenBody, resp *http.Response) {
				AssertUnauthorizedResponse(t, resp)
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if the user is unauthorized",
			ReqFn: func(t *testing.T) (dtos.OAuthTokenBody, string) {
				body := dtos.OAuthTokenBody{
					Code:        utils.Base62UUID(),
					RedirectURI: "https://kiwiscript.com/auth/callback",
				}
				return body, ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ dtos.OAuthTokenBody, resp *http.Response) {
				AssertUnauthorizedResponse(t, resp)
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if the redirect URI is invalid",
			ReqFn: func(t *testing.T) (dtos.OAuthTokenBody, string) {
				_, oauthToken := generateUserAndToken(t)
				body := dtos.OAuthTokenBody{
					Code:        utils.Base62UUID(),
					RedirectURI: "https://not-kiwiscript.com/auth/callback",
				}
				return body, oauthToken
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ dtos.OAuthTokenBody, resp *http.Response) {
				AssertUnauthorizedResponse(t, resp)
			},
		},
		{
			Name: "Should return 400 BAD REQUEST if body is invalid",
			ReqFn: func(t *testing.T) (dtos.OAuthTokenBody, string) {
				_, oauthToken := generateUserAndToken(t)
				return dtos.OAuthTokenBody{
					Code:        "",
					RedirectURI: "not-a-url",
				}, oauthToken
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, _ dtos.OAuthTokenBody, resp *http.Response) {
				AssertValidationErrorResponse(t, resp, []ValidationErrorAssertion{
					{Param: "code", Message: "must be provided"},
					{Param: "redirectURI", Message: "must be a valid URL"},
				})
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodPost, baseExtAuthPath+"/token", tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}
