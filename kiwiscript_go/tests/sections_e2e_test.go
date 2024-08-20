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
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodPost, tc.Path, tc)
		})
	}

	t.Cleanup(languagesCleanUp(t))
	t.Cleanup(userCleanUp(t))
}
