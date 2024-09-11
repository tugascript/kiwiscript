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

func TestCreateOrUpdateLessonProgress(t *testing.T) {
	languagesCleanUp(t)()
	staffUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)

	var sectionID, lessonID int32
	func() {
		testDb := GetTestDatabase(t)
		testServices := GetTestServices(t)
		ctx := context.Background()
		requestID := uuid.NewString()

		prms := db.CreateLanguageParams{
			Name:     "Rust",
			Icon:     strings.TrimSpace(languageIcons["Rust"]),
			AuthorID: staffUser.ID,
			Slug:     "rust",
		}
		if _, err := testDb.CreateLanguage(ctx, prms); err != nil {
			t.Fatal("Failed to create language", err)
		}

		serPrms := db.CreateSeriesParams{
			LanguageSlug: "rust",
			Title:        "Rust Series",
			Slug:         "rust-series",
			AuthorID:     staffUser.ID,
			Description:  "Some cool rust series",
		}
		if _, err := testDb.CreateSeries(ctx, serPrms); err != nil {
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

		sectionID = section.ID
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

		lessonID = lesson.ID
		artOpts := services.CreateLessonArticleOptions{
			RequestID:    requestID,
			UserID:       staffUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
			SectionID:    section.ID,
			LessonID:     lesson.ID,
			Content:      strings.Repeat("Some cool rust lesson ", 10),
		}
		if _, serviceErr := testServices.CreateLessonArticle(ctx, artOpts); serviceErr != nil {
			t.Fatal("Failed to create lesson article", "serviceErr", serviceErr)
		}

		pubLessonOpts := services.UpdateLessonIsPublishedOptions{
			RequestID:    requestID,
			UserID:       staffUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
			SectionID:    section.ID,
			LessonID:     lesson.ID,
			IsPublished:  true,
		}
		if _, serviceErr := testServices.UpdateLessonIsPublished(ctx, pubLessonOpts); serviceErr != nil {
			t.Fatal("Failed to update lesson is published", "serviceErr", serviceErr)
		}

		pubSecOpts := services.UpdateSectionIsPublishedOptions{
			RequestID:    requestID,
			UserID:       staffUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
			SectionID:    section.ID,
			IsPublished:  true,
		}
		if _, serviceErr := testServices.UpdateSectionIsPublished(ctx, pubSecOpts); serviceErr != nil {
			t.Fatal("Failed to update section is published", "serviceErr", serviceErr)
		}

		pubSerOpts := services.UpdateSeriesIsPublishedOptions{
			RequestID:    requestID,
			UserID:       staffUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
			IsPublished:  true,
		}
		if _, serviceErr := testServices.UpdateSeriesIsPublished(ctx, pubSerOpts); serviceErr != nil {
			t.Fatal("Failed to update series is published", "serviceErr", serviceErr)
		}
	}()

	beforeEach := func(t *testing.T) {
		testServices := GetTestServices(t)
		ctx := context.Background()
		requestID := uuid.NewString()

		langOpts := services.CreateOrUpdateLanguageProgressOptions{
			RequestID:    requestID,
			UserID:       testUser.ID,
			LanguageSlug: "rust",
		}
		if _, _, _, serviceErr := testServices.CreateOrUpdateLanguageProgress(ctx, langOpts); serviceErr != nil {
			t.Fatal("Failed to create language progress", "serviceErr", serviceErr)
		}

		serOpts := services.CreateOrUpdateSeriesProgressOptions{
			RequestID:    requestID,
			UserID:       testUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
		}
		if _, _, _, seviceErr := testServices.CreateOrUpdateSeriesProgress(ctx, serOpts); seviceErr != nil {
			t.Fatal("Failed to create series progress", "serviceErr", seviceErr)
		}

		secOpts := services.CreateOrUpdateSectionProgressOptions{
			RequestID:    requestID,
			UserID:       testUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
			SectionID:    sectionID,
		}
		if _, _, _, serviceErr := testServices.CreateOrUpdateSectionProgress(ctx, secOpts); serviceErr != nil {
			t.Fatal("Failed to create section progress", "serviceErr", serviceErr)
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
			Name: "Should return 201 CREATED and a lesson progress",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusCreated,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.LessonResponse{})
				AssertNotEmpty(t, resBody.ViewedAt)
				AssertEqual(t, resBody.IsCompleted, false)
				afterEach(t)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/rust-series/sections/%d/lessons/%d/progress",
				baseLanguagesPath,
				sectionID,
				lessonID,
			),
		},
		{
			Name: "Should return 200 OK when a lesson progress is updated",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				testServices := GetTestServices(t)
				ctx := context.Background()
				requestID := uuid.NewString()
				opts := services.CreateOrUpdateLessonProgressOptions{
					RequestID:    requestID,
					UserID:       testUser.ID,
					LanguageSlug: "rust",
					SeriesSlug:   "rust-series",
					SectionID:    sectionID,
					LessonID:     lessonID,
				}
				if _, _, _, serviceErr := testServices.CreateOrUpdateLessonProgress(ctx, opts); serviceErr != nil {
					t.Fatal("Failed to create lesson progress", "serviceErr", serviceErr)
				}

				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.LessonResponse{})
				AssertNotEmpty(t, resBody.ViewedAt)
				AssertEqual(t, resBody.IsCompleted, false)
				afterEach(t)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/rust-series/sections/%d/lessons/%d/progress",
				baseLanguagesPath,
				sectionID,
				lessonID,
			),
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
			Path: fmt.Sprintf(
				"%s/rust/series/rust-series/sections/%d/lessons/%d/progress",
				baseLanguagesPath,
				sectionID,
				lessonID,
			),
		},
		{
			Name: "Should return 403 FORBIDDEN when the user is a staff",
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
			Path: fmt.Sprintf(
				"%s/rust/series/rust-series/sections/%d/lessons/%d/progress",
				baseLanguagesPath,
				sectionID,
				lessonID,
			),
		},
		{
			Name: "Should return 404 NOT FOUND when the lesson is not found",
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
			Path: fmt.Sprintf(
				"%s/rust/series/rust-series/sections/%d/lessons/987654321/progress",
				baseLanguagesPath,
				sectionID,
			),
		},
		{
			Name: "Should return 404 NOT FOUND when the sections is not found",
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
			Path: fmt.Sprintf(
				"%s/rust/series/rust-series/sections/987654321/lessons/%d/progress",
				baseLanguagesPath,
				lessonID,
			),
		},
		{
			Name: "Should return 404 NOT FOUND when the series is not found",
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
			Path: fmt.Sprintf(
				"%s/rust/series/python-series/sections/%d/lessons/%d/progress",
				baseLanguagesPath,
				sectionID,
				lessonID,
			),
		},
		{
			Name: "Should return 404 NOT FOUND when the language is not found",
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
			Path: fmt.Sprintf(
				"%s/python/series/rust-series/sections/%d/lessons/%d/progress",
				baseLanguagesPath,
				sectionID,
				lessonID,
			),
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

func TestCompleteLessonProgress(t *testing.T) {
	languagesCleanUp(t)()
	staffUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)

	var sectionID, lessonID, lesson2ID int32
	func() {
		testDb := GetTestDatabase(t)
		testServices := GetTestServices(t)
		ctx := context.Background()
		requestID := uuid.NewString()

		prms := db.CreateLanguageParams{
			Name:     "Rust",
			Icon:     strings.TrimSpace(languageIcons["Rust"]),
			AuthorID: staffUser.ID,
			Slug:     "rust",
		}
		if _, err := testDb.CreateLanguage(ctx, prms); err != nil {
			t.Fatal("Failed to create language", err)
		}

		serPrms := db.CreateSeriesParams{
			LanguageSlug: "rust",
			Title:        "Rust Series",
			Slug:         "rust-series",
			AuthorID:     staffUser.ID,
			Description:  "Some cool rust series",
		}
		if _, err := testDb.CreateSeries(ctx, serPrms); err != nil {
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

		sectionID = section.ID
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

		lessonID = lesson.ID
		artOpts := services.CreateLessonArticleOptions{
			RequestID:    requestID,
			UserID:       staffUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
			SectionID:    section.ID,
			LessonID:     lesson.ID,
			Content:      strings.Repeat("Some cool rust lesson ", 10),
		}
		if _, serviceErr := testServices.CreateLessonArticle(ctx, artOpts); serviceErr != nil {
			t.Fatal("Failed to create lesson article", "serviceErr", serviceErr)
		}

		pubLessonOpts := services.UpdateLessonIsPublishedOptions{
			RequestID:    requestID,
			UserID:       staffUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
			SectionID:    section.ID,
			LessonID:     lesson.ID,
			IsPublished:  true,
		}
		if _, serviceErr := testServices.UpdateLessonIsPublished(ctx, pubLessonOpts); serviceErr != nil {
			t.Fatal("Failed to update lesson is published", "serviceErr", serviceErr)
		}

		lesson2, err := testDb.CreateLesson(ctx, db.CreateLessonParams{
			Title:        "Cool advanced rust lesson",
			AuthorID:     staffUser.ID,
			SectionID:    section.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
		})
		if err != nil {
			t.Fatal("Failed to create second lesson", err)
		}

		lesson2ID = lesson2.ID
		art2Opts := services.CreateLessonArticleOptions{
			RequestID:    requestID,
			UserID:       staffUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
			SectionID:    section.ID,
			LessonID:     lesson2.ID,
			Content:      strings.Repeat("Some cool python lesson ", 10),
		}
		if _, serviceErr := testServices.CreateLessonArticle(ctx, art2Opts); serviceErr != nil {
			t.Fatal("Failed to create lesson article", "serviceErr", serviceErr)
		}

		pubLesson2Opts := services.UpdateLessonIsPublishedOptions{
			RequestID:    requestID,
			UserID:       staffUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
			SectionID:    section.ID,
			LessonID:     lesson2.ID,
			IsPublished:  true,
		}
		if _, serviceErr := testServices.UpdateLessonIsPublished(ctx, pubLesson2Opts); serviceErr != nil {
			t.Fatal("Failed to update lesson is published", "serviceErr", serviceErr)
		}

		pubSecOpts := services.UpdateSectionIsPublishedOptions{
			RequestID:    requestID,
			UserID:       staffUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
			SectionID:    section.ID,
			IsPublished:  true,
		}
		if _, serviceErr := testServices.UpdateSectionIsPublished(ctx, pubSecOpts); serviceErr != nil {
			t.Fatal("Failed to update section is published", "serviceErr", serviceErr)
		}

		pubSerOpts := services.UpdateSeriesIsPublishedOptions{
			RequestID:    requestID,
			UserID:       staffUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
			IsPublished:  true,
		}
		if _, serviceErr := testServices.UpdateSeriesIsPublished(ctx, pubSerOpts); serviceErr != nil {
			t.Fatal("Failed to update series is published", "serviceErr", serviceErr)
		}
	}()

	beforeEach := func(t *testing.T, lessonID int32) {
		testServices := GetTestServices(t)
		ctx := context.Background()
		requestID := uuid.NewString()

		langOpts := services.CreateOrUpdateLanguageProgressOptions{
			RequestID:    requestID,
			UserID:       testUser.ID,
			LanguageSlug: "rust",
		}
		if _, _, _, serviceErr := testServices.CreateOrUpdateLanguageProgress(ctx, langOpts); serviceErr != nil {
			t.Fatal("Failed to create language progress", "serviceErr", serviceErr)
		}

		serOpts := services.CreateOrUpdateSeriesProgressOptions{
			RequestID:    requestID,
			UserID:       testUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
		}
		if _, _, _, seviceErr := testServices.CreateOrUpdateSeriesProgress(ctx, serOpts); seviceErr != nil {
			t.Fatal("Failed to create series progress", "serviceErr", seviceErr)
		}

		secOpts := services.CreateOrUpdateSectionProgressOptions{
			RequestID:    requestID,
			UserID:       testUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
			SectionID:    sectionID,
		}
		if _, _, _, serviceErr := testServices.CreateOrUpdateSectionProgress(ctx, secOpts); serviceErr != nil {
			t.Fatal("Failed to create section progress", "serviceErr", serviceErr)
		}

		lesOpts := services.CreateOrUpdateLessonProgressOptions{
			RequestID:    requestID,
			UserID:       testUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
			SectionID:    sectionID,
			LessonID:     lessonID,
		}
		if _, _, _, serviceErr := testServices.CreateOrUpdateLessonProgress(ctx, lesOpts); serviceErr != nil {
			t.Fatal("Failed to create lesson progress", "serviceErr", serviceErr)
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
			Name: "Should return 200 OK and complete a lesson progress",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t, lessonID)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.LessonResponse{})
				AssertNotEmpty(t, resBody.ViewedAt)
				AssertEqual(t, resBody.IsCompleted, true)
				AssertEqual(t, resBody.Embedded, nil)
				afterEach(t)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/rust-series/sections/%d/lessons/%d/progress/complete",
				baseLanguagesPath,
				sectionID,
				lessonID,
			),
		},
		{
			Name: "Should return 200 OK and complete the last lesson progress and return a certificate",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t, lessonID)
				service := GetTestServices(t)
				ctx := context.Background()
				requestID := uuid.NewString()

				opts := services.CreateOrUpdateLessonProgressOptions{
					RequestID:    requestID,
					UserID:       testUser.ID,
					LanguageSlug: "rust",
					SeriesSlug:   "rust-series",
					SectionID:    sectionID,
					LessonID:     lesson2ID,
				}
				if _, _, _, serviceErr := service.CreateOrUpdateLessonProgress(ctx, opts); serviceErr != nil {
					t.Fatal("Failed to create lesson progress", "serviceErr", serviceErr)
				}

				completeOpts := services.CompleteLessonProgressOptions{
					RequestID:    requestID,
					UserID:       testUser.ID,
					LanguageSlug: "rust",
					SeriesSlug:   "rust-series",
					SectionID:    sectionID,
					LessonID:     lesson2ID,
				}
				if _, _, _, serviceErr := service.CompleteLessonProgress(ctx, completeOpts); serviceErr != nil {
					t.Fatal("Failed to complete lesson progress", "serviceErr", serviceErr)
				}

				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.LessonResponse{})
				AssertNotEmpty(t, resBody.ViewedAt)
				AssertEqual(t, resBody.IsCompleted, true)
				AssertNotEmpty(t, resBody.Embedded.Certificate)
				AssertEqual(t, resBody.Embedded.Certificate.SeriesTitle, "Rust Series")
				afterEach(t)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/rust-series/sections/%d/lessons/%d/progress/complete",
				baseLanguagesPath,
				sectionID,
				lessonID,
			),
		},
		{
			Name: "Should return 401 UNAUTHORIZED when the user is not authenticated",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t, lessonID)
				return "", ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertUnauthorizedResponse(t, resp)
				afterEach(t)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/rust-series/sections/%d/lessons/%d/progress/complete",
				baseLanguagesPath,
				sectionID,
				lessonID,
			),
		},
		{
			Name: "Should return 403 FORBIDDEN when the user is a staff",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t, lessonID)
				staffUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, staffUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusForbidden,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertForbiddenResponse(t, resp)
				afterEach(t)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/rust-series/sections/%d/lessons/%d/progress/complete",
				baseLanguagesPath,
				sectionID,
				lessonID,
			),
		},
		{
			Name: "Should return 404 NOT FOUND when the lesson is not found",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t, lessonID)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
				afterEach(t)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/rust-series/sections/%d/lessons/%d/progress/complete",
				baseLanguagesPath,
				sectionID,
				987654321,
			),
		},
		{
			Name: "Should return 404 NOT FOUND when the section is not found",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t, lessonID)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
				afterEach(t)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/rust-series/sections/%d/lessons/%d/progress/complete",
				baseLanguagesPath,
				987654321,
				lessonID,
			),
		},
		{
			Name: "Should return 404 NOT FOUND when the series is not found",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t, lessonID)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
				afterEach(t)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/python-series/sections/%d/lessons/%d/progress/complete",
				baseLanguagesPath,
				sectionID,
				lessonID,
			),
		},
		{
			Name: "Should return 404 NOT FOUND when the language is not found",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t, lessonID)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
				afterEach(t)
			},
			Path: fmt.Sprintf(
				"%s/python/series/rust-series/sections/%d/lessons/%d/progress/complete",
				baseLanguagesPath,
				sectionID,
				lessonID,
			),
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

func TestResetLessonProgress(t *testing.T) {
	languagesCleanUp(t)()
	staffUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)

	var sectionID, lessonID, lesson2ID int32
	func() {
		testDb := GetTestDatabase(t)
		testServices := GetTestServices(t)
		ctx := context.Background()
		requestID := uuid.NewString()

		prms := db.CreateLanguageParams{
			Name:     "Rust",
			Icon:     strings.TrimSpace(languageIcons["Rust"]),
			AuthorID: staffUser.ID,
			Slug:     "rust",
		}
		if _, err := testDb.CreateLanguage(ctx, prms); err != nil {
			t.Fatal("Failed to create language", err)
		}

		serPrms := db.CreateSeriesParams{
			LanguageSlug: "rust",
			Title:        "Rust Series",
			Slug:         "rust-series",
			AuthorID:     staffUser.ID,
			Description:  "Some cool rust series",
		}
		if _, err := testDb.CreateSeries(ctx, serPrms); err != nil {
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

		sectionID = section.ID
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

		lessonID = lesson.ID
		artOpts := services.CreateLessonArticleOptions{
			RequestID:    requestID,
			UserID:       staffUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
			SectionID:    section.ID,
			LessonID:     lesson.ID,
			Content:      strings.Repeat("Some cool rust lesson ", 10),
		}
		if _, serviceErr := testServices.CreateLessonArticle(ctx, artOpts); serviceErr != nil {
			t.Fatal("Failed to create lesson article", "serviceErr", serviceErr)
		}

		pubLessonOpts := services.UpdateLessonIsPublishedOptions{
			RequestID:    requestID,
			UserID:       staffUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
			SectionID:    section.ID,
			LessonID:     lesson.ID,
			IsPublished:  true,
		}
		if _, serviceErr := testServices.UpdateLessonIsPublished(ctx, pubLessonOpts); serviceErr != nil {
			t.Fatal("Failed to update lesson is published", "serviceErr", serviceErr)
		}

		lesson2, err := testDb.CreateLesson(ctx, db.CreateLessonParams{
			Title:        "Cool advanced rust lesson",
			AuthorID:     staffUser.ID,
			SectionID:    section.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
		})
		if err != nil {
			t.Fatal("Failed to create second lesson", err)
		}

		lesson2ID = lesson2.ID
		art2Opts := services.CreateLessonArticleOptions{
			RequestID:    requestID,
			UserID:       staffUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
			SectionID:    section.ID,
			LessonID:     lesson2.ID,
			Content:      strings.Repeat("Some cool python lesson ", 10),
		}
		if _, serviceErr := testServices.CreateLessonArticle(ctx, art2Opts); serviceErr != nil {
			t.Fatal("Failed to create lesson article", "serviceErr", serviceErr)
		}

		pubLesson2Opts := services.UpdateLessonIsPublishedOptions{
			RequestID:    requestID,
			UserID:       staffUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
			SectionID:    section.ID,
			LessonID:     lesson2.ID,
			IsPublished:  true,
		}
		if _, serviceErr := testServices.UpdateLessonIsPublished(ctx, pubLesson2Opts); serviceErr != nil {
			t.Fatal("Failed to update lesson is published", "serviceErr", serviceErr)
		}

		pubSecOpts := services.UpdateSectionIsPublishedOptions{
			RequestID:    requestID,
			UserID:       staffUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
			SectionID:    section.ID,
			IsPublished:  true,
		}
		if _, serviceErr := testServices.UpdateSectionIsPublished(ctx, pubSecOpts); serviceErr != nil {
			t.Fatal("Failed to update section is published", "serviceErr", serviceErr)
		}

		pubSerOpts := services.UpdateSeriesIsPublishedOptions{
			RequestID:    requestID,
			UserID:       staffUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
			IsPublished:  true,
		}
		if _, serviceErr := testServices.UpdateSeriesIsPublished(ctx, pubSerOpts); serviceErr != nil {
			t.Fatal("Failed to update series is published", "serviceErr", serviceErr)
		}
	}()

	beforeEach := func(t *testing.T) {
		testServices := GetTestServices(t)
		ctx := context.Background()
		requestID := uuid.NewString()

		langOpts := services.CreateOrUpdateLanguageProgressOptions{
			RequestID:    requestID,
			UserID:       testUser.ID,
			LanguageSlug: "rust",
		}
		if _, _, _, serviceErr := testServices.CreateOrUpdateLanguageProgress(ctx, langOpts); serviceErr != nil {
			t.Fatal("Failed to create language progress", "serviceErr", serviceErr)
		}

		serOpts := services.CreateOrUpdateSeriesProgressOptions{
			RequestID:    requestID,
			UserID:       testUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
		}
		if _, _, _, seviceErr := testServices.CreateOrUpdateSeriesProgress(ctx, serOpts); seviceErr != nil {
			t.Fatal("Failed to create series progress", "serviceErr", seviceErr)
		}

		secOpts := services.CreateOrUpdateSectionProgressOptions{
			RequestID:    requestID,
			UserID:       testUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
			SectionID:    sectionID,
		}
		if _, _, _, serviceErr := testServices.CreateOrUpdateSectionProgress(ctx, secOpts); serviceErr != nil {
			t.Fatal("Failed to create section progress", "serviceErr", serviceErr)
		}

		lesOpts := services.CreateOrUpdateLessonProgressOptions{
			RequestID:    requestID,
			UserID:       testUser.ID,
			LanguageSlug: "rust",
			SeriesSlug:   "rust-series",
			SectionID:    sectionID,
			LessonID:     lessonID,
		}
		if _, _, _, serviceErr := testServices.CreateOrUpdateLessonProgress(ctx, lesOpts); serviceErr != nil {
			t.Fatal("Failed to create lesson progress", "serviceErr", serviceErr)
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
			Name: "Should return 204 NO CONTENT when a lesson progress is reset",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNoContent,
			AssertFn: func(t *testing.T, _ string, _ *http.Response) {
				afterEach(t)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/rust-series/sections/%d/lessons/%d/progress",
				baseLanguagesPath, sectionID, lessonID,
			),
		},
		{
			Name: "Should return 204 NO CONTENT when a completed lesson progress is reset",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				service := GetTestServices(t)
				ctx := context.Background()
				requestID := uuid.NewString()
				opts := services.CompleteLessonProgressOptions{
					RequestID:    requestID,
					UserID:       testUser.ID,
					LanguageSlug: "rust",
					SeriesSlug:   "rust-series",
					SectionID:    sectionID,
					LessonID:     lessonID,
				}
				if _, _, _, serviceErr := service.CompleteLessonProgress(ctx, opts); serviceErr != nil {
					t.Fatal("Failed to complete lesson progress", "serviceErr", serviceErr)
				}

				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNoContent,
			AssertFn: func(t *testing.T, _ string, _ *http.Response) {
				afterEach(t)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/rust-series/sections/%d/lessons/%d/progress",
				baseLanguagesPath, sectionID, lessonID,
			),
		},
		{
			Name: "Should return 204 NO CONTENT when a lesson progress in a completed series is reset",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				testServices := GetTestServices(t)
				ctx := context.Background()
				requestID := uuid.NewString()
				cmpOpts := services.CompleteLessonProgressOptions{
					RequestID:    requestID,
					UserID:       testUser.ID,
					LanguageSlug: "rust",
					SeriesSlug:   "rust-series",
					SectionID:    sectionID,
					LessonID:     lessonID,
				}
				if _, _, _, serviceErr := testServices.CompleteLessonProgress(ctx, cmpOpts); serviceErr != nil {
					t.Fatal("Failed to complete lesson progress", "serviceErr", serviceErr)
				}

				progOpts2 := services.CreateOrUpdateLessonProgressOptions{
					RequestID:    requestID,
					UserID:       testUser.ID,
					LanguageSlug: "rust",
					SeriesSlug:   "rust-series",
					SectionID:    sectionID,
					LessonID:     lesson2ID,
				}
				if _, _, _, serviceErr := testServices.CreateOrUpdateLessonProgress(ctx, progOpts2); serviceErr != nil {
					t.Fatal("Failed to create lesson progress", "serviceErr", serviceErr)
				}

				cmpOpts2 := services.CompleteLessonProgressOptions{
					RequestID:    requestID,
					UserID:       testUser.ID,
					LanguageSlug: "rust",
					SeriesSlug:   "rust-series",
					SectionID:    sectionID,
					LessonID:     lesson2ID,
				}
				if _, _, _, serviceErr := testServices.CompleteLessonProgress(ctx, cmpOpts2); serviceErr != nil {
					t.Fatal("Failed to complete lesson progress", "serviceErr", serviceErr)
				}

				seriesProgress, serviceErr := testServices.FindSeriesProgress(ctx, services.FindSeriesProgressOptions{
					RequestID:    requestID,
					UserID:       testUser.ID,
					LanguageSlug: "rust",
					SeriesSlug:   "rust-series",
				})
				if serviceErr != nil {
					t.Fatal("Failed to find series progress", "serviceErr", serviceErr)
				}
				AssertEqual(t, seriesProgress.CompletedAt.Valid, true)

				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNoContent,
			AssertFn: func(t *testing.T, _ string, _ *http.Response) {
				testServices := GetTestServices(t)
				ctx := context.Background()
				requestID := uuid.NewString()
				seriesProgress, serviceErr := testServices.FindSeriesProgress(ctx, services.FindSeriesProgressOptions{
					RequestID:    requestID,
					UserID:       testUser.ID,
					LanguageSlug: "rust",
					SeriesSlug:   "rust-series",
				})

				if serviceErr != nil {
					t.Fatal("Failed to find series progress", "serviceErr", serviceErr)
				}
				AssertEqual(t, seriesProgress.CompletedAt.Valid, false)

				afterEach(t)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/rust-series/sections/%d/lessons/%d/progress",
				baseLanguagesPath, sectionID, lessonID,
			),
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
			Path: fmt.Sprintf(
				"%s/rust/series/rust-series/sections/%d/lessons/%d/progress",
				baseLanguagesPath, sectionID, lessonID,
			),
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
			Path: fmt.Sprintf(
				"%s/rust/series/rust-series/sections/%d/lessons/%d/progress",
				baseLanguagesPath, sectionID, lessonID,
			),
		},
		{
			Name: "Should return 404 NOT FOUND when the lesson is not found",
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
			Path: fmt.Sprintf(
				"%s/rust/series/rust-series/sections/%d/lessons/%d/progress",
				baseLanguagesPath,
				sectionID,
				987654321,
			),
		},
		{
			Name: "Should return 404 NOT FOUND when the section is not found",
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
			Path: fmt.Sprintf(
				"%s/rust/series/rust-series/sections/%d/lessons/%d/progress",
				baseLanguagesPath,
				987654321,
				lessonID,
			),
		},
		{
			Name: "Should return 404 NOT FOUND when the series is not found",
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
			Path: fmt.Sprintf(
				"%s/rust/series/python-series/sections/%d/lessons/%d/progress",
				baseLanguagesPath,
				sectionID,
				lessonID,
			),
		},
		{
			Name: "Should return 404 NOT FOUND when the language is not found",
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
			Path: fmt.Sprintf(
				"%s/python/series/rust-series/sections/%d/lessons/%d/progress",
				baseLanguagesPath,
				sectionID,
				lessonID,
			),
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
