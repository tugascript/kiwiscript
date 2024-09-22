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
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kiwiscript/kiwiscript_go/dtos"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/services"
	"net/http"
	"strings"
	"testing"
)

func TestUploadSeriesPicture(t *testing.T) {
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

	beforeEach := func(t *testing.T) {
		testServices := GetTestServices(t)

		rmvOpts := services.DeletePictureOptions{
			RequestID:    uuid.NewString(),
			UserID:       testUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "existing-series",
		}
		if err := testServices.DeleteSeriesPicture(context.Background(), rmvOpts); err != nil {
			t.Log("Failed to remove existing picture", err)
		}
	}

	testCases := []TestRequestCase[FormFileBody]{
		{
			Name: "Should return 201 CREATED and convert PNG to JPEG",
			ReqFn: func(t *testing.T) (FormFileBody, string) {
				beforeEach(t)
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return ImageUploadForm(t, "png"), accessToken
			},
			ExpStatus: fiber.StatusCreated,
			AssertFn: func(t *testing.T, _ FormFileBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.SeriesPictureResponse{})
				AssertNotEmpty(t, resBody.URL)
				AssertStringContains(t, resBody.URL, ".jpeg")
			},
			Path: baseLanguagesPath + "/rust/series/existing-series/picture",
		},
		{
			Name: "Should return 201 CREATED and compress JPEG",
			ReqFn: func(t *testing.T) (FormFileBody, string) {
				beforeEach(t)
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return ImageUploadForm(t, "jpg"), accessToken
			},
			ExpStatus: fiber.StatusCreated,
			AssertFn: func(t *testing.T, _ FormFileBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.SeriesPictureResponse{})
				AssertNotEmpty(t, resBody.URL)
				AssertStringContains(t, resBody.URL, ".jpeg")
			},
			Path: baseLanguagesPath + "/rust/series/existing-series/picture",
		},
		{
			Name: "Should return 400 Bad Request when mime type is not supported",
			ReqFn: func(t *testing.T) (FormFileBody, string) {
				beforeEach(t)
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return ImageUploadForm(t, "webp"), accessToken
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, _ FormFileBody, resp *http.Response) {
				AssertValidationErrorWithoutFieldsResponse(t, resp, "Invalid file type")
			},
			Path: baseLanguagesPath + "/rust/series/existing-series/picture",
		},
		{
			Name: "Should return 409 Conflict when picture already exists",
			ReqFn: func(t *testing.T) (FormFileBody, string) {
				beforeEach(t)
				testDb := GetTestDatabase(t)
				ctx := context.Background()

				series, err := testDb.FindSeriesBySlugAndLanguageSlug(ctx, db.FindSeriesBySlugAndLanguageSlugParams{
					Slug:         "existing-series",
					LanguageSlug: "rust",
				})
				if err != nil {
					t.Fatal("Failed to find series", err)
				}

				serPicParams := db.CreateSeriesPictureParams{
					ID:       uuid.New(),
					SeriesID: series.ID,
					AuthorID: testUser.ID,
					Ext:      "jpeg",
				}
				if _, err := testDb.CreateSeriesPicture(ctx, serPicParams); err != nil {
					t.Fatal("Failed to create series picture", err)
				}

				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return ImageUploadForm(t, "png"), accessToken
			},
			ExpStatus: fiber.StatusConflict,
			AssertFn: func(t *testing.T, _ FormFileBody, resp *http.Response) {
				AssertConflictResponse(t, resp, "Series picture already exists")
			},
			Path: baseLanguagesPath + "/rust/series/existing-series/picture",
		},
		{
			Name: "Should return 403 Forbidden when user is not staff",
			ReqFn: func(t *testing.T) (FormFileBody, string) {
				beforeEach(t)
				testUser.IsStaff = false
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return ImageUploadForm(t, "png"), accessToken
			},
			ExpStatus: fiber.StatusForbidden,
			AssertFn: func(t *testing.T, _ FormFileBody, resp *http.Response) {
				AssertForbiddenResponse(t, resp)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series/picture",
		},
		{
			Name: "Should return 401 Unauthorized when user is not staff",
			ReqFn: func(t *testing.T) (FormFileBody, string) {
				beforeEach(t)
				return ImageUploadForm(t, "png"), ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ FormFileBody, resp *http.Response) {
				AssertUnauthorizedResponse(t, resp)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series/picture",
		},
		{
			Name: "Should return 404 Not Found when the series does not exist",
			ReqFn: func(t *testing.T) (FormFileBody, string) {
				beforeEach(t)
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return ImageUploadForm(t, "png"), accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ FormFileBody, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: baseLanguagesPath + "/rust/series/non-existing-series/picture",
		},
		{
			Name: "Should return 404 Not Found when the series does not exist",
			ReqFn: func(t *testing.T) (FormFileBody, string) {
				beforeEach(t)
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return ImageUploadForm(t, "png"), accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ FormFileBody, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: baseLanguagesPath + "/python/series/existing-series/picture",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCaseWithForm(t, tc)
		})
	}

	t.Cleanup(languagesCleanUp(t))
	t.Cleanup(userCleanUp(t))
}

func TestDeleteSeriesPicture(t *testing.T) {
	t.Cleanup(languagesCleanUp(t))
	t.Cleanup(userCleanUp(t))

	// Set up a test user with staff privileges
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	testUser.IsStaff = true

	// Create a language and series, and add a picture to the series
	func() {
		testDb := GetTestDatabase(t)
		ctx := context.Background()

		// Create language "Rust"
		langParams := db.CreateLanguageParams{
			Name:     "Rust",
			Icon:     strings.TrimSpace(languageIcons["Rust"]),
			AuthorID: testUser.ID,
			Slug:     "rust",
		}
		if _, err := testDb.CreateLanguage(ctx, langParams); err != nil {
			t.Fatal("Failed to create language", err)
		}

		// Create series "Existing Series"
		serParams := db.CreateSeriesParams{
			Title:        "Existing Series",
			Slug:         "existing-series",
			Description:  "Some description",
			LanguageSlug: "rust",
			AuthorID:     testUser.ID,
		}
		series, err := testDb.CreateSeries(ctx, serParams)
		if err != nil {
			t.Fatal("Failed to create series", err)
		}

		// Add a picture to the series
		serPicParams := db.CreateSeriesPictureParams{
			ID:       uuid.New(),
			SeriesID: series.ID,
			AuthorID: testUser.ID,
			Ext:      "jpeg",
		}
		if _, err := testDb.CreateSeriesPicture(ctx, serPicParams); err != nil {
			t.Fatal("Failed to create series picture", err)
		}
	}()

	beforeEach := func(t *testing.T) {
		testDb := GetTestDatabase(t)
		ctx := context.Background()

		// Ensure the series picture exists before each test
		series, err := testDb.FindSeriesBySlugAndLanguageSlug(ctx, db.FindSeriesBySlugAndLanguageSlugParams{
			Slug:         "existing-series",
			LanguageSlug: "rust",
		})
		if err != nil {
			t.Fatal("Failed to find series", err)
		}

		_, err = testDb.FindSeriesPictureBySeriesID(ctx, series.ID)
		if err != nil {
			// Create series picture if it doesn't exist
			serPicParams := db.CreateSeriesPictureParams{
				ID:       uuid.New(),
				SeriesID: series.ID,
				AuthorID: testUser.ID,
				Ext:      "jpeg",
			}
			if _, err := testDb.CreateSeriesPicture(ctx, serPicParams); err != nil {
				t.Fatal("Failed to create series picture", err)
			}
		}
	}

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 204 NO CONTENT if the series picture is deleted successfully",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNoContent,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				// Verify that the picture is actually deleted
				testDb := GetTestDatabase(t)
				ctx := context.Background()
				series, err := testDb.FindSeriesBySlugAndLanguageSlug(ctx, db.FindSeriesBySlugAndLanguageSlugParams{
					Slug:         "existing-series",
					LanguageSlug: "rust",
				})
				if err != nil {
					t.Fatal("Failed to find series", err)
				}
				_, err = testDb.FindSeriesPictureBySeriesID(ctx, series.ID)
				if err == nil {
					t.Fatal("Expected series picture to be deleted, but it still exists")
				}
			},
			Path: baseLanguagesPath + "/rust/series/existing-series/picture",
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
				AssertForbiddenResponse(t, resp)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series/picture",
		},
		{
			Name: "Should return 401 Unauthorized when user is not authenticated",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				return "", "" // No access token
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				AssertUnauthorizedResponse(t, resp)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series/picture",
		},
		{
			Name: "Should return 404 NOT FOUND if the series picture does not exist",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				// Delete the picture to simulate it not existing
				testDb := GetTestDatabase(t)
				ctx := context.Background()
				series, err := testDb.FindSeriesBySlugAndLanguageSlug(ctx, db.FindSeriesBySlugAndLanguageSlugParams{
					Slug:         "existing-series",
					LanguageSlug: "rust",
				})
				if err != nil {
					t.Fatal("Failed to find series", err)
				}

				seriesPicture, err := testDb.FindSeriesPictureBySeriesID(ctx, series.ID)
				if err != nil {
					t.Fatal("Failed to find series picture", err)
				}

				if err := testDb.DeleteSeriesPicture(ctx, seriesPicture.ID); err != nil {
					t.Fatal("Failed to delete series picture", err)
				}
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series/picture",
		},
		{
			Name: "Should return 404 NOT FOUND if the series does not exist",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: baseLanguagesPath + "/rust/series/non-existing-series/picture",
		},
		{
			Name: "Should return 404 NOT FOUND if the language does not exist",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: baseLanguagesPath + "/python/series/existing-series/picture",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodDelete, tc.Path, tc)
		})
	}
}
