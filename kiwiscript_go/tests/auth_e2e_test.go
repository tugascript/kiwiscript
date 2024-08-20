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
	"github.com/kiwiscript/kiwiscript_go/dtos"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
	"github.com/kiwiscript/kiwiscript_go/controllers"
	cc "github.com/kiwiscript/kiwiscript_go/providers/cache"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/providers/tokens"
	"github.com/kiwiscript/kiwiscript_go/services"
)

func userCleanUp(t *testing.T) func() {
	return func() {
		dbProv := GetTestDatabase(t)
		cacheProv := GetTestCache(t)

		if err := dbProv.DeleteAllUsers(context.Background()); err != nil {
			t.Fatal("Failed to delete all users", err)
		}
		if err := cacheProv.ResetCache(); err != nil {
			t.Fatal("Failed to reset cache", err)
		}
	}
}

type fakeRegisterData struct {
	Email     string `faker:"email"`
	FirstName string `faker:"first_name"`
	LastName  string `faker:"last_name"`
	Location  string `faker:"oneof: NZL, AUS, NAM, EUR, OTH"`
	BirthDate string `faker:"date"`
	Password  string `faker:"oneof: Pas@w0rd123, P@sW0rd456, P@ssw0rd789, P@ssW0rd012, P@ssw0rd!345"`
}

func generateEmailToken(t *testing.T, tokenType string, user *db.User) string {
	tokensProv := GetTestTokens(t)
	emailToken, err := tokensProv.CreateEmailToken(tokenType, user)

	if err != nil {
		t.Fatal("Failed to generate test emailToken", err)
	}

	return emailToken
}

func confirmTestUser(t *testing.T, userID int32) *db.User {
	serv := GetTestServices(t)
	user, err := serv.ConfirmUser(context.Background(), userID)

	if err != nil {
		t.Fatal(err)
	}

	return user
}

func assertOAuthResponse(t *testing.T, resp *http.Response) {
	resBody := AssertTestResponseBody(t, resp, dtos.AuthResponse{})
	AssertEqual(t, "Bearer", resBody.TokenType)
	AssertNotEmpty(t, resBody.AccessToken)
	AssertNotEmpty(t, resBody.RefreshToken)
	AssertGreaterThan(t, resBody.ExpiresIn, 0)
}

func assertUnauthorizeError(t *testing.T, resp *http.Response) {
	resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
	AssertEqual(t, services.MessageUnauthorized, resBody.Message)
	AssertEqual(t, controllers.StatusUnauthorized, resBody.Code)
}

func performCookieRequest(t *testing.T, app *fiber.App, path, accessToken, refreshToken string) *http.Response {
	config := GetTestConfig(t)
	req := httptest.NewRequest(http.MethodPost, path, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if refreshToken != "" {
		encryptedRefreshToken, err := encryptcookie.EncryptCookie(refreshToken, config.CookieSecret)
		if err != nil {
			t.Fatal("Failed to encrypt cookie", err)
		}

		req.AddCookie(&http.Cookie{
			Name:  config.RefreshCookieName,
			Value: encryptedRefreshToken,
			Path:  "/api/auth",
		})
	}
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal("Failed to perform request", err)
	}

	return resp
}

func TestRegister(t *testing.T) {
	const registerPath = "/api/auth/register"

	generateFakeRegisterData := func(t *testing.T) dtos.SignUpBody {
		fakeData := fakeRegisterData{}
		if err := faker.FakeData(&fakeData); err != nil {
			t.Fatal("Failed to generate fake data", err)
		}
		return dtos.SignUpBody{
			Email:     fakeData.Email,
			FirstName: fakeData.FirstName,
			LastName:  fakeData.LastName,
			Location:  fakeData.Location,
			Password1: fakeData.Password,
			Password2: fakeData.Password,
		}
	}

	testCases := []TestRequestCase[dtos.SignUpBody]{
		{
			Name: "Should return 200 OK registering a user",
			ReqFn: func(t *testing.T) (dtos.SignUpBody, string) {
				return generateFakeRegisterData(t), ""
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ dtos.SignUpBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.MessageResponse{})
				AssertEqual(t, "Confirmation email has been sent", resBody.Message)
				AssertNotEmpty(t, resBody.ID)
			},
			DelayMs: 50,
		},
		{
			Name: "Should return 400 BAD REQUEST if request validation fails",
			ReqFn: func(t *testing.T) (dtos.SignUpBody, string) {
				reqBody := generateFakeRegisterData(t)
				reqBody.Email = "notAnEmail"
				return reqBody, ""
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req dtos.SignUpBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
				AssertEqual(t, 1, len(resBody.Fields))
				AssertEqual(t, "email", resBody.Fields[0].Param)
				AssertEqual(t, controllers.StrFieldErrMessageEmail, resBody.Fields[0].Message)
				AssertEqual(t, req.Email, resBody.Fields[0].Value.(string))
			},
		},
		{
			Name: "Should return 400 BAD REQUEST if password missmatch",
			ReqFn: func(t *testing.T) (dtos.SignUpBody, string) {
				reqBody := generateFakeRegisterData(t)
				reqBody.Password2 = "differentPassword"
				return reqBody, ""
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req dtos.SignUpBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
				AssertEqual(t, 1, len(resBody.Fields))
				AssertEqual(t, "password2", resBody.Fields[0].Param)
				AssertEqual(t, controllers.FieldErrMessageEqField, resBody.Fields[0].Message)
				AssertEqual(t, req.Password2, resBody.Fields[0].Value.(string))
			},
		},
		{
			Name: "Should return 409 CONFLICT if email already exists",
			ReqFn: func(t *testing.T) (dtos.SignUpBody, string) {
				testUser := CreateTestUser(t, nil)
				reqBody := generateFakeRegisterData(t)
				reqBody.Email = testUser.Email
				return reqBody, ""
			},
			ExpStatus: fiber.StatusConflict,
			AssertFn: func(t *testing.T, _ dtos.SignUpBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, "Email already exists", resBody.Message)
				AssertEqual(t, controllers.StatusConflict, resBody.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodPost, registerPath, tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}

func TestConfirmEmail(t *testing.T) {
	const confirmEmailPath = "/api/auth/confirm-email"

	generateTestConfirmEmailData := func(t *testing.T, user *db.User) dtos.ConfirmBody {
		emailToken := generateEmailToken(t, tokens.EmailTokenConfirmation, user)

		return dtos.ConfirmBody{
			ConfirmationToken: emailToken,
		}
	}
	unauthorizedFunc := func(t *testing.T, _ dtos.ConfirmBody, resp *http.Response) {
		assertUnauthorizeError(t, resp)
	}

	testCases := []TestRequestCase[dtos.ConfirmBody]{
		{
			Name: "Should return 200 OK with OAuth response",
			ReqFn: func(t *testing.T) (dtos.ConfirmBody, string) {
				testUser := CreateTestUser(t, nil)
				return generateTestConfirmEmailData(t, testUser), ""
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ dtos.ConfirmBody, resp *http.Response) {
				assertOAuthResponse(t, resp)
			},
		},
		{
			Name: "Should return 400 BAD REQUEST if request validation fails",
			ReqFn: func(t *testing.T) (dtos.ConfirmBody, string) {
				return dtos.ConfirmBody{ConfirmationToken: "invalidToken"}, ""
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req dtos.ConfirmBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
				AssertEqual(t, 1, len(resBody.Fields))
				AssertEqual(t, "confirmationToken", resBody.Fields[0].Param)
				AssertEqual(t, controllers.StrFieldErrMessageJWT, resBody.Fields[0].Message)
				AssertEqual(t, req.ConfirmationToken, resBody.Fields[0].Value.(string))
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if user not found",
			ReqFn: func(t *testing.T) (dtos.ConfirmBody, string) {
				fakeUser := CreateFakeTestUser(t)
				return generateTestConfirmEmailData(t, fakeUser), ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn:  unauthorizedFunc,
		},
		{
			Name: "Should return 401 UNAUTHORIZED if user already confirmed",
			ReqFn: func(t *testing.T) (dtos.ConfirmBody, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				return generateTestConfirmEmailData(t, testUser), ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn:  unauthorizedFunc,
		},
		{
			Name: "Should return 401 UNAUTHORIZED for version missmatch",
			ReqFn: func(t *testing.T) (dtos.ConfirmBody, string) {
				testUser := CreateTestUser(t, nil)
				testUser.Version = math.MaxInt16
				return generateTestConfirmEmailData(t, testUser), ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn:  unauthorizedFunc,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodPost, confirmEmailPath, tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}

func TestLogin(t *testing.T) {
	const loginPath = "/api/auth/login"

	testCases := []TestRequestCase[dtos.SignInBody]{
		{
			Name: "Should return 200 OK with message response",
			ReqFn: func(t *testing.T) (dtos.SignInBody, string) {
				fakeUserData := GenerateFakeUserData(t)
				testUser := confirmTestUser(t, CreateTestUser(t, &fakeUserData).ID)
				reqBody := dtos.SignInBody{Email: testUser.Email, Password: fakeUserData.Password}
				return reqBody, ""
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ dtos.SignInBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.MessageResponse{})
				AssertEqual(t, "Confirmation code has been sent to your email", resBody.Message)
				AssertNotEmpty(t, resBody.ID)
			},
			DelayMs: 50,
		},
		{
			Name: "Should return 400 BAD REQUEST if request validation fails",
			ReqFn: func(t *testing.T) (dtos.SignInBody, string) {
				return dtos.SignInBody{Email: "notAnEmail", Password: ""}, ""
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req dtos.SignInBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
				AssertEqual(t, 2, len(resBody.Fields))
				AssertEqual(t, "email", resBody.Fields[0].Param)
				AssertEqual(t, controllers.StrFieldErrMessageEmail, resBody.Fields[0].Message)
				AssertEqual(t, req.Email, resBody.Fields[0].Value.(string))
				AssertEqual(t, "password", resBody.Fields[1].Param)
				AssertEqual(t, controllers.FieldErrMessageRequired, resBody.Fields[1].Message)
				AssertEqual(t, req.Password, resBody.Fields[1].Value.(string))
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if user not found",
			ReqFn: func(t *testing.T) (dtos.SignInBody, string) {
				fakeEmail := faker.Email()
				fakePassword := faker.Password()
				return dtos.SignInBody{Email: fakeEmail, Password: fakePassword}, ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ dtos.SignInBody, resp *http.Response) {
				AssertTestStatusCode(t, resp, fiber.StatusUnauthorized)
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, services.MessageUnauthorized, resBody.Message)
				AssertEqual(t, controllers.StatusUnauthorized, resBody.Code)
			},
		},
		{
			Name: "Should return 400 BAD REQUEST if user not confirmed",
			ReqFn: func(t *testing.T) (dtos.SignInBody, string) {
				fakeUserData := GenerateFakeUserData(t)
				testUser := CreateTestUser(t, &fakeUserData)
				return dtos.SignInBody{Email: testUser.Email, Password: fakeUserData.Password}, ""
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, _ dtos.SignInBody, resp *http.Response) {
				AssertTestStatusCode(t, resp, fiber.StatusBadRequest)
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, "User not confirmed", resBody.Message)
				AssertEqual(t, controllers.StatusValidation, resBody.Code)
			},
			DelayMs: 50,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodPost, loginPath, tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}

func TestLoginConfirm(t *testing.T) {
	const loginConfirmPath = "/api/auth/login/confirm"

	generateTestTwoFactorCode := func(t *testing.T, user *db.User) string {
		cacheProv := GetTestCache(t)
		code, err := cacheProv.AddTwoFactorCode(user.ID)

		if err != nil {
			t.Fatal("Failed to generate two factor code", err)
		}

		return code
	}

	testCases := []TestRequestCase[dtos.ConfirmSignInBody]{
		{
			Name: "Should return 200 OK with OAuth response",
			ReqFn: func(t *testing.T) (dtos.ConfirmSignInBody, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				code := generateTestTwoFactorCode(t, testUser)
				return dtos.ConfirmSignInBody{Email: testUser.Email, Code: code}, ""
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ dtos.ConfirmSignInBody, resp *http.Response) {
				assertOAuthResponse(t, resp)
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if code is wrong",
			ReqFn: func(t *testing.T) (dtos.ConfirmSignInBody, string) {
				fakeUser := CreateFakeTestUser(t)
				return dtos.ConfirmSignInBody{Email: fakeUser.Email, Code: "123456"}, ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ dtos.ConfirmSignInBody, resp *http.Response) {
				assertUnauthorizeError(t, resp)
			},
		},
		{
			Name: "Should return 400 BAD REQUEST if request validation fails",
			ReqFn: func(t *testing.T) (dtos.ConfirmSignInBody, string) {
				return dtos.ConfirmSignInBody{Email: "notAnEmail", Code: ""}, ""
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req dtos.ConfirmSignInBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
				AssertEqual(t, 2, len(resBody.Fields))
				AssertEqual(t, "email", resBody.Fields[0].Param)
				AssertEqual(t, controllers.StrFieldErrMessageEmail, resBody.Fields[0].Message)
				AssertEqual(t, req.Email, resBody.Fields[0].Value.(string))
				AssertEqual(t, "code", resBody.Fields[1].Param)
				AssertEqual(t, controllers.FieldErrMessageRequired, resBody.Fields[1].Message)
				AssertEqual(t, req.Code, resBody.Fields[1].Value.(string))
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if user not found",
			ReqFn: func(t *testing.T) (dtos.ConfirmSignInBody, string) {
				fakeUser := CreateFakeTestUser(t)
				return dtos.ConfirmSignInBody{Email: fakeUser.Email, Code: "123456"}, ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ dtos.ConfirmSignInBody, resp *http.Response) {
				assertUnauthorizeError(t, resp)
			},
		},
		{
			Name: "Should return 400 BAD REQUEST if user not confirmed",
			ReqFn: func(t *testing.T) (dtos.ConfirmSignInBody, string) {
				testUser := CreateTestUser(t, nil)
				code := generateTestTwoFactorCode(t, testUser)
				return dtos.ConfirmSignInBody{Email: testUser.Email, Code: code}, ""
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, _ dtos.ConfirmSignInBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, "User not confirmed", resBody.Message)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodPost, loginConfirmPath, tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}

type cookieTestCase struct {
	Name      string
	ExpStatus int
	TokenFn   func(t *testing.T) (string, string)
	AssertFn  func(t *testing.T, resp *http.Response)
}

func TestLogout(t *testing.T) {
	const logoutPath = "/api/auth/logout"

	bodyTestCases := []TestRequestCase[dtos.SignOutBody]{
		{
			Name: "Should return 204 NO CONTENT",
			ReqFn: func(t *testing.T) (dtos.SignOutBody, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				accessToken, refreshToken := GenerateTestAuthTokens(t, testUser)
				return dtos.SignOutBody{RefreshToken: refreshToken}, accessToken
			},
			ExpStatus: fiber.StatusNoContent,
			AssertFn:  func(t *testing.T, _ dtos.SignOutBody, resp *http.Response) {},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if access token is invalid",
			ReqFn: func(t *testing.T) (dtos.SignOutBody, string) {
				testUser := CreateFakeTestUser(t)
				_, refreshToken := GenerateTestAuthTokens(t, testUser)
				return dtos.SignOutBody{RefreshToken: refreshToken}, "invalid"
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ dtos.SignOutBody, resp *http.Response) {
				assertUnauthorizeError(t, resp)
			},
		},
		{
			Name: "Should return 400 BAD REQUEST if request validation fails",
			ReqFn: func(t *testing.T) (dtos.SignOutBody, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.SignOutBody{RefreshToken: "not-a-jwt"}, accessToken
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req dtos.SignOutBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
				AssertEqual(t, 1, len(resBody.Fields))
				AssertEqual(t, "refreshToken", resBody.Fields[0].Param)
				AssertEqual(t, controllers.StrFieldErrMessageJWT, resBody.Fields[0].Message)
				AssertEqual(t, req.RefreshToken, resBody.Fields[0].Value.(string))
			},
		},
	}

	for _, tc := range bodyTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodPost, logoutPath, tc)
		})
	}

	cookieTestCases := []cookieTestCase{
		{
			Name:      "Should return 204 NO CONTENT - cookie test",
			ExpStatus: fiber.StatusNoContent,
			TokenFn: func(t *testing.T) (string, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				accessToken, refreshToken := GenerateTestAuthTokens(t, testUser)
				return accessToken, refreshToken
			},
			AssertFn: func(t *testing.T, resp *http.Response) {},
		},
		{
			Name:      "Should return 401 UNAUTHORIZED if access token is invalid - cookie test",
			ExpStatus: fiber.StatusUnauthorized,
			TokenFn: func(t *testing.T) (string, string) {
				testUser := CreateFakeTestUser(t)
				_, refreshToken := GenerateTestAuthTokens(t, testUser)
				return "invalid", refreshToken
			},
			AssertFn: func(t *testing.T, resp *http.Response) {
				assertUnauthorizeError(t, resp)
			},
		},
		{
			Name:      "Should return 400 BAD REQUEST if refresh token is invalid - cookie test",
			ExpStatus: fiber.StatusBadRequest,
			TokenFn: func(t *testing.T) (string, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return accessToken, ""
			},
			AssertFn: func(t *testing.T, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.EmptyRequestValidationError{})
				AssertEqual(t, controllers.RequestValidationMessage, resBody.Message)
			},
		},
	}

	for _, tc := range cookieTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			accessToken, refreshToken := tc.TokenFn(t)
			app := GetTestApp(t)

			resp := performCookieRequest(t, app, logoutPath, accessToken, refreshToken)
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Fatal(err)
				}
			}()

			AssertTestStatusCode(t, resp, tc.ExpStatus)
			tc.AssertFn(t, resp)
		})
	}

	t.Cleanup(userCleanUp(t))
}

func TestRefresh(t *testing.T) {
	const refreshPath = "/api/auth/refresh"

	blackListRefreshToken := func(t *testing.T, refreshToken string) {
		cacheProv := GetTestCache(t)
		tokensProv := GetTestTokens(t)

		_, id, exp, err := tokensProv.VerifyRefreshToken(refreshToken)
		if err != nil {
			t.Fatal("Failed to verify refresh token", err)
		}

		opts := cc.AddBlackListOptions{
			ID:  id,
			Exp: exp,
		}
		if err := cacheProv.AddBlackList(opts); err != nil {
			t.Fatal("Failed to add token to black list", err)
		}
	}

	bodyTestCases := []TestRequestCase[dtos.RefreshBody]{
		{
			Name: "Should return 200 OK",
			ReqFn: func(t *testing.T) (dtos.RefreshBody, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				_, refreshToken := GenerateTestAuthTokens(t, testUser)
				return dtos.RefreshBody{RefreshToken: refreshToken}, ""
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ dtos.RefreshBody, resp *http.Response) {
				assertOAuthResponse(t, resp)
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if refresh token is blacklisted",
			ReqFn: func(t *testing.T) (dtos.RefreshBody, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				_, refreshToken := GenerateTestAuthTokens(t, testUser)
				blackListRefreshToken(t, refreshToken)
				return dtos.RefreshBody{RefreshToken: refreshToken}, ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ dtos.RefreshBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, services.MessageUnauthorized, resBody.Message)
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if user version mismatches",
			ReqFn: func(t *testing.T) (dtos.RefreshBody, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				testUser.Version = math.MaxInt16
				_, refreshToken := GenerateTestAuthTokens(t, testUser)
				return dtos.RefreshBody{RefreshToken: refreshToken}, ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ dtos.RefreshBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, services.MessageUnauthorized, resBody.Message)
			},
		},
		{
			Name: "Should return 400 BAD REQUEST if refresh token is invalid",
			ReqFn: func(t *testing.T) (dtos.RefreshBody, string) {
				return dtos.RefreshBody{RefreshToken: "not-jwt"}, ""
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req dtos.RefreshBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
				AssertEqual(t, 1, len(resBody.Fields))
				AssertEqual(t, "refreshToken", resBody.Fields[0].Param)
				AssertEqual(t, controllers.StrFieldErrMessageJWT, resBody.Fields[0].Message)
				AssertEqual(t, req.RefreshToken, resBody.Fields[0].Value.(string))
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if user is deleted",
			ReqFn: func(t *testing.T) (dtos.RefreshBody, string) {
				testUser := CreateFakeTestUser(t)
				_, refreshToken := GenerateTestAuthTokens(t, testUser)
				return dtos.RefreshBody{RefreshToken: refreshToken}, ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ dtos.RefreshBody, resp *http.Response) {
				assertUnauthorizeError(t, resp)
			},
		},
	}

	for _, tc := range bodyTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodPost, refreshPath, tc)
		})
	}

	cookieTestCases := []cookieTestCase{
		{
			Name:      "Should return 200 OK - cookie test",
			ExpStatus: fiber.StatusOK,
			TokenFn: func(t *testing.T) (string, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				_, refreshToken := GenerateTestAuthTokens(t, testUser)
				return "", refreshToken
			},
			AssertFn: func(t *testing.T, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.AuthResponse{})
				AssertEqual(t, "Bearer", resBody.TokenType)
			},
		},
		{
			Name:      "Should return 401 UNAUTHORIZED if refresh token is blacklisted - cookie test",
			ExpStatus: fiber.StatusUnauthorized,
			TokenFn: func(t *testing.T) (string, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				_, refreshToken := GenerateTestAuthTokens(t, testUser)
				blackListRefreshToken(t, refreshToken)
				return "", refreshToken
			},
			AssertFn: func(t *testing.T, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, services.MessageUnauthorized, resBody.Message)
			},
		},
		{
			Name:      "Should return 401 UNAUTHORIZED if user version mismatches - cookie test",
			ExpStatus: fiber.StatusUnauthorized,
			TokenFn: func(t *testing.T) (string, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				testUser.Version = math.MaxInt16
				_, refreshToken := GenerateTestAuthTokens(t, testUser)
				return "", refreshToken
			},
			AssertFn: func(t *testing.T, resp *http.Response) {
				assertUnauthorizeError(t, resp)
			},
		},
		{
			Name:      "Should return 401 UNAUTHORIZED if user is deleted - cookie test",
			ExpStatus: fiber.StatusUnauthorized,
			TokenFn: func(t *testing.T) (string, string) {
				testUser := CreateFakeTestUser(t)
				_, refreshToken := GenerateTestAuthTokens(t, testUser)
				return "", refreshToken
			},
			AssertFn: func(t *testing.T, resp *http.Response) {
				assertUnauthorizeError(t, resp)
			},
		},
	}

	for _, tc := range cookieTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			accessToken, refreshToken := tc.TokenFn(t)
			app := GetTestApp(t)

			resp := performCookieRequest(t, app, refreshPath, accessToken, refreshToken)
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Fatal(err)
				}
			}()

			AssertTestStatusCode(t, resp, tc.ExpStatus)
			tc.AssertFn(t, resp)
		})
	}

	t.Cleanup(userCleanUp(t))
}

func TestForgotPassword(t *testing.T) {
	const forgotPasswordPath = "/api/auth/forgot-password"

	bodyTestCases := []TestRequestCase[dtos.ForgotPasswordBody]{
		{
			Name: "Should return 200 OK with existing user",
			ReqFn: func(t *testing.T) (dtos.ForgotPasswordBody, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				return dtos.ForgotPasswordBody{Email: testUser.Email}, ""
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ dtos.ForgotPasswordBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.MessageResponse{})
				AssertEqual(t, "If the email exists, a password reset email has been sent", resBody.Message)
			},
		},
		{
			Name: "Should return 200 OK with non-existing user",
			ReqFn: func(t *testing.T) (dtos.ForgotPasswordBody, string) {
				fakeEmail := faker.Email()
				return dtos.ForgotPasswordBody{Email: fakeEmail}, ""
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ dtos.ForgotPasswordBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.MessageResponse{})
				AssertEqual(t, "If the email exists, a password reset email has been sent", resBody.Message)
			},
		},
		{
			Name: "Should return 400 BAD REQUEST if request validation fails",
			ReqFn: func(t *testing.T) (dtos.ForgotPasswordBody, string) {
				return dtos.ForgotPasswordBody{Email: "notAnEmail"}, ""
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req dtos.ForgotPasswordBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
				AssertEqual(t, 1, len(resBody.Fields))
				AssertEqual(t, "email", resBody.Fields[0].Param)
				AssertEqual(t, controllers.StrFieldErrMessageEmail, resBody.Fields[0].Message)
				AssertEqual(t, req.Email, resBody.Fields[0].Value.(string))
			},
		},
	}

	for _, tc := range bodyTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodPost, forgotPasswordPath, tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}

func TestResetPassword(t *testing.T) {
	const resetPasswordPath = "/api/auth/reset-password"
	generateTestResetPasswordData := func(t *testing.T, user *db.User) dtos.ResetPasswordBody {
		emailToken := generateEmailToken(t, tokens.EmailTokenReset, user)
		password := faker.Name() + "123!"

		return dtos.ResetPasswordBody{
			ResetToken: emailToken,
			Password1:  password,
			Password2:  password,
		}
	}

	testCases := []TestRequestCase[dtos.ResetPasswordBody]{
		{
			Name: "Should return 200 OK",
			ReqFn: func(t *testing.T) (dtos.ResetPasswordBody, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				reqBody := generateTestResetPasswordData(t, testUser)
				return reqBody, ""
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ dtos.ResetPasswordBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.MessageResponse{})
				AssertEqual(t, "Password reset successfully", resBody.Message)
			},
		},
		{
			Name: "Should return 400 BAD REQUEST if request validation fails",
			ReqFn: func(t *testing.T) (dtos.ResetPasswordBody, string) {
				return dtos.ResetPasswordBody{ResetToken: "invalid", Password1: "a", Password2: "cb"}, ""
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req dtos.ResetPasswordBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
				AssertEqual(t, 3, len(resBody.Fields))
				AssertEqual(t, "resetToken", resBody.Fields[0].Param)
				AssertEqual(t, controllers.StrFieldErrMessageJWT, resBody.Fields[0].Message)
				AssertEqual(t, req.ResetToken, resBody.Fields[0].Value.(string))
				AssertEqual(t, "password1", resBody.Fields[1].Param)
				AssertEqual(t, controllers.StrFieldErrMessageMin, resBody.Fields[1].Message)
				AssertEqual(t, req.Password1, resBody.Fields[1].Value.(string))
				AssertEqual(t, "password2", resBody.Fields[2].Param)
				AssertEqual(t, controllers.FieldErrMessageEqField, resBody.Fields[2].Message)
				AssertEqual(t, req.Password2, resBody.Fields[2].Value.(string))
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if user version mismatches",
			ReqFn: func(t *testing.T) (dtos.ResetPasswordBody, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				testUser.Version = math.MaxInt16
				reqBody := generateTestResetPasswordData(t, testUser)
				return reqBody, ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ dtos.ResetPasswordBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, services.MessageUnauthorized, resBody.Message)
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if user is deleted",
			ReqFn: func(t *testing.T) (dtos.ResetPasswordBody, string) {
				testUser := CreateFakeTestUser(t)
				reqBody := generateTestResetPasswordData(t, testUser)
				return reqBody, ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ dtos.ResetPasswordBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, services.MessageUnauthorized, resBody.Message)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodPost, resetPasswordPath, tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}

func TestUpdatePassword(t *testing.T) {
	const updatePasswordPath = "/api/auth/update-password"

	generateTestUpdatePasswordData := func(_ *testing.T, oldPassword string) dtos.UpdatePasswordBody {
		newPassword := faker.Name() + "123!"
		return dtos.UpdatePasswordBody{
			OldPassword: oldPassword,
			Password1:   newPassword,
			Password2:   newPassword,
		}
	}

	testCases := []TestRequestCase[dtos.UpdatePasswordBody]{
		{
			Name: "Should return 200 OK",
			ReqFn: func(t *testing.T) (dtos.UpdatePasswordBody, string) {
				fakeUserData := GenerateFakeUserData(t)
				testUser := confirmTestUser(t, CreateTestUser(t, &fakeUserData).ID)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				reqBody := generateTestUpdatePasswordData(t, fakeUserData.Password)
				return reqBody, accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ dtos.UpdatePasswordBody, resp *http.Response) {
				assertOAuthResponse(t, resp)
			},
		},
		{
			Name: "Should return 400 BAD REQUEST if request validation fails",
			ReqFn: func(t *testing.T) (dtos.UpdatePasswordBody, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.UpdatePasswordBody{OldPassword: "", Password1: "a", Password2: "b"}, accessToken
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req dtos.UpdatePasswordBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
				AssertEqual(t, 3, len(resBody.Fields))
				AssertEqual(t, "oldPassword", resBody.Fields[0].Param)
				AssertEqual(t, controllers.FieldErrMessageRequired, resBody.Fields[0].Message)
				AssertEqual(t, req.OldPassword, resBody.Fields[0].Value.(string))
				AssertEqual(t, "password1", resBody.Fields[1].Param)
				AssertEqual(t, controllers.StrFieldErrMessageMin, resBody.Fields[1].Message)
				AssertEqual(t, req.Password1, resBody.Fields[1].Value.(string))
				AssertEqual(t, "password2", resBody.Fields[2].Param)
				AssertEqual(t, controllers.FieldErrMessageEqField, resBody.Fields[2].Message)
			},
		},
		{
			Name: "Should return 400 BAD REQUEST if user users wrong password",
			ReqFn: func(t *testing.T) (dtos.UpdatePasswordBody, string) {
				fakeUserData := GenerateFakeUserData(t)
				testUser := confirmTestUser(t, CreateTestUser(t, &fakeUserData).ID)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				reqBody := generateTestUpdatePasswordData(t, fakeUserData.Password+"wrong")
				return reqBody, accessToken
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, _ dtos.UpdatePasswordBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, "Old password is incorrect", resBody.Message)
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if user version mismatches",
			ReqFn: func(t *testing.T) (dtos.UpdatePasswordBody, string) {
				fakeUserData := GenerateFakeUserData(t)
				testUser := confirmTestUser(t, CreateTestUser(t, &fakeUserData).ID)
				testUser.Version = math.MaxInt16
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				reqBody := generateTestUpdatePasswordData(t, fakeUserData.Password)
				return reqBody, accessToken
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ dtos.UpdatePasswordBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, services.MessageUnauthorized, resBody.Message)
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if user is deleted",
			ReqFn: func(t *testing.T) (dtos.UpdatePasswordBody, string) {
				fakeUserData := GenerateFakeUserData(t)
				testUser := CreateFakeTestUser(t)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				reqBody := generateTestUpdatePasswordData(t, fakeUserData.Password)
				return reqBody, accessToken
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ dtos.UpdatePasswordBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, services.MessageUnauthorized, resBody.Message)
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if access token is missing",
			ReqFn: func(t *testing.T) (dtos.UpdatePasswordBody, string) {
				reqBody := generateTestUpdatePasswordData(t, faker.Password())
				return reqBody, ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ dtos.UpdatePasswordBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, services.MessageUnauthorized, resBody.Message)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodPost, updatePasswordPath, tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}

func TestUpdateEmail(t *testing.T) {
	const updateEmailPath = "/api/auth/update-email"

	generateTestUpdateEmailData := func(_ *testing.T, password string) dtos.UpdateEmailBody {
		return dtos.UpdateEmailBody{
			Email:    faker.Email(),
			Password: password,
		}
	}

	testCases := []TestRequestCase[dtos.UpdateEmailBody]{
		{
			Name: "Should return 200 OK",
			ReqFn: func(t *testing.T) (dtos.UpdateEmailBody, string) {
				fakeUserData := GenerateFakeUserData(t)
				testUser := confirmTestUser(t, CreateTestUser(t, &fakeUserData).ID)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				reqBody := generateTestUpdateEmailData(t, fakeUserData.Password)
				return reqBody, accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ dtos.UpdateEmailBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.AuthResponse{})
				AssertEqual(t, "Bearer", resBody.TokenType)
			},
		},
		{
			Name: "Should return 400 BAD REQUEST if request validation fails",
			ReqFn: func(t *testing.T) (dtos.UpdateEmailBody, string) {
				fakeUserData := GenerateFakeUserData(t)
				testUser := confirmTestUser(t, CreateTestUser(t, &fakeUserData).ID)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.UpdateEmailBody{Email: "", Password: fakeUserData.Password}, accessToken
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req dtos.UpdateEmailBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
				AssertEqual(t, 1, len(resBody.Fields))
				AssertEqual(t, "email", resBody.Fields[0].Param)
				AssertEqual(t, controllers.FieldErrMessageRequired, resBody.Fields[0].Message)
				AssertEqual(t, req.Email, resBody.Fields[0].Value.(string))
			},
		},
		{
			Name: "Should return 400 BAD REQUEST if user users wrong password",
			ReqFn: func(t *testing.T) (dtos.UpdateEmailBody, string) {
				fakeUserData := GenerateFakeUserData(t)
				testUser := confirmTestUser(t, CreateTestUser(t, &fakeUserData).ID)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				reqBody := generateTestUpdateEmailData(t, fakeUserData.Password+"wrong")
				return reqBody, accessToken
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, _ dtos.UpdateEmailBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, "Invalid password", resBody.Message)
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if user version mismatches",
			ReqFn: func(t *testing.T) (dtos.UpdateEmailBody, string) {
				fakeUserData := GenerateFakeUserData(t)
				testUser := confirmTestUser(t, CreateTestUser(t, &fakeUserData).ID)
				testUser.Version = math.MaxInt16
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				reqBody := generateTestUpdateEmailData(t, fakeUserData.Password)
				return reqBody, accessToken
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ dtos.UpdateEmailBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, services.MessageUnauthorized, resBody.Message)
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if user is deleted",
			ReqFn: func(t *testing.T) (dtos.UpdateEmailBody, string) {
				fakeUserData := GenerateFakeUserData(t)
				testUser := CreateFakeTestUser(t)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				reqBody := generateTestUpdateEmailData(t, fakeUserData.Password)
				return reqBody, accessToken
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ dtos.UpdateEmailBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, services.MessageUnauthorized, resBody.Message)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodPost, updateEmailPath, tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}
