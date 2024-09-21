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
	"github.com/kiwiscript/kiwiscript_go/exceptions"
	"github.com/kiwiscript/kiwiscript_go/providers/oauth"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

var _testConfig *app.Config
var _testServices *services.Services
var _testApp *fiber.App
var _testTokens *tokens.Tokens
var _testDatabase *db.Database
var _testCache *cc.Cache

func initTestServicesAndApp(t *testing.T) {
	log := app.DefaultLogger()
	_testConfig = app.NewConfig(log, "../.env")
	ctx := context.Background()
	log = app.GetLogger(_testConfig.Logger.Env, _testConfig.Logger.Debug)

	// Build storages/models
	log.Info("Building redis connection...")
	storage := redis.New(redis.Config{
		URL: _testConfig.RedisURL,
	})
	log.Info("Finished building redis connection")

	// Build database connection
	log.Info("Building database connection...")
	testPostgresURL := os.Getenv("DATABASE_TEST_URL")
	if testPostgresURL == "" {
		t.Fatal("DATABASE_TEST_URL is not set")
	}
	dbConnPool, err := pgxpool.New(ctx, testPostgresURL)
	if err != nil {
		log.ErrorContext(ctx, "Failed to connect to database", "error", err)
		t.Fatal("Failed to connect to database", err)
	}
	log.Info("Finished building database connection")

	// Build s3 client
	log.Info("Building s3 client...")
	s3Cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(_testConfig.ObjectStorage.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(_testConfig.ObjectStorage.AccessKey, _testConfig.ObjectStorage.SecretKey, "")),
	)
	if err != nil {
		log.ErrorContext(ctx, "Failed to load s3 config", "error", err)
		t.Fatal("Failed to load s3 config", err)
	}

	s3Client := s3.NewFromConfig(s3Cfg, func(o *s3.Options) {
		o.UsePathStyle = true
		o.BaseEndpoint = aws.String("http://" + _testConfig.ObjectStorage.Host)
	})
	log.Info("Finished building s3 client")

	_testTokens = tokens.NewTokens(
		tokens.NewTokenSecretData(_testConfig.Tokens.Access.PublicKey, _testConfig.Tokens.Access.PrivateKey, _testConfig.Tokens.Access.TtlSec),
		tokens.NewTokenSecretData(_testConfig.Tokens.Refresh.PublicKey, _testConfig.Tokens.Refresh.PrivateKey, _testConfig.Tokens.Refresh.TtlSec),
		tokens.NewTokenSecretData(_testConfig.Tokens.Email.PublicKey, _testConfig.Tokens.Email.PrivateKey, _testConfig.Tokens.Email.TtlSec),
		tokens.NewTokenSecretData(_testConfig.Tokens.OAuth.PublicKey, _testConfig.Tokens.OAuth.PrivateKey, _testConfig.Tokens.OAuth.TtlSec),
		"https://"+_testConfig.BackendDomain,
	)
	mailer := email.NewMail(
		log,
		_testConfig.Email.Username,
		_testConfig.Email.Password,
		_testConfig.Email.Port,
		_testConfig.Email.Host,
		_testConfig.Email.Name,
		_testConfig.FrontendDomain,
	)
	_testDatabase = db.NewDatabase(dbConnPool)
	_testCache = cc.NewCache(log, storage)
	testObjectStorage := stg.NewObjectStorage(log, s3Client, _testConfig.ObjectStorage.Bucket)
	testOAuthProvider := oauth.NewOAuthProviders(
		log,
		_testConfig.OAuthProviders.GitHub.ClientID,
		_testConfig.OAuthProviders.GitHub.ClientSecret,
		_testConfig.OAuthProviders.Google.ClientID,
		_testConfig.OAuthProviders.Google.ClientSecret,
		_testConfig.BackendDomain,
	)
	_testServices = services.NewServices(
		log,
		_testDatabase,
		_testCache,
		testObjectStorage,
		mailer,
		_testTokens,
		testOAuthProvider,
	)
	_testApp = app.CreateApp(
		log,
		storage,
		dbConnPool,
		s3Client,
		&_testConfig.Email,
		&_testConfig.Tokens,
		&_testConfig.Limiter,
		&_testConfig.OAuthProviders,
		_testConfig.ObjectStorage.Bucket,
		_testConfig.BackendDomain,
		_testConfig.FrontendDomain,
		_testConfig.RefreshCookieName,
		_testConfig.CookieSecret,
	)
}

func GetTestConfig(t *testing.T) *app.Config {
	if _testConfig == nil {
		initTestServicesAndApp(t)
	}

	return _testConfig
}

func GetTestServices(t *testing.T) *services.Services {
	if _testServices == nil {
		initTestServicesAndApp(t)
	}

	return _testServices
}

func GetTestApp(t *testing.T) *fiber.App {
	if _testApp == nil {
		initTestServicesAndApp(t)
	}

	return _testApp
}

func GetTestTokens(t *testing.T) *tokens.Tokens {
	if _testTokens == nil {
		initTestServicesAndApp(t)
	}

	return _testTokens
}

func GetTestDatabase(t *testing.T) *db.Database {
	if _testDatabase == nil {
		initTestServicesAndApp(t)
	}

	return _testDatabase
}

func GetTestCache(t *testing.T) *cc.Cache {
	if _testCache == nil {
		initTestServicesAndApp(t)
	}

	return _testCache
}

func CreateTestJSONRequestBody(t *testing.T, reqBody interface{}) *bytes.Reader {
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatal("Failed to marshal JSON", err)
	}

	return bytes.NewReader(jsonBody)
}

func PerformTestRequest(t *testing.T, app *fiber.App, delayMs int, method, path, accessToken, contentType string, body io.Reader) *http.Response {
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "application/json")

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	resp, err := app.Test(req, 2000)
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
	if expected >= actual {
		t.Fatalf("Actual: %v, Expected: %v", actual, expected)
	}
}

func AssertNotEmpty[V comparable](t *testing.T, actual V) {
	var empty V
	if actual == empty {
		t.Fatal("Value is empty")
	}
}

func AssertStringContains(t *testing.T, actual string, expected string) {
	if !strings.Contains(actual, expected) {
		t.Fatalf("Actual: %s, Expected: %s", actual, expected)
	}
}

func assertRequestErrorResponse(t *testing.T, resp *http.Response, code, message string) {
	resBody := AssertTestResponseBody(t, resp, exceptions.RequestError{})
	AssertEqual(t, resBody.Code, code)
	AssertEqual(t, resBody.Message, message)
}

func AssertForbiddenResponse(t *testing.T, resp *http.Response) {
	assertRequestErrorResponse(t, resp, exceptions.StatusForbidden, exceptions.StatusForbidden)
}

func AssertUnauthorizedResponse(t *testing.T, resp *http.Response) {
	assertRequestErrorResponse(t, resp, exceptions.StatusUnauthorized, exceptions.StatusUnauthorized)
}

func AssertNotFoundResponse(t *testing.T, resp *http.Response) {
	assertRequestErrorResponse(t, resp, exceptions.StatusNotFound, exceptions.MessageNotFound)
}

func AssertConflictResponse(t *testing.T, resp *http.Response, message string) {
	assertRequestErrorResponse(t, resp, exceptions.StatusConflict, message)
}

func AssertConflictDuplicateKeyResponse(t *testing.T, resp *http.Response) {
	AssertConflictResponse(t, resp, exceptions.MessageDuplicateKey)
}

type ValidationErrorAssertion struct {
	Param   string
	Message string
}

func AssertValidationErrorResponse(t *testing.T, resp *http.Response, assertions []ValidationErrorAssertion) {
	resBody := AssertTestResponseBody(t, resp, exceptions.RequestValidationError{})
	AssertEqual(t, resBody.Code, exceptions.StatusValidation)
	AssertEqual(t, resBody.Message, exceptions.RequestValidationMessage)

	for i, a := range assertions {
		AssertEqual(t, resBody.Fields[i].Param, a.Param)
		AssertEqual(t, resBody.Fields[i].Message, a.Message)
	}
}

func AssertValidationErrorWithoutFieldsResponse(t *testing.T, resp *http.Response, message string) {
	assertRequestErrorResponse(t, resp, exceptions.StatusValidation, message)
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
	PathFn    func() string
	Method    string
}

func PerformTestRequestCase[R any](t *testing.T, method, path string, tc TestRequestCase[R]) {
	// Arrange
	reqBody, accessToken := tc.ReqFn(t)
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	fiberApp := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, fiberApp, tc.DelayMs, method, path, accessToken, "application/json", jsonBody)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	// Assert
	AssertTestStatusCode(t, resp, tc.ExpStatus)
	tc.AssertFn(t, reqBody, resp)
}

func PerformTestRequestCaseWithForm(t *testing.T, tc TestRequestCase[FormFileBody]) {
	// Arrange
	reqBody, accessToken := tc.ReqFn(t)
	fiberApp := GetTestApp(t)

	resp := PerformTestRequest(t, fiberApp, tc.DelayMs, http.MethodPost, tc.Path, accessToken, reqBody.ContentType, bytes.NewReader(reqBody.Body.Bytes()))
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	// Assert
	AssertTestStatusCode(t, resp, tc.ExpStatus)
	tc.AssertFn(t, reqBody, resp)
}

func PerformTestRequestCaseWithPathFn[R any](t *testing.T, method string, tc TestRequestCase[R]) {
	// Arrange
	reqBody, accessToken := tc.ReqFn(t)
	jsonBody := CreateTestJSONRequestBody(t, reqBody)
	fiberApp := GetTestApp(t)

	// Act
	resp := PerformTestRequest(t, fiberApp, tc.DelayMs, method, tc.PathFn(), accessToken, "application/json", jsonBody)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	// Assert
	AssertTestStatusCode(t, resp, tc.ExpStatus)
	tc.AssertFn(t, reqBody, resp)
}

func FileUploadMock(t *testing.T) *multipart.FileHeader {
	// Create a buffer to hold the file and form data
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// Add file to the form data
	part, err := writer.CreateFormFile("file", "lorem-ipsum-test.pdf")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	// Open a file to simulate file upload
	file, err := os.Open("./fixtures/lorem-ipsum-test.pdf")
	if err != nil {
		t.Fatal("Failed to open file", "error", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Fatal("Failed to close file", "error", err)
		}
	}()

	// Copy the file content to the multipart writer
	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatalf("Failed to copy file content: %v", err)
	}

	// Close the writer to finalize the multipart form data
	err = writer.Close()
	if err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}

	// Now parse the multipart form from the buffer
	reader := multipart.NewReader(body, writer.Boundary())
	form, err := reader.ReadForm(10 << 20) // Limit to 10 MB
	if err != nil {
		t.Fatalf("Failed to parse multipart form: %v", err)
	}

	// Extract the file header
	fileHeader := form.File["file"][0]

	return fileHeader
}

func ImageUploadMock(t *testing.T) *multipart.FileHeader {
	// Create a buffer to hold the file and form data
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// Add file to the form data
	part, err := writer.CreateFormFile("file", "image.jpg")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	// Open a file to simulate file upload
	file, err := os.Open("./fixtures/image.jpg")
	if err != nil {
		t.Fatal("Failed to open file", "error", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Fatal("Failed to close file", "error", err)
		}
	}()

	// Copy the file content to the multipart writer
	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatalf("Failed to copy file content: %v", err)
	}

	// Close the writer to finalize the multipart form data
	err = writer.Close()
	if err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}

	// Now parse the multipart form from the buffer
	reader := multipart.NewReader(body, writer.Boundary())
	form, err := reader.ReadForm(10 << 20) // Limit to 10 MB
	if err != nil {
		t.Fatalf("Failed to parse multipart form: %v", err)
	}

	// Extract the file header
	fileHeader := form.File["file"][0]

	return fileHeader
}

type FormFileBody struct {
	Body        *bytes.Buffer
	ContentType string
}

func ImageUploadForm(t *testing.T, ext string) FormFileBody {
	// Create a buffer to hold the file and form data
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// Add file to the form data
	part, err := writer.CreateFormFile("file", "image."+ext)
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	// Open a file to simulate file upload
	file, err := os.Open("./fixtures/image." + ext)
	if err != nil {
		t.Fatal("Failed to open file", "error", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Fatal("Failed to close file", "error", err)
		}
	}()

	// Copy the file content to the multipart writer
	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatalf("Failed to copy file content: %v", err)
	}

	contentType := writer.FormDataContentType()
	// Close the writer to finalize the multipart form data
	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}

	return FormFileBody{
		Body:        body,
		ContentType: contentType,
	}
}

func FileUploadForm(t *testing.T, name, ext string) FormFileBody {
	// Create a buffer to hold the file and form data
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// Add file to the form data
	part, err := writer.CreateFormFile("file", "lorem-ipsum-test."+ext)
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	// Open a file to simulate file upload
	file, err := os.Open("./fixtures/lorem-ipsum-test." + ext)
	if err != nil {
		t.Fatal("Failed to open file", "error", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Fatal("Failed to close file", "error", err)
		}
	}()

	// Copy the file content to the multipart writer
	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatalf("Failed to copy file content: %v", err)
	}

	fileName, err := writer.CreateFormField("name")
	if err != nil {
		t.Fatalf("Failed to create form field: %v", err)
	}

	_, err = fileName.Write([]byte(name))
	if err != nil {
		t.Fatalf("Failed to write form field: %v", err)
	}

	contentType := writer.FormDataContentType()
	// Close the writer to finalize the multipart form data
	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}

	return FormFileBody{
		Body:        body,
		ContentType: contentType,
	}
}
