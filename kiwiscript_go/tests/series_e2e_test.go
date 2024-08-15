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
	"github.com/go-faker/faker/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/kiwiscript/kiwiscript_go/controllers"
	"github.com/kiwiscript/kiwiscript_go/dtos"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/services"
	"github.com/kiwiscript/kiwiscript_go/utils"
	"net/http"
	"strings"
	"testing"
)

func TestCreateSeries(t *testing.T) {
	languagesCleanUp(t)()
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	func() {
		testDb := GetTestDatabase(t)

		params := db.CreateLanguageParams{
			Name:     "Rust",
			Icon:     strings.TrimSpace(languageIcons["Rust"]),
			AuthorID: testUser.ID,
			Slug:     "rust",
		}
		if _, err := testDb.CreateLanguage(context.Background(), params); err != nil {
			t.Fatal("Failed to create language", err)
		}
	}()

	generateFakeSeriesData := func(t *testing.T) dtos.CreateSeriesBody {
		title := faker.Name()
		description := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Aenean consequat nisl vel rutrum congue. Quisque mattis id massa id tincidunt. Nulla fringilla enim id dignissim dignissim. Morbi consequat, dui vel auctor pharetra, tortor tortor tristique nibh, et malesuada ante nisl egestas lectus. Donec vitae mollis enim, non aliquam tortor. Nulla."
		return dtos.CreateSeriesBody{
			Title:       title,
			Description: description,
		}
	}

	testCases := []TestRequestCase[dtos.CreateSeriesBody]{
		{
			Name: "Should return 201 CREATED when a new series is created",
			ReqFn: func(t *testing.T) (dtos.CreateSeriesBody, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return generateFakeSeriesData(t), accessToken
			},
			ExpStatus: fiber.StatusCreated,
			AssertFn: func(t *testing.T, req dtos.CreateSeriesBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.SeriesResponse{})
				AssertEqual(t, req.Title, resBody.Title)
				AssertEqual(t, req.Description, resBody.Description)
				AssertEqual(
					t,
					"https://api.kiwiscript.com/api/v1/languages/rust",
					resBody.Links.Language.Href,
				)
				AssertEqual(t, utils.Slugify(req.Title), resBody.Slug)
				AssertEqual(
					t,
					"https://api.kiwiscript.com/api/v1/languages/rust/series/"+resBody.Slug,
					resBody.Links.Self.Href,
				)
			},
			Path: baseLanguagesPath + "/rust/series",
		},
		{
			Name: "Should return 400 BAD REQUEST when a new series name",
			ReqFn: func(t *testing.T) (dtos.CreateSeriesBody, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				fakeData := generateFakeSeriesData(t)
				fakeData.Title = strings.Repeat(fakeData.Title, 100)
				return fakeData, accessToken
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req dtos.CreateSeriesBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
				AssertEqual(t, 1, len(resBody.Fields))
				AssertEqual(t, "title", resBody.Fields[0].Param)
				AssertEqual(t, controllers.StrFieldErrMessageMax, resBody.Fields[0].Message)
			},
			Path: baseLanguagesPath + "/rust/series",
		},
		{
			Name: "Should return 409 FORBIDDEN when the user is not staff",
			ReqFn: func(t *testing.T) (dtos.CreateSeriesBody, string) {
				testUser.IsStaff = false
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return generateFakeSeriesData(t), accessToken
			},
			ExpStatus: fiber.StatusForbidden,
			AssertFn: func(t *testing.T, req dtos.CreateSeriesBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, resBody.Code, controllers.StatusForbidden)
				AssertEqual(t, resBody.Message, controllers.StatusForbidden)
			},
			Path: baseLanguagesPath + "/rust/series",
		},
		{
			Name: "Should return 404 NOT FOUND when the language doesn't exist",
			ReqFn: func(t *testing.T) (dtos.CreateSeriesBody, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return generateFakeSeriesData(t), accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, req dtos.CreateSeriesBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, resBody.Code, controllers.StatusNotFound)
				AssertEqual(t, resBody.Message, services.MessageNotFound)
			},
			Path: baseLanguagesPath + "/not-a-language/series",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, MethodPost, tc.Path, tc)
		})
	}

	t.Cleanup(languagesCleanUp(t))
	t.Cleanup(userCleanUp(t))
}

func TestUpdateSeries(t *testing.T) {
	languagesCleanUp(t)()
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	beforeEach := func(t *testing.T) {
		testDb := GetTestDatabase(t)

		langParams := db.CreateLanguageParams{
			Name:     "Rust",
			Icon:     strings.TrimSpace(languageIcons["Rust"]),
			AuthorID: testUser.ID,
			Slug:     "rust",
		}
		if _, err := testDb.CreateLanguage(context.Background(), langParams); err != nil {
			t.Fatal("Failed to create language", err)
		}

		serParams := db.CreateSeriesParams{
			Title:        "Existing Series",
			Slug:         "existing-series",
			Description:  "Some description",
			LanguageSlug: "rust",
			AuthorID:     testUser.ID,
		}
		if _, err := testDb.CreateSeries(context.Background(), serParams); err != nil {
			t.Fatal("Failed to create series", err)
		}
	}

	generateFakeSeriesData := func(t *testing.T) dtos.UpdateSeriesBody {
		title := faker.Name()
		description := "Some other description compared"
		return dtos.UpdateSeriesBody{
			Title:       title,
			Description: description,
		}
	}

	testCases := []TestRequestCase[dtos.UpdateSeriesBody]{
		{
			Name: "Should return 200 OK when a series is updated",
			ReqFn: func(t *testing.T) (dtos.UpdateSeriesBody, string) {
				beforeEach(t)
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return generateFakeSeriesData(t), accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, req dtos.UpdateSeriesBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.SeriesResponse{})
				AssertEqual(t, req.Title, resBody.Title)
				AssertEqual(t, req.Description, resBody.Description)
				AssertEqual(
					t,
					"https://api.kiwiscript.com/api/v1/languages/rust",
					resBody.Links.Language.Href,
				)
				AssertEqual(t, utils.Slugify(req.Title), resBody.Slug)
				AssertEqual(
					t,
					"https://api.kiwiscript.com/api/v1/languages/rust/series/"+resBody.Slug,
					resBody.Links.Self.Href,
				)
				t.Cleanup(languagesCleanUp(t))
			},
			Path: baseLanguagesPath + "/rust/series/existing-series",
		},
		{
			Name: "Should return 400 BAD REQUEST when a series name is invalid",
			ReqFn: func(t *testing.T) (dtos.UpdateSeriesBody, string) {
				beforeEach(t)
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				fakeData := generateFakeSeriesData(t)
				fakeData.Title = strings.Repeat(fakeData.Title, 100)
				return fakeData, accessToken
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req dtos.UpdateSeriesBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
				AssertEqual(t, 1, len(resBody.Fields))
				AssertEqual(t, "title", resBody.Fields[0].Param)
				AssertEqual(t, controllers.StrFieldErrMessageMax, resBody.Fields[0].Message)
				t.Cleanup(languagesCleanUp(t))
			},
			Path: baseLanguagesPath + "/rust/series/existing-series",
		},
		{
			Name: "Should return 403 FORBIDDEN when the user is not staff",
			ReqFn: func(t *testing.T) (dtos.UpdateSeriesBody, string) {
				beforeEach(t)
				testUser.IsStaff = false
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return generateFakeSeriesData(t), accessToken
			},
			ExpStatus: fiber.StatusForbidden,
			AssertFn: func(t *testing.T, req dtos.UpdateSeriesBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, resBody.Code, controllers.StatusForbidden)
				AssertEqual(t, resBody.Message, controllers.StatusForbidden)
				t.Cleanup(languagesCleanUp(t))
			},
			Path: baseLanguagesPath + "/rust/series/existing-series",
		},
		{
			Name: "Should return 403 FORBIDDEN when the user is not the owner of the series",
			ReqFn: func(t *testing.T) (dtos.UpdateSeriesBody, string) {
				beforeEach(t)
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return generateFakeSeriesData(t), accessToken
			},
			ExpStatus: fiber.StatusForbidden,
			AssertFn: func(t *testing.T, req dtos.UpdateSeriesBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, resBody.Code, controllers.StatusForbidden)
				AssertEqual(t, resBody.Message, controllers.StatusForbidden)
				t.Cleanup(languagesCleanUp(t))
			},
			Path: baseLanguagesPath + "/rust/series/existing-series",
		},
		{
			Name: "Should return 404 NOT FOUND when the series doesn't exist",
			ReqFn: func(t *testing.T) (dtos.UpdateSeriesBody, string) {
				beforeEach(t)
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return generateFakeSeriesData(t), accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, req dtos.UpdateSeriesBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, resBody.Code, controllers.StatusNotFound)
				AssertEqual(t, resBody.Message, services.MessageNotFound)
				t.Cleanup(languagesCleanUp(t))
			},
			Path: baseLanguagesPath + "/rust/series/non-existing-series",
		},
		{
			Name: "Should return 404 NOT FOUND when the language doesn't exist",
			ReqFn: func(t *testing.T) (dtos.UpdateSeriesBody, string) {
				beforeEach(t)
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return generateFakeSeriesData(t), accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, req dtos.UpdateSeriesBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, resBody.Code, controllers.StatusNotFound)
				AssertEqual(t, resBody.Message, services.MessageNotFound)
				t.Cleanup(languagesCleanUp(t))
			},
			Path: baseLanguagesPath + "/non-existing/series/non-existing-series",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, MethodPut, tc.Path, tc)
		})
	}

	t.Cleanup(languagesCleanUp(t))
	t.Cleanup(userCleanUp(t))
}

func TestGetSeries(t *testing.T) {
	languagesCleanUp(t)()
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	func() {
		testDb := GetTestDatabase(t)

		langParams := db.CreateLanguageParams{
			Name:     "Rust",
			Icon:     strings.TrimSpace(languageIcons["Rust"]),
			AuthorID: testUser.ID,
			Slug:     "rust",
		}
		if _, err := testDb.CreateLanguage(context.Background(), langParams); err != nil {
			t.Fatal("Failed to create language", err)
		}

		serParams := db.CreateSeriesParams{
			Title:        "Existing Series",
			Slug:         "existing-series",
			Description:  "Some description",
			LanguageSlug: "rust",
			AuthorID:     testUser.ID,
		}
		if _, err := testDb.CreateSeries(context.Background(), serParams); err != nil {
			t.Fatal("Failed to create series", err)
		}
	}()

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 200 OK when a series is found and not published when the user is staff",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.SeriesResponse{})
				AssertNotEmpty(t, resBody.Title)
				AssertNotEmpty(t, resBody.Description)
				AssertEqual(t, resBody.IsPublished, false)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series",
		},
		{
			Name: "Should return 404 NOT FOUND when the series is not published and the user is not staff",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = false
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, resBody.Code, controllers.StatusNotFound)
				AssertEqual(t, resBody.Message, services.MessageNotFound)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series",
		},
		{
			Name: "Should return 404 NOT FOUND when the series is not published and the user is not staff",
			ReqFn: func(t *testing.T) (string, string) {
				return "", ""
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, resBody.Code, controllers.StatusNotFound)
				AssertEqual(t, resBody.Message, services.MessageNotFound)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodGet, tc.Path, tc)
		})
	}

	t.Cleanup(languagesCleanUp(t))
	t.Cleanup(userCleanUp(t))
}
