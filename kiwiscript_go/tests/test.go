package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/storage/redis/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kiwiscript/kiwiscript_go/app"
	"github.com/kiwiscript/kiwiscript_go/controllers"
	cc "github.com/kiwiscript/kiwiscript_go/providers/cache"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/providers/email"
	"github.com/kiwiscript/kiwiscript_go/providers/tokens"
	"github.com/kiwiscript/kiwiscript_go/services"
)

var testServices *services.Services
var testApp *fiber.App
var testTokens *tokens.Tokens

const (
	MethodPost  string = "POST"
	MethodGet   string = "GET"
	MethodPut   string = "PUT"
	MethodPatch string = "PATCH"
	MethodDel   string = "DELETE"
)

func initTestServicesAndApp(t *testing.T) {
	log := app.DefaultLogger()
	config := app.NewConfig(log, "../.env")
	log = app.GetLogger(config.Logger.Env, config.Logger.Debug)

	// Build storages/models
	log.Info("Building redis connection...")
	storage := redis.New(redis.Config{
		URL: config.RedisURL,
	})
	log.Info("Finished building redis connection")

	// Build database connection
	log.Info("Building database connection...")
	ctx := context.Background()
	dbConnPool, err := pgxpool.New(ctx, config.PostgresURL)
	if err != nil {
		log.ErrorContext(ctx, "Failed to connect to database", "error", err)
		t.Fatal("Failed to connect to database", err)
	}
	log.Info("Finished building database connection")

	testTokens = tokens.NewTokens(
		tokens.NewTokenSecretData(config.Tokens.Access.PublicKey, config.Tokens.Access.PrivateKey, config.Tokens.Access.TtlSec),
		tokens.NewTokenSecretData(config.Tokens.Refresh.PublicKey, config.Tokens.Refresh.PrivateKey, config.Tokens.Refresh.TtlSec),
		tokens.NewTokenSecretData(config.Tokens.Email.PublicKey, config.Tokens.Email.PrivateKey, config.Tokens.Email.TtlSec),
		"https://"+config.BackendDomain,
	)
	mailer := email.NewMail(
		config.Email.Username,
		config.Email.Password,
		config.Email.Port,
		config.Email.Host,
		config.Email.Name,
		config.FrontendDomain,
	)

	testServices = services.NewServices(
		db.NewDatabase(dbConnPool),
		cc.NewCache(storage),
		mailer,
		testTokens,
		log,
	)
	testApp = app.CreateApp(
		log,
		storage,
		dbConnPool,
		&config.Email,
		&config.Tokens,
		config.BackendDomain,
		config.FrontendDomain,
		config.RefreshCookieName,
	)
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

func CreateTestJSONRequestBody(t *testing.T, reqBody interface{}) *bytes.Reader {
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatal("Failed to marshal JSON", err)
	}

	return bytes.NewReader(jsonBody)
}

func PerformTestRequest(t *testing.T, app *fiber.App, delayMs int, method, path string, body io.Reader) *http.Response {
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
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
		t.Fatal("Failed to register user")
	}
}

func AssertTestResponseBody(t *testing.T, resp *http.Response, expectedBody interface{}) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Failed to read response body", err)
	}

	if err := json.Unmarshal(body, &expectedBody); err != nil {
		t.Logf("Body: %s", body)
		t.Fatal("Failed to register user")
	}
}

type fakeUserData struct {
	Email     string `faker:"email"`
	FirstName string `faker:"first_name"`
	LastName  string `faker:"last_name"`
	Location  string `faker:"oneof: NZL, AUS, NAM, EUR, OTH"`
	BirthDate string `faker:"date"`
	Password  string `faker:"oneof: Pas@w0rd123, P@sW0rd456, P@ssw0rd789, P@ssW0rd012, P@ssw0rd!345"`
}

func GenerateFakeUserData(t *testing.T) services.CreateUserOptions {
	fakeData := fakeUserData{}
	if err := faker.FakeData(&fakeData); err != nil {
		t.Fatal("Failed to generate fake data", err)
	}

	birthDate, err := time.Parse(time.DateOnly, fakeData.BirthDate)
	if err != nil {
		t.Fatal("Failed to parse birth date", err)
	}

	return services.CreateUserOptions{
		FirstName: controllers.Capitalized(fakeData.FirstName),
		LastName:  controllers.Capitalized(fakeData.LastName),
		Location:  controllers.Uppercased(fakeData.Location),
		Email:     controllers.Lowered(fakeData.Email),
		BirthDate: birthDate,
		Password:  fakeData.Password,
		Provider:  services.ProviderEmail,
	}
}

func CreateTestUser(t *testing.T) db.User {
	opts := GenerateFakeUserData(t)
	ser := GetTestServices(t)

	user, err := ser.CreateUser(context.Background(), opts)
	if err != nil {
		t.Fatal("Failed to create test user", err)
	}

	return user
}

func CreateFakeTestUser(t *testing.T) db.User {
	opts := GenerateFakeUserData(t)
	return db.User{
		ID:        math.MaxInt32,
		FirstName: opts.FirstName,
		LastName:  opts.LastName,
		Location:  opts.Location,
		Email:     opts.Email,
		BirthDate: pgtype.Date{Time: opts.BirthDate, Valid: true},
		Version:   1,
		Password:  pgtype.Text{String: opts.Password, Valid: true},
	}
}
