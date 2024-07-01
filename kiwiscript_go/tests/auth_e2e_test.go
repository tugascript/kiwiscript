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

// TODO: Add string type matches (e.g. UUID, jwt, etc.)

type fakeRegisterData struct {
	Email     string `faker:"email"`
	FirstName string `faker:"first_name"`
	LastName  string `faker:"last_name"`
	Location  string `faker:"oneof: NZL, AUS, NAM, EUR, OTH"`
	BirthDate string `faker:"date"`
	Password  string `faker:"oneof: Pas@w0rd123, P@sW0rd456, P@ssw0rd789, P@ssW0rd012, P@ssw0rd!345"`
}

func generateFakeRegisterData(t *testing.T) controllers.SignUpRequest {
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

	t.Run("Should register a user", func(t *testing.T) {
		// Arrange
		reqBody := generateFakeRegisterData(t)
		jsonBody := CreateTestJSONRequestBody(t, reqBody)
		app := GetTestApp(t)
		delayMs := 50
		method := MethodPost

		// Act
		resp := PerformTestRequest(t, app, delayMs, method, registerPath, "", jsonBody)
		defer resp.Body.Close()

		// Assert
		AssertTestStatusCode(t, resp, fiber.StatusCreated)
		resBody := AssertTestResponseBody(t, resp, controllers.MessageResponse{})
		AssertEqual(t, "Confirmation email has been sent", resBody.Message)
	})

	t.Run("Should return 400 BAD REQUEST if request validation fails", func(t *testing.T) {
		// Arrange
		reqBody := generateFakeRegisterData(t)
		reqBody.Email = "notAnEmail"
		jsonBody := CreateTestJSONRequestBody(t, reqBody)
		app := GetTestApp(t)
		delayMs := 0
		method := MethodPost

		// Act
		resp := PerformTestRequest(t, app, delayMs, method, registerPath, "", jsonBody)
		defer resp.Body.Close()

		// Assert
		AssertTestStatusCode(t, resp, fiber.StatusBadRequest)
		resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
		AssertEqual(t, 1, len(resBody.Fields))
		AssertEqual(t, "email", resBody.Fields[0].Param)
		AssertEqual(t, "Field validation for 'Email' failed on the 'email' tag", resBody.Fields[0].Error)
		AssertEqual(t, reqBody.Email, resBody.Fields[0].Value.(string))
	})

	t.Run("Should return 409 CONFLICT if email already exists", func(t *testing.T) {
		// Arrange
		testUser := CreateTestUser(t, nil)
		reqBody := generateFakeRegisterData(t)
		reqBody.Email = testUser.Email
		jsonBody := CreateTestJSONRequestBody(t, reqBody)
		app := GetTestApp(t)
		delayMs := 0
		method := MethodPost

		// Act
		resp := PerformTestRequest(t, app, delayMs, method, registerPath, "", jsonBody)
		defer resp.Body.Close()

		// Assert
		AssertTestStatusCode(t, resp, fiber.StatusConflict)
		resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
		AssertEqual(t, "Email already exists", resBody.Message)
	})

	t.Cleanup(func() {
		// Clean up test data
		dbProv := GetTestDatabase(t)
		dbProv.DeleteAllUsers(context.Background())
	})
}

// ----- Test Confirm Email -----
const confirmEmailPath = "/api/auth/confirm-email"

func generateTestConfirmEmailData(t *testing.T, user db.User) controllers.ConfirmRequest {
	emailToken := generateEmailToken(t, tokens.EmailTokenConfirmation, user)

	return controllers.ConfirmRequest{
		ConfirmationToken: emailToken,
	}
}

// Happy Path
func TestConfirmEmail(t *testing.T) {
	unauthorizedFunc := func(t *testing.T, _ controllers.ConfirmRequest, resp *http.Response) {
		resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
		AssertEqual(t, services.MessageUnauthorized, resBody.Message)
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
				AssertEqual(t, "Field validation for 'ConfirmationToken' failed on the 'jwt' tag", resBody.Fields[0].Error)
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

	t.Cleanup(func() {
		// Clean up test data
		dbProv := GetTestDatabase(t)
		dbProv.DeleteAllUsers(context.Background())
	})
}

// ----- Test Login -----
const loginPath = "/api/auth/login"

// Happy Path
func TestLogin(t *testing.T) {
	// Arrange
	fakeUserData := GenerateFakeUserData(t)
	testUser := confirmTestUser(t, CreateTestUser(t, &fakeUserData).ID)
	reqBody := controllers.SignInRequest{Email: testUser.Email, Password: fakeUserData.Password}
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 50, MethodPost, loginPath, "", jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusOK)
	AssertTestResponseBody(t, resp, controllers.MessageResponse{})
}

// Sad Path
func TestLoginValidationErr(t *testing.T) {
	// Arrange
	reqBody := controllers.SignInRequest{Email: "notAnEmail", Password: ""}
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, loginPath, "", jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusBadRequest)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
	AssertEqual(t, 2, len(resBody.Fields))
	AssertEqual(t, "email", resBody.Fields[0].Param)
	AssertEqual(t, "Field validation for 'Email' failed on the 'email' tag", resBody.Fields[0].Error)
	AssertEqual(t, reqBody.Email, resBody.Fields[0].Value.(string))
	AssertEqual(t, "password", resBody.Fields[1].Param)
	AssertEqual(t, "Field validation for 'Password' failed on the 'required' tag", resBody.Fields[1].Error)
	AssertEqual(t, reqBody.Password, resBody.Fields[1].Value.(string))
}

func TestLoginNotFound(t *testing.T) {
	// Arrange
	fakeEmail := faker.Email()
	fakePassword := faker.Password()
	reqBody := controllers.SignInRequest{Email: fakeEmail, Password: fakePassword}
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, loginPath, "", jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusUnauthorized)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
	AssertEqual(t, services.MessageUnauthorized, resBody.Message)
}

func TestLoginUserNotConfirmed(t *testing.T) {
	// Arrange
	fakeUserData := GenerateFakeUserData(t)
	testUser := CreateTestUser(t, &fakeUserData)
	reqBody := controllers.SignInRequest{Email: testUser.Email, Password: fakeUserData.Password}
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, loginPath, "", jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusBadRequest)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
	AssertEqual(t, "User not confirmed", resBody.Message)
}

// ----- Test Login Confirm -----
const loginConfirmPath = "/api/auth/login/confirm"

func generateTestTwoFactorCode(t *testing.T, user db.User) string {
	cacheProv := GetTestCache(t)
	code, err := cacheProv.AddTwoFactorCode(user.ID)

	if err != nil {
		t.Fatal("Failed to generate two factor code", err)
	}

	return code
}

func TestLoginConfirm(t *testing.T) {
	// Arrange
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	app := GetTestApp(t)
	code := generateTestTwoFactorCode(t, testUser)
	reqBody := controllers.ConfirmSignInRequest{Email: testUser.Email, Code: code}
	jsonBody := CreateTestJSONRequestBody(t, reqBody)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, loginConfirmPath, "", jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusOK)
	resBody := AssertTestResponseBody(t, resp, controllers.AuthResponse{})
	AssertEqual(t, "Bearer", resBody.TokenType)
}

func TestLoginConfirmUnauthorizedErr(t *testing.T) {
	// Arrange
	testUser := CreateFakeTestUser(t)
	app := GetTestApp(t)
	code := "123456"
	reqBody := controllers.ConfirmSignInRequest{Email: testUser.Email, Code: code}
	jsonBody := CreateTestJSONRequestBody(t, reqBody)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, loginConfirmPath, "", jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusUnauthorized)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
	AssertEqual(t, services.MessageUnauthorized, resBody.Message)
}

func TestLoginConfirmValidationErr(t *testing.T) {
	// Arrange
	reqBody := controllers.ConfirmSignInRequest{Email: "notAnEmail", Code: ""}
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, loginConfirmPath, "", jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusBadRequest)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
	AssertEqual(t, 2, len(resBody.Fields))
	AssertEqual(t, "email", resBody.Fields[0].Param)
	AssertEqual(t, "Field validation for 'Email' failed on the 'email' tag", resBody.Fields[0].Error)
	AssertEqual(t, reqBody.Email, resBody.Fields[0].Value.(string))
	AssertEqual(t, "code", resBody.Fields[1].Param)
	AssertEqual(t, "Field validation for 'Code' failed on the 'required' tag", resBody.Fields[1].Error)
	AssertEqual(t, reqBody.Code, resBody.Fields[1].Value.(string))
}

func TestLoginConfirmNotFound(t *testing.T) {
	// Arrange
	fakeUser := CreateFakeTestUser(t)
	reqBody := controllers.ConfirmSignInRequest{Email: fakeUser.Email, Code: "123456"}
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, loginConfirmPath, "", jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusUnauthorized)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
	AssertEqual(t, services.MessageUnauthorized, resBody.Message)
}

func TestLoginConfirmUserNotConfirmed(t *testing.T) {
	// Arrange
	testUser := CreateTestUser(t, nil)
	app := GetTestApp(t)
	code := generateTestTwoFactorCode(t, testUser)
	reqBody := controllers.ConfirmSignInRequest{Email: testUser.Email, Code: code}
	jsonBody := CreateTestJSONRequestBody(t, reqBody)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, loginConfirmPath, "", jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusBadRequest)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
	AssertEqual(t, "User not confirmed", resBody.Message)
}

// ----- Logout -----
const logoutPath = "/api/auth/logout"

// TODO: use table tests for logout body and cookie

func TestBodyLogout(t *testing.T) {
	// Arrange
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	accessToken, refreshToken := GenerateTestAuthTokens(t, testUser)
	app := GetTestApp(t)
	reqBody := controllers.SignOutRequest{RefreshToken: refreshToken}
	jsonBody := CreateTestJSONRequestBody(t, reqBody)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, logoutPath, accessToken, jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusNoContent)
}

func TestCookieLogout(t *testing.T) {
	// Arrange
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	accessToken, refreshToken := GenerateTestAuthTokens(t, testUser)
	app := GetTestApp(t)

	// Act
	resp := performCookieRequest(t, app, logoutPath, accessToken, refreshToken)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusNoContent)
}

func TestBodyLogoutUnauthorizedErr(t *testing.T) {
	// Arrange
	testUser := CreateFakeTestUser(t)
	_, refreshToken := GenerateTestAuthTokens(t, testUser)
	app := GetTestApp(t)
	reqBody := controllers.SignOutRequest{RefreshToken: refreshToken}
	jsonBody := CreateTestJSONRequestBody(t, reqBody)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, logoutPath, "invalid", jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusUnauthorized)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
	AssertEqual(t, services.MessageUnauthorized, resBody.Message)
}

func TestCookieLogoutUnauthorizedErr(t *testing.T) {
	// Arrange
	testUser := CreateFakeTestUser(t)
	_, refreshToken := GenerateTestAuthTokens(t, testUser)
	app := GetTestApp(t)

	// Act
	resp := performCookieRequest(t, app, logoutPath, "invalid", refreshToken)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusUnauthorized)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
	AssertEqual(t, services.MessageUnauthorized, resBody.Message)
}

func TestBodyLogoutValidationErr(t *testing.T) {
	// Arrange
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	accessToken, _ := GenerateTestAuthTokens(t, testUser)
	app := GetTestApp(t)
	reqBody := controllers.SignOutRequest{RefreshToken: ""}
	jsonBody := CreateTestJSONRequestBody(t, reqBody)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, logoutPath, accessToken, jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusBadRequest)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
	AssertEqual(t, 1, len(resBody.Fields))
	AssertEqual(t, "refresh_token", resBody.Fields[0].Param)
	AssertEqual(t, "Field validation for 'RefreshToken' failed on the 'required' tag", resBody.Fields[0].Error)
	AssertEqual(t, reqBody.RefreshToken, resBody.Fields[0].Value.(string))
}

func TestCookieLogoutValidationErr(t *testing.T) {
	// Arrange
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	accessToken, _ := GenerateTestAuthTokens(t, testUser)
	app := GetTestApp(t)

	// Act
	resp := performCookieRequest(t, app, logoutPath, accessToken, "")
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusBadRequest)
	resBody := AssertTestResponseBody(t, resp, controllers.EmptyRequestValidationError{})
	AssertEqual(t, controllers.RequestValidationMessage, resBody.Message)
}

// ----- Test Refresh -----
const refreshPath = "/api/auth/refresh"

// TODO: use table tests for refresh body and cookie

func blackListRefreshToken(t *testing.T, refreshToken string) {
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

func TestBodyRefresh(t *testing.T) {
	// Arrange
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	_, refreshToken := GenerateTestAuthTokens(t, testUser)
	app := GetTestApp(t)
	reqBody := controllers.RefreshRequest{RefreshToken: refreshToken}
	jsonBody := CreateTestJSONRequestBody(t, reqBody)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, refreshPath, "", jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusOK)
	resBody := AssertTestResponseBody(t, resp, controllers.AuthResponse{})
	AssertEqual(t, "Bearer", resBody.TokenType)
}

func TestCookieRefresh(t *testing.T) {
	// Arrange
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	_, refreshToken := GenerateTestAuthTokens(t, testUser)
	app := GetTestApp(t)

	// Act
	resp := performCookieRequest(t, app, refreshPath, "", refreshToken)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusOK)
	resBody := AssertTestResponseBody(t, resp, controllers.AuthResponse{})
	AssertEqual(t, "Bearer", resBody.TokenType)
}

func TestCookieRefreshVersionMissmatchErr(t *testing.T) {
	// Arrange
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	testUser.Version = math.MaxInt16
	_, refreshToken := GenerateTestAuthTokens(t, testUser)
	app := GetTestApp(t)

	// Act
	resp := performCookieRequest(t, app, refreshPath, "", refreshToken)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusUnauthorized)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
	AssertEqual(t, services.MessageUnauthorized, resBody.Message)
}

func TestBodyRefreshVersionMissmatchErr(t *testing.T) {
	// Arrange
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	testUser.Version = math.MaxInt16
	_, refreshToken := GenerateTestAuthTokens(t, testUser)
	app := GetTestApp(t)
	reqBody := controllers.RefreshRequest{RefreshToken: refreshToken}
	jsonBody := CreateTestJSONRequestBody(t, reqBody)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, refreshPath, "", jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusUnauthorized)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
	AssertEqual(t, services.MessageUnauthorized, resBody.Message)
}

func TestCookieRefreshBlacklistedErr(t *testing.T) {
	// Arrange
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	_, refreshToken := GenerateTestAuthTokens(t, testUser)
	blackListRefreshToken(t, refreshToken)
	app := GetTestApp(t)

	// Act
	resp := performCookieRequest(t, app, refreshPath, "", refreshToken)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusUnauthorized)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
	AssertEqual(t, services.MessageUnauthorized, resBody.Message)
}

func TestBodyRefreshBlacklistedErr(t *testing.T) {
	// Arrange
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	_, refreshToken := GenerateTestAuthTokens(t, testUser)
	blackListRefreshToken(t, refreshToken)
	app := GetTestApp(t)
	reqBody := controllers.RefreshRequest{RefreshToken: refreshToken}
	jsonBody := CreateTestJSONRequestBody(t, reqBody)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, refreshPath, "", jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusUnauthorized)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
	AssertEqual(t, services.MessageUnauthorized, resBody.Message)
}

func TestBodyRefreshDeletedUserErr(t *testing.T) {
	// Arrange
	testUser := CreateFakeTestUser(t)
	_, refreshToken := GenerateTestAuthTokens(t, testUser)
	app := GetTestApp(t)
	reqBody := controllers.RefreshRequest{RefreshToken: refreshToken}
	jsonBody := CreateTestJSONRequestBody(t, reqBody)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, refreshPath, "", jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusUnauthorized)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
	AssertEqual(t, services.MessageUnauthorized, resBody.Message)
}

func TestCookieRefreshDeletedUserErr(t *testing.T) {
	// Arrange
	testUser := CreateFakeTestUser(t)
	_, refreshToken := GenerateTestAuthTokens(t, testUser)
	app := GetTestApp(t)

	// Act
	resp := performCookieRequest(t, app, refreshPath, "", refreshToken)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusUnauthorized)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
	AssertEqual(t, services.MessageUnauthorized, resBody.Message)
}

func TestBodyRefreshValidationErr(t *testing.T) {
	// Arrange
	app := GetTestApp(t)
	reqBody := controllers.RefreshRequest{RefreshToken: "not-jwt"}
	jsonBody := CreateTestJSONRequestBody(t, reqBody)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, refreshPath, "", jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusBadRequest)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
	AssertEqual(t, 1, len(resBody.Fields))
	AssertEqual(t, "refresh_token", resBody.Fields[0].Param)
	AssertEqual(t, "Field validation for 'RefreshToken' failed on the 'jwt' tag", resBody.Fields[0].Error)
	AssertEqual(t, reqBody.RefreshToken, resBody.Fields[0].Value.(string))
}

// ----- Test Forgot Password -----
const forgotPasswordPath = "/api/auth/forgot-password"

func TestForgotPasswordWithExistingUser(t *testing.T) {
	// Arrange
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	reqBody := controllers.ForgotPasswordRequest{Email: testUser.Email}
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, forgotPasswordPath, "", jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusOK)
	resBody := AssertTestResponseBody(t, resp, controllers.MessageResponse{})
	AssertEqual(t, "If the email exists, a password reset email has been sent", resBody.Message)
}

func TestForgotPasswordWithNonExistingUser(t *testing.T) {
	// Arrange
	fakeEmail := faker.Email()
	reqBody := controllers.ForgotPasswordRequest{Email: fakeEmail}
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, forgotPasswordPath, "", jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusOK)
	resBody := AssertTestResponseBody(t, resp, controllers.MessageResponse{})
	AssertEqual(t, "If the email exists, a password reset email has been sent", resBody.Message)
}

func TestForgotPasswordValidationErr(t *testing.T) {
	// Arrange
	reqBody := controllers.ForgotPasswordRequest{Email: "notAnEmail"}
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, forgotPasswordPath, "", jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusBadRequest)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
	AssertEqual(t, 1, len(resBody.Fields))
	AssertEqual(t, "email", resBody.Fields[0].Param)
	AssertEqual(t, "Field validation for 'Email' failed on the 'email' tag", resBody.Fields[0].Error)
	AssertEqual(t, reqBody.Email, resBody.Fields[0].Value.(string))
}

// ----- Test Reset Password -----
const resetPasswordPath = "/api/auth/reset-password"

func generateTestResetPasswordData(t *testing.T, user db.User) controllers.ResetPasswordRequest {
	emailToken := generateEmailToken(t, tokens.EmailTokenReset, user)
	password := faker.Name() + "123!"

	return controllers.ResetPasswordRequest{
		ResetToken: emailToken,
		Password1:  password,
		Password2:  password,
	}
}

func TestResetPassword(t *testing.T) {
	// Arrange
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	reqBody := generateTestResetPasswordData(t, testUser)
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, resetPasswordPath, "", jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusOK)
	resBody := AssertTestResponseBody(t, resp, controllers.MessageResponse{})
	AssertEqual(t, "Password reseted successfully", resBody.Message)
}

func TestResetPasswordValidationErr(t *testing.T) {
	// Arrange
	reqBody := controllers.ResetPasswordRequest{ResetToken: "invalid", Password1: "", Password2: ""}
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, resetPasswordPath, "", jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusBadRequest)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
	AssertEqual(t, 3, len(resBody.Fields))
	AssertEqual(t, "reset_token", resBody.Fields[0].Param)
	AssertEqual(t, "Field validation for 'ResetToken' failed on the 'jwt' tag", resBody.Fields[0].Error)
	AssertEqual(t, reqBody.ResetToken, resBody.Fields[0].Value.(string))
	AssertEqual(t, "password1", resBody.Fields[1].Param)
	AssertEqual(t, "Field validation for 'Password1' failed on the 'required' tag", resBody.Fields[1].Error)
	AssertEqual(t, reqBody.Password1, resBody.Fields[1].Value.(string))
	AssertEqual(t, "password2", resBody.Fields[2].Param)
	AssertEqual(t, "Field validation for 'Password2' failed on the 'required' tag", resBody.Fields[2].Error)
	AssertEqual(t, reqBody.Password2, resBody.Fields[2].Value.(string))
}

func TestResetPasswordInvalidVersionErr(t *testing.T) {
	// Arrange
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	testUser.Version = math.MaxInt16
	reqBody := generateTestResetPasswordData(t, testUser)
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, resetPasswordPath, "", jsonBody)

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusUnauthorized)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
	AssertEqual(t, services.MessageUnauthorized, resBody.Message)
}

func TestResetPasswordInvalidDeletedUserErr(t *testing.T) {
	// Arrange
	testUser := CreateFakeTestUser(t)
	reqBody := generateTestResetPasswordData(t, testUser)
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, resetPasswordPath, "", jsonBody)

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusUnauthorized)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
	AssertEqual(t, services.MessageUnauthorized, resBody.Message)
}

// ----- Test Update Password -----
const updatePasswordPath = "/api/auth/update-password"

func generateTestUpdatePasswordData(_ *testing.T, oldPassword string) controllers.UpdatePasswordRequest {
	newPassword := faker.Name() + "123!"

	return controllers.UpdatePasswordRequest{
		OldPassword: oldPassword,
		Password1:   newPassword,
		Password2:   newPassword,
	}
}

func TestUpdatePassword(t *testing.T) {
	// Arrange
	fakeUserData := GenerateFakeUserData(t)
	testUser := confirmTestUser(t, CreateTestUser(t, &fakeUserData).ID)
	accessToken, _ := GenerateTestAuthTokens(t, testUser)
	reqBody := generateTestUpdatePasswordData(t, fakeUserData.Password)
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, updatePasswordPath, accessToken, jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusOK)
	resBody := AssertTestResponseBody(t, resp, controllers.AuthResponse{})
	AssertEqual(t, "Bearer", resBody.TokenType)
}

func TestUpdatePasswordValidationErr(t *testing.T) {
	// Arrange
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	accessToken, _ := GenerateTestAuthTokens(t, testUser)
	app := GetTestApp(t)
	reqBody := controllers.UpdatePasswordRequest{OldPassword: "", Password1: "", Password2: ""}
	jsonBody := CreateTestJSONRequestBody(t, reqBody)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, updatePasswordPath, accessToken, jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusBadRequest)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
	AssertEqual(t, 3, len(resBody.Fields))
	AssertEqual(t, "old_password", resBody.Fields[0].Param)
	AssertEqual(t, "Field validation for 'OldPassword' failed on the 'required' tag", resBody.Fields[0].Error)
	AssertEqual(t, reqBody.OldPassword, resBody.Fields[0].Value.(string))
	AssertEqual(t, "password1", resBody.Fields[1].Param)
	AssertEqual(t, "Field validation for 'Password1' failed on the 'required' tag", resBody.Fields[1].Error)
	AssertEqual(t, reqBody.Password1, resBody.Fields[1].Value.(string))
	AssertEqual(t, "password2", resBody.Fields[2].Param)
	AssertEqual(t, "Field validation for 'Password2' failed on the 'required' tag", resBody.Fields[2].Error)
}

func TestUpdatePasswordInvalidVersionErr(t *testing.T) {
	// Arrange
	fakeUserData := GenerateFakeUserData(t)
	testUser := confirmTestUser(t, CreateTestUser(t, &fakeUserData).ID)
	testUser.Version = math.MaxInt16
	accessToken, _ := GenerateTestAuthTokens(t, testUser)
	reqBody := generateTestUpdatePasswordData(t, fakeUserData.Password)
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, updatePasswordPath, accessToken, jsonBody)

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusUnauthorized)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
	AssertEqual(t, services.MessageUnauthorized, resBody.Message)
}

func TestUpdatePasswordInvalidDeletedUserErr(t *testing.T) {
	// Arrange
	fakeUserData := GenerateFakeUserData(t)
	testUser := CreateFakeTestUser(t)
	accessToken, _ := GenerateTestAuthTokens(t, testUser)
	reqBody := generateTestUpdatePasswordData(t, fakeUserData.Password)
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, updatePasswordPath, accessToken, jsonBody)

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusUnauthorized)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
	AssertEqual(t, services.MessageUnauthorized, resBody.Message)
}

func TestUpdatePasswordUnauthorizedErr(t *testing.T) {
	// Arrange
	app := GetTestApp(t)
	reqBody := generateTestUpdatePasswordData(t, faker.Password())
	jsonBody := CreateTestJSONRequestBody(t, reqBody)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, updatePasswordPath, "", jsonBody)

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusUnauthorized)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
	AssertEqual(t, services.MessageUnauthorized, resBody.Message)
}

// ----- Test Update Email -----
const updateEmailPath = "/api/auth/update-email"

func generateTestUpdateEmailData(_ *testing.T, password string) controllers.UpdateEmailRequest {

	return controllers.UpdateEmailRequest{
		Email:    faker.Email(),
		Password: password,
	}
}

func TestUpdateEmail(t *testing.T) {
	// Arrange
	fakeUserData := GenerateFakeUserData(t)
	testUser := confirmTestUser(t, CreateTestUser(t, &fakeUserData).ID)
	accessToken, _ := GenerateTestAuthTokens(t, testUser)
	reqBody := generateTestUpdateEmailData(t, fakeUserData.Password)
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, updateEmailPath, accessToken, jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusOK)
	resBody := AssertTestResponseBody(t, resp, controllers.AuthResponse{})
	AssertEqual(t, "Bearer", resBody.TokenType)
}

func TestUpdateEmailValidationErr(t *testing.T) {
	// Arrange
	fakeUserData := GenerateFakeUserData(t)
	testUser := confirmTestUser(t, CreateTestUser(t, &fakeUserData).ID)
	accessToken, _ := GenerateTestAuthTokens(t, testUser)
	app := GetTestApp(t)
	reqBody := controllers.UpdateEmailRequest{Email: "", Password: fakeUserData.Password}
	jsonBody := CreateTestJSONRequestBody(t, reqBody)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, updateEmailPath, accessToken, jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusBadRequest)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
	AssertEqual(t, 1, len(resBody.Fields))
	AssertEqual(t, "email", resBody.Fields[0].Param)
	AssertEqual(t, "Field validation for 'Email' failed on the 'required' tag", resBody.Fields[0].Error)
}

func TestUpdateEmailInvalidVersionErr(t *testing.T) {
	// Arrange
	fakeUserData := GenerateFakeUserData(t)
	testUser := confirmTestUser(t, CreateTestUser(t, &fakeUserData).ID)
	testUser.Version = math.MaxInt16
	accessToken, _ := GenerateTestAuthTokens(t, testUser)
	reqBody := generateTestUpdateEmailData(t, fakeUserData.Password)
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, updateEmailPath, accessToken, jsonBody)

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusUnauthorized)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
	AssertEqual(t, services.MessageUnauthorized, resBody.Message)
}

func TestUpdateEmailInvalidDeletedUserErr(t *testing.T) {
	// Arrange
	fakeUserData := GenerateFakeUserData(t)
	testUser := CreateFakeTestUser(t)
	accessToken, _ := GenerateTestAuthTokens(t, testUser)
	reqBody := generateTestUpdateEmailData(t, fakeUserData.Password)
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, updateEmailPath, accessToken, jsonBody)

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusUnauthorized)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
	AssertEqual(t, services.MessageUnauthorized, resBody.Message)
}
