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
	"bytes"
	"context"
	"encoding/json"
	"github.com/kiwiscript/kiwiscript_go/controllers"
	"github.com/kiwiscript/kiwiscript_go/providers/oauth"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-faker/faker/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/storage/redis/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kiwiscript/kiwiscript_go/app"
	cc "github.com/kiwiscript/kiwiscript_go/providers/cache"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/providers/email"
	stg "github.com/kiwiscript/kiwiscript_go/providers/object_storage"
	"github.com/kiwiscript/kiwiscript_go/providers/tokens"
	"github.com/kiwiscript/kiwiscript_go/services"
	"github.com/kiwiscript/kiwiscript_go/utils"
)

var testConfig *app.Config
var testServices *services.Services
var testApp *fiber.App
var testTokens *tokens.Tokens
var testDatabase *db.Database
var testCache *cc.Cache

func initTestServicesAndApp(t *testing.T) {
	log := app.DefaultLogger()
	testConfig = app.NewConfig(log, "../.env")
	log = app.GetLogger(testConfig.Logger.Env, testConfig.Logger.Debug)

	// Build storages/models
	log.Info("Building redis connection...")
	storage := redis.New(redis.Config{
		URL: testConfig.RedisURL,
	})
	log.Info("Finished building redis connection")

	// Build database connection
	log.Info("Building database connection...")
	ctx := context.Background()
	dbConnPool, err := pgxpool.New(ctx, testConfig.PostgresURL)
	if err != nil {
		log.ErrorContext(ctx, "Failed to connect to database", "error", err)
		t.Fatal("Failed to connect to database", err)
	}
	log.Info("Finished building database connection")

	// Build s3 client
	log.Info("Building s3 client...")
	s3Cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(testConfig.ObjectStorage.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(testConfig.ObjectStorage.AccessKey, testConfig.ObjectStorage.SecretKey, "")),
	)
	if err != nil {
		log.ErrorContext(ctx, "Failed to load s3 config", "error", err)
		t.Fatal("Failed to load s3 config", err)
	}
	s3Client := s3.NewFromConfig(s3Cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String("http://" + testConfig.ObjectStorage.Host)
	})
	log.Info("Finished building s3 client")

	testTokens = tokens.NewTokens(
		tokens.NewTokenSecretData(testConfig.Tokens.Access.PublicKey, testConfig.Tokens.Access.PrivateKey, testConfig.Tokens.Access.TtlSec),
		tokens.NewTokenSecretData(testConfig.Tokens.Refresh.PublicKey, testConfig.Tokens.Refresh.PrivateKey, testConfig.Tokens.Refresh.TtlSec),
		tokens.NewTokenSecretData(testConfig.Tokens.Email.PublicKey, testConfig.Tokens.Email.PrivateKey, testConfig.Tokens.Email.TtlSec),
		tokens.NewTokenSecretData(testConfig.Tokens.OAuth.PublicKey, testConfig.Tokens.OAuth.PrivateKey, testConfig.Tokens.OAuth.TtlSec),
		"https://"+testConfig.BackendDomain,
	)
	mailer := email.NewMail(
		testConfig.Email.Username,
		testConfig.Email.Password,
		testConfig.Email.Port,
		testConfig.Email.Host,
		testConfig.Email.Name,
		testConfig.FrontendDomain,
	)
	testDatabase = db.NewDatabase(dbConnPool)
	testCache = cc.NewCache(storage)
	testObjectStorage := stg.NewObjectStorage(log, s3Client, testConfig.ObjectStorage.Bucket)
	testOAuthProvider := oauth.NewOAuthProviders(
		log,
		testConfig.OAuthProviders.GitHub.ClientID,
		testConfig.OAuthProviders.GitHub.ClientSecret,
		testConfig.OAuthProviders.Google.ClientID,
		testConfig.OAuthProviders.Google.ClientSecret,
		testConfig.BackendDomain,
	)
	testServices = services.NewServices(
		log,
		testDatabase,
		testCache,
		testObjectStorage,
		mailer,
		testTokens,
		testOAuthProvider,
	)
	testApp = app.CreateApp(
		log,
		storage,
		dbConnPool,
		s3Client,
		&testConfig.Email,
		&testConfig.Tokens,
		&testConfig.Limiter,
		&testConfig.OAuthProviders,
		testConfig.ObjectStorage.Bucket,
		testConfig.BackendDomain,
		testConfig.FrontendDomain,
		testConfig.RefreshCookieName,
		testConfig.CookieSecret,
	)
}

func GetTestConfig(t *testing.T) *app.Config {
	if testConfig == nil {
		initTestServicesAndApp(t)
	}

	return testConfig
}

func GetTestServices(t *testing.T) *services.Services {
	if testServices == nil {
		initTestServicesAndApp(t)
	}

	return testServices
}

func GetTestApp(t *testing.T) *fiber.App {
	if testApp == nil {
		initTestServicesAndApp(t)
	}

	return testApp
}

func GetTestTokens(t *testing.T) *tokens.Tokens {
	if testTokens == nil {
		initTestServicesAndApp(t)
	}

	return testTokens
}

func GetTestDatabase(t *testing.T) *db.Database {
	if testDatabase == nil {
		initTestServicesAndApp(t)
	}

	return testDatabase
}

func GetTestCache(t *testing.T) *cc.Cache {
	if testCache == nil {
		initTestServicesAndApp(t)
	}

	return testCache
}

func CreateTestJSONRequestBody(t *testing.T, reqBody interface{}) *bytes.Reader {
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatal("Failed to marshal JSON", err)
	}

	return bytes.NewReader(jsonBody)
}

func PerformTestRequest(t *testing.T, app *fiber.App, delayMs int, method, path, accessToken string, body io.Reader) *http.Response {
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal("Failed to perform request", err)
	}

	if delayMs > 0 {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}

	return resp
}

func AssertTestStatusCode(t *testing.T, resp *http.Response, expectedStatusCode int) {
	if resp.StatusCode != expectedStatusCode {
		t.Logf("Status Code: %d", resp.StatusCode)
		t.Fatal("Failed to assert status code")
	}
}

func AssertTestResponseBody[V interface{}](t *testing.T, resp *http.Response, expectedBody V) V {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Failed to read response body", err)
	}

	if err := json.Unmarshal(body, &expectedBody); err != nil {
		t.Logf("Body: %s", body)
		t.Fatal("Failed to register user")
	}
	return expectedBody
}

func AssertEqual[V comparable](t *testing.T, actual, expected V) {
	if expected != actual {
		t.Fatalf("Actual: %v, Expected: %v", actual, expected)
	}
}

type ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~string
}

func AssertGreaterThan[V ordered](t *testing.T, actual, expected V) {
	if expected > actual {
		t.Fatalf("Actual: %v, Expected: %v", actual, expected)
	}
}

func AssertNotEmpty[V comparable](t *testing.T, actual V) {
	var empty V
	if actual == empty {
		t.Fatal("Value is empty")
	}
}

func assertRequestErrorResponse(t *testing.T, resp *http.Response, code, message string) {
	resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
	AssertEqual(t, resBody.Code, code)
	AssertEqual(t, resBody.Message, message)
}

func AssertForbiddenResponse(t *testing.T, resp *http.Response) {
	assertRequestErrorResponse(t, resp, controllers.StatusForbidden, controllers.StatusForbidden)
}

func AssertUnauthorizedResponse(t *testing.T, resp *http.Response) {
	assertRequestErrorResponse(t, resp, controllers.StatusUnauthorized, controllers.StatusUnauthorized)
}

func AssertNotFoundResponse(t *testing.T, resp *http.Response) {
	assertRequestErrorResponse(t, resp, controllers.StatusNotFound, services.MessageNotFound)
}

func AssertConflictDuplicateKeyResponse(t *testing.T, resp *http.Response) {
	assertRequestErrorResponse(t, resp, controllers.StatusConflict, services.MessageDuplicateKey)
}

type ValidationErrorAssertion struct {
	Param   string
	Message string
}

func AssertValidationErrorResponse(t *testing.T, resp *http.Response, assertions []ValidationErrorAssertion) {
	resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
	AssertEqual(t, resBody.Code, controllers.StatusValidation)
	AssertEqual(t, resBody.Message, controllers.RequestValidationMessage)

	for i, a := range assertions {
		AssertEqual(t, resBody.Fields[i].Param, a.Param)
		AssertEqual(t, resBody.Fields[i].Message, a.Message)
	}
}

type fakeUserData struct {
	Email     string `faker:"email"`
	FirstName string `faker:"first_name"`
	LastName  string `faker:"last_name"`
	Location  string `faker:"oneof: NZL, AUS, NAM, EUR, OTH"`
	Password  string `faker:"oneof: Pas@w0rd123, P@sW0rd456, P@ssw0rd789, P@ssW0rd012, P@ssw0rd!345"`
}

func GenerateFakeUserData(t *testing.T) services.CreateUserOptions {
	fakeData := fakeUserData{}
	if err := faker.FakeData(&fakeData); err != nil {
		t.Fatal("Failed to generate fake data", err)
	}

	return services.CreateUserOptions{
		FirstName: utils.Capitalized(fakeData.FirstName),
		LastName:  utils.Capitalized(fakeData.LastName),
		Location:  utils.Uppercased(fakeData.Location),
		Email:     utils.Lowered(fakeData.Email),
		Password:  fakeData.Password,
		Provider:  utils.ProviderEmail,
	}
}

func CreateTestUser(t *testing.T, userData *services.CreateUserOptions) *db.User {
	var opts services.CreateUserOptions
	if userData == nil {
		opts = GenerateFakeUserData(t)
	} else {
		opts = *userData
	}

	ser := GetTestServices(t)

	passwordHash, err := utils.HashPassword(opts.Password)
	if err != nil {
		t.Fatal("Failed to hash password", err)
	}

	opts.Password = passwordHash
	user, serErr := ser.CreateUser(context.Background(), opts)
	if serErr != nil {
		t.Fatal("Failed to create test user", serErr)
	}

	return user
}

func CreateFakeTestUser(t *testing.T) *db.User {
	opts := GenerateFakeUserData(t)
	return &db.User{
		ID:        math.MaxInt32,
		FirstName: opts.FirstName,
		LastName:  opts.LastName,
		Location:  opts.Location,
		Email:     opts.Email,
		Version:   1,
		Password:  pgtype.Text{String: opts.Password, Valid: true},
	}
}

func GenerateTestAuthTokens(t *testing.T, user *db.User) (accessToken string, refreshToken string) {
	tks := GetTestTokens(t)
	accessToken, err := tks.CreateAccessToken(user)

	if err != nil {
		t.Fatal("Failed to create access token", err)
	}

	refreshToken, err = tks.CreateRefreshToken(user)
	if err != nil {
		t.Fatal("Failed to create refresh token", err)
	}

	return accessToken, refreshToken
}

type TestRequestCase[R any] struct {
	Name      string
	ReqFn     func(t *testing.T) (R, string)
	ExpStatus int
	AssertFn  func(t *testing.T, req R, resp *http.Response)
	DelayMs   int
	Path      string
	Method    string
}

func PerformTestRequestCase[R any](t *testing.T, method, path string, tc TestRequestCase[R]) {
	// Arrange
	reqBody, accessToken := tc.ReqFn(t)
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	fiberApp := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, fiberApp, tc.DelayMs, method, path, accessToken, jsonBody)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	// Assert
	AssertTestStatusCode(t, resp, tc.ExpStatus)
	tc.AssertFn(t, reqBody, resp)
}
