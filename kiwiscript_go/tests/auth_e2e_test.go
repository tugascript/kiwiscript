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
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/providers/tokens"
	"github.com/kiwiscript/kiwiscript_go/services"
)

// TODO: Parse response bodies

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

// ----- Test Register -----
const registerPath = "/api/auth/register"

// Happy Path
func TestRegister(t *testing.T) {
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
}

// Sad Path
func TestRegisterValidatorErr(t *testing.T) {
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
	AssertEqual(t, reqBody.Email, resBody.Fields[0].Value)
}

func TestRegisterDuplicateEmail(t *testing.T) {
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
	// Arrange
	testUser := CreateTestUser(t, nil)
	reqBody := generateTestConfirmEmailData(t, testUser)
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, confirmEmailPath, "", jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusOK)
	resBody := AssertTestResponseBody(t, resp, controllers.AuthResponse{})
	AssertEqual(t, "Bearer", resBody.TokenType)
}

// Sad Path
func TestConfirmEmailValidationErr(t *testing.T) {
	// Arrange
	reqBody := controllers.ConfirmRequest{ConfirmationToken: "invalidToken"}
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, confirmEmailPath, "", jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusBadRequest)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
	AssertEqual(t, 1, len(resBody.Fields))
	AssertEqual(t, "confirmation_token", resBody.Fields[0].Param)
	AssertEqual(t, "Field validation for 'ConfirmationToken' failed on the 'jwt' tag", resBody.Fields[0].Error)
	AssertEqual(t, reqBody.ConfirmationToken, resBody.Fields[0].Value)
}

func TestConfirmEmailNotFound(t *testing.T) {
	// Arrange
	fakeUser := CreateFakeTestUser(t)
	reqBody := generateTestConfirmEmailData(t, fakeUser)
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, confirmEmailPath, "", jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusUnauthorized)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
	AssertEqual(t, services.MessageUnauthorized, resBody.Message)
}

func TestUserAlreadyConfirmed(t *testing.T) {
	// Arrange
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	reqBody := generateTestConfirmEmailData(t, testUser)
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, confirmEmailPath, "", jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusUnauthorized)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
	AssertEqual(t, services.MessageUnauthorized, resBody.Message)
}

func TestUserVersionMissmatch(t *testing.T) {
	// Arrange
	testUser := CreateTestUser(t, nil)
	testUser.Version = math.MaxInt16
	reqBody := generateTestConfirmEmailData(t, testUser)
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, confirmEmailPath, "", jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusUnauthorized)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
	AssertEqual(t, services.MessageUnauthorized, resBody.Message)
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
	AssertEqual(t, reqBody.Email, resBody.Fields[0].Value)
	AssertEqual(t, "password", resBody.Fields[1].Param)
	AssertEqual(t, "Field validation for 'Password' failed on the 'required' tag", resBody.Fields[1].Error)
	AssertEqual(t, reqBody.Password, resBody.Fields[1].Value)
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
	AssertEqual(t, reqBody.Email, resBody.Fields[0].Value)
	AssertEqual(t, "code", resBody.Fields[1].Param)
	AssertEqual(t, "Field validation for 'Code' failed on the 'required' tag", resBody.Fields[1].Error)
	AssertEqual(t, reqBody.Code, resBody.Fields[1].Value)
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
const logoutConfirmPath = "/api/auth/logout"

func performCookieLogoutRequest(t *testing.T, app *fiber.App, accessToken, refreshToken string) *http.Response {
	config := GetTestConfig(t)
	req := httptest.NewRequest(MethodPost, logoutConfirmPath, nil)
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

func TestBodyLogout(t *testing.T) {
	// Arrange
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	accessToken, refreshToken := GenerateTestAuthTokens(t, testUser)
	app := GetTestApp(t)
	reqBody := controllers.SignOutRequest{RefreshToken: refreshToken}
	jsonBody := CreateTestJSONRequestBody(t, reqBody)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, logoutConfirmPath, accessToken, jsonBody)
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
	resp := performCookieLogoutRequest(t, app, accessToken, refreshToken)
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
	resp := PerformTestRequest(t, app, 0, MethodPost, logoutConfirmPath, "invalid", jsonBody)
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
	resp := performCookieLogoutRequest(t, app, "invalid", refreshToken)
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
	resp := PerformTestRequest(t, app, 0, MethodPost, logoutConfirmPath, accessToken, jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusBadRequest)
	resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
	AssertEqual(t, 1, len(resBody.Fields))
	AssertEqual(t, "refresh_token", resBody.Fields[0].Param)
	AssertEqual(t, "Field validation for 'RefreshToken' failed on the 'required' tag", resBody.Fields[0].Error)
	AssertEqual(t, reqBody.RefreshToken, resBody.Fields[0].Value)
}

func TestCookieLogoutValidationErr(t *testing.T) {
	// Arrange
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	accessToken, _ := GenerateTestAuthTokens(t, testUser)
	app := GetTestApp(t)

	// Act
	resp := performCookieLogoutRequest(t, app, accessToken, "")
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusBadRequest)
	resBody := AssertTestResponseBody(t, resp, controllers.EmptyRequestValidationError{})
	AssertEqual(t, controllers.RequestValidationMessage, resBody.Message)
}
