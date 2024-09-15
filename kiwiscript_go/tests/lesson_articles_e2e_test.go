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
	"github.com/google/uuid"
	"github.com/kiwiscript/kiwiscript_go/dtos"
	"github.com/kiwiscript/kiwiscript_go/exceptions"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/services"
	"net/http"
	"strings"
	"testing"
)

func TestCreateLessonArticle(t *testing.T) {
	languagesCleanUp(t)()
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)

	var sectionID, lessonID int32
	func() {
		testDb := GetTestDatabase(t)
		testServices := GetTestServices(t)

		// Create language
		params := db.CreateLanguageParams{
			Name:     "Rust",
			Icon:     strings.TrimSpace(languageIcons["Rust"]),
			AuthorID: testUser.ID,
			Slug:     "rust",
		}
		if _, err := testDb.CreateLanguage(context.Background(), params); err != nil {
			t.Fatal("Failed to create language", err)
		}

		// Create series
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

		// Publish series
		isPubPrms := db.UpdateSeriesIsPublishedParams{
			IsPublished: true,
			ID:          series.ID,
		}
		if _, err := testDb.UpdateSeriesIsPublished(context.Background(), isPubPrms); err != nil {
			t.Fatal("Failed to update series is published", "error", err)
		}

		// Create section
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

		// Publish section
		isPubSecPrms := db.UpdateSectionIsPublishedParams{
			IsPublished: true,
			ID:          sectionID,
		}
		if _, err := testDb.UpdateSectionIsPublished(context.Background(), isPubSecPrms); err != nil {
			t.Fatal("Failed to update section is published", "error", err)
		}

		// Create lesson
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

	generateFakeLessonArticleData := func(t *testing.T) dtos.LessonArticleBody {
		return dtos.LessonArticleBody{Content: faker.Paragraph()}
	}

	testCases := []TestRequestCase[dtos.LessonArticleBody]{
		{
			Name: "Should return 201 CREATED when a lesson article is created",
			ReqFn: func(t *testing.T) (dtos.LessonArticleBody, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return generateFakeLessonArticleData(t), accessToken
			},
			ExpStatus: fiber.StatusCreated,
			AssertFn: func(t *testing.T, req dtos.LessonArticleBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.LessonArticleResponse{})
				AssertEqual(t, req.Content, resBody.Content)
				AssertEqual(t, resBody.ReadTime, services.CalculateReadingTime(req.Content))
			},
			Path: fmt.Sprintf("%s/rust/series/existing-series/sections/%d/lessons/%d/article",
				baseLanguagesPath, sectionID, lessonID),
		},
		{
			Name: "Should return 400 BAD REQUEST when the content is empty",
			ReqFn: func(t *testing.T) (dtos.LessonArticleBody, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.LessonArticleBody{Content: ""}, accessToken
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, _ dtos.LessonArticleBody, resp *http.Response) {
				AssertValidationErrorResponse(t, resp, []ValidationErrorAssertion{{
					Param:   "content",
					Message: exceptions.FieldErrMessageRequired,
				}})
			},
			Path: fmt.Sprintf("%s/rust/series/existing-series/sections/%d/lessons/%d/article",
				baseLanguagesPath, sectionID, lessonID),
		},
		{
			Name: "Should return 403 FORBIDDEN when the user is not staff",
			ReqFn: func(t *testing.T) (dtos.LessonArticleBody, string) {
				testUser.IsStaff = false
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return generateFakeLessonArticleData(t), accessToken
			},
			ExpStatus: fiber.StatusForbidden,
			AssertFn: func(t *testing.T, _ dtos.LessonArticleBody, resp *http.Response) {
				AssertForbiddenResponse(t, resp)
			},
			Path: fmt.Sprintf("%s/rust/series/existing-series/sections/%d/lessons/%d/article",
				baseLanguagesPath, sectionID, lessonID),
		},
		{
			Name: "Should return 401 UNAUTHORIZED when the user is not authenticated",
			ReqFn: func(t *testing.T) (dtos.LessonArticleBody, string) {
				return generateFakeLessonArticleData(t), ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ dtos.LessonArticleBody, resp *http.Response) {
				AssertUnauthorizedResponse(t, resp)
			},
			Path: fmt.Sprintf("%s/rust/series/existing-series/sections/%d/lessons/%d/article",
				baseLanguagesPath, sectionID, lessonID),
		},
		{
			Name: "Should return 404 NOT FOUND when the lesson does not exist",
			ReqFn: func(t *testing.T) (dtos.LessonArticleBody, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return generateFakeLessonArticleData(t), accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ dtos.LessonArticleBody, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf("%s/rust/series/existing-series/sections/%d/lessons/9999999/article",
				baseLanguagesPath, sectionID),
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

func TestGetLessonArticle(t *testing.T) {
	languagesCleanUp(t)()
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)

	var sectionID, lessonID int32
	func() {
		testDb := GetTestDatabase(t)
		testServices := GetTestServices(t)

		// Create language
		params := db.CreateLanguageParams{
			Name:     "Rust",
			Icon:     strings.TrimSpace(languageIcons["Rust"]),
			AuthorID: testUser.ID,
			Slug:     "rust",
		}
		if _, err := testDb.CreateLanguage(context.Background(), params); err != nil {
			t.Fatal("Failed to create language", err)
		}

		// Create series
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

		// Publish series
		isPubPrms := db.UpdateSeriesIsPublishedParams{
			IsPublished: true,
			ID:          series.ID,
		}
		if _, err := testDb.UpdateSeriesIsPublished(context.Background(), isPubPrms); err != nil {
			t.Fatal("Failed to update series is published", "error", err)
		}

		// Create section
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

		// Publish section
		isPubSecPrms := db.UpdateSectionIsPublishedParams{
			IsPublished: true,
			ID:          sectionID,
		}
		if _, err := testDb.UpdateSectionIsPublished(context.Background(), isPubSecPrms); err != nil {
			t.Fatal("Failed to update section is published", "error", err)
		}

		// Create lesson
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

		// Create lesson article
		_, serviceErr = testServices.CreateLessonArticle(context.Background(), services.CreateLessonArticleOptions{
			UserID:       testUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "existing-series",
			SectionID:    sectionID,
			LessonID:     lessonID,
			Content:      "Sample article content",
		})
		if serviceErr != nil {
			t.Fatal("Failed to create lesson article", "serviceError", serviceErr)
		}
	}()

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 200 OK when the unpublished lesson article is found and user is staff",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.LessonArticleResponse{})
				AssertEqual(t, resBody.Content, "Sample article content")
			},
			Path: fmt.Sprintf("%s/rust/series/existing-series/sections/%d/lessons/%d/article",
				baseLanguagesPath, sectionID, lessonID),
		},
		{
			Name: "Should return 200 OK when the published lesson article is found and user is not unauthenticated",
			ReqFn: func(t *testing.T) (string, string) {
				testServices := GetTestServices(t)
				ctx := context.Background()
				requestID := uuid.NewString()

				pubLessonOpts := services.UpdateLessonIsPublishedOptions{
					RequestID:    requestID,
					UserID:       testUser.ID,
					LanguageSlug: "rust",
					SeriesSlug:   "existing-series",
					SectionID:    sectionID,
					LessonID:     lessonID,
					IsPublished:  true,
				}
				if _, serviceErr := testServices.UpdateLessonIsPublished(ctx, pubLessonOpts); serviceErr != nil {
					t.Fatal("Failed to update lesson is published", "serviceErr", serviceErr)
				}

				pubSecOpts := services.UpdateSectionIsPublishedOptions{
					RequestID:    requestID,
					UserID:       testUser.ID,
					LanguageSlug: "rust",
					SeriesSlug:   "existing-series",
					SectionID:    sectionID,
					IsPublished:  true,
				}
				if _, serviceErr := testServices.UpdateSectionIsPublished(ctx, pubSecOpts); serviceErr != nil {
					t.Fatal("Failed to update section is published", "serviceErr", serviceErr)
				}

				pubSerOpts := services.UpdateSeriesIsPublishedOptions{
					RequestID:    requestID,
					UserID:       testUser.ID,
					LanguageSlug: "rust",
					SeriesSlug:   "existing-series",
					IsPublished:  true,
				}
				if _, serviceErr := testServices.UpdateSeriesIsPublished(ctx, pubSerOpts); serviceErr != nil {
					t.Fatal("Failed to update series is published", "serviceErr", serviceErr)
				}

				return "", ""
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.LessonArticleResponse{})
				AssertEqual(t, resBody.Content, "Sample article content")
			},
			Path: fmt.Sprintf("%s/rust/series/existing-series/sections/%d/lessons/%d/article",
				baseLanguagesPath, sectionID, lessonID),
		},
		{
			Name: "Should return 404 NOT FOUND when the lesson article does not exist",
			ReqFn: func(t *testing.T) (string, string) {
				// Delete the article
				testDb := GetTestDatabase(t)
				ctx := context.Background()

				lessonArticle, err := testDb.GetLessonArticleByLessonID(ctx, lessonID)
				if err != nil {
					t.Fatal("Failed to get lesson article", "error", err)
				}

				if err := testDb.DeleteLessonArticle(ctx, lessonArticle.ID); err != nil {
					t.Fatal("Failed to delete lesson article", "error", err)
				}
				return "", ""
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf("%s/rust/series/existing-series/sections/%d/lessons/%d/article",
				baseLanguagesPath, sectionID, lessonID),
		},
		{
			Name: "Should return 404 NOT FOUND when the lesson does not exist",
			ReqFn: func(t *testing.T) (string, string) {
				return "", ""
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf("%s/rust/series/existing-series/sections/%d/lessons/987654321/article",
				baseLanguagesPath, sectionID),
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

func TestUpdateLessonArticle(t *testing.T) {
	languagesCleanUp(t)()
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)

	var sectionID, lessonID int32
	func() {
		testDb := GetTestDatabase(t)
		testServices := GetTestServices(t)

		// Create language
		params := db.CreateLanguageParams{
			Name:     "Rust",
			Icon:     strings.TrimSpace(languageIcons["Rust"]),
			AuthorID: testUser.ID,
			Slug:     "rust",
		}
		if _, err := testDb.CreateLanguage(context.Background(), params); err != nil {
			t.Fatal("Failed to create language", err)
		}

		// Create series
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

		// Publish series
		isPubPrms := db.UpdateSeriesIsPublishedParams{
			IsPublished: true,
			ID:          series.ID,
		}
		if _, err := testDb.UpdateSeriesIsPublished(context.Background(), isPubPrms); err != nil {
			t.Fatal("Failed to update series is published", "error", err)
		}

		// Create section
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

		// Publish section
		isPubSecPrms := db.UpdateSectionIsPublishedParams{
			IsPublished: true,
			ID:          sectionID,
		}
		if _, err := testDb.UpdateSectionIsPublished(context.Background(), isPubSecPrms); err != nil {
			t.Fatal("Failed to update section is published", "error", err)
		}

		// Create lesson
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

		// Create lesson article
		_, serviceErr = testServices.CreateLessonArticle(context.Background(), services.CreateLessonArticleOptions{
			UserID:       testUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "existing-series",
			SectionID:    sectionID,
			LessonID:     lessonID,
			Content:      "Original article content",
		})
		if serviceErr != nil {
			t.Fatal("Failed to create lesson article", "serviceError", serviceErr)
		}
	}()

	generateUpdatedLessonArticleData := func(t *testing.T) dtos.LessonArticleBody {
		return dtos.LessonArticleBody{Content: "Updated article content"}
	}

	testCases := []TestRequestCase[dtos.LessonArticleBody]{
		{
			Name: "Should return 200 OK when the lesson article is updated",
			ReqFn: func(t *testing.T) (dtos.LessonArticleBody, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return generateUpdatedLessonArticleData(t), accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, req dtos.LessonArticleBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.LessonArticleResponse{})
				AssertEqual(t, req.Content, resBody.Content)
			},
			Path: fmt.Sprintf("%s/rust/series/existing-series/sections/%d/lessons/%d/article",
				baseLanguagesPath, sectionID, lessonID),
		},
		{
			Name: "Should return 404 NOT FOUND when the lesson article does not exist",
			ReqFn: func(t *testing.T) (dtos.LessonArticleBody, string) {
				// Delete the article
				testDb := GetTestDatabase(t)
				ctx := context.Background()

				lessonArticle, err := testDb.GetLessonArticleByLessonID(ctx, lessonID)
				if err != nil {
					t.Fatal("Failed to get lesson article", "error", err)
				}

				if err := testDb.DeleteLessonArticle(ctx, lessonArticle.ID); err != nil {
					t.Fatal("Failed to delete lesson article", "error", err)
				}

				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return generateUpdatedLessonArticleData(t), accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ dtos.LessonArticleBody, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf("%s/rust/series/existing-series/sections/%d/lessons/%d/article",
				baseLanguagesPath, sectionID, lessonID),
		},
		{
			Name: "Should return 403 FORBIDDEN when the user is not staff",
			ReqFn: func(t *testing.T) (dtos.LessonArticleBody, string) {
				testUser.IsStaff = false
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return generateUpdatedLessonArticleData(t), accessToken
			},
			ExpStatus: fiber.StatusForbidden,
			AssertFn: func(t *testing.T, _ dtos.LessonArticleBody, resp *http.Response) {
				AssertForbiddenResponse(t, resp)
			},
			Path: fmt.Sprintf("%s/rust/series/existing-series/sections/%d/lessons/%d/article",
				baseLanguagesPath, sectionID, lessonID),
		},
		{
			Name: "Should return 401 UNAUTHORIZED when the user is not authenticated",
			ReqFn: func(t *testing.T) (dtos.LessonArticleBody, string) {
				return generateUpdatedLessonArticleData(t), ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ dtos.LessonArticleBody, resp *http.Response) {
				AssertUnauthorizedResponse(t, resp)
			},
			Path: fmt.Sprintf("%s/rust/series/existing-series/sections/%d/lessons/%d/article",
				baseLanguagesPath, sectionID, lessonID),
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

func TestDeleteLessonArticle(t *testing.T) {
	languagesCleanUp(t)()
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)

	var sectionID, lessonID int32
	func() {
		testDb := GetTestDatabase(t)
		testServices := GetTestServices(t)

		// Create language
		params := db.CreateLanguageParams{
			Name:     "Rust",
			Icon:     strings.TrimSpace(languageIcons["Rust"]),
			AuthorID: testUser.ID,
			Slug:     "rust",
		}
		if _, err := testDb.CreateLanguage(context.Background(), params); err != nil {
			t.Fatal("Failed to create language", err)
		}

		// Create series
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

		// Publish series
		isPubPrms := db.UpdateSeriesIsPublishedParams{
			IsPublished: true,
			ID:          series.ID,
		}
		if _, err := testDb.UpdateSeriesIsPublished(context.Background(), isPubPrms); err != nil {
			t.Fatal("Failed to update series is published", "error", err)
		}

		// Create section
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

		// Publish section
		isPubSecPrms := db.UpdateSectionIsPublishedParams{
			IsPublished: true,
			ID:          sectionID,
		}
		if _, err := testDb.UpdateSectionIsPublished(context.Background(), isPubSecPrms); err != nil {
			t.Fatal("Failed to update section is published", "error", err)
		}

		// Create lesson
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

	beforeEach := func(t *testing.T) {
		testServices := GetTestServices(t)
		ctx := context.Background()
		opts := services.CreateLessonArticleOptions{
			UserID:       testUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "existing-series",
			SectionID:    sectionID,
			LessonID:     lessonID,
			Content:      "Original article content",
		}
		if _, serviceErr := testServices.CreateLessonArticle(ctx, opts); serviceErr != nil {
			t.Fatal("Failed to create lesson article", "serviceError", serviceErr)
		}
	}

	afterEach := func(t *testing.T) {
		testDb := GetTestDatabase(t)
		ctx := context.Background()

		lessonArticle, err := testDb.GetLessonArticleByLessonID(ctx, lessonID)
		if err != nil {
			t.Log("Failed to get lesson article", "error", err)
			return
		}

		if err := testDb.DeleteLessonArticle(ctx, lessonArticle.ID); err != nil {
			t.Fatal("Failed to delete lesson article", "error", err)
		}
	}

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 204 NO CONTENT when the lesson article is deleted",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNoContent,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				afterEach(t)
			},
			Path: fmt.Sprintf("%s/rust/series/existing-series/sections/%d/lessons/%d/article",
				baseLanguagesPath, sectionID, lessonID),
		},
		{
			Name: "Should return 404 NOT FOUND when the lesson article does not exist",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf("%s/rust/series/existing-series/sections/%d/lessons/%d/article",
				baseLanguagesPath, sectionID, lessonID),
		},
		{
			Name: "Should return 403 FORBIDDEN when the user is not staff",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = false
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusForbidden,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertForbiddenResponse(t, resp)
			},
			Path: fmt.Sprintf("%s/rust/series/existing-series/sections/%d/lessons/%d/article",
				baseLanguagesPath, sectionID, lessonID),
		},
		{
			Name: "Should return 401 UNAUTHORIZED when the user is not authenticated",
			ReqFn: func(t *testing.T) (string, string) {
				return "", ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertUnauthorizedResponse(t, resp)
			},
			Path: fmt.Sprintf("%s/rust/series/existing-series/sections/%d/lessons/%d/article",
				baseLanguagesPath, sectionID, lessonID),
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
