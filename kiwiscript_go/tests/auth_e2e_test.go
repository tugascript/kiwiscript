package tests

import (
	"context"
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

func generateEmailToken(t *testing.T, tokenType string, user db.User) string {
	tokensProv := GetTestTokens(t)
	emailToken, err := tokensProv.CreateEmailToken(tokenType, user)

	if err != nil {
		t.Fatal("Failed to generate test emailToken", err)
	}

	return emailToken
}

func confirmTestUser(t *testing.T, userID int32) db.User {
	serv := GetTestServices(t)
	user, err := serv.ConfirmUser(context.Background(), userID)

	if err != nil {
		t.Fatal(err)
	}

	return user
}

func assertOAuthResponse(t *testing.T, resp *http.Response) {
	resBody := AssertTestResponseBody(t, resp, controllers.AuthResponse{})
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
	req := httptest.NewRequest(MethodPost, path, nil)
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

	generateFakeRegisterData := func(t *testing.T) controllers.SignUpRequest {
		fakeData := fakeRegisterData{}
		if err := faker.FakeData(&fakeData); err != nil {
			t.Fatal("Failed to generate fake data", err)
		}
		return controllers.SignUpRequest{
			Email:     fakeData.Email,
			FirstName: fakeData.FirstName,
			LastName:  fakeData.LastName,
			Location:  fakeData.Location,
			BirthDate: fakeData.BirthDate,
			Password1: fakeData.Password,
			Password2: fakeData.Password,
		}
	}

	testCases := []TestRequestCase[controllers.SignUpRequest]{
		{
			Name: "Should return 200 OK registering a user",
			ReqFn: func(t *testing.T) (controllers.SignUpRequest, string) {
				return generateFakeRegisterData(t), ""
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ controllers.SignUpRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.MessageResponse{})
				AssertEqual(t, "Confirmation email has been sent", resBody.Message)
				AssertNotEmpty(t, resBody.ID)
			},
			DelayMs: 50,
		},
		{
			Name: "Should return 400 BAD REQUEST if request validation fails",
			ReqFn: func(t *testing.T) (controllers.SignUpRequest, string) {
				reqBody := generateFakeRegisterData(t)
				reqBody.Email = "notAnEmail"
				return reqBody, ""
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req controllers.SignUpRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
				AssertEqual(t, 1, len(resBody.Fields))
				AssertEqual(t, "email", resBody.Fields[0].Param)
				AssertEqual(t, controllers.StrFieldErrMessageEmail, resBody.Fields[0].Message)
				AssertEqual(t, req.Email, resBody.Fields[0].Value.(string))
			},
		},
		{
			Name: "Should return 400 BAD REQUEST if password missmatch",
			ReqFn: func(t *testing.T) (controllers.SignUpRequest, string) {
				reqBody := generateFakeRegisterData(t)
				reqBody.Password2 = "differentPassword"
				return reqBody, ""
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req controllers.SignUpRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
				AssertEqual(t, 1, len(resBody.Fields))
				AssertEqual(t, "password2", resBody.Fields[0].Param)
				AssertEqual(t, controllers.FieldErrMessageEqField, resBody.Fields[0].Message)
				AssertEqual(t, req.Password2, resBody.Fields[0].Value.(string))
			},
		},
		{
			Name: "Should return 409 CONFLICT if email already exists",
			ReqFn: func(t *testing.T) (controllers.SignUpRequest, string) {
				testUser := CreateTestUser(t, nil)
				reqBody := generateFakeRegisterData(t)
				reqBody.Email = testUser.Email
				return reqBody, ""
			},
			ExpStatus: fiber.StatusConflict,
			AssertFn: func(t *testing.T, _ controllers.SignUpRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, "Email already exists", resBody.Message)
				AssertEqual(t, controllers.StatusDuplicateKey, resBody.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, registerPath, tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}

func TestConfirmEmail(t *testing.T) {
	const confirmEmailPath = "/api/auth/confirm-email"

	generateTestConfirmEmailData := func(t *testing.T, user db.User) controllers.ConfirmRequest {
		emailToken := generateEmailToken(t, tokens.EmailTokenConfirmation, user)

		return controllers.ConfirmRequest{
			ConfirmationToken: emailToken,
		}
	}
	unauthorizedFunc := func(t *testing.T, _ controllers.ConfirmRequest, resp *http.Response) {
		assertUnauthorizeError(t, resp)
	}

	testCases := []TestRequestCase[controllers.ConfirmRequest]{
		{
			Name: "Should return 200 OK with OAuth response",
			ReqFn: func(t *testing.T) (controllers.ConfirmRequest, string) {
				testUser := CreateTestUser(t, nil)
				return generateTestConfirmEmailData(t, testUser), ""
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ controllers.ConfirmRequest, resp *http.Response) {
				assertOAuthResponse(t, resp)
			},
		},
		{
			Name: "Should return 400 BAD REQUEST if request validation fails",
			ReqFn: func(t *testing.T) (controllers.ConfirmRequest, string) {
				return controllers.ConfirmRequest{ConfirmationToken: "invalidToken"}, ""
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req controllers.ConfirmRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
				AssertEqual(t, 1, len(resBody.Fields))
				AssertEqual(t, "confirmation_token", resBody.Fields[0].Param)
				AssertEqual(t, controllers.StrFieldErrMessageJWT, resBody.Fields[0].Message)
				AssertEqual(t, req.ConfirmationToken, resBody.Fields[0].Value.(string))
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if user not found",
			ReqFn: func(t *testing.T) (controllers.ConfirmRequest, string) {
				fakeUser := CreateFakeTestUser(t)
				return generateTestConfirmEmailData(t, fakeUser), ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn:  unauthorizedFunc,
		},
		{
			Name: "Should return 401 UNAUTHORIZED if user already confirmed",
			ReqFn: func(t *testing.T) (controllers.ConfirmRequest, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				return generateTestConfirmEmailData(t, testUser), ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn:  unauthorizedFunc,
		},
		{
			Name: "Should return 401 UNAUTHORIZED for version missmatch",
			ReqFn: func(t *testing.T) (controllers.ConfirmRequest, string) {
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
			PerformTestRequestCase(t, confirmEmailPath, tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}

func TestLogin(t *testing.T) {
	const loginPath = "/api/auth/login"

	testCases := []TestRequestCase[controllers.SignInRequest]{
		{
			Name: "Should return 200 OK with message response",
			ReqFn: func(t *testing.T) (controllers.SignInRequest, string) {
				fakeUserData := GenerateFakeUserData(t)
				testUser := confirmTestUser(t, CreateTestUser(t, &fakeUserData).ID)
				reqBody := controllers.SignInRequest{Email: testUser.Email, Password: fakeUserData.Password}
				return reqBody, ""
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ controllers.SignInRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.MessageResponse{})
				AssertEqual(t, "Confirmation code has been sent to your email", resBody.Message)
				AssertNotEmpty(t, resBody.ID)
			},
			DelayMs: 50,
		},
		{
			Name: "Should return 400 BAD REQUEST if request validation fails",
			ReqFn: func(t *testing.T) (controllers.SignInRequest, string) {
				return controllers.SignInRequest{Email: "notAnEmail", Password: ""}, ""
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req controllers.SignInRequest, resp *http.Response) {
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
			ReqFn: func(t *testing.T) (controllers.SignInRequest, string) {
				fakeEmail := faker.Email()
				fakePassword := faker.Password()
				return controllers.SignInRequest{Email: fakeEmail, Password: fakePassword}, ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ controllers.SignInRequest, resp *http.Response) {
				AssertTestStatusCode(t, resp, fiber.StatusUnauthorized)
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, services.MessageUnauthorized, resBody.Message)
				AssertEqual(t, controllers.StatusUnauthorized, resBody.Code)
			},
		},
		{
			Name: "Should return 400 BAD REQUEST if user not confirmed",
			ReqFn: func(t *testing.T) (controllers.SignInRequest, string) {
				fakeUserData := GenerateFakeUserData(t)
				testUser := CreateTestUser(t, &fakeUserData)
				return controllers.SignInRequest{Email: testUser.Email, Password: fakeUserData.Password}, ""
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, _ controllers.SignInRequest, resp *http.Response) {
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
			PerformTestRequestCase(t, loginPath, tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}

func TestLoginConfirm(t *testing.T) {
	const loginConfirmPath = "/api/auth/login/confirm"

	generateTestTwoFactorCode := func(t *testing.T, user db.User) string {
		cacheProv := GetTestCache(t)
		code, err := cacheProv.AddTwoFactorCode(user.ID)

		if err != nil {
			t.Fatal("Failed to generate two factor code", err)
		}

		return code
	}

	testCases := []TestRequestCase[controllers.ConfirmSignInRequest]{
		{
			Name: "Should return 200 OK with OAuth response",
			ReqFn: func(t *testing.T) (controllers.ConfirmSignInRequest, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				code := generateTestTwoFactorCode(t, testUser)
				return controllers.ConfirmSignInRequest{Email: testUser.Email, Code: code}, ""
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ controllers.ConfirmSignInRequest, resp *http.Response) {
				assertOAuthResponse(t, resp)
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if code is wrong",
			ReqFn: func(t *testing.T) (controllers.ConfirmSignInRequest, string) {
				fakeUser := CreateFakeTestUser(t)
				return controllers.ConfirmSignInRequest{Email: fakeUser.Email, Code: "123456"}, ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ controllers.ConfirmSignInRequest, resp *http.Response) {
				assertUnauthorizeError(t, resp)
			},
		},
		{
			Name: "Should return 400 BAD REQUEST if request validation fails",
			ReqFn: func(t *testing.T) (controllers.ConfirmSignInRequest, string) {
				return controllers.ConfirmSignInRequest{Email: "notAnEmail", Code: ""}, ""
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req controllers.ConfirmSignInRequest, resp *http.Response) {
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
			ReqFn: func(t *testing.T) (controllers.ConfirmSignInRequest, string) {
				fakeUser := CreateFakeTestUser(t)
				return controllers.ConfirmSignInRequest{Email: fakeUser.Email, Code: "123456"}, ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ controllers.ConfirmSignInRequest, resp *http.Response) {
				assertUnauthorizeError(t, resp)
			},
		},
		{
			Name: "Should return 400 BAD REQUEST if user not confirmed",
			ReqFn: func(t *testing.T) (controllers.ConfirmSignInRequest, string) {
				testUser := CreateTestUser(t, nil)
				code := generateTestTwoFactorCode(t, testUser)
				return controllers.ConfirmSignInRequest{Email: testUser.Email, Code: code}, ""
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, _ controllers.ConfirmSignInRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, "User not confirmed", resBody.Message)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, loginConfirmPath, tc)
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

	bodyTestCases := []TestRequestCase[controllers.SignOutRequest]{
		{
			Name: "Should return 204 NO CONTENT",
			ReqFn: func(t *testing.T) (controllers.SignOutRequest, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				accessToken, refreshToken := GenerateTestAuthTokens(t, testUser)
				return controllers.SignOutRequest{RefreshToken: refreshToken}, accessToken
			},
			ExpStatus: fiber.StatusNoContent,
			AssertFn:  func(t *testing.T, _ controllers.SignOutRequest, resp *http.Response) {},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if access token is invalid",
			ReqFn: func(t *testing.T) (controllers.SignOutRequest, string) {
				testUser := CreateFakeTestUser(t)
				_, refreshToken := GenerateTestAuthTokens(t, testUser)
				return controllers.SignOutRequest{RefreshToken: refreshToken}, "invalid"
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ controllers.SignOutRequest, resp *http.Response) {
				assertUnauthorizeError(t, resp)
			},
		},
		{
			Name: "Should return 400 BAD REQUEST if request validation fails",
			ReqFn: func(t *testing.T) (controllers.SignOutRequest, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return controllers.SignOutRequest{RefreshToken: "not-a-jwt"}, accessToken
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req controllers.SignOutRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
				AssertEqual(t, 1, len(resBody.Fields))
				AssertEqual(t, "refresh_token", resBody.Fields[0].Param)
				AssertEqual(t, controllers.StrFieldErrMessageJWT, resBody.Fields[0].Message)
				AssertEqual(t, req.RefreshToken, resBody.Fields[0].Value.(string))
			},
		},
	}

	for _, tc := range bodyTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, logoutPath, tc)
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
			defer resp.Body.Close()

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

	bodyTestCases := []TestRequestCase[controllers.RefreshRequest]{
		{
			Name: "Should return 200 OK",
			ReqFn: func(t *testing.T) (controllers.RefreshRequest, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				_, refreshToken := GenerateTestAuthTokens(t, testUser)
				return controllers.RefreshRequest{RefreshToken: refreshToken}, ""
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ controllers.RefreshRequest, resp *http.Response) {
				assertOAuthResponse(t, resp)
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if refresh token is blacklisted",
			ReqFn: func(t *testing.T) (controllers.RefreshRequest, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				_, refreshToken := GenerateTestAuthTokens(t, testUser)
				blackListRefreshToken(t, refreshToken)
				return controllers.RefreshRequest{RefreshToken: refreshToken}, ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ controllers.RefreshRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, services.MessageUnauthorized, resBody.Message)
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if user version mismatches",
			ReqFn: func(t *testing.T) (controllers.RefreshRequest, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				testUser.Version = math.MaxInt16
				_, refreshToken := GenerateTestAuthTokens(t, testUser)
				return controllers.RefreshRequest{RefreshToken: refreshToken}, ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ controllers.RefreshRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, services.MessageUnauthorized, resBody.Message)
			},
		},
		{
			Name: "Should return 400 BAD REQUEST if refresh token is invalid",
			ReqFn: func(t *testing.T) (controllers.RefreshRequest, string) {
				return controllers.RefreshRequest{RefreshToken: "not-jwt"}, ""
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req controllers.RefreshRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
				AssertEqual(t, 1, len(resBody.Fields))
				AssertEqual(t, "refresh_token", resBody.Fields[0].Param)
				AssertEqual(t, controllers.StrFieldErrMessageJWT, resBody.Fields[0].Message)
				AssertEqual(t, req.RefreshToken, resBody.Fields[0].Value.(string))
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if user is deleted",
			ReqFn: func(t *testing.T) (controllers.RefreshRequest, string) {
				testUser := CreateFakeTestUser(t)
				_, refreshToken := GenerateTestAuthTokens(t, testUser)
				return controllers.RefreshRequest{RefreshToken: refreshToken}, ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ controllers.RefreshRequest, resp *http.Response) {
				assertUnauthorizeError(t, resp)
			},
		},
	}

	for _, tc := range bodyTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, refreshPath, tc)
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
				resBody := AssertTestResponseBody(t, resp, controllers.AuthResponse{})
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
			defer resp.Body.Close()

			AssertTestStatusCode(t, resp, tc.ExpStatus)
			tc.AssertFn(t, resp)
		})
	}

	t.Cleanup(userCleanUp(t))
}

func TestForgotPassword(t *testing.T) {
	const forgotPasswordPath = "/api/auth/forgot-password"

	bodyTestCases := []TestRequestCase[controllers.ForgotPasswordRequest]{
		{
			Name: "Should return 200 OK with existing user",
			ReqFn: func(t *testing.T) (controllers.ForgotPasswordRequest, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				return controllers.ForgotPasswordRequest{Email: testUser.Email}, ""
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ controllers.ForgotPasswordRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.MessageResponse{})
				AssertEqual(t, "If the email exists, a password reset email has been sent", resBody.Message)
			},
		},
		{
			Name: "Should return 200 OK with non-existing user",
			ReqFn: func(t *testing.T) (controllers.ForgotPasswordRequest, string) {
				fakeEmail := faker.Email()
				return controllers.ForgotPasswordRequest{Email: fakeEmail}, ""
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ controllers.ForgotPasswordRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.MessageResponse{})
				AssertEqual(t, "If the email exists, a password reset email has been sent", resBody.Message)
			},
		},
		{
			Name: "Should return 400 BAD REQUEST if request validation fails",
			ReqFn: func(t *testing.T) (controllers.ForgotPasswordRequest, string) {
				return controllers.ForgotPasswordRequest{Email: "notAnEmail"}, ""
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req controllers.ForgotPasswordRequest, resp *http.Response) {
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
			PerformTestRequestCase(t, forgotPasswordPath, tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}

func TestResetPassword(t *testing.T) {
	const resetPasswordPath = "/api/auth/reset-password"
	generateTestResetPasswordData := func(t *testing.T, user db.User) controllers.ResetPasswordRequest {
		emailToken := generateEmailToken(t, tokens.EmailTokenReset, user)
		password := faker.Name() + "123!"

		return controllers.ResetPasswordRequest{
			ResetToken: emailToken,
			Password1:  password,
			Password2:  password,
		}
	}

	testCases := []TestRequestCase[controllers.ResetPasswordRequest]{
		{
			Name: "Should return 200 OK",
			ReqFn: func(t *testing.T) (controllers.ResetPasswordRequest, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				reqBody := generateTestResetPasswordData(t, testUser)
				return reqBody, ""
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ controllers.ResetPasswordRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.MessageResponse{})
				AssertEqual(t, "Password reseted successfully", resBody.Message)
			},
		},
		{
			Name: "Should return 400 BAD REQUEST if request validation fails",
			ReqFn: func(t *testing.T) (controllers.ResetPasswordRequest, string) {
				return controllers.ResetPasswordRequest{ResetToken: "invalid", Password1: "a", Password2: "cb"}, ""
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req controllers.ResetPasswordRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
				AssertEqual(t, 3, len(resBody.Fields))
				AssertEqual(t, "reset_token", resBody.Fields[0].Param)
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
			ReqFn: func(t *testing.T) (controllers.ResetPasswordRequest, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				testUser.Version = math.MaxInt16
				reqBody := generateTestResetPasswordData(t, testUser)
				return reqBody, ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ controllers.ResetPasswordRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, services.MessageUnauthorized, resBody.Message)
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if user is deleted",
			ReqFn: func(t *testing.T) (controllers.ResetPasswordRequest, string) {
				testUser := CreateFakeTestUser(t)
				reqBody := generateTestResetPasswordData(t, testUser)
				return reqBody, ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ controllers.ResetPasswordRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, services.MessageUnauthorized, resBody.Message)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, resetPasswordPath, tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}

func TestUpdatePassword(t *testing.T) {
	const updatePasswordPath = "/api/auth/update-password"

	generateTestUpdatePasswordData := func(_ *testing.T, oldPassword string) controllers.UpdatePasswordRequest {
		newPassword := faker.Name() + "123!"
		return controllers.UpdatePasswordRequest{
			OldPassword: oldPassword,
			Password1:   newPassword,
			Password2:   newPassword,
		}
	}

	testCases := []TestRequestCase[controllers.UpdatePasswordRequest]{
		{
			Name: "Should return 200 OK",
			ReqFn: func(t *testing.T) (controllers.UpdatePasswordRequest, string) {
				fakeUserData := GenerateFakeUserData(t)
				testUser := confirmTestUser(t, CreateTestUser(t, &fakeUserData).ID)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				reqBody := generateTestUpdatePasswordData(t, fakeUserData.Password)
				return reqBody, accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ controllers.UpdatePasswordRequest, resp *http.Response) {
				assertOAuthResponse(t, resp)
			},
		},
		{
			Name: "Should return 400 BAD REQUEST if request validation fails",
			ReqFn: func(t *testing.T) (controllers.UpdatePasswordRequest, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return controllers.UpdatePasswordRequest{OldPassword: "", Password1: "a", Password2: "b"}, accessToken
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req controllers.UpdatePasswordRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
				AssertEqual(t, 3, len(resBody.Fields))
				AssertEqual(t, "old_password", resBody.Fields[0].Param)
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
			ReqFn: func(t *testing.T) (controllers.UpdatePasswordRequest, string) {
				fakeUserData := GenerateFakeUserData(t)
				testUser := confirmTestUser(t, CreateTestUser(t, &fakeUserData).ID)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				reqBody := generateTestUpdatePasswordData(t, fakeUserData.Password+"wrong")
				return reqBody, accessToken
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, _ controllers.UpdatePasswordRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, "Old password is incorrect", resBody.Message)
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if user version mismatches",
			ReqFn: func(t *testing.T) (controllers.UpdatePasswordRequest, string) {
				fakeUserData := GenerateFakeUserData(t)
				testUser := confirmTestUser(t, CreateTestUser(t, &fakeUserData).ID)
				testUser.Version = math.MaxInt16
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				reqBody := generateTestUpdatePasswordData(t, fakeUserData.Password)
				return reqBody, accessToken
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ controllers.UpdatePasswordRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, services.MessageUnauthorized, resBody.Message)
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if user is deleted",
			ReqFn: func(t *testing.T) (controllers.UpdatePasswordRequest, string) {
				fakeUserData := GenerateFakeUserData(t)
				testUser := CreateFakeTestUser(t)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				reqBody := generateTestUpdatePasswordData(t, fakeUserData.Password)
				return reqBody, accessToken
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ controllers.UpdatePasswordRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, services.MessageUnauthorized, resBody.Message)
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if access token is missing",
			ReqFn: func(t *testing.T) (controllers.UpdatePasswordRequest, string) {
				reqBody := generateTestUpdatePasswordData(t, faker.Password())
				return reqBody, ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ controllers.UpdatePasswordRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, services.MessageUnauthorized, resBody.Message)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, updatePasswordPath, tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}

func TestUpdateEmail(t *testing.T) {
	const updateEmailPath = "/api/auth/update-email"

	generateTestUpdateEmailData := func(_ *testing.T, password string) controllers.UpdateEmailRequest {
		return controllers.UpdateEmailRequest{
			Email:    faker.Email(),
			Password: password,
		}
	}

	testCases := []TestRequestCase[controllers.UpdateEmailRequest]{
		{
			Name: "Should return 200 OK",
			ReqFn: func(t *testing.T) (controllers.UpdateEmailRequest, string) {
				fakeUserData := GenerateFakeUserData(t)
				testUser := confirmTestUser(t, CreateTestUser(t, &fakeUserData).ID)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				reqBody := generateTestUpdateEmailData(t, fakeUserData.Password)
				return reqBody, accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ controllers.UpdateEmailRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.AuthResponse{})
				AssertEqual(t, "Bearer", resBody.TokenType)
			},
		},
		{
			Name: "Should return 400 BAD REQUEST if request validation fails",
			ReqFn: func(t *testing.T) (controllers.UpdateEmailRequest, string) {
				fakeUserData := GenerateFakeUserData(t)
				testUser := confirmTestUser(t, CreateTestUser(t, &fakeUserData).ID)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return controllers.UpdateEmailRequest{Email: "", Password: fakeUserData.Password}, accessToken
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req controllers.UpdateEmailRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
				AssertEqual(t, 1, len(resBody.Fields))
				AssertEqual(t, "email", resBody.Fields[0].Param)
				AssertEqual(t, controllers.FieldErrMessageRequired, resBody.Fields[0].Message)
				AssertEqual(t, req.Email, resBody.Fields[0].Value.(string))
			},
		},
		{
			Name: "Should return 400 BAD REQUEST if user users wrong password",
			ReqFn: func(t *testing.T) (controllers.UpdateEmailRequest, string) {
				fakeUserData := GenerateFakeUserData(t)
				testUser := confirmTestUser(t, CreateTestUser(t, &fakeUserData).ID)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				reqBody := generateTestUpdateEmailData(t, fakeUserData.Password+"wrong")
				return reqBody, accessToken
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, _ controllers.UpdateEmailRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, "Invalid password", resBody.Message)
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if user version mismatches",
			ReqFn: func(t *testing.T) (controllers.UpdateEmailRequest, string) {
				fakeUserData := GenerateFakeUserData(t)
				testUser := confirmTestUser(t, CreateTestUser(t, &fakeUserData).ID)
				testUser.Version = math.MaxInt16
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				reqBody := generateTestUpdateEmailData(t, fakeUserData.Password)
				return reqBody, accessToken
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ controllers.UpdateEmailRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, services.MessageUnauthorized, resBody.Message)
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if user is deleted",
			ReqFn: func(t *testing.T) (controllers.UpdateEmailRequest, string) {
				fakeUserData := GenerateFakeUserData(t)
				testUser := CreateFakeTestUser(t)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				reqBody := generateTestUpdateEmailData(t, fakeUserData.Password)
				return reqBody, accessToken
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ controllers.UpdateEmailRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, services.MessageUnauthorized, resBody.Message)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, updateEmailPath, tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}
