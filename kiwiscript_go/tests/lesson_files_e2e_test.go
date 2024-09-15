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
	"github.com/kiwiscript/kiwiscript_go/exceptions"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/services"
	"mime/multipart"
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

func TestGetLessonFiles(t *testing.T) {
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

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 200 OK when getting lesson files with 0 files",
			ReqFn: func(t *testing.T) (string, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, []dtos.LessonFileResponse{})
				AssertEqual(t, len(resBody), 0)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d/files",
				baseLanguagesPath,
				sectionID,
				lessonID,
			),
		},
		{
			Name: "Should return 200 OK when getting lesson files with multiple files",
			ReqFn: func(t *testing.T) (string, string) {
				testServices := GetTestServices(t)
				ctx := context.Background()
				fheaders := [4]*multipart.FileHeader{
					FileUploadMock(t),
					FileUploadMock(t),
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
					if _, err := testServices.UploadLessonFile(ctx, uploadOpts); err != nil {
						t.Fatal("Failed to create lesson file", "error", err)
					}
				}

				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, []dtos.LessonFileResponse{})
				AssertEqual(t, len(resBody), 4)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d/files",
				baseLanguagesPath,
				sectionID,
				lessonID,
			),
		},
		{
			Name: "Should return 403 FORBIDDEN when getting lesson files without an access token when the lesson is not published",
			ReqFn: func(t *testing.T) (string, string) {
				return "", ""
			},
			ExpStatus: fiber.StatusForbidden,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
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
			Name: "Should return 200 OK when getting lesson files without an access token when the lesson is published",
			ReqFn: func(t *testing.T) (string, string) {
				testDb := GetTestDatabase(t)
				ctx := context.Background()

				// Publish lesson
				isPubPrms := db.UpdateLessonIsPublishedParams{
					IsPublished: true,
					ID:          lessonID,
				}
				if _, err := testDb.UpdateLessonIsPublished(ctx, isPubPrms); err != nil {
					t.Fatal("Failed to update lesson is published", "error", err)
				}

				return "", ""
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, []dtos.LessonFileResponse{})
				AssertGreaterThan(t, len(resBody), -1)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d/files",
				baseLanguagesPath,
				sectionID,
				lessonID,
			),
		},
		{
			Name: "Should return 404 if the lesson is not found",
			ReqFn: func(t *testing.T) (string, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
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
			Name: "Should return 404 if the section is not found",
			ReqFn: func(t *testing.T) (string, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
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
			Name: "Should return 404 if the series is not found",
			ReqFn: func(t *testing.T) (string, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
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
			Name: "Should return 404 if the language is not found",
			ReqFn: func(t *testing.T) (string, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
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
			PerformTestRequestCase(t, http.MethodGet, tc.Path, tc)
		})
	}

	t.Cleanup(languagesCleanUp(t))
	t.Cleanup(userCleanUp(t))
}

func TestGetLessonFile(t *testing.T) {
	languagesCleanUp(t)()
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	testUser.IsStaff = true

	var sectionID, lessonID int32
	var fileID uuid.UUID
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

		fh := FileUploadMock(t)
		lessonFile, serviceErr := testServices.UploadLessonFile(ctx, services.UploadLessonFileOptions{
			RequestID:    uuid.NewString(),
			UserID:       testUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "existing-series",
			SectionID:    sectionID,
			LessonID:     lessonID,
			Name:         "test-file.pdf",
			FileHeader:   fh,
		})
		if serviceErr != nil {
			t.Fatal("Failed to upload lesson file", "error", serviceErr)
		}
		fileID = lessonFile.ID
	}()

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 200 OK when getting a lesson file",
			ReqFn: func(t *testing.T) (string, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.LessonFileResponse{})
				AssertEqual(t, resBody.Name, "test-file.pdf")
				AssertEqual(t, resBody.Ext, "pdf")
				AssertStringContains(t, resBody.URL, fileID.String()+".pdf")
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
					fileID.String(),
				))
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d/files/%s",
				baseLanguagesPath,
				sectionID,
				lessonID,
				fileID,
			),
		},
		{
			Name: "Should return 403 FORBIDDEN if the lesson is not published and the user is not staff",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = false
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusForbidden,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertForbiddenResponse(t, resp)
				testUser.IsStaff = true
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d/files/%s",
				baseLanguagesPath,
				sectionID,
				lessonID,
				fileID,
			),
		},
		{
			Name: "Should return 200 OK if the lesson is published and the user is not staff",
			ReqFn: func(t *testing.T) (string, string) {
				testDb := GetTestDatabase(t)
				ctx := context.Background()

				// Publish lesson
				isPubPrms := db.UpdateLessonIsPublishedParams{
					IsPublished: true,
					ID:          lessonID,
				}
				if _, err := testDb.UpdateLessonIsPublished(ctx, isPubPrms); err != nil {
					t.Fatal("Failed to update lesson is published", "error", err)
				}

				testUser.IsStaff = false
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.LessonFileResponse{})
				AssertEqual(t, resBody.Name, "test-file.pdf")
				AssertEqual(t, resBody.Ext, "pdf")
				AssertStringContains(t, resBody.URL, fileID.String()+".pdf")
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
					fileID.String(),
				))
				testUser.IsStaff = true
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d/files/%s",
				baseLanguagesPath,
				sectionID,
				lessonID,
				fileID,
			),
		},
		{
			Name: "Should return 404 if the lesson file is not found",
			ReqFn: func(t *testing.T) (string, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d/files/%s",
				baseLanguagesPath,
				sectionID,
				lessonID,
				uuid.NewString(),
			),
		},
		{
			Name: "Should return 404 if the lesson is not found",
			ReqFn: func(t *testing.T) (string, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d/files/%s",
				baseLanguagesPath,
				sectionID,
				987654321,
				fileID.String(),
			),
		},
		{
			Name: "Should return 404 if the section is not found",
			ReqFn: func(t *testing.T) (string, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d/files/%s",
				baseLanguagesPath,
				987654321,
				lessonID,
				fileID.String(),
			),
		},
		{
			Name: "Should return 404 if the series is not found",
			ReqFn: func(t *testing.T) (string, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/non-existing-series/sections/%d/lessons/%d/files/%s",
				baseLanguagesPath,
				sectionID,
				lessonID,
				fileID.String(),
			),
		},
		{
			Name: "Should return 404 if the language is not found",
			ReqFn: func(t *testing.T) (string, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf(
				"%s/python/series/existing-series/sections/%d/lessons/%d/files/%s",
				baseLanguagesPath,
				sectionID,
				lessonID,
				fileID.String(),
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

func TestDeleteLessonFile(t *testing.T) {
	languagesCleanUp(t)()
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	testUser.IsStaff = true

	var sectionID, lessonID int32
	var fileID uuid.UUID
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

	beforeEach := func(t *testing.T) {
		testServices := GetTestServices(t)
		ctx := context.Background()
		fh := FileUploadMock(t)
		lessonFile, serviceErr := testServices.UploadLessonFile(ctx, services.UploadLessonFileOptions{
			RequestID:    uuid.NewString(),
			UserID:       testUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "existing-series",
			SectionID:    sectionID,
			LessonID:     lessonID,
			Name:         "test-file.pdf",
			FileHeader:   fh,
		})
		if serviceErr != nil {
			t.Fatal("Failed to upload lesson file", "error", serviceErr)
		}
		fileID = lessonFile.ID
	}

	afterEach := func(t *testing.T) {
		testDb := GetTestDatabase(t)
		ctx := context.Background()
		if err := testDb.DeleteLessonFile(ctx, fileID); err != nil {
			t.Log("Failed to delete lesson file", "error", err)
		}
	}

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 204 NO CONTENT when getting a lesson file",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNoContent,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				afterEach(t)
			},
			PathFn: func() string {
				return fmt.Sprintf(
					"%s/rust/series/existing-series/sections/%d/lessons/%d/files/%s",
					baseLanguagesPath,
					sectionID,
					lessonID,
					fileID,
				)
			},
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
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertForbiddenResponse(t, resp)
				testUser.IsStaff = true
				afterEach(t)
			},
			PathFn: func() string {
				return fmt.Sprintf(
					"%s/rust/series/existing-series/sections/%d/lessons/%d/files/%s",
					baseLanguagesPath,
					sectionID,
					lessonID,
					fileID,
				)
			},
		},
		{
			Name: "Should return 404 if the lesson file is not found",
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
			PathFn: func() string {
				return fmt.Sprintf(
					"%s/rust/series/existing-series/sections/%d/lessons/%d/files/%s",
					baseLanguagesPath,
					sectionID,
					lessonID,
					uuid.NewString(),
				)
			},
		},
		{
			Name: "Should return 404 if the lesson is not found",
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
			PathFn: func() string {
				return fmt.Sprintf(
					"%s/rust/series/existing-series/sections/%d/lessons/%d/files/%s",
					baseLanguagesPath,
					sectionID,
					987654321,
					fileID.String(),
				)
			},
		},
		{
			Name: "Should return 404 if the section is not found",
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
			PathFn: func() string {
				return fmt.Sprintf(
					"%s/rust/series/existing-series/sections/%d/lessons/%d/files/%s",
					baseLanguagesPath,
					987654321,
					lessonID,
					fileID.String(),
				)
			},
		},
		{
			Name: "Should return 404 if the series is not found",
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
			PathFn: func() string {
				return fmt.Sprintf(
					"%s/rust/series/non-existing-series/sections/%d/lessons/%d/files/%s",
					baseLanguagesPath,
					sectionID,
					lessonID,
					fileID.String(),
				)
			},
		},
		{
			Name: "Should return 404 if the language is not found",
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
			PathFn: func() string {
				return fmt.Sprintf(
					"%s/python/series/existing-series/sections/%d/lessons/%d/files/%s",
					baseLanguagesPath,
					sectionID,
					lessonID,
					fileID.String(),
				)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCaseWithPathFn(t, http.MethodDelete, tc)
		})
	}

	t.Cleanup(languagesCleanUp(t))
	t.Cleanup(userCleanUp(t))
}

func TestUpdateLessonFile(t *testing.T) {
	languagesCleanUp(t)()
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	testUser.IsStaff = true

	var sectionID, lessonID int32
	var fileID uuid.UUID
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

		fh := FileUploadMock(t)
		lessonFile, serviceErr := testServices.UploadLessonFile(ctx, services.UploadLessonFileOptions{
			RequestID:    uuid.NewString(),
			UserID:       testUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "existing-series",
			SectionID:    sectionID,
			LessonID:     lessonID,
			Name:         "test-file.pdf",
			FileHeader:   fh,
		})
		if serviceErr != nil {
			t.Fatal("Failed to upload lesson file", "error", serviceErr)
		}
		fileID = lessonFile.ID
	}()

	testCases := []TestRequestCase[dtos.LessonFileBody]{
		{
			Name: "Should return 200 OK when updating lesson file",
			ReqFn: func(t *testing.T) (dtos.LessonFileBody, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.LessonFileBody{Name: "New File Name"}, accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, req dtos.LessonFileBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.LessonFileResponse{})
				AssertEqual(t, resBody.Name, req.Name)
				AssertEqual(t, resBody.Ext, "pdf")
				AssertStringContains(t, resBody.URL, fileID.String()+".pdf")
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
					fileID.String(),
				))
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d/files/%s",
				baseLanguagesPath,
				sectionID,
				lessonID,
				fileID,
			),
		},
		{
			Name: "Should return 400 BAD REQUEST if the request body is invalid",
			ReqFn: func(t *testing.T) (dtos.LessonFileBody, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.LessonFileBody{Name: "N"}, accessToken
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req dtos.LessonFileBody, resp *http.Response) {
				AssertValidationErrorResponse(t, resp, []ValidationErrorAssertion{
					{
						Param:   "name",
						Message: exceptions.StrFieldErrMessageMin,
					},
				})
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d/files/%s",
				baseLanguagesPath,
				sectionID,
				lessonID,
				fileID,
			),
		},
		{
			Name: "Should return 403 FORBIDDEN if the user is not staff",
			ReqFn: func(t *testing.T) (dtos.LessonFileBody, string) {
				testUser.IsStaff = false
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.LessonFileBody{Name: "New Name"}, accessToken
			},
			ExpStatus: fiber.StatusForbidden,
			AssertFn: func(t *testing.T, _ dtos.LessonFileBody, resp *http.Response) {
				AssertForbiddenResponse(t, resp)
				testUser.IsStaff = true
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d/files/%s",
				baseLanguagesPath,
				sectionID,
				lessonID,
				fileID,
			),
		},
		{
			Name: "Should return 404 if the lesson file is not found",
			ReqFn: func(t *testing.T) (dtos.LessonFileBody, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.LessonFileBody{Name: "New Name"}, accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ dtos.LessonFileBody, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d/files/%s",
				baseLanguagesPath,
				sectionID,
				lessonID,
				uuid.NewString(),
			),
		},
		{
			Name: "Should return 404 if the lesson is not found",
			ReqFn: func(t *testing.T) (dtos.LessonFileBody, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.LessonFileBody{Name: "New Name"}, accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ dtos.LessonFileBody, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d/files/%s",
				baseLanguagesPath,
				sectionID,
				987654321,
				fileID.String(),
			),
		},
		{
			Name: "Should return 404 if the section is not found",
			ReqFn: func(t *testing.T) (dtos.LessonFileBody, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.LessonFileBody{Name: "New Name"}, accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ dtos.LessonFileBody, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d/files/%s",
				baseLanguagesPath,
				987654321,
				lessonID,
				fileID.String(),
			),
		},
		{
			Name: "Should return 404 if the series is not found",
			ReqFn: func(t *testing.T) (dtos.LessonFileBody, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.LessonFileBody{Name: "New Name"}, accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ dtos.LessonFileBody, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/non-existing-series/sections/%d/lessons/%d/files/%s",
				baseLanguagesPath,
				sectionID,
				lessonID,
				fileID.String(),
			),
		},
		{
			Name: "Should return 404 if the language is not found",
			ReqFn: func(t *testing.T) (dtos.LessonFileBody, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.LessonFileBody{Name: "New Name"}, accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ dtos.LessonFileBody, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf(
				"%s/python/series/existing-series/sections/%d/lessons/%d/files/%s",
				baseLanguagesPath,
				sectionID,
				lessonID,
				fileID.String(),
			),
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
