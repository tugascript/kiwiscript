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
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kiwiscript/kiwiscript_go/dtos"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/services"
	"net/http"
	"strings"
	"testing"
)

func TestUploadLessonFile(t *testing.T) {
	languagesCleanUp(t)()
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	testUser.IsStaff = true

	var sectionID, lessonID int32
	func() {
		testDb := GetTestDatabase(t)
		testServices := GetTestServices(t)
		ctx := context.Background()

		// Create language
		params := db.CreateLanguageParams{
			Name:     "Rust",
			Icon:     strings.TrimSpace(languageIcons["Rust"]),
			AuthorID: testUser.ID,
			Slug:     "rust",
		}
		if _, err := testDb.CreateLanguage(ctx, params); err != nil {
			t.Fatal("Failed to create language", err)
		}

		// Create series
		series, err := testDb.CreateSeries(ctx, db.CreateSeriesParams{
			Title:        "Existing Series",
			Slug:         "existing-series",
			Description:  "Some description",
			LanguageSlug: "rust",
			AuthorID:     testUser.ID,
		})
		if err != nil {
			t.Fatal("Failed to create series", "error", err)
		}

		// Publish series
		isPubPrms := db.UpdateSeriesIsPublishedParams{
			IsPublished: true,
			ID:          series.ID,
		}
		if _, err := testDb.UpdateSeriesIsPublished(ctx, isPubPrms); err != nil {
			t.Fatal("Failed to update series is published", "error", err)
		}

		// Create section
		section, serviceErr := testServices.CreateSection(ctx, services.CreateSectionOptions{
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

		// Publish section
		isPubSecPrms := db.UpdateSectionIsPublishedParams{
			IsPublished: true,
			ID:          sectionID,
		}
		if _, err := testDb.UpdateSectionIsPublished(ctx, isPubSecPrms); err != nil {
			t.Fatal("Failed to update section is published", "error", err)
		}

		// Create lesson
		lesson, serviceErr := testServices.CreateLesson(ctx, services.CreateLessonOptions{
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

	testCases := []TestRequestCase[FormFileBody]{
		{
			Name: "Should return 201 CREATED when uploading a file",
			ReqFn: func(t *testing.T) (FormFileBody, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return FileUploadForm(t, "Some Rust Cheatsheet", "pdf"), accessToken
			},
			ExpStatus: fiber.StatusCreated,
			AssertFn: func(t *testing.T, req FormFileBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.LessonFileResponse{})
				AssertEqual(t, resBody.Name, "Some Rust Cheatsheet")
				AssertEqual(t, resBody.Ext, "pdf")
				AssertStringContains(t, resBody.URL, resBody.ID.String()+".pdf")
				AssertStringContains(t, resBody.Links.Lesson.Href, fmt.Sprintf(
					"%s/rust/series/existing-series/sections/%d/lessons/%d",
					baseLanguagesPath,
					sectionID,
					lessonID,
				))
				AssertStringContains(t, resBody.Links.LessonFiles.Href, fmt.Sprintf(
					"%s/rust/series/existing-series/sections/%d/lessons/%d/files",
					baseLanguagesPath,
					sectionID,
					lessonID,
				))
				AssertStringContains(t, resBody.Links.Self.Href, fmt.Sprintf(
					"%s/rust/series/existing-series/sections/%d/lessons/%d/files/%s",
					baseLanguagesPath,
					sectionID,
					lessonID,
					resBody.ID.String(),
				))
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d/files",
				baseLanguagesPath,
				sectionID,
				lessonID,
			),
		},
		{
			Name: "Should return 400 BAD REQUEST when uploading a file with an invalid extension",
			ReqFn: func(t *testing.T) (FormFileBody, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return FileUploadForm(t, "Some Rust Cheatsheet 2", "ods"), accessToken
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, _ FormFileBody, resp *http.Response) {
				AssertValidationErrorWithoutFieldsResponse(t, resp, "File type not supported")
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d/files",
				baseLanguagesPath,
				sectionID,
				lessonID,
			),
		},
		{
			Name: "Should return 409 Conflict if file name already exists",
			ReqFn: func(t *testing.T) (FormFileBody, string) {
				testServices := GetTestServices(t)
				fheader := FileUploadMock(t)
				ctx := context.Background()
				requestID := uuid.NewString()

				uploadOpts := services.UploadLessonFileOptions{
					RequestID:    requestID,
					UserID:       testUser.ID,
					LanguageSlug: "rust",
					SeriesSlug:   "existing-series",
					SectionID:    sectionID,
					LessonID:     lessonID,
					Name:         "Existing name",
					FileHeader:   fheader,
				}
				if _, serviceErr := testServices.UploadLessonFile(ctx, uploadOpts); serviceErr != nil {
					t.Fatal("Failed to upload lesson file", "serviceError", serviceErr)
				}

				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return FileUploadForm(t, "Existing name", "pdf"), accessToken
			},
			ExpStatus: fiber.StatusConflict,
			AssertFn: func(t *testing.T, _ FormFileBody, resp *http.Response) {
				AssertConflictDuplicateKeyResponse(t, resp)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d/files",
				baseLanguagesPath,
				sectionID,
				lessonID,
			),
		},
		{
			Name: "Should return 401 UNAUTHORIZED when uploading a file without an access token",
			ReqFn: func(t *testing.T) (FormFileBody, string) {
				return FileUploadForm(t, "Some Rust Cheatsheet 3", "pdf"), ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ FormFileBody, resp *http.Response) {
				AssertUnauthorizedResponse(t, resp)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d/files",
				baseLanguagesPath,
				sectionID,
				lessonID,
			),
		},
		{
			Name: "Should return 403 FORBIDDEN when uploading a file with a user that is not staff",
			ReqFn: func(t *testing.T) (FormFileBody, string) {
				testUser.IsStaff = false
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return FileUploadForm(t, "Some Rust Cheatsheet 4", "pdf"), accessToken
			},
			ExpStatus: fiber.StatusForbidden,
			AssertFn: func(t *testing.T, _ FormFileBody, resp *http.Response) {
				AssertForbiddenResponse(t, resp)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d/files",
				baseLanguagesPath,
				sectionID,
				lessonID,
			),
		},
		{
			Name: "Should return 404 NOT FOUND when uploading a file with a non-existing lesson",
			ReqFn: func(t *testing.T) (FormFileBody, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return FileUploadForm(t, "Some Rust Cheatsheet 5", "pdf"), accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ FormFileBody, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d/files",
				baseLanguagesPath,
				sectionID,
				987654321,
			),
		},
		{
			Name: "Should return 404 NOT FOUND when uploading a file with a non-existing section",
			ReqFn: func(t *testing.T) (FormFileBody, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return FileUploadForm(t, "Some Rust Cheatsheet 6", "pdf"), accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ FormFileBody, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d/files",
				baseLanguagesPath,
				987654321,
				lessonID,
			),
		},
		{
			Name: "Should return 404 NOT FOUND when uploading a file with a non-existing series",
			ReqFn: func(t *testing.T) (FormFileBody, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return FileUploadForm(t, "Some Rust Cheatsheet 7", "pdf"), accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ FormFileBody, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/non-existing-series/sections/%d/lessons/%d/files",
				baseLanguagesPath,
				sectionID,
				lessonID,
			),
		},
		{
			Name: "Should return 404 NOT FOUND when uploading a file with a non-existing language",
			ReqFn: func(t *testing.T) (FormFileBody, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return FileUploadForm(t, "Some Rust Cheatsheet 8", "pdf"), accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ FormFileBody, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf(
				"%s/python/series/existing-series/sections/%d/lessons/%d/files",
				baseLanguagesPath,
				sectionID,
				lessonID,
			),
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

//func TestGetLessonFiles(t *testing.T) {
//	languagesCleanUp(t)()
//	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
//	testUser.IsStaff = true
//
//	var sectionID, lessonID int32
//	func() {
//		testDb := GetTestDatabase(t)
//		testServices := GetTestServices(t)
//		ctx := context.Background()
//
//		// Create language
//		params := db.CreateLanguageParams{
//			Name:     "Rust",
//			Icon:     strings.TrimSpace(languageIcons["Rust"]),
//			AuthorID: testUser.ID,
//			Slug:     "rust",
//		}
//		if _, err := testDb.CreateLanguage(ctx, params); err != nil {
//			t.Fatal("Failed to create language", err)
//		}
//
//		// Create series
//		series, err := testDb.CreateSeries(ctx, db.CreateSeriesParams{
//			Title:        "Existing Series",
//			Slug:         "existing-series",
//			Description:  "Some description",
//			LanguageSlug: "rust",
//			AuthorID:     testUser.ID,
//		})
//		if err != nil {
//			t.Fatal("Failed to create series", "error", err)
//		}
//
//		// Publish series
//		isPubPrms := db.UpdateSeriesIsPublishedParams{
//			IsPublished: true,
//			ID:          series.ID,
//		}
//		if _, err := testDb.UpdateSeriesIsPublished(ctx, isPubPrms); err != nil {
//			t.Fatal("Failed to update series is published", "error", err)
//		}
//
//		// Create section
//		section, serviceErr := testServices.CreateSection(ctx, services.CreateSectionOptions{
//			UserID:       testUser.ID,
//			Title:        "Some Section",
//			LanguageSlug: "rust",
//			SeriesSlug:   "existing-series",
//			Description:  "Some description",
//		})
//		if serviceErr != nil {
//			t.Fatal("Failed to create section", "serviceError", serviceErr)
//		}
//		sectionID = section.ID
//
//		// Publish section
//		isPubSecPrms := db.UpdateSectionIsPublishedParams{
//			IsPublished: true,
//			ID:          sectionID,
//		}
//		if _, err := testDb.UpdateSectionIsPublished(ctx, isPubSecPrms); err != nil {
//			t.Fatal("Failed to update section is published", "error", err)
//		}
//
//		// Create lesson
//		lesson, serviceErr := testServices.CreateLesson(ctx, services.CreateLessonOptions{
//			UserID:       testUser.ID,
//			LanguageSlug: "rust",
//			SeriesSlug:   "existing-series",
//			SectionID:    sectionID,
//			Title:        "Some lesson",
//		})
//		if serviceErr != nil {
//			t.Fatal("Failed to create lesson", "serviceError", serviceErr)
//		}
//		lessonID = lesson.ID
//	}()
//}
