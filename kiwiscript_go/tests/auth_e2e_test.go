package tests

import (
	"context"
	"math"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/kiwiscript/kiwiscript_go/controllers"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/providers/tokens"
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
	resp := PerformTestRequest(t, app, delayMs, method, registerPath, jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusCreated)
	AssertTestResponseBody(t, resp, controllers.MessageResponse{})
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
	resp := PerformTestRequest(t, app, delayMs, method, registerPath, jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusBadRequest)
	AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
}

func TestRegisterDuplicateEmail(t *testing.T) {
	// Arrange
	testUser := CreateTestUser(t)
	reqBody := generateFakeRegisterData(t)
	reqBody.Email = testUser.Email
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)
	delayMs := 0
	method := MethodPost

	// Act
	resp := PerformTestRequest(t, app, delayMs, method, registerPath, jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusConflict)
	AssertTestResponseBody(t, resp, controllers.RequestError{})
}

// ----- Test Confirm Email -----
const confirmEmailPath = "/api/auth/confirm-email"

func generateTestConfirmEmailData(t *testing.T, user db.User) controllers.ConfirmRequest {
	emailToken := generateEmailToken(t, tokens.EmailTokenConfirmation, user)

	return controllers.ConfirmRequest{
		ConfirmationToken: emailToken,
	}
}

func confirmTestUser(t *testing.T, userID int32) db.User {
	serv := GetTestServices(t)
	user, err := serv.ConfirmUser(context.Background(), userID)

	if err != nil {
		t.Fatal(err)
	}

	return user
}

// Happy Path
func TestConfirmEmail(t *testing.T) {
	// Arrange
	testUser := CreateTestUser(t)
	reqBody := generateTestConfirmEmailData(t, testUser)
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, confirmEmailPath, jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusOK)
	AssertTestResponseBody(t, resp, controllers.AuthResponse{})
}

// Sad Path
func TestConfirmEmailValidationErr(t *testing.T) {
	// Arrange
	reqBody := controllers.ConfirmRequest{ConfirmationToken: "invalidToken"}
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, confirmEmailPath, jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusBadRequest)
	AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
}

func TestConfirmEmailNotFound(t *testing.T) {
	// Arrange
	fakeUser := CreateFakeTestUser(t)
	reqBody := generateTestConfirmEmailData(t, fakeUser)
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, confirmEmailPath, jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusUnauthorized)
	AssertTestResponseBody(t, resp, controllers.RequestError{})
}

func TestUserAlreadyConfirmed(t *testing.T) {
	// Arrange
	testUser := confirmTestUser(t, CreateTestUser(t).ID)
	reqBody := generateTestConfirmEmailData(t, testUser)
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, confirmEmailPath, jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusUnauthorized)
	AssertTestResponseBody(t, resp, controllers.RequestError{})
}

func TestUserVersionMissmatch(t *testing.T) {
	// Arrange
	testUser := CreateTestUser(t)
	testUser.Version = math.MaxInt16
	reqBody := generateTestConfirmEmailData(t, testUser)
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	app := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, app, 0, MethodPost, confirmEmailPath, jsonBody)
	defer resp.Body.Close()

	// Assert
	AssertTestStatusCode(t, resp, fiber.StatusUnauthorized)
	AssertTestResponseBody(t, resp, controllers.RequestError{})
}
