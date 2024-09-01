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
					Message: exceptions.StrFieldErrMessageMax,
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
		{
			Name: "Should return 200 OK if the series is published and no user is logged in",
			ReqFn: func(t *testing.T) (string, string) {
				testDb := GetTestDatabase(t)
				ctx := context.Background()
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
				resBody := AssertTestResponseBody(t, resp, dtos.LessonResponse{})
				AssertNotEmpty(t, resBody.Title)
				AssertEqual(t, resBody.IsCompleted, false)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d",
				baseLanguagesPath, sectionID, lessonID,
			),
		},
		{
			Name: "Should return 200 OK with user progress if the user is not staff",
			ReqFn: func(t *testing.T) (string, string) {
				testDb := GetTestDatabase(t)
				ctx := context.Background()
				progUser := confirmTestUser(t, CreateTestUser(t, nil).ID)

				isPubPrms := db.UpdateLessonIsPublishedParams{
					IsPublished: true,
					ID:          lessonID,
				}
				if _, err := testDb.UpdateLessonIsPublished(ctx, isPubPrms); err != nil {
					t.Fatal("Failed to update lesson is published", "error", err)
				}

				langProg, err := testDb.CreateLanguageProgress(context.Background(), db.CreateLanguageProgressParams{
					LanguageSlug: "rust",
					UserID:       testUser.ID,
				})
				if err != nil {
					t.Fatal("Failed to create language progress", "error", err)
				}

				seriesProg, err := testDb.CreateSeriesProgress(context.Background(), db.CreateSeriesProgressParams{
					LanguageSlug:       "rust",
					SeriesSlug:         "existing-series",
					LanguageProgressID: langProg.ID,
					UserID:             testUser.ID,
				})
				if err != nil {
					t.Fatal("Failed to create series progress", "error", err)
				}

				secProg, err := testDb.CreateSectionProgress(context.Background(), db.CreateSectionProgressParams{
					LanguageSlug:       "rust",
					SeriesSlug:         "existing-series",
					SectionID:          sectionID,
					LanguageProgressID: langProg.ID,
					SeriesProgressID:   seriesProg.ID,
					UserID:             testUser.ID,
				})
				if err != nil {
					t.Fatal("Failed to create section progress", "error", err)
				}

				lProg, err := testDb.CreateLessonProgress(context.Background(), db.CreateLessonProgressParams{
					LanguageSlug:       "rust",
					SeriesSlug:         "existing-series",
					SectionID:          sectionID,
					LessonID:           lessonID,
					LanguageProgressID: langProg.ID,
					SeriesProgressID:   seriesProg.ID,
					SectionProgressID:  secProg.ID,
					UserID:             progUser.ID,
				})
				if err != nil {
					t.Fatal("Failed to create lesson progress", "error", err)
				}

				if _, err := testDb.CompleteLessonProgress(ctx, lProg.ID); err != nil {
					t.Fatal("Failed to complete lesson progress", "error", err)
				}

				accessToken, _ := GenerateTestAuthTokens(t, progUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.LessonResponse{})
				AssertNotEmpty(t, resBody.Title)
				AssertEqual(t, resBody.IsCompleted, true)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d",
				baseLanguagesPath, sectionID, lessonID,
			),
		},
		{
			Name: "should return 404 NOT FOUND when the lesson is not found",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/987654321",
				baseLanguagesPath, sectionID,
			),
		},
		{
			Name: "Should return 404 NOT FOUND when the section is not found",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/987654321/lessons/%d",
				baseLanguagesPath, lessonID,
			),
		},
		{
			Name: "Should return 404 NOT FOUND when the series is not found",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/non-existing-series/sections/%d/lessons/%d",
				baseLanguagesPath, sectionID, lessonID,
			),
		},
		{
			Name: "Should return 404 NOT FOUND when the language is not found",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf(
				"%s/python/series/existing-series/sections/%d/lessons/%d",
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

func TestUpdateLesson(t *testing.T) {
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

	testCases := []TestRequestCase[dtos.UpdateLessonBody]{
		{
			Name: "Should return 200 OK when a lesson is updated",
			ReqFn: func(t *testing.T) (dtos.UpdateLessonBody, string) {
				// create multiple lessons
				testServices := GetTestServices(t)
				titleMap := make(map[string]bool)
				for count := 0; count < 5; {
					fakeTitle := faker.Name()
					if _, ok := titleMap[fakeTitle]; !ok {
						lesson, serviceErr := testServices.CreateLesson(context.Background(), services.CreateLessonOptions{
							UserID:       testUser.ID,
							LanguageSlug: "rust",
							SeriesSlug:   "existing-series",
							SectionID:    sectionID,
							Title:        fakeTitle,
						})
						if serviceErr != nil {
							t.Fatal("Failed to create lesson", "serviceError", serviceErr)
						}
						t.Log("Created lesson with position:", lesson.Position)
						titleMap[fakeTitle] = true
						count++
					}
				}

				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.UpdateLessonBody{Title: faker.Name(), Position: 3}, accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, req dtos.UpdateLessonBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.LessonResponse{})
				AssertEqual(t, req.Title, resBody.Title)
				AssertEqual(t, req.Position, resBody.Position)

				// Get the lessons to check if the position is correct
				testDb := GetTestDatabase(t)
				lessons, err := testDb.FindLessonsBySectionID(context.Background(), sectionID)
				if err != nil {
					t.Fatal("Failed to get lessons", "error", err)
				}

				for i, lesson := range lessons {
					if i < 3 {
						AssertGreaterThan(t, resBody.Position, lesson.Position)
					} else if lesson.ID != resBody.ID {
						AssertGreaterThan(t, lesson.Position, resBody.Position)
					}
				}

			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d",
				baseLanguagesPath, sectionID, lessonID,
			),
		},
		{
			Name: "Should return 400 BAD REQUEST when the title is too long",
			ReqFn: func(t *testing.T) (dtos.UpdateLessonBody, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.UpdateLessonBody{Title: strings.Repeat("too-long", 100)}, accessToken
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, _ dtos.UpdateLessonBody, resp *http.Response) {
				AssertValidationErrorResponse(t, resp, []ValidationErrorAssertion{{
					Param:   "title",
					Message: exceptions.StrFieldErrMessageMax,
				}})
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d",
				baseLanguagesPath, sectionID, lessonID,
			),
		},
		{
			Name: "Should return 400 BAD REQUEST when the position is too low",
			ReqFn: func(t *testing.T) (dtos.UpdateLessonBody, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.UpdateLessonBody{Title: faker.Name(), Position: -1}, accessToken
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, _ dtos.UpdateLessonBody, resp *http.Response) {
				AssertValidationErrorResponse(t, resp, []ValidationErrorAssertion{{
					Param:   "position",
					Message: exceptions.IntFieldErrMessageGte,
				}})
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d",
				baseLanguagesPath, sectionID, lessonID,
			),
		},
		{
			Name: "Should return 409 CONFLICT when the title is already taken",
			ReqFn: func(t *testing.T) (dtos.UpdateLessonBody, string) {
				testServices := GetTestServices(t)
				fakeTitle := faker.Name()
				_, serviceErr := testServices.CreateLesson(context.Background(), services.CreateLessonOptions{
					UserID:       testUser.ID,
					LanguageSlug: "rust",
					SeriesSlug:   "existing-series",
					SectionID:    sectionID,
					Title:        fakeTitle,
				})
				if serviceErr != nil {
					t.Fatal("Failed to create lesson", "serviceError", serviceErr)
				}

				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.UpdateLessonBody{Title: fakeTitle, Position: 1}, accessToken
			},
			ExpStatus: fiber.StatusConflict,
			AssertFn: func(t *testing.T, _ dtos.UpdateLessonBody, resp *http.Response) {
				AssertConflictDuplicateKeyResponse(t, resp)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/%d",
				baseLanguagesPath, sectionID, lessonID,
			),
		},
		{
			Name: "Should return 404 NOT FOUND when the lesson is not found",
			ReqFn: func(t *testing.T) (dtos.UpdateLessonBody, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.UpdateLessonBody{Title: faker.Name(), Position: 1}, accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ dtos.UpdateLessonBody, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons/987654321",
				baseLanguagesPath, sectionID,
			),
		},
		{
			Name: "Should return 404 NOT FOUND when the section is not found",
			ReqFn: func(t *testing.T) (dtos.UpdateLessonBody, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.UpdateLessonBody{Title: faker.Name(), Position: 1}, accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ dtos.UpdateLessonBody, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/9876543231/lessons/%d",
				baseLanguagesPath, lessonID,
			),
		},
		{
			Name: "Should return 404 NOT FOUND when the series is not found",
			ReqFn: func(t *testing.T) (dtos.UpdateLessonBody, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.UpdateLessonBody{Title: faker.Name(), Position: 1}, accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ dtos.UpdateLessonBody, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/non-existing-series/sections/%d/lessons/%d",
				baseLanguagesPath, sectionID, lessonID,
			),
		},
		{
			Name: "Should return 404 NOT FOUND when the language is not found",
			ReqFn: func(t *testing.T) (dtos.UpdateLessonBody, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.UpdateLessonBody{Title: faker.Name(), Position: 1}, accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ dtos.UpdateLessonBody, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf(
				"%s/python/series/existing-series/sections/%d/lessons/%d",
				baseLanguagesPath, sectionID, lessonID,
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

func TestGetLessons(t *testing.T) {
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

		isPubSecPrms := db.UpdateSectionIsPublishedParams{
			IsPublished: true,
			ID:          sectionID,
		}
		if _, err := testDb.UpdateSectionIsPublished(context.Background(), isPubSecPrms); err != nil {
			t.Fatal("Failed to update section is published", "error", err)
		}

		for count := 0; count < 10; {
			fakeTitle := faker.Name()
			titleMap := make(map[string]bool)
			if _, ok := titleMap[fakeTitle]; !ok {
				opts := services.CreateLessonOptions{
					RequestID:    uuid.NewString(),
					UserID:       testUser.ID,
					LanguageSlug: "rust",
					SeriesSlug:   "existing-series",
					SectionID:    sectionID,
					Title:        fakeTitle,
				}
				if _, err := testServices.CreateLesson(context.Background(), opts); err != nil {
					t.Fatal("Failed to create lesson", "error", err)
				}
				titleMap[fakeTitle] = true
				count++
			}
		}
	}()

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 200 OK when lessons are found and the user is staff",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.PaginatedResponse[dtos.LessonResponse]{})
				AssertEqual(t, len(resBody.Results), 10)
				AssertEqual(t, resBody.Count, 10)
				AssertGreaterThan(t, resBody.Results[1].Position, resBody.Results[0].Position)
				AssertEqual(
					t,
					resBody.Links.Self.Href,
					fmt.Sprintf(
						"https://api.kiwiscript.com/api/v1/languages/rust/series/existing-series/sections/%d/lessons?limit=25&offset=0",
						sectionID,
					),
				)
				AssertEqual(t, resBody.Links.Next, nil)
				AssertEqual(t, resBody.Links.Prev, nil)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons",
				baseLanguagesPath, sectionID,
			),
		},
		{
			Name: "Should return 200 OK when paginated lessons are found and the user is staff",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.PaginatedResponse[dtos.LessonResponse]{})
				AssertEqual(t, len(resBody.Results), 5)
				AssertEqual(t, resBody.Count, 10)
				AssertGreaterThan(t, resBody.Results[1].Position, resBody.Results[0].Position)
				AssertEqual(
					t,
					resBody.Links.Self.Href,
					fmt.Sprintf(
						"https://api.kiwiscript.com/api/v1/languages/rust/series/existing-series/sections/%d/lessons?limit=5&offset=2",
						sectionID,
					),
				)
				AssertEqual(
					t,
					resBody.Links.Next.Href,
					fmt.Sprintf(
						"https://api.kiwiscript.com/api/v1/languages/rust/series/existing-series/sections/%d/lessons?limit=5&offset=7",
						sectionID,
					),
				)
				AssertEqual(
					t,
					resBody.Links.Prev.Href,
					fmt.Sprintf(
						"https://api.kiwiscript.com/api/v1/languages/rust/series/existing-series/sections/%d/lessons?limit=5&offset=0",
						sectionID,
					),
				)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons?offset=2&limit=5",
				baseLanguagesPath, sectionID,
			),
		},
		{
			Name: "Should return 200 OK with no lessons when the user is not staff and the lessons are not published",
			ReqFn: func(t *testing.T) (string, string) {
				newUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				accessToken, _ := GenerateTestAuthTokens(t, newUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.PaginatedResponse[dtos.LessonResponse]{})
				AssertEqual(t, len(resBody.Results), 0)
				AssertEqual(t, resBody.Count, 0)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons",
				baseLanguagesPath, sectionID,
			),
		},
		{
			Name: "Should return 200 OK with lessons if they are published and there is no user",
			ReqFn: func(t *testing.T) (string, string) {
				testDb := GetTestDatabase(t)
				ctx := context.Background()
				lessons, err := testDb.FindLessonsBySectionID(ctx, sectionID)
				if err != nil {
					t.Fatal("Failed to get lessons", "error", err)
				}

				for _, lesson := range lessons {
					isPubPrms := db.UpdateLessonIsPublishedParams{
						IsPublished: true,
						ID:          lesson.ID,
					}
					if _, err := testDb.UpdateLessonIsPublished(ctx, isPubPrms); err != nil {
						t.Fatal("Failed to update lesson is published", "error", err)
					}
				}

				return "", ""
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.PaginatedResponse[dtos.LessonResponse]{})
				AssertEqual(t, len(resBody.Results), 10)
				AssertEqual(t, resBody.Count, 10)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons",
				baseLanguagesPath, sectionID,
			),
		},
		{
			Name: "Should return 200 OK with lessons if they are published and the user has progress",
			ReqFn: func(t *testing.T) (string, string) {
				testDb := GetTestDatabase(t)
				ctx := context.Background()
				lessons, err := testDb.FindLessonsBySectionID(ctx, sectionID)
				if err != nil {
					t.Fatal("Failed to get lessons", "error", err)
				}

				for _, lesson := range lessons {
					isPubPrms := db.UpdateLessonIsPublishedParams{
						IsPublished: true,
						ID:          lesson.ID,
					}
					if _, err := testDb.UpdateLessonIsPublished(ctx, isPubPrms); err != nil {
						t.Fatal("Failed to update lesson is published", "error", err)
					}
				}

				progUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				langProg, err := testDb.CreateLanguageProgress(context.Background(), db.CreateLanguageProgressParams{
					LanguageSlug: "rust",
					UserID:       progUser.ID,
				})
				if err != nil {
					t.Fatal("Failed to create language progress", "error", err)
				}

				seriesProg, err := testDb.CreateSeriesProgress(context.Background(), db.CreateSeriesProgressParams{
					LanguageSlug:       "rust",
					SeriesSlug:         "existing-series",
					LanguageProgressID: langProg.ID,
					UserID:             progUser.ID,
				})
				if err != nil {
					t.Fatal("Failed to create series progress", "error", err)
				}

				secProg, err := testDb.CreateSectionProgress(context.Background(), db.CreateSectionProgressParams{
					LanguageSlug:       "rust",
					SeriesSlug:         "existing-series",
					SectionID:          sectionID,
					LanguageProgressID: langProg.ID,
					SeriesProgressID:   seriesProg.ID,
					UserID:             progUser.ID,
				})
				if err != nil {
					t.Fatal("Failed to create section progress", "error", err)
				}

				for i, lesson := range lessons {
					lProg, err := testDb.CreateLessonProgress(context.Background(), db.CreateLessonProgressParams{
						LanguageSlug:       "rust",
						SeriesSlug:         "existing-series",
						SectionID:          sectionID,
						LessonID:           lesson.ID,
						LanguageProgressID: langProg.ID,
						SeriesProgressID:   seriesProg.ID,
						SectionProgressID:  secProg.ID,
						UserID:             progUser.ID,
					})
					if err != nil {
						t.Fatal("Failed to create lesson progress", "error", err)
					}

					if i < 3 {
						if _, err := testDb.CompleteLessonProgress(ctx, lProg.ID); err != nil {
							t.Fatal("Failed to complete lesson progress", "error", err)
						}
					}
				}

				accessToken, _ := GenerateTestAuthTokens(t, progUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.PaginatedResponse[dtos.LessonResponse]{})
				AssertEqual(t, len(resBody.Results), 10)
				AssertEqual(t, resBody.Count, 10)

				for i, lesson := range resBody.Results {
					if i < 3 {
						AssertEqual(t, lesson.IsCompleted, true)
					} else {
						AssertEqual(t, lesson.IsCompleted, false)
					}
				}
			},
			Path: fmt.Sprintf(
				"%s/rust/series/existing-series/sections/%d/lessons",
				baseLanguagesPath, sectionID,
			),
		},
		{
			Name: "Should return 404 not found if the section is not found",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series/sections/987654321/lessons",
		},
		{
			Name: "Should return 404 not found if the series is not found",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf(
				"%s/rust/series/non-existing-series/sections/%d/lessons",
				baseLanguagesPath, sectionID,
			),
		},
		{
			Name: "Should return 404 not found if the language is not found",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf(
				"%s/python/series/existing-series/sections/%d/lessons",
				baseLanguagesPath, sectionID,
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

func TestDeleteLesson(t *testing.T) {
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
	}()

	beforeEach := func(t *testing.T) {
		testServices := GetTestServices(t)
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
	}

	afterEach := func(t *testing.T) {
		testDb := GetTestDatabase(t)
		if err := testDb.DeleteLessonByID(context.Background(), lessonID); err != nil {
			t.Fatal("Failed to delete lesson", "error", err)
		}
	}

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 204 NO CONTENT if the lesson is deleted",
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
			PathFn: func() string {
				return fmt.Sprintf(
					"%s/rust/series/existing-series/sections/%d/lessons/%d",
					baseLanguagesPath, sectionID, lessonID,
				)
			},
		},
		{
			Name: "Should return 409 CONFLICT if the lesson has students",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				testDb := GetTestDatabase(t)
				progUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				ctx := context.Background()
				langProg, err := testDb.CreateLanguageProgress(ctx, db.CreateLanguageProgressParams{
					LanguageSlug: "rust",
					UserID:       progUser.ID,
				})
				if err != nil {
					t.Fatal("Failed to create language progress", "error", err)
				}

				seriesProg, err := testDb.CreateSeriesProgress(ctx, db.CreateSeriesProgressParams{
					LanguageSlug:       "rust",
					SeriesSlug:         "existing-series",
					LanguageProgressID: langProg.ID,
					UserID:             progUser.ID,
				})
				if err != nil {
					t.Fatal("Failed to create series progress", "error", err)
				}

				secProg, err := testDb.CreateSectionProgress(ctx, db.CreateSectionProgressParams{
					LanguageSlug:       "rust",
					SeriesSlug:         "existing-series",
					SectionID:          sectionID,
					LanguageProgressID: langProg.ID,
					SeriesProgressID:   seriesProg.ID,
					UserID:             progUser.ID,
				})
				if err != nil {
					t.Fatal("Failed to create section progress", "error", err)
				}

				lPrms := db.CreateLessonProgressParams{
					LanguageSlug:       "rust",
					SeriesSlug:         "existing-series",
					SectionID:          sectionID,
					LessonID:           lessonID,
					LanguageProgressID: langProg.ID,
					SeriesProgressID:   seriesProg.ID,
					SectionProgressID:  secProg.ID,
					UserID:             progUser.ID,
				}
				if _, err := testDb.CreateLessonProgress(ctx, lPrms); err != nil {
					t.Fatal("Failed to create lesson progress", "error", err)
				}

				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusConflict,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertConflictResponse(t, resp, "Lesson has students")
				afterEach(t)
			},
			PathFn: func() string {
				return fmt.Sprintf(
					"%s/rust/series/existing-series/sections/%d/lessons/%d",
					baseLanguagesPath, sectionID, lessonID,
				)
			},
		},
		{
			Name: "Should return 403 FORBIDDEN if the user is staff but not the owner",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				newUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				newUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, newUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusForbidden,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertForbiddenResponse(t, resp)
				afterEach(t)
			},
			PathFn: func() string {
				return fmt.Sprintf(
					"%s/rust/series/existing-series/sections/%d/lessons/%d",
					baseLanguagesPath, sectionID, lessonID,
				)
			},
		},
		{
			Name: "Should return 403 FORBIDDEN if the user is not staff",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				newUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				accessToken, _ := GenerateTestAuthTokens(t, newUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusForbidden,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertForbiddenResponse(t, resp)
				afterEach(t)
			},
			PathFn: func() string {
				return fmt.Sprintf(
					"%s/rust/series/existing-series/sections/%d/lessons/%d",
					baseLanguagesPath, sectionID, lessonID,
				)
			},
		},
		{
			Name: "Should return 404 NOT FOUND if the lesson is not found",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				testUser.IsStaff = true
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
					"%s/rust/series/existing-series/sections/%d/lessons/987654321",
					baseLanguagesPath, sectionID,
				)
			},
		},
		{
			Name: "Should return 404 NOT FOUND if the section is not found",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				testUser.IsStaff = true
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
					"%s/rust/series/existing-series/sections/987654321/lessons/%d",
					baseLanguagesPath, lessonID,
				)
			},
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
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
				afterEach(t)
			},
			PathFn: func() string {
				return fmt.Sprintf(
					"%s/rust/series/non-existing-series/sections/%d/lessons/%d",
					baseLanguagesPath, sectionID, lessonID,
				)
			},
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
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
				afterEach(t)
			},
			PathFn: func() string {
				return fmt.Sprintf(
					"%s/python/series/existing-series/sections/%d/lessons/%d",
					baseLanguagesPath, sectionID, lessonID,
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

func TestPublishLesson(t *testing.T) {
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
	}()

	beforeEach := func(t *testing.T) {
		testServices := GetTestServices(t)
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
	}

	afterEach := func(t *testing.T) {
		testDb := GetTestDatabase(t)
		if err := testDb.DeleteLessonByID(context.Background(), lessonID); err != nil {
			t.Fatal("Failed to delete lesson", "error", err)
		}
	}

	testCases := []TestRequestCase[dtos.UpdateIsPublishedBody]{
		{
			Name: "Should return 200 OK if the lesson is published with article",
			ReqFn: func(t *testing.T) (dtos.UpdateIsPublishedBody, string) {
				beforeEach(t)
				testService := GetTestServices(t)
				ctx := context.Background()
				artOpts := services.CreateLessonArticleOptions{
					RequestID:    uuid.NewString(),
					UserID:       testUser.ID,
					LanguageSlug: "rust",
					SeriesSlug:   "existing-series",
					SectionID:    sectionID,
					LessonID:     lessonID,
					Content:      strings.Repeat("Lorem ipsum dolor sit amet, ", 100),
				}
				_, serviceErr := testService.CreateLessonArticle(ctx, artOpts)
				if serviceErr != nil {
					t.Fatal("Failed to create lesson article", "serviceError", serviceErr)
				}

				lesson, err := testService.FindLessonBySlugsAndIDs(ctx, services.FindLessonOptions{
					RequestID:    uuid.NewString(),
					LanguageSlug: "rust",
					SeriesSlug:   "existing-series",
					SectionID:    sectionID,
					LessonID:     lessonID,
				})
				if err != nil {
					t.Fatal("Failed to find lesson", "error", err)
				}

				article, err := testService.FindLessonArticleByLessonID(ctx, services.FindLessonArticleByLessonIDOptions{
					RequestID: uuid.NewString(),
					LessonID:  lesson.ID,
				})
				if err != nil {
					t.Fatal("Failed to find lesson article", "error", err)
				}
				AssertEqual(t, lesson.ReadTimeSeconds, article.ReadTimeSeconds)

				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.UpdateIsPublishedBody{IsPublished: true}, accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ dtos.UpdateIsPublishedBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.LessonResponse{})
				AssertEqual(t, resBody.IsPublished, true)
				AssertEqual(t, resBody.ReadTime, services.CalculateReadingTime(
					strings.Repeat("Lorem ipsum dolor sit amet, ", 100),
				))
				afterEach(t)
			},
			PathFn: func() string {
				return fmt.Sprintf(
					"%s/rust/series/existing-series/sections/%d/lessons/%d/publish",
					baseLanguagesPath, sectionID, lessonID,
				)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCaseWithPathFn(t, http.MethodPatch, tc)
		})
	}

	t.Cleanup(languagesCleanUp(t))
	t.Cleanup(userCleanUp(t))
}
