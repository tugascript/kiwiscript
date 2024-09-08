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
	"github.com/kiwiscript/kiwiscript_go/exceptions"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/services"
	"net/http"
	"strings"
	"testing"
)

func TestCreateOrUpdateLanguageProgress(t *testing.T) {
	languagesCleanUp(t)()
	staffUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)

	beforeEach := func(t *testing.T) {
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
	}

	afterEach := func(t *testing.T) {
		languagesCleanUp(t)()
	}

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 201 CREATED when a lesson progress is created",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusCreated,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.LanguageResponse{})
				AssertEqual(t, resBody.Name, "Rust")
				AssertNotEmpty(t, resBody.ViewedAt)
				afterEach(t)
			},
			Path: baseLanguagesPath + "/rust/progress",
		},
		{
			Name: "Should return 200 OK when a lesson progress is updated",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)

				testServices := GetTestServices(t)
				ctx := context.Background()
				opts := services.CreateOrUpdateLanguageProgressOptions{
					RequestID:    uuid.NewString(),
					UserID:       testUser.ID,
					LanguageSlug: "rust",
				}
				if _, _, _, serviceErr := testServices.CreateOrUpdateLanguageProgress(ctx, opts); serviceErr != nil {
					t.Fatal("Failed to create language progress", serviceErr)
				}

				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.LanguageResponse{})
				AssertEqual(t, resBody.Name, "Rust")
				AssertNotEmpty(t, resBody.ViewedAt)
				afterEach(t)
			},
			Path: baseLanguagesPath + "/rust/progress",
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
			Path: baseLanguagesPath + "/rust/progress",
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
			Path: baseLanguagesPath + "/rust/progress",
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
			Path: baseLanguagesPath + "/python/progress",
		},
		{
			Name: "Should return 400 BAD REQUEST when the language slug is invalid",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertValidationErrorResponse(t, resp, []ValidationErrorAssertion{
					{Param: "languageSlug", Message: exceptions.StrFieldErrMessageSlug},
				})
				afterEach(t)
			},
			Path: baseLanguagesPath + "/c..p..p/progress",
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

func TestDeleteLessonProgress(t *testing.T) {
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
			t.Fatal("Failed to create language progress", serviceErr)
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
			t.Fatal("Failed to delete language progress", serviceErr)
		}
	}

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 204 NO CONTENT when a lesson progress is deleted",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNoContent,
			AssertFn: func(t *testing.T, _ string, _ *http.Response) {
				return
			},
			Path: baseLanguagesPath + "/rust/progress",
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
			Path: baseLanguagesPath + "/rust/progress",
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
			Path: baseLanguagesPath + "/rust/progress",
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
			Path: baseLanguagesPath + "/python/progress",
		},
		{
			Name: "Should return 400 BAD REQUEST when the language slug is invalid",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertValidationErrorResponse(t, resp, []ValidationErrorAssertion{
					{Param: "languageSlug", Message: exceptions.StrFieldErrMessageSlug},
				})
				afterEach(t)
			},
			Path: baseLanguagesPath + "/c..p..p/progress",
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
