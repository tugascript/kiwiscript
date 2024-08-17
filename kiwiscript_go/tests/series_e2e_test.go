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
			PerformTestRequestCase(t, http.MethodPost, tc.Path, tc)
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
			PerformTestRequestCase(t, http.MethodPut, tc.Path, tc)
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
		{
			Name: "Should return 200 OK when the series is published",
			ReqFn: func(t *testing.T) (string, string) {
				testDB := GetTestDatabase(t)
				series, err := testDB.FindSeriesBySlugAndLanguageSlug(context.Background(), db.FindSeriesBySlugAndLanguageSlugParams{
					Slug:         "existing-series",
					LanguageSlug: "rust",
				})
				if err != nil {
					t.Fatal("Failed to find series", err)
				}
				opts := db.UpdateSeriesIsPublishedParams{
					IsPublished: true,
					ID:          series.ID,
				}
				if _, err := testDB.UpdateSeriesIsPublished(context.Background(), opts); err != nil {
					t.Fatal("Failed to publish series", err)
				}

				return "", ""
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.SeriesResponse{})
				AssertNotEmpty(t, resBody.Title)
				AssertNotEmpty(t, resBody.Description)
				AssertEqual(t, resBody.IsPublished, true)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series",
		},
		{
			Name: "Should return 200 OK with progress if the user is not staff",
			ReqFn: func(t *testing.T) (string, string) {
				testDB := GetTestDatabase(t)
				series, err := testDB.FindSeriesBySlugAndLanguageSlug(context.Background(), db.FindSeriesBySlugAndLanguageSlugParams{
					Slug:         "existing-series",
					LanguageSlug: "rust",
				})
				if err != nil {
					t.Fatal("Failed to find series", err)
				}
				opts := db.UpdateSeriesIsPublishedParams{
					IsPublished: true,
					ID:          series.ID,
				}
				if _, err := testDB.UpdateSeriesIsPublished(context.Background(), opts); err != nil {
					t.Fatal("Failed to publish series", err)
				}

				langProg, err := testDB.CreateLanguageProgress(context.Background(), db.CreateLanguageProgressParams{
					LanguageSlug: "rust",
					UserID:       testUser.ID,
				})
				if err != nil {
					t.Fatal("Failed to create language progress", "error", err)
				}

				seriesProg, err := testDB.CreateSeriesProgress(context.Background(), db.CreateSeriesProgressParams{
					LanguageSlug:       "rust",
					SeriesSlug:         "existing-series",
					LanguageProgressID: langProg.ID,
					UserID:             testUser.ID,
				})
				if err != nil {
					t.Fatal("Failed to create series progress", "error", err)
				}

				if _, err := testDB.IncrementSeriesProgressCompletedSections(context.Background(), seriesProg.ID); err != nil {
					t.Fatal("Failed to increment series progress competed sections", "error", err)
				}

				testUser.IsStaff = false
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.SeriesResponse{})
				AssertNotEmpty(t, resBody.Title)
				AssertNotEmpty(t, resBody.Description)
				AssertEqual(t, resBody.IsPublished, true)
				AssertEqual(t, resBody.CompletedLessons, 1)
				AssertEqual(t, resBody.CompletedSections, 1)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series",
		},
		{
			Name: "Should return 404 NOT FOUND when the series is not found",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, resBody.Code, controllers.StatusNotFound)
				AssertEqual(t, resBody.Message, services.MessageNotFound)
			},
			Path: baseLanguagesPath + "/rust/series/non-existing-series",
		},
		{
			Name: "Should return 404 NOT FOUND when the language is not found",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, resBody.Code, controllers.StatusNotFound)
				AssertEqual(t, resBody.Message, services.MessageNotFound)
			},
			Path: baseLanguagesPath + "/python/series/existing-series",
		},
		{
			Name: "Should return 400 BAD REQUEST when the series slug is invalid",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
				AssertEqual(t, resBody.Code, controllers.StatusValidation)
				AssertEqual(t, 1, len(resBody.Fields))
				AssertEqual(t, "seriesSlug", resBody.Fields[0].Param)
				AssertEqual(t, controllers.StrFieldErrMessageSlug, resBody.Fields[0].Message)
			},
			Path: baseLanguagesPath + "/rust/series/invalid--slug",
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

func TestGetPaginatedSeries(t *testing.T) {
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

		count := 0
		titleMap := map[string]bool{}
		for count < 20 {
			fakeTitle := faker.Name()
			if _, ok := titleMap[fakeTitle]; !ok {
				seriesParam := db.CreateSeriesParams{
					Title:        fakeTitle,
					Slug:         utils.Slugify(fakeTitle),
					Description:  "Lorem ipsum lorem ipsum",
					LanguageSlug: "rust",
					AuthorID:     testUser.ID,
				}
				if _, err := testDb.CreateSeries(context.Background(), seriesParam); err != nil {
					t.Fatal("Failed to create series", err)
				}
				count++
			}
		}
	}()

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 200 OK with all series if the user is staff",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.PaginatedResponse[dtos.SeriesResponse]{})
				AssertEqual(t, resBody.Count, 20)
				AssertEqual(t, len(resBody.Results), 20)
				AssertEqual(t, resBody.Links.Next, nil)
				AssertEqual(t, resBody.Links.Prev, nil)
			},
			Path: baseLanguagesPath + "/rust/series",
		},
		{
			Name: "Should return 200 OK with 5 rows with a limit of 5",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.PaginatedResponse[dtos.SeriesResponse]{})
				AssertEqual(t, resBody.Count, 20)
				AssertEqual(t, len(resBody.Results), 5)
				AssertGreaterThan(t, resBody.Results[0].ID, resBody.Results[1].ID)
				AssertEqual(
					t,
					resBody.Links.Next.Href,
					"https://api.kiwiscript.com/api/v1/languages/rust/series?sortBy=date&limit=5&offset=5",
				)
				AssertEqual(t, resBody.Links.Prev, nil)
			},
			Path: baseLanguagesPath + "/rust/series?limit=5",
		},
		{
			Name: "Should return 200 OK with 5 rows with a limit of 5 and offset of 5, and sort by slug",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.PaginatedResponse[dtos.SeriesResponse]{})
				AssertEqual(t, resBody.Count, 20)
				AssertEqual(t, len(resBody.Results), 5)
				AssertGreaterThan(t, resBody.Results[1].Slug, resBody.Results[0].Slug)
				AssertEqual(
					t,
					resBody.Links.Next.Href,
					"https://api.kiwiscript.com/api/v1/languages/rust/series?sortBy=slug&limit=5&offset=10",
				)
				AssertEqual(
					t,
					resBody.Links.Prev.Href,
					"https://api.kiwiscript.com/api/v1/languages/rust/series?sortBy=slug&limit=5&offset=0",
				)
			},
			Path: baseLanguagesPath + "/rust/series?limit=5&offset=5&sortBy=slug",
		},
		{
			Name: "Should return 200 OK with 0 rows if non of the series are published and no user logged in",
			ReqFn: func(t *testing.T) (string, string) {
				return "", ""
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.PaginatedResponse[dtos.SeriesResponse]{})
				AssertEqual(t, resBody.Count, 0)
				AssertEqual(t, len(resBody.Results), 0)
				AssertEqual(t, resBody.Links.Next, nil)
				AssertEqual(t, resBody.Links.Prev, nil)
			},
			Path: baseLanguagesPath + "/rust/series",
		},
		{
			Name: "Should return 200 OK with 4 rows if only 5 series are published, offset is 1 and no user logged in",
			ReqFn: func(t *testing.T) (string, string) {
				testDb := GetTestDatabase(t)
				series, err := testDb.FindPaginatedSeriesWithAuthorSortBySlug(
					context.Background(),
					db.FindPaginatedSeriesWithAuthorSortBySlugParams{
						LanguageSlug: "rust",
						Limit:        5,
						Offset:       10,
					},
				)
				if err != nil {
					t.Fatal("Failed to get series", "error", err)
				}

				for i := 0; i < 5; i++ {
					params := db.UpdateSeriesIsPublishedParams{
						IsPublished: true,
						ID:          series[i].ID,
					}
					if _, err := testDb.UpdateSeriesIsPublished(context.Background(), params); err != nil {
						t.Fatal("Failed to update series is published", "error", err)
					}
				}

				return "", ""
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.PaginatedResponse[dtos.SeriesResponse]{})
				AssertEqual(t, resBody.Count, 5)
				AssertEqual(t, len(resBody.Results), 4)
				AssertEqual(t, resBody.Links.Next, nil)
				AssertEqual(
					t,
					resBody.Links.Prev.Href,
					"https://api.kiwiscript.com/api/v1/languages/rust/series?sortBy=date&limit=25&offset=0",
				)
			},
			Path: baseLanguagesPath + "/rust/series?offset=1",
		},
		{
			Name: "Should return 200 OK with user's progress if the user is not staff",
			ReqFn: func(t *testing.T) (string, string) {
				testDb := GetTestDatabase(t)
				series, err := testDb.FindPaginatedSeriesWithAuthorSortByID(
					context.Background(),
					db.FindPaginatedSeriesWithAuthorSortByIDParams{
						LanguageSlug: "rust",
						Limit:        2,
					},
				)
				if err != nil {
					t.Fatal("Failed to find paginated series", "error", err)
				}

				for i := 0; i < 2; i++ {
					publishedPrms := db.UpdateSeriesIsPublishedParams{
						IsPublished: true,
						ID:          series[i].ID,
					}
					if _, err := testDb.UpdateSeriesIsPublished(context.Background(), publishedPrms); err != nil {
						t.Fatal("Failed to update series is published", "error", err)
					}
				}

				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				testUser.IsStaff = false
				accessToken, _ := GenerateTestAuthTokens(t, testUser)

				langProg, err := testDb.CreateLanguageProgress(context.Background(), db.CreateLanguageProgressParams{
					LanguageSlug: "rust",
					UserID:       testUser.ID,
				})
				if err != nil {
					t.Fatal("Failed to create language progress", "error", err)
				}

				seriesProg, err := testDb.CreateSeriesProgress(context.Background(), db.CreateSeriesProgressParams{
					LanguageSlug:       series[0].LanguageSlug,
					SeriesSlug:         series[0].Slug,
					LanguageProgressID: langProg.ID,
					UserID:             testUser.ID,
				})
				if err != nil {
					t.Fatal("Failed to create series progress", "error", err)
				}

				if _, err := testDb.IncrementSeriesProgressCompletedSections(context.Background(), seriesProg.ID); err != nil {
					t.Fatal("Failed to increment series progress completed sections", "error", err)
				}

				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.PaginatedResponse[dtos.SeriesResponse]{})
				AssertEqual(t, resBody.Results[0].CompletedLessons, 1)
				AssertEqual(t, resBody.Results[0].CompletedSections, 1)
				AssertEqual(t, resBody.Results[1].CompletedSections, 0)
				AssertEqual(t, resBody.Results[1].CompletedSections, 0)
			},
			Path: baseLanguagesPath + "/rust/series",
		},
		{
			Name: "Should return 200 OK with user's progress if the user is not staff sortedBy slug",
			ReqFn: func(t *testing.T) (string, string) {
				testDb := GetTestDatabase(t)
				series, err := testDb.FindPaginatedSeriesWithAuthorSortBySlug(
					context.Background(),
					db.FindPaginatedSeriesWithAuthorSortBySlugParams{
						LanguageSlug: "rust",
						Limit:        2,
					},
				)
				if err != nil {
					t.Fatal("Failed to find paginated series", "error", err)
				}

				for i := 0; i < 2; i++ {
					publishedPrms := db.UpdateSeriesIsPublishedParams{
						IsPublished: true,
						ID:          series[i].ID,
					}
					if _, err := testDb.UpdateSeriesIsPublished(context.Background(), publishedPrms); err != nil {
						t.Fatal("Failed to update series is published", "error", err)
					}
				}

				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				testUser.IsStaff = false
				accessToken, _ := GenerateTestAuthTokens(t, testUser)

				langProg, err := testDb.CreateLanguageProgress(context.Background(), db.CreateLanguageProgressParams{
					LanguageSlug: "rust",
					UserID:       testUser.ID,
				})
				if err != nil {
					t.Fatal("Failed to create language progress", "error", err)
				}

				seriesProg, err := testDb.CreateSeriesProgress(context.Background(), db.CreateSeriesProgressParams{
					LanguageSlug:       series[0].LanguageSlug,
					SeriesSlug:         series[0].Slug,
					LanguageProgressID: langProg.ID,
					UserID:             testUser.ID,
				})
				if err != nil {
					t.Fatal("Failed to create series progress", "error", err)
				}

				if _, err := testDb.IncrementSeriesProgressCompletedSections(context.Background(), seriesProg.ID); err != nil {
					t.Fatal("Failed to increment series progress completed sections", "error", err)
				}

				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.PaginatedResponse[dtos.SeriesResponse]{})
				AssertEqual(t, resBody.Results[0].CompletedLessons, 1)
				AssertEqual(t, resBody.Results[0].CompletedSections, 1)
				AssertEqual(t, resBody.Results[1].CompletedSections, 0)
				AssertEqual(t, resBody.Results[1].CompletedSections, 0)
			},
			Path: baseLanguagesPath + "/rust/series?sortBy=slug",
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

func TestDeleteSeries(t *testing.T) {
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
	}()

	beforeEach := func(t *testing.T) {
		testDb := GetTestDatabase(t)
		serParams := db.CreateSeriesParams{
			Title:        "Existing Series",
			Slug:         "existing-series",
			Description:  "Some description",
			LanguageSlug: "rust",
			AuthorID:     testUser.ID,
		}
		if _, err := testDb.CreateSeries(context.Background(), serParams); err != nil {
			t.Fatal("Failed to create series", "error", err)
		}
	}

	afterEach := func(t *testing.T) {
		testDb := GetTestDatabase(t)
		if err := testDb.DeleteAllLanguageSeries(context.Background(), "rust"); err != nil {
			t.Fatal("Failed to delete all language series", "error", err)
		}
	}

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 204 NO CONTENT if the series is deleted successfully",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNoContent,
			AssertFn: func(t *testing.T, _ string, _ *http.Response) {
				afterEach(t)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series",
		},
		{
			Name: "Should return 403 FORBIDDEN if the user is not staff",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				testUser.IsStaff = false
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusForbidden,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, resBody.Code, controllers.StatusForbidden)
				AssertEqual(t, resBody.Message, controllers.StatusForbidden)
				afterEach(t)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series",
		},
		{
			Name: "Should return 404 NOT FOUND if the series is not found",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, resBody.Code, controllers.StatusNotFound)
				AssertEqual(t, resBody.Message, services.MessageNotFound)
				afterEach(t)
			},
			Path: baseLanguagesPath + "/rust/series/non-existing-series",
		},
		{
			Name: "Should return 404 NOT FOUND if the language is not found",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, resBody.Code, controllers.StatusNotFound)
				AssertEqual(t, resBody.Message, services.MessageNotFound)
				afterEach(t)
			},
			Path: baseLanguagesPath + "/python/series/existing-series",
		},
		{
			Name: "Should return 409 CONFLICT if series has students",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				testDb := GetTestDatabase(t)
				series, err := testDb.FindSeriesBySlugAndLanguageSlug(
					context.Background(),
					db.FindSeriesBySlugAndLanguageSlugParams{
						Slug:         "existing-series",
						LanguageSlug: "rust",
					},
				)
				if err != nil {
					t.Fatal("Failed to find series", "error", err)
				}

				isPubPrms := db.UpdateSeriesIsPublishedParams{
					IsPublished: true,
					ID:          series.ID,
				}
				if _, err := testDb.UpdateSeriesIsPublished(context.Background(), isPubPrms); err != nil {
					t.Fatal("Failed to publish series", "error", err)
				}

				progUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				langProg, err := testDb.CreateLanguageProgress(context.Background(), db.CreateLanguageProgressParams{
					LanguageSlug: "rust",
					UserID:       progUser.ID,
				})
				if err != nil {
					t.Fatal("Failed to create language progress", "error", err)
				}

				sProgPrms := db.CreateSeriesProgressParams{
					LanguageSlug:       "rust",
					SeriesSlug:         "existing-series",
					LanguageProgressID: langProg.ID,
					UserID:             progUser.ID,
				}
				if _, err := testDb.CreateSeriesProgress(context.Background(), sProgPrms); err != nil {
					t.Fatal("Failed to create series progress", "error", err)
				}

				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusConflict,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, resBody.Code, controllers.StatusConflict)
				AssertEqual(t, resBody.Message, "Series has students")
				afterEach(t)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodDelete, tc.Path, tc)
		})
	}

	t.Cleanup(languagesCleanUp(t))
	t.Cleanup(userCleanUp(t))
}

func TestPublishSeries(t *testing.T) {
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
	}()

	beforeEach := func(t *testing.T) {
		testDb := GetTestDatabase(t)
		serParams := db.CreateSeriesParams{
			Title:        "Existing Series",
			Slug:         "existing-series",
			Description:  "Some description",
			LanguageSlug: "rust",
			AuthorID:     testUser.ID,
		}
		if _, err := testDb.CreateSeries(context.Background(), serParams); err != nil {
			t.Fatal("Failed to create series", "error", err)
		}
	}

	afterEach := func(t *testing.T) {
		testDb := GetTestDatabase(t)
		series, err := testDb.FindSeriesBySlugAndLanguageSlug(
			context.Background(),
			db.FindSeriesBySlugAndLanguageSlugParams{
				Slug:         "existing-series",
				LanguageSlug: "rust",
			},
		)
		if err != nil {
			t.Fatal("Failed to get series", "error", err)
		}

		if err := testDb.DeleteSeriesById(context.Background(), series.ID); err != nil {
			t.Fatal("Failed to delete series", "error", err)
		}
	}

	testCases := []TestRequestCase[dtos.UpdateIsPublishedBody]{
		{
			Name: "Should return 200 OK publishing a series if it has sections",
			ReqFn: func(t *testing.T) (dtos.UpdateIsPublishedBody, string) {
				beforeEach(t)
				testDb := GetTestDatabase(t)

				addPrms := db.AddSeriesSectionsCountParams{
					Slug:             "existing-series",
					LessonsCount:     2,
					ReadTimeSeconds:  100,
					WatchTimeSeconds: 1000,
				}
				if err := testDb.AddSeriesSectionsCount(context.Background(), addPrms); err != nil {
					t.Fatal("Failed to add series sections count", "error", err)
				}

				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.UpdateIsPublishedBody{IsPublished: true}, accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, req dtos.UpdateIsPublishedBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.SeriesResponse{})
				AssertNotEmpty(t, resBody.Title)
				AssertNotEmpty(t, resBody.Description)
				AssertEqual(t, resBody.IsPublished, true)
				AssertEqual(t, resBody.TotalLessons, 2)
				AssertEqual(t, resBody.TotalSections, 1)
				afterEach(t)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series/publish",
		},
		{
			Name: "Should return 200 OK unpublishing a series with no students",
			ReqFn: func(t *testing.T) (dtos.UpdateIsPublishedBody, string) {
				beforeEach(t)
				testDb := GetTestDatabase(t)

				addPrms := db.AddSeriesSectionsCountParams{
					Slug:             "existing-series",
					LessonsCount:     2,
					ReadTimeSeconds:  100,
					WatchTimeSeconds: 1000,
				}
				if err := testDb.AddSeriesSectionsCount(context.Background(), addPrms); err != nil {
					t.Fatal("Failed to add series sections count", "error", err)
				}

				series, err := testDb.FindSeriesBySlugAndLanguageSlug(
					context.Background(),
					db.FindSeriesBySlugAndLanguageSlugParams{
						Slug:         "existing-series",
						LanguageSlug: "rust",
					},
				)
				if err != nil {
					t.Fatal("Failed to find series", "error", err)
				}

				isPubPrms := db.UpdateSeriesIsPublishedParams{
					IsPublished: true,
					ID:          series.ID,
				}
				if _, err := testDb.UpdateSeriesIsPublished(context.Background(), isPubPrms); err != nil {
					t.Fatal("Failed to publish series", "error", err)
				}

				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.UpdateIsPublishedBody{IsPublished: false}, accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, req dtos.UpdateIsPublishedBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.SeriesResponse{})
				AssertNotEmpty(t, resBody.Title)
				AssertNotEmpty(t, resBody.Description)
				AssertEqual(t, resBody.IsPublished, false)
				AssertEqual(t, resBody.TotalLessons, 2)
				AssertEqual(t, resBody.TotalSections, 1)
				afterEach(t)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series/publish",
		},
		{
			Name: "Should return 400 BAD REQUEST if the series has no sections",
			ReqFn: func(t *testing.T) (dtos.UpdateIsPublishedBody, string) {
				beforeEach(t)
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.UpdateIsPublishedBody{IsPublished: true}, accessToken
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req dtos.UpdateIsPublishedBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, resBody.Code, controllers.StatusValidation)
				AssertEqual(t, resBody.Message, "Series must have sections to be published")
				afterEach(t)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series/publish",
		},
		{
			Name: "Should return 409 CONFLICT unpublishing a series with students",
			ReqFn: func(t *testing.T) (dtos.UpdateIsPublishedBody, string) {
				beforeEach(t)
				testDb := GetTestDatabase(t)

				addPrms := db.AddSeriesSectionsCountParams{
					Slug:             "existing-series",
					LessonsCount:     2,
					ReadTimeSeconds:  100,
					WatchTimeSeconds: 1000,
				}
				if err := testDb.AddSeriesSectionsCount(context.Background(), addPrms); err != nil {
					t.Fatal("Failed to add series sections count", "error", err)
				}

				series, err := testDb.FindSeriesBySlugAndLanguageSlug(
					context.Background(),
					db.FindSeriesBySlugAndLanguageSlugParams{
						Slug:         "existing-series",
						LanguageSlug: "rust",
					},
				)
				if err != nil {
					t.Fatal("Failed to find series", "error", err)
				}

				isPubPrms := db.UpdateSeriesIsPublishedParams{
					IsPublished: true,
					ID:          series.ID,
				}
				if _, err := testDb.UpdateSeriesIsPublished(context.Background(), isPubPrms); err != nil {
					t.Fatal("Failed to publish series", "error", err)
				}

				langProg, err := testDb.CreateLanguageProgress(context.Background(), db.CreateLanguageProgressParams{
					LanguageSlug: "rust",
					UserID:       testUser.ID,
				})
				if err != nil {
					t.Fatal("Failed to create language progress", "error", err)
				}

				sProgPrms := db.CreateSeriesProgressParams{
					LanguageSlug:       "rust",
					SeriesSlug:         "existing-series",
					LanguageProgressID: langProg.ID,
					UserID:             testUser.ID,
				}
				if _, err := testDb.CreateSeriesProgress(context.Background(), sProgPrms); err != nil {
					t.Fatal("Failed to create series progress", "error", err)
				}

				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.UpdateIsPublishedBody{IsPublished: false}, accessToken
			},
			ExpStatus: fiber.StatusConflict,
			AssertFn: func(t *testing.T, req dtos.UpdateIsPublishedBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, resBody.Code, controllers.StatusConflict)
				AssertEqual(t, resBody.Message, "Series has students")
				afterEach(t)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series/publish",
		},
		{
			Name: "Should return 403 FORBIDDEN if the user is not staff",
			ReqFn: func(t *testing.T) (dtos.UpdateIsPublishedBody, string) {
				beforeEach(t)
				testUser.IsStaff = false
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.UpdateIsPublishedBody{IsPublished: true}, accessToken
			},
			ExpStatus: fiber.StatusForbidden,
			AssertFn: func(t *testing.T, req dtos.UpdateIsPublishedBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, resBody.Code, controllers.StatusForbidden)
				AssertEqual(t, resBody.Message, controllers.StatusForbidden)
				afterEach(t)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series/publish",
		},
		{
			Name: "Should return 404 NOT FOUND if the series is not found",
			ReqFn: func(t *testing.T) (dtos.UpdateIsPublishedBody, string) {
				beforeEach(t)
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.UpdateIsPublishedBody{IsPublished: true}, accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, req dtos.UpdateIsPublishedBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, resBody.Code, controllers.StatusNotFound)
				AssertEqual(t, resBody.Message, services.MessageNotFound)
				afterEach(t)
			},
			Path: baseLanguagesPath + "/rust/series/non-existing-series/publish",
		},
		{
			Name: "Should return 404 NOT FOUND if the language is not found",
			ReqFn: func(t *testing.T) (dtos.UpdateIsPublishedBody, string) {
				beforeEach(t)
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.UpdateIsPublishedBody{IsPublished: true}, accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, req dtos.UpdateIsPublishedBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, resBody.Code, controllers.StatusNotFound)
				AssertEqual(t, resBody.Message, services.MessageNotFound)
				afterEach(t)
			},
			Path: baseLanguagesPath + "/python/series/existing-series/publish",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodPatch, tc.Path, tc)
		})
	}

	t.Cleanup(languagesCleanUp(t))
	t.Cleanup(userCleanUp(t))
}
