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
	"github.com/kiwiscript/kiwiscript_go/paths"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"net/http"
	"strings"
	"testing"
)

const baseCertificatesPath = "/api" + paths.CertificatesV1

func TestGetCertificates(t *testing.T) {
	languagesCleanUp(t)()
	staffUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	staffUser.IsStaff = true
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)

	func() {
		testUser2 := confirmTestUser(t, CreateTestUser(t, nil).ID)
		testDb := GetTestDatabase(t)
		ctx := context.Background()

		langParams := db.CreateLanguageParams{
			Name:     "Rust",
			Icon:     strings.TrimSpace(languageIcons["Rust"]),
			AuthorID: staffUser.ID,
			Slug:     "rust",
		}
		if _, err := testDb.CreateLanguage(ctx, langParams); err != nil {
			t.Fatal("Failed to create language", err)
		}

		for i := 0; i < 5; i++ {
			title := fmt.Sprintf("Rust series %d", i)
			slug := fmt.Sprintf("rust-series-%d", i)
			serParams := db.CreateSeriesParams{
				Title:        title,
				Slug:         slug,
				Description:  "Some description",
				LanguageSlug: "rust",
				AuthorID:     staffUser.ID,
			}
			if _, err := testDb.CreateSeries(ctx, serParams); err != nil {
				t.Fatal("Failed to create series", err)
			}

			certParams := db.CreateCertificateParams{
				ID:               uuid.New(),
				UserID:           testUser.ID,
				LanguageSlug:     "rust",
				SeriesSlug:       slug,
				SeriesTitle:      title,
				Lessons:          5,
				WatchTimeSeconds: 100,
				ReadTimeSeconds:  1000,
			}
			if _, err := testDb.CreateCertificate(ctx, certParams); err != nil {
				t.Fatal("Failed to create certificate", err)
			}

			certParams2 := db.CreateCertificateParams{
				ID:               uuid.New(),
				UserID:           testUser2.ID,
				LanguageSlug:     "rust",
				SeriesSlug:       slug,
				SeriesTitle:      title,
				Lessons:          5,
				WatchTimeSeconds: 100,
				ReadTimeSeconds:  1000,
			}
			if _, err := testDb.CreateCertificate(ctx, certParams2); err != nil {
				t.Fatal("Failed to create certificate", err)
			}
		}
	}()

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 200 OK with the current user certificates",
			ReqFn: func(t *testing.T) (string, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.PaginatedResponse[dtos.CertificateResponse]{})
				AssertEqual(t, resBody.Count, 5)
				AssertEqual(t, len(resBody.Results), 5)
				AssertEqual(t, resBody.Links.Next, nil)
				AssertEqual(t, resBody.Links.Prev, nil)
			},
			Path: baseCertificatesPath,
		},
		{
			Name: "Should return 200 OK with the current user certificates with limit and offset",
			ReqFn: func(t *testing.T) (string, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.PaginatedResponse[dtos.CertificateResponse]{})
				AssertEqual(t, resBody.Count, 5)
				AssertEqual(t, len(resBody.Results), 2)
				AssertStringContains(t, resBody.Links.Next.Href, baseCertificatesPath+"?limit=2&offset=3")
				AssertStringContains(t, resBody.Links.Prev.Href, baseCertificatesPath+"?limit=2&offset=0")
			},
			Path: baseCertificatesPath + "?limit=2&offset=1",
		},
		{
			Name: "Should return 401 UNAUTHORIZED if the user is not signed in",
			ReqFn: func(t *testing.T) (string, string) {
				return "", ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				assertUnauthorizeError(t, resp)
			},
			Path: baseCertificatesPath,
		},
		{
			Name: "Should return 403 FORBIDDEN if the user is staff",
			ReqFn: func(t *testing.T) (string, string) {
				accessToken, _ := GenerateTestAuthTokens(t, staffUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusForbidden,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertForbiddenResponse(t, resp)
			},
			Path: baseCertificatesPath,
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

func TestGetCertificate(t *testing.T) {
	languagesCleanUp(t)()
	staffUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	staffUser.IsStaff = true
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)

	certificateID := uuid.New()
	func() {
		testDb := GetTestDatabase(t)
		ctx := context.Background()

		langParams := db.CreateLanguageParams{
			Name:     "Rust",
			Icon:     strings.TrimSpace(languageIcons["Rust"]),
			AuthorID: staffUser.ID,
			Slug:     "rust",
		}
		if _, err := testDb.CreateLanguage(ctx, langParams); err != nil {
			t.Fatal("Failed to create language", err)
		}

		title := "Rust series"
		slug := "rust-series"
		serParams := db.CreateSeriesParams{
			Title:        title,
			Slug:         slug,
			Description:  "Some description",
			LanguageSlug: "rust",
			AuthorID:     staffUser.ID,
		}
		if _, err := testDb.CreateSeries(ctx, serParams); err != nil {
			t.Fatal("Failed to create series", err)
		}

		certParams := db.CreateCertificateParams{
			ID:               certificateID,
			UserID:           testUser.ID,
			LanguageSlug:     "rust",
			SeriesSlug:       slug,
			SeriesTitle:      title,
			Lessons:          5,
			WatchTimeSeconds: 100,
			ReadTimeSeconds:  1000,
		}
		if _, err := testDb.CreateCertificate(ctx, certParams); err != nil {
			t.Fatal("Failed to create certificate", err)
		}
	}()

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 200 OK with the certificate when user is not authenticated",
			ReqFn: func(t *testing.T) (string, string) {
				return "", ""
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.CertificateResponse{})
				AssertEqual(t, resBody.ID, certificateID)
				AssertEqual(t, resBody.FirstName, testUser.FirstName)
				AssertEqual(t, resBody.LastName, testUser.LastName)
				AssertEqual(t, resBody.SeriesTitle, "Rust series")
				AssertEqual(t, resBody.Lessons, int16(5))
				AssertEqual(t, resBody.WatchTime, int32(100))
				AssertEqual(t, resBody.ReadTime, int32(1000))
				AssertStringContains(t, resBody.Links.Language.Href, baseLanguagesPath+"/rust")
				AssertStringContains(t, resBody.Links.Series.Href, baseLanguagesPath+"/rust"+paths.SeriesPath+"/rust-series")
				AssertStringContains(t, resBody.Links.Self.Href, baseCertificatesPath+"/"+certificateID.String())
			},
			Path: baseCertificatesPath + "/" + certificateID.String(),
		},
		{
			Name: "Should return 200 OK with the certificate when user is authenticated",
			ReqFn: func(t *testing.T) (string, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.CertificateResponse{})
				AssertEqual(t, resBody.ID, certificateID)
				AssertEqual(t, resBody.FirstName, testUser.FirstName)
				AssertEqual(t, resBody.LastName, testUser.LastName)
				AssertEqual(t, resBody.SeriesTitle, "Rust series")
				AssertEqual(t, resBody.Lessons, int16(5))
				AssertEqual(t, resBody.WatchTime, int32(100))
				AssertEqual(t, resBody.ReadTime, int32(1000))
				AssertStringContains(t, resBody.Links.Language.Href, baseLanguagesPath+"/rust")
				AssertStringContains(t, resBody.Links.Series.Href, baseLanguagesPath+"/rust"+paths.SeriesPath+"/rust-series")
				AssertStringContains(t, resBody.Links.Self.Href, baseCertificatesPath+"/"+certificateID.String())
			},
			Path: baseCertificatesPath + "/" + certificateID.String(),
		},
		{
			Name: "Should return 200 OK with the certificate when user is authenticated and Staff",
			ReqFn: func(t *testing.T) (string, string) {
				accessToken, _ := GenerateTestAuthTokens(t, staffUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.CertificateResponse{})
				AssertEqual(t, resBody.ID, certificateID)
				AssertEqual(t, resBody.FirstName, testUser.FirstName)
				AssertEqual(t, resBody.LastName, testUser.LastName)
				AssertEqual(t, resBody.SeriesTitle, "Rust series")
				AssertEqual(t, resBody.Lessons, int16(5))
				AssertEqual(t, resBody.WatchTime, int32(100))
				AssertEqual(t, resBody.ReadTime, int32(1000))
				AssertStringContains(t, resBody.Links.Language.Href, baseLanguagesPath+"/rust")
				AssertStringContains(t, resBody.Links.Series.Href, baseLanguagesPath+"/rust"+paths.SeriesPath+"/rust-series")
				AssertStringContains(t, resBody.Links.Self.Href, baseCertificatesPath+"/"+certificateID.String())
			},
			Path: baseCertificatesPath + "/" + certificateID.String(),
		},
		{
			Name: "Should return 404 NOT FOUND when the certificate is not found",
			ReqFn: func(t *testing.T) (string, string) {
				return "", ""
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: baseCertificatesPath + "/" + uuid.NewString(),
		},
		{
			Name: "Should return 400 BAD REQUEST when the certificate ID is not an UUID",
			ReqFn: func(t *testing.T) (string, string) {
				return "", ""
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertValidationErrorResponse(t, resp, []ValidationErrorAssertion{
					{
						Param:   "certificateID",
						Message: exceptions.StrFieldErrMessageUUID,
					},
				})
			},
			Path: baseCertificatesPath + "/" + "123-456-789",
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
