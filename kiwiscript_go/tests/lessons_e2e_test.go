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
	"fmt"
	"github.com/go-faker/faker/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/kiwiscript/kiwiscript_go/controllers"
	"github.com/kiwiscript/kiwiscript_go/dtos"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/services"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
)

func TestCreateLesson(t *testing.T) {
	languagesCleanUp(t)()
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)

	var sectionID int32
	func() {
		testDb := GetTestDatabase(t)
		testServices := GetTestServices(t)

		params := db.CreateLanguageParams{
			Name:     "Rust",
			Icon:     strings.TrimSpace(languageIcons["Rust"]),
			AuthorID: testUser.ID,
			Slug:     "rust",
		}
		if _, err := testDb.CreateLanguage(context.Background(), params); err != nil {
			t.Fatal("Failed to create language", err)
		}

		series, err := testDb.CreateSeries(context.Background(), db.CreateSeriesParams{
			Title:        "Existing Series",
			Slug:         "existing-series",
			Description:  "Some description",
			LanguageSlug: "rust",
			AuthorID:     testUser.ID,
		})
		if err != nil {
			t.Fatal("Failed to create series", "error", err)
		}

		isPubPrms := db.UpdateSeriesIsPublishedParams{
			IsPublished: true,
			ID:          series.ID,
		}
		if _, err := testDb.UpdateSeriesIsPublished(context.Background(), isPubPrms); err != nil {
			t.Fatal("Failed to update series is published", "error", err)
		}

		section, serviceErr := testServices.CreateSection(context.Background(), services.CreateSectionOptions{
			UserID:       testUser.ID,
			Title:        "Some Section",
			LanguageSlug: "rust",
			SeriesSlug:   "existing-series",
			Description:  "Some description",
		})
		if serviceErr != nil {
			t.Fatal("Failed to create section", "serviceError", serviceErr)
		}
		sectionID = section.ID
	}()

	generateFakeLessonsData := func(t *testing.T) dtos.CreateLessonBody {
		return dtos.CreateLessonBody{Title: faker.Name()}
	}

	testCases := []TestRequestCase[dtos.CreateLessonBody]{
		{
			Name: "Should return 201 CREATED when a lesson is created",
			ReqFn: func(t *testing.T) (dtos.CreateLessonBody, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return generateFakeLessonsData(t), accessToken
			},
			ExpStatus: fiber.StatusCreated,
			AssertFn: func(t *testing.T, req dtos.CreateLessonBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.LessonResponse{})
				AssertEqual(t, req.Title, resBody.Title)
			},
			Path: fmt.Sprintf("%s/rust/series/existing-series/sections/%d/lessons", baseLanguagesPath, sectionID),
		},
		{
			Name: "Should return 400 BAD REQUEST when the title is too long",
			ReqFn: func(t *testing.T) (dtos.CreateLessonBody, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.CreateLessonBody{Title: strings.Repeat("too-long", 100)}, accessToken
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, _ dtos.CreateLessonBody, resp *http.Response) {
				AssertValidationErrorResponse(t, resp, []ValidationErrorAssertion{{
					Param:   "title",
					Message: controllers.StrFieldErrMessageMax,
				}})
			},
			Path: fmt.Sprintf("%s/rust/series/existing-series/sections/%d/lessons", baseLanguagesPath, sectionID),
		},
		{
			Name: "Should return 404 NOT FOUND when the section is not found",
			ReqFn: func(t *testing.T) (dtos.CreateLessonBody, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return generateFakeLessonsData(t), accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ dtos.CreateLessonBody, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series/sections/987654321/lessons",
		},
		{
			Name: "Should return 404 NOT FOUND when the series is not found",
			ReqFn: func(t *testing.T) (dtos.CreateLessonBody, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return generateFakeLessonsData(t), accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ dtos.CreateLessonBody, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf("%s/rust/series/non-existing-series/sections/%d/lessons", baseLanguagesPath, sectionID),
		},
		{
			Name: "Should return 404 NOT FOUND when the language is not found",
			ReqFn: func(t *testing.T) (dtos.CreateLessonBody, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return generateFakeLessonsData(t), accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ dtos.CreateLessonBody, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf("%s/python/series/existing-series/sections/%d/lessons", baseLanguagesPath, sectionID),
		},
		{
			Name: "Should return 403 FORBIDDEN when the user is not staff",
			ReqFn: func(t *testing.T) (dtos.CreateLessonBody, string) {
				testUser.IsStaff = false
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return generateFakeLessonsData(t), accessToken
			},
			ExpStatus: fiber.StatusForbidden,
			AssertFn: func(t *testing.T, _ dtos.CreateLessonBody, resp *http.Response) {
				AssertForbiddenResponse(t, resp)
			},
			Path: fmt.Sprintf("%s/rust/series/existing-series/sections/%d/lessons", baseLanguagesPath, sectionID),
		},
		{
			Name: "Should return 401 UNAUTHORIZED when the user is not authenticated",
			ReqFn: func(t *testing.T) (dtos.CreateLessonBody, string) {
				return generateFakeLessonsData(t), ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ dtos.CreateLessonBody, resp *http.Response) {
				AssertUnauthorizedResponse(t, resp)
			},
			Path: fmt.Sprintf("%s/rust/series/existing-series/sections/%d/lessons", baseLanguagesPath, sectionID),
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

func TestGetLesson(t *testing.T) {
	languagesCleanUp(t)()
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)

	var sectionID, lessonID int32
	func() {
		testDb := GetTestDatabase(t)
		testServices := GetTestServices(t)

		params := db.CreateLanguageParams{
			Name:     "Rust",
			Icon:     strings.TrimSpace(languageIcons["Rust"]),
			AuthorID: testUser.ID,
			Slug:     "rust",
		}
		if _, err := testDb.CreateLanguage(context.Background(), params); err != nil {
			t.Fatal("Failed to create language", err)
		}

		series, err := testDb.CreateSeries(context.Background(), db.CreateSeriesParams{
			Title:        "Existing Series",
			Slug:         "existing-series",
			Description:  "Some description",
			LanguageSlug: "rust",
			AuthorID:     testUser.ID,
		})
		if err != nil {
			t.Fatal("Failed to create series", "error", err)
		}

		isPubPrms := db.UpdateSeriesIsPublishedParams{
			IsPublished: true,
			ID:          series.ID,
		}
		if _, err := testDb.UpdateSeriesIsPublished(context.Background(), isPubPrms); err != nil {
			t.Fatal("Failed to update series is published", "error", err)
		}

		section, serviceErr := testServices.CreateSection(context.Background(), services.CreateSectionOptions{
			UserID:       testUser.ID,
			Title:        "Some Section",
			LanguageSlug: "rust",
			SeriesSlug:   "existing-series",
			Description:  "Some description",
		})
		if serviceErr != nil {
			t.Fatal("Failed to create section", "serviceError", serviceErr)
		}
		sectionID = section.ID

		isPubSecPrms := db.UpdateSectionIsPublishedParams{
			IsPublished: true,
			ID:          sectionID,
		}
		if _, err := testDb.UpdateSectionIsPublished(context.Background(), isPubSecPrms); err != nil {
			t.Fatal("Failed to update section is published", "error", err)
		}

		lesson, serviceErr := testServices.CreateLesson(context.Background(), services.CreateLessonOptions{
			UserID:       testUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "existing-series",
			SectionID:    sectionID,
			Title:        "Some lesson",
		})
		if serviceErr != nil {
			t.Fatal("Failed to create lesson", "serviceError", serviceErr)
		}
		lessonID = lesson.ID
	}()

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 200 OK when a non-published lesson is found and the user is staff",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.LessonResponse{})
				AssertNotEmpty(t, resBody.Title)
				AssertEqual(t, resBody.IsCompleted, false)
				AssertEqual(
					t,
					resBody.Links.Self.Href,
					fmt.Sprintf(
						"https://api.kiwiscript.com/api/v1/languages/rust/series/existing-series/sections/%d/lessons/%d",
						sectionID, lessonID,
					),
				)
				AssertEqual(
					t,
					resBody.Links.Section.Href,
					fmt.Sprintf(
						"https://api.kiwiscript.com/api/v1/languages/rust/series/existing-series/sections/%d",
						sectionID,
					),
				)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d",
				baseLanguagesPath, sectionID, lessonID,
			),
		},
		{
			Name: "Should return 200 OK with files",
			ReqFn: func(t *testing.T) (string, string) {
				testServices := GetTestServices(t)
				fheaders := [2]*multipart.FileHeader{
					FileUploadMock(t),
					FileUploadMock(t),
				}

				for i, fh := range fheaders {
					uploadOpts := services.UploadLessonFileOptions{
						UserID:       testUser.ID,
						LanguageSlug: "rust",
						SeriesSlug:   "existing-series",
						SectionID:    sectionID,
						LessonID:     lessonID,
						Name:         fmt.Sprintf("test-file-%d.pdf", i),
						FileHeader:   fh,
					}
					if _, err := testServices.UploadLessonFile(context.Background(), uploadOpts); err != nil {
						t.Fatal("Failed to create lesson file", "error", err)
					}
				}

				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.LessonResponse{})
				AssertNotEmpty(t, resBody.Title)
				AssertEqual(t, resBody.IsCompleted, false)
				AssertEqual(
					t,
					resBody.Links.Self.Href,
					fmt.Sprintf(
						"https://api.kiwiscript.com/api/v1/languages/rust/series/existing-series/sections/%d/lessons/%d",
						sectionID, lessonID,
					),
				)
				AssertEqual(
					t,
					resBody.Links.Section.Href,
					fmt.Sprintf(
						"https://api.kiwiscript.com/api/v1/languages/rust/series/existing-series/sections/%d",
						sectionID,
					),
				)
				AssertEqual(t, len(resBody.Embedded.Files), 2)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d",
				baseLanguagesPath, sectionID, lessonID,
			),
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
