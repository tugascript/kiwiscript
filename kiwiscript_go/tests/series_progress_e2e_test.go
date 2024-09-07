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

func TestCreateOrUpdateSeriesProgress(t *testing.T) {
	languagesCleanUp(t)()
	staffUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)

	func() {
		testDb := GetTestDatabase(t)
		ctx := context.Background()

		prms := db.CreateLanguageParams{
			Name:     "Rust",
			Icon:     strings.TrimSpace(languageIcons["Rust"]),
			AuthorID: staffUser.ID,
			Slug:     "rust",
		}
		if _, err := testDb.CreateLanguage(ctx, prms); err != nil {
			t.Fatal("Failed to create language", err)
		}

		series, err := testDb.CreateSeries(ctx, db.CreateSeriesParams{
			LanguageSlug: "rust",
			Title:        "Rust Series",
			Slug:         "rust-series",
			AuthorID:     staffUser.ID,
			Description:  "Some cool rust series",
		})
		if err != nil {
			t.Fatal("Failed to create series", err)
		}

		section, err := testDb.CreateSection(ctx, db.CreateSectionParams{
			Title:        "Rust Section",
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
			Description:  "Some section",
			AuthorID:     staffUser.ID,
		})
		if err != nil {
			t.Fatal("Failed to create section", err)
		}

		lesson, err := testDb.CreateLesson(ctx, db.CreateLessonParams{
			Title:        "Cool rust lesson",
			AuthorID:     staffUser.ID,
			SectionID:    section.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
		})
		if err != nil {
			t.Fatal("Failed to create lesson", err)
		}

		isPubLesPrms := db.UpdateLessonIsPublishedParams{
			IsPublished: true,
			ID:          lesson.ID,
		}
		if _, err := testDb.UpdateLessonIsPublished(ctx, isPubLesPrms); err != nil {
			t.Fatal("Failed to update lesson is published", err)
		}

		isPubSecPrms := db.UpdateSectionIsPublishedParams{
			IsPublished: true,
			ID:          section.ID,
		}
		if _, err := testDb.UpdateSectionIsPublished(ctx, isPubSecPrms); err != nil {
			t.Fatal("Failed to update section is published", err)
		}

		isPubSerPrms := db.UpdateSeriesIsPublishedParams{
			IsPublished: true,
			ID:          series.ID,
		}
		if _, err := testDb.UpdateSeriesIsPublished(ctx, isPubSerPrms); err != nil {
			t.Fatal("Failed to update series is published", err)
		}
	}()

	beforeEach := func(t *testing.T) {
		testServices := GetTestServices(t)
		ctx := context.Background()

		opts := services.CreateOrUpdateLanguageProgressOptions{
			RequestID:    uuid.NewString(),
			UserID:       testUser.ID,
			LanguageSlug: "rust",
		}
		if _, _, _, serviceErr := testServices.CreateOrUpdateLanguageProgress(ctx, opts); serviceErr != nil {
			t.Fatal("Failed to create language progress", "serviceErr", serviceErr)
		}
	}

	afterEach := func(t *testing.T) {
		testServices := GetTestServices(t)
		ctx := context.Background()

		opts := services.DeleteLanguageProgressOptions{
			RequestID:    uuid.NewString(),
			UserID:       testUser.ID,
			LanguageSlug: "rust",
		}
		if serviceErr := testServices.DeleteLanguageProgress(ctx, opts); serviceErr != nil {
			t.Fatal("Failed to delete language progress", "serviceErr", serviceErr)
		}
	}

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 201 CREATED when a series progress is created",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusCreated,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.SeriesResponse{})
				AssertEqual(t, resBody.Title, "Rust Series")
				AssertNotEmpty(t, resBody.ViewedAt)
				afterEach(t)
			},
			Path: baseLanguagesPath + "/rust/series/rust-series/progress",
		},
		{
			Name: "Should return 200 OK when a series progress is updated",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				testServices := GetTestServices(t)
				ctx := context.Background()

				opts := services.CreateOrUpdateSeriesProgressOptions{
					RequestID:    uuid.NewString(),
					UserID:       testUser.ID,
					LanguageSlug: "rust",
					SeriesSlug:   "rust-series",
				}
				if _, _, _, serviceErr := testServices.CreateOrUpdateSeriesProgress(ctx, opts); serviceErr != nil {
					t.Fatal("Failed to create series progress", serviceErr)
				}

				accessToken, _ := GenerateTestAuthTokens(t, testUser)

				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.SeriesResponse{})
				AssertEqual(t, resBody.Title, "Rust Series")
				AssertNotEmpty(t, resBody.ViewedAt)
				afterEach(t)
			},
			Path: baseLanguagesPath + "/rust/series/rust-series/progress",
		},
		{
			Name: "Should return 401 UNAUTHORIZED when the user is not authenticated",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				return "", ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertUnauthorizedResponse(t, resp)
				afterEach(t)
			},
			Path: baseLanguagesPath + "/rust/series/rust-series/progress",
		},
		{
			Name: "Should return 403 FORBIDDEN when the user is a staff user",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				staffUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, staffUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusForbidden,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertForbiddenResponse(t, resp)
				afterEach(t)
			},
			Path: baseLanguagesPath + "/rust/series/rust-series/progress",
		},
		{
			Name: "Should return 404 NOT FOUND when the series does not exist",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
				afterEach(t)
			},
			Path: baseLanguagesPath + "/rust/series/python-series/progress",
		},
		{
			Name: "Should return 404 NOT FOUND when the language does not exist",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
				afterEach(t)
			},
			Path: baseLanguagesPath + "/python/series/rust-series/progress",
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

func TestDeleteSeriesProgress(t *testing.T) {
	languagesCleanUp(t)()
	staffUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)

	func() {
		testDb := GetTestDatabase(t)
		ctx := context.Background()

		prms := db.CreateLanguageParams{
			Name:     "Rust",
			Icon:     strings.TrimSpace(languageIcons["Rust"]),
			AuthorID: staffUser.ID,
			Slug:     "rust",
		}
		if _, err := testDb.CreateLanguage(ctx, prms); err != nil {
			t.Fatal("Failed to create language", err)
		}

		series, err := testDb.CreateSeries(ctx, db.CreateSeriesParams{
			LanguageSlug: "rust",
			Title:        "Rust Series",
			Slug:         "rust-series",
			AuthorID:     staffUser.ID,
			Description:  "Some cool rust series",
		})
		if err != nil {
			t.Fatal("Failed to create series", err)
		}

		section, err := testDb.CreateSection(ctx, db.CreateSectionParams{
			Title:        "Rust Section",
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
			Description:  "Some section",
			AuthorID:     staffUser.ID,
		})
		if err != nil {
			t.Fatal("Failed to create section", err)
		}

		lesson, err := testDb.CreateLesson(ctx, db.CreateLessonParams{
			Title:        "Cool rust lesson",
			AuthorID:     staffUser.ID,
			SectionID:    section.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
		})
		if err != nil {
			t.Fatal("Failed to create lesson", err)
		}

		isPubLesPrms := db.UpdateLessonIsPublishedParams{
			IsPublished: true,
			ID:          lesson.ID,
		}
		if _, err := testDb.UpdateLessonIsPublished(ctx, isPubLesPrms); err != nil {
			t.Fatal("Failed to update lesson is published", err)
		}

		isPubSecPrms := db.UpdateSectionIsPublishedParams{
			IsPublished: true,
			ID:          section.ID,
		}
		if _, err := testDb.UpdateSectionIsPublished(ctx, isPubSecPrms); err != nil {
			t.Fatal("Failed to update section is published", err)
		}

		isPubSerPrms := db.UpdateSeriesIsPublishedParams{
			IsPublished: true,
			ID:          series.ID,
		}
		if _, err := testDb.UpdateSeriesIsPublished(ctx, isPubSerPrms); err != nil {
			t.Fatal("Failed to update series is published", err)
		}
	}()

	beforeEach := func(t *testing.T) {
		testServices := GetTestServices(t)
		ctx := context.Background()
		requestID := uuid.NewString()

		langProgOpts := services.CreateOrUpdateLanguageProgressOptions{
			RequestID:    requestID,
			UserID:       testUser.ID,
			LanguageSlug: "rust",
		}
		if _, _, _, serviceErr := testServices.CreateOrUpdateLanguageProgress(ctx, langProgOpts); serviceErr != nil {
			t.Fatal("Failed to create language progress", "serviceErr", serviceErr)
		}

		serProgOpts := services.CreateOrUpdateSeriesProgressOptions{
			RequestID:    requestID,
			UserID:       testUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
		}
		if _, _, _, serviceErr := testServices.CreateOrUpdateSeriesProgress(ctx, serProgOpts); serviceErr != nil {
			t.Fatal("Failed to create series progress", "serviceErr", serviceErr)
		}
	}

	afterEach := func(t *testing.T) {
		testServices := GetTestServices(t)
		ctx := context.Background()

		opts := services.DeleteLanguageProgressOptions{
			RequestID:    uuid.NewString(),
			UserID:       testUser.ID,
			LanguageSlug: "rust",
		}
		if serviceErr := testServices.DeleteLanguageProgress(ctx, opts); serviceErr != nil {
			t.Fatal("Failed to delete language progress", "serviceErr", serviceErr)
		}
	}

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 204 NO CONTENT when a series progress is deleted",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNoContent,
			AssertFn: func(t *testing.T, _ string, _ *http.Response) {
				afterEach(t)
			},
			Path: baseLanguagesPath + "/rust/series/rust-series/progress",
		},
		{
			Name: "Should return 401 UNAUTHORIZED when the user is not authenticated",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				return "", ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertUnauthorizedResponse(t, resp)
				afterEach(t)
			},
			Path: baseLanguagesPath + "/rust/series/rust-series/progress",
		},
		{
			Name: "Should return 403 FORBIDDEN when the user is staff",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				staffUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, staffUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusForbidden,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertForbiddenResponse(t, resp)
				afterEach(t)
			},
			Path: baseLanguagesPath + "/rust/series/rust-series/progress",
		},
		{
			Name: "Should return 404 NOT FOUND when the series does not exist",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
				afterEach(t)
			},
			Path: baseLanguagesPath + "/rust/series/python-series/progress",
		},
		{
			Name: "Should return 404 NOT FOUND when the series does not exist",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
				afterEach(t)
			},
			Path: baseLanguagesPath + "/python/series/rust-series/progress",
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
