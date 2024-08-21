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
	"net/http"
	"strings"
	"testing"
)

func TestCreateSections(t *testing.T) {
	languagesCleanUp(t)()
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)

	generateFakeSectionData := func(t *testing.T) dtos.CreateSectionBody {
		title := faker.Name()
		description := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Aenean consequat nisl vel rutrum congue. Quisque mattis id massa id tincidunt. Nulla fringilla enim id dignissim dignissim. Morbi consequat, dui vel auctor pharetra, tortor tortor tristique nibh, et malesuada ante nisl egestas lectus. Donec vitae mollis enim, non aliquam tortor. Nulla."
		return dtos.CreateSectionBody{
			Title:       title,
			Description: description,
		}
	}

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

	testCases := []TestRequestCase[dtos.CreateSectionBody]{
		{
			Name: "Should return 201 CREATED when a section is created",
			ReqFn: func(t *testing.T) (dtos.CreateSectionBody, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return generateFakeSectionData(t), accessToken
			},
			ExpStatus: fiber.StatusCreated,
			AssertFn: func(t *testing.T, req dtos.CreateSectionBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.SectionResponse{})
				AssertEqual(t, req.Title, resBody.Title)
				AssertEqual(t, req.Description, resBody.Description)
				AssertEqual(
					t,
					"https://api.kiwiscript.com/api/v1/languages/rust/series/existing-series",
					resBody.Links.Series.Href,
				)
				AssertEqual(
					t,
					fmt.Sprintf(
						"https://api.kiwiscript.com/api/v1/languages/rust/series/existing-series/sections/%d",
						resBody.ID,
					),
					resBody.Links.Self.Href,
				)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series/sections",
		},
		{
			Name: "Should return 400 BAD REQUEST if the title is too long and description is empty",
			ReqFn: func(t *testing.T) (dtos.CreateSectionBody, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.CreateSectionBody{
					Title:       strings.Repeat(faker.Name(), 100),
					Description: "",
				}, accessToken
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req dtos.CreateSectionBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
				AssertEqual(t, resBody.Code, controllers.StatusValidation)
				AssertEqual(t, resBody.Message, "Invalid request")
				AssertEqual(t, len(resBody.Fields), 2)
				AssertEqual(t, resBody.Fields[0].Param, "title")
				AssertEqual(t, resBody.Fields[0].Message, controllers.StrFieldErrMessageMax)
				AssertEqual(t, resBody.Fields[1].Param, "description")
				AssertEqual(t, resBody.Fields[1].Message, controllers.FieldErrMessageRequired)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series/sections",
		},
		{
			Name: "Should return 403 FORBIDDEN if the user is not staff",
			ReqFn: func(t *testing.T) (dtos.CreateSectionBody, string) {
				testUser.IsStaff = false
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return generateFakeSectionData(t), accessToken
			},
			ExpStatus: fiber.StatusForbidden,
			AssertFn: func(t *testing.T, req dtos.CreateSectionBody, resp *http.Response) {
				AssertForbiddenResponse(t, resp)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series/sections",
		},
		{
			Name: "Should return 401 UNAUTHORIZED if the user is not authenticated",
			ReqFn: func(t *testing.T) (dtos.CreateSectionBody, string) {
				return generateFakeSectionData(t), ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, req dtos.CreateSectionBody, resp *http.Response) {
				AssertUnauthorizedResponse(t, resp)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series/sections",
		},
		{
			Name: "Should return 404 NOT FOUND if the series does not exist",
			ReqFn: func(t *testing.T) (dtos.CreateSectionBody, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return generateFakeSectionData(t), accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, req dtos.CreateSectionBody, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: baseLanguagesPath + "/rust/series/non-existing-series/sections",
		},
		{
			Name: "Should return 404 NOT FOUND if the language does not exist",
			ReqFn: func(t *testing.T) (dtos.CreateSectionBody, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return generateFakeSectionData(t), accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, req dtos.CreateSectionBody, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: baseLanguagesPath + "/python/series/existing-series/sections",
		},
		{
			Name: "Should return 409 CONFLICT if the section already exists",
			ReqFn: func(t *testing.T) (dtos.CreateSectionBody, string) {
				fakeData := generateFakeSectionData(t)
				testServices := GetTestServices(t)

				secOpts := services.CreateSectionOptions{
					UserID:       testUser.ID,
					Title:        fakeData.Title,
					LanguageSlug: "rust",
					SeriesSlug:   "existing-series",
					Description:  "Some description",
				}
				if _, serviceErr := testServices.CreateSection(context.Background(), secOpts); serviceErr != nil {
					t.Fatal("Failed to create section", "serviceError", serviceErr)
				}

				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return fakeData, accessToken
			},
			ExpStatus: fiber.StatusConflict,
			AssertFn: func(t *testing.T, req dtos.CreateSectionBody, resp *http.Response) {
				AssertConflictDuplicateKeyResponse(t, resp)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series/sections",
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

func TestGetSection(t *testing.T) {
	languagesCleanUp(t)()
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	var sectionID int32
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
		if err != nil {
			t.Fatal("Failed to create section", "serviceError", serviceErr)
		}
		sectionID = section.ID
	}()

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 200 OK when a non-published section is found and the user is staff",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.SectionResponse{})
				AssertNotEmpty(t, resBody.Title)
				AssertNotEmpty(t, resBody.Description)
				AssertEqual(t, resBody.TotalLessons, 0)
				AssertEqual(t, resBody.CompletedLessons, 0)
				AssertEqual(
					t,
					"https://api.kiwiscript.com/api/v1/languages/rust/series/existing-series",
					resBody.Links.Series.Href,
				)
				AssertEqual(
					t,
					fmt.Sprintf(
						"https://api.kiwiscript.com/api/v1/languages/rust/series/existing-series/sections/%d",
						resBody.ID,
					),
					resBody.Links.Self.Href,
				)
			},
			Path: fmt.Sprintf("%s/rust/series/existing-series/sections/%d", baseLanguagesPath, sectionID),
		},
		{
			Name: "Should return 200 OK when a published section is found and the user with progress is not staff",
			ReqFn: func(t *testing.T) (string, string) {
				testDb := GetTestDatabase(t)

				for i := 0; i < 3; i++ {
					incLCountPrms := db.IncrementSectionLessonsCountParams{
						ID:               sectionID,
						WatchTimeSeconds: 100 + int32(i),
						ReadTimeSeconds:  100 + int32(i),
					}
					if err := testDb.IncrementSectionLessonsCount(context.Background(), incLCountPrms); err != nil {
						t.Fatal("Failed to increment section lessons count", "error", err)
					}
				}

				isPubPrms := db.UpdateSectionIsPublishedParams{
					IsPublished: true,
					ID:          sectionID,
				}
				if _, err := testDb.UpdateSectionIsPublished(context.Background(), isPubPrms); err != nil {
					t.Fatal("Failed to update section is published", "error", err)
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

				if _, err := testDb.IncrementSectionProgressCompletedLessons(context.Background(), secProg.ID); err != nil {
					t.Fatal("Failed to increment series progress completed lessons", "error", err)
				}

				testUser.IsStaff = false
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.SectionResponse{})
				AssertNotEmpty(t, resBody.Title)
				AssertNotEmpty(t, resBody.Description)
				AssertEqual(t, resBody.TotalLessons, 3)
				AssertEqual(t, resBody.CompletedLessons, 1)
				AssertEqual(
					t,
					"https://api.kiwiscript.com/api/v1/languages/rust/series/existing-series",
					resBody.Links.Series.Href,
				)
				AssertEqual(
					t,
					fmt.Sprintf(
						"https://api.kiwiscript.com/api/v1/languages/rust/series/existing-series/sections/%d",
						resBody.ID,
					),
					resBody.Links.Self.Href,
				)
			},
			Path: fmt.Sprintf("%s/rust/series/existing-series/sections/%d", baseLanguagesPath, sectionID),
		},
		{
			Name: "Should return 404 NOT FOUND if the section does not exist",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series/sections/12345678910",
		},
		{
			Name: "Should return 404 NOT FOUND if the language does not exist",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf("%s/python/series/existing-series/sections/%d", baseLanguagesPath, sectionID),
		},
		{
			Name: "Should return 404 NOT FOUND if the series does not exist",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf("%s/python/series/non-existing-series/sections/%d", baseLanguagesPath, sectionID),
		},
		{
			Name: "Should return 404 NOT FOUND if the series is not published and the user is not staff",
			ReqFn: func(t *testing.T) (string, string) {
				testDb := GetTestDatabase(t)

				isPubPrms := db.UpdateSectionIsPublishedParams{
					IsPublished: false,
					ID:          sectionID,
				}
				if _, err := testDb.UpdateSectionIsPublished(context.Background(), isPubPrms); err != nil {
					t.Fatal("Failed to update section is published", "error", err)
				}

				testUser.IsStaff = false
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf("%s/rust/series/existing-series/sections/%d", baseLanguagesPath, sectionID),
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

func TestGetPaginatedSection(t *testing.T) {
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

		titleMap := map[string]bool{}
		for count := 0; count < 20; {
			fakeTitle := faker.Name()
			if _, ok := titleMap[fakeTitle]; !ok {
				secOpts := services.CreateSectionOptions{
					UserID:       testUser.ID,
					Title:        fakeTitle,
					LanguageSlug: "rust",
					SeriesSlug:   "existing-series",
					Description:  "Some description",
				}
				if _, serviceErr := testServices.CreateSection(context.Background(), secOpts); serviceErr != nil {
					t.Fatal("Failed to create section", "serviceError", serviceErr)
				}
				count++
			}
		}
	}()

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 200 OK with all sections if the user is staff",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.PaginatedResponse[dtos.SectionResponse]{})
				AssertEqual(t, resBody.Count, 20)
				AssertEqual(t, len(resBody.Results), 20)
				AssertEqual(t, resBody.Links.Next, nil)
				AssertEqual(t, resBody.Links.Prev, nil)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series/sections",
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
				resBody := AssertTestResponseBody(t, resp, dtos.PaginatedResponse[dtos.SectionResponse]{})
				AssertEqual(t, resBody.Count, 20)
				AssertEqual(t, len(resBody.Results), 5)
				AssertGreaterThan(t, resBody.Results[1].Position, resBody.Results[0].Position)
				AssertEqual(
					t,
					resBody.Links.Next.Href,
					"https://api.kiwiscript.com/api/v1/languages/rust/series/existing-series/sections?limit=5&offset=5",
				)
				AssertEqual(t, resBody.Links.Prev, nil)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series/sections?limit=5",
		},
		{
			Name: "Should return 200 OK with 3 rows if only 6 sections are published, offset is 3 and no user is logged in",
			ReqFn: func(t *testing.T) (string, string) {
				testDb := GetTestDatabase(t)
				sections, err := testDb.FindPaginatedSectionsBySlugs(context.Background(), db.FindPaginatedSectionsBySlugsParams{
					LanguageSlug: "rust",
					SeriesSlug:   "existing-series",
					Limit:        25,
					Offset:       0,
				})
				if err != nil {
					t.Fatal("Failed to find paginated sections", "error", err)
				}

				for i := 0; i < 6; i++ {
					params := db.UpdateSectionIsPublishedParams{
						IsPublished: true,
						ID:          sections[i].ID,
					}
					if _, err := testDb.UpdateSectionIsPublished(context.Background(), params); err != nil {
						t.Fatal("Failed to update section is published", "error", err)
					}
				}

				return "", ""
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.PaginatedResponse[dtos.SectionResponse]{})
				AssertEqual(t, resBody.Count, 6)
				AssertEqual(t, len(resBody.Results), 3)
				AssertEqual(t, resBody.Links.Next, nil)
				AssertEqual(
					t,
					resBody.Links.Prev.Href,
					"https://api.kiwiscript.com/api/v1/languages/rust/series/existing-series/sections?limit=25&offset=0",
				)
			},
			Path: baseLanguagesPath + "/rust/series/existing-series/sections?offset=3",
		},
		{
			Name: "Should return 200 OK with user progress",
			ReqFn: func(t *testing.T) (string, string) {
				testDb := GetTestDatabase(t)
				sections, err := testDb.FindPaginatedSectionsBySlugs(context.Background(), db.FindPaginatedSectionsBySlugsParams{
					LanguageSlug: "rust",
					SeriesSlug:   "existing-series",
					Limit:        25,
					Offset:       0,
				})
				if err != nil {
					t.Fatal("Failed to find paginated sections", "error", err)
				}

				for i := 0; i < 6; i++ {
					params := db.UpdateSectionIsPublishedParams{
						IsPublished: true,
						ID:          sections[i].ID,
					}
					if _, err := testDb.UpdateSectionIsPublished(context.Background(), params); err != nil {
						t.Fatal("Failed to update section is published", "error", err)
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
					LanguageSlug:       "rust",
					SeriesSlug:         "existing-series",
					LanguageProgressID: langProg.ID,
					UserID:             testUser.ID,
				})
				if err != nil {
					t.Fatal("Failed to create series progress", "error", err)
				}

				for i := 0; i < 3; i++ {
					prms := db.IncrementSectionLessonsCountParams{
						ID:               sections[0].ID,
						WatchTimeSeconds: 100 + int32(i*2),
						ReadTimeSeconds:  110 + int32(i*3),
					}
					if err := testDb.IncrementSectionLessonsCount(context.Background(), prms); err != nil {
						t.Fatal("Failed to increment section lesson count", "error", err)
					}
				}

				sectionProg, err := testDb.CreateSectionProgress(context.Background(), db.CreateSectionProgressParams{
					LanguageSlug:       "rust",
					SeriesSlug:         "existing-series",
					SectionID:          sections[0].ID,
					LanguageProgressID: langProg.ID,
					SeriesProgressID:   seriesProg.ID,
					UserID:             testUser.ID,
				})
				if err != nil {
					t.Fatal("Failed to create section progress", "error", err)
				}

				if _, err := testDb.IncrementSectionProgressCompletedLessons(context.Background(), sectionProg.ID); err != nil {
					t.Fatal("Failed to increment section progress completed lessons", "error", err)
				}
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.PaginatedResponse[dtos.SectionResponse]{})
				AssertEqual(t, resBody.Count, 6)
				AssertEqual(t, len(resBody.Results), 6)
				AssertEqual(t, resBody.Results[0].CompletedLessons, 1)
				AssertEqual(t, resBody.Results[0].TotalLessons, 3)
				AssertEqual(t, resBody.Links.Next, nil)
				AssertEqual(t, resBody.Links.Prev, nil)

			},
			Path: baseLanguagesPath + "/rust/series/existing-series/sections",
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
