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
	"github.com/google/uuid"
	"github.com/kiwiscript/kiwiscript_go/dtos"
	"github.com/kiwiscript/kiwiscript_go/exceptions"
	"github.com/kiwiscript/kiwiscript_go/paths"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/services"
	"github.com/kiwiscript/kiwiscript_go/utils"
	"net/http"
	"strings"
	"testing"
)

const (
	userPath = "/api" + paths.UsersPathV1
	mePath   = userPath + "/me"
)

func TestGetMe(t *testing.T) {
	userCleanUp(t)()
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 200 OK with only user data if the user is not staff",
			ReqFn: func(t *testing.T) (string, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: http.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.UserResponse{})
				AssertEqual(t, resBody.ID, testUser.ID)
				AssertEqual(t, resBody.FirstName, testUser.FirstName)
				AssertEqual(t, resBody.LastName, testUser.LastName)
				AssertEqual(t, resBody.Location, testUser.Location)
				AssertEqual(t, resBody.IsAdmin, testUser.IsAdmin)
				AssertEqual(t, resBody.IsStaff, testUser.IsStaff)
				AssertEqual(t, resBody.Embedded, nil)
				AssertStringContains(t, resBody.Links.Self.Href, fmt.Sprintf("%s/%d", userPath, testUser.ID))
			},
		},
		{
			Name: "Should return 200 OK with user and profile if the user is staff",
			ReqFn: func(t *testing.T) (string, string) {
				testServices := GetTestServices(t)
				testDb := GetTestDatabase(t)
				ctx := context.Background()
				requestID := uuid.NewString()

				profOpts := services.UserProfileOptions{
					RequestID: requestID,
					UserID:    testUser.ID,
					Bio:       "Lorem ipsum",
					GitHub:    "https://github.com/johndoe",
					LinkedIn:  "https://www.linkedin.com/in/john-doe",
					Website:   "https://johndoe.com",
				}
				if _, serviceErr := testServices.CreateUserProfile(ctx, profOpts); serviceErr != nil {
					t.Fatal("Failed to create user profile", "serviceError", serviceErr)
				}

				staffPrms := db.UpdateUserIsStaffParams{
					IsStaff: true,
					ID:      testUser.ID,
				}
				if err := testDb.UpdateUserIsStaff(ctx, staffPrms); err != nil {
					t.Fatal("Failed to update user is staff", "error", err)
				}

				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: http.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.UserResponse{})
				AssertEqual(t, resBody.ID, testUser.ID)
				AssertEqual(t, resBody.FirstName, testUser.FirstName)
				AssertEqual(t, resBody.LastName, testUser.LastName)
				AssertEqual(t, resBody.Location, testUser.Location)
				AssertStringContains(t, resBody.Links.Self.Href, fmt.Sprintf("%s/%d", userPath, testUser.ID))
				AssertStringContains(t,
					resBody.Links.Profile.Href,
					fmt.Sprintf("%s/%d/profile", userPath, testUser.ID),
				)
				AssertEqual(t, resBody.Links.Picture, nil)
				AssertEqual(t, resBody.Embedded.Profile.Bio, "Lorem ipsum")
				AssertEqual(t, resBody.Embedded.Profile.GitHub, "https://github.com/johndoe")
				AssertEqual(t, resBody.Embedded.Profile.LinkedIn, "https://www.linkedin.com/in/john-doe")
				AssertEqual(t, resBody.Embedded.Profile.Website, "https://johndoe.com")
				AssertEqual(t,
					resBody.Embedded.Profile.Links.Self.Href,
					fmt.Sprintf("https://api.kiwiscript.com%s/%d/profile", userPath, testUser.ID),
				)
				AssertEqual(t, resBody.Embedded.Picture, nil)
			},
		},
		{
			Name: "Should return 200 OK with user if the user is staff",
			ReqFn: func(t *testing.T) (string, string) {
				testUser = confirmTestUser(t, CreateTestUser(t, nil).ID)
				testDb := GetTestDatabase(t)
				ctx := context.Background()

				staffPrms := db.UpdateUserIsStaffParams{
					IsStaff: true,
					ID:      testUser.ID,
				}
				if err := testDb.UpdateUserIsStaff(ctx, staffPrms); err != nil {
					t.Fatal("Failed to update user is staff", "error", err)
				}

				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: http.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.UserResponse{})
				AssertEqual(t, resBody.ID, testUser.ID)
				AssertEqual(t, resBody.FirstName, testUser.FirstName)
				AssertEqual(t, resBody.LastName, testUser.LastName)
				AssertEqual(t, resBody.Location, testUser.Location)
				AssertEqual(t, resBody.IsAdmin, testUser.IsAdmin)
				AssertEqual(t, resBody.IsStaff, testUser.IsStaff)
				AssertEqual(t, resBody.Embedded, nil)
				AssertStringContains(t, resBody.Links.Self.Href, fmt.Sprintf("%s/%d", userPath, testUser.ID))
			},
		},
		{
			Name: "Should return 200 OK with user and picture if the user is staff",
			ReqFn: func(t *testing.T) (string, string) {
				testUser = confirmTestUser(t, CreateTestUser(t, nil).ID)
				testServices := GetTestServices(t)
				testDb := GetTestDatabase(t)
				ctx := context.Background()
				requestID := uuid.NewString()

				picOpts := services.UploadUserPictureOptions{
					RequestID:  requestID,
					UserID:     testUser.ID,
					FileHeader: ImageUploadMock(t),
				}
				if _, serviceErr := testServices.UploadUserPicture(ctx, picOpts); serviceErr != nil {
					t.Fatal("Failed to upload user picture", "serviceError", serviceErr)
				}

				staffPrms := db.UpdateUserIsStaffParams{
					IsStaff: true,
					ID:      testUser.ID,
				}
				if err := testDb.UpdateUserIsStaff(ctx, staffPrms); err != nil {
					t.Fatal("Failed to update user is staff", "error", err)
				}

				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: http.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.UserResponse{})
				AssertEqual(t, resBody.ID, testUser.ID)
				AssertEqual(t, resBody.FirstName, testUser.FirstName)
				AssertEqual(t, resBody.LastName, testUser.LastName)
				AssertEqual(t, resBody.Location, testUser.Location)
				AssertStringContains(t, resBody.Links.Self.Href, fmt.Sprintf("%s/%d", userPath, testUser.ID))
				AssertStringContains(t,
					resBody.Links.Picture.Href,
					fmt.Sprintf("%s/%d/picture", userPath, testUser.ID),
				)
				AssertEqual(t, resBody.Links.Profile, nil)
				AssertNotEmpty(t, resBody.Embedded.Picture.EXT)
				AssertNotEmpty(t, resBody.Embedded.Picture.URL)
				AssertEqual(t,
					resBody.Embedded.Picture.Links.Self.Href,
					fmt.Sprintf("https://api.kiwiscript.com%s/%d/picture", userPath, testUser.ID),
				)
				AssertEqual(t, resBody.Embedded.Profile, nil)
			},
		},
		{
			Name: "Should return 200 OK with user and profile if the user is staff",
			ReqFn: func(t *testing.T) (string, string) {
				testUser = confirmTestUser(t, CreateTestUser(t, nil).ID)
				testServices := GetTestServices(t)
				testDb := GetTestDatabase(t)
				ctx := context.Background()
				requestID := uuid.NewString()

				profOpts := services.UserProfileOptions{
					RequestID: requestID,
					UserID:    testUser.ID,
					Bio:       "Lorem ipsum",
					GitHub:    "https://github.com/johndoe",
					LinkedIn:  "https://www.linkedin.com/in/john-doe",
					Website:   "https://johndoe.com",
				}
				if _, serviceErr := testServices.CreateUserProfile(ctx, profOpts); serviceErr != nil {
					t.Fatal("Failed to create user profile", "serviceError", serviceErr)
				}

				picOpts := services.UploadUserPictureOptions{
					RequestID:  requestID,
					UserID:     testUser.ID,
					FileHeader: ImageUploadMock(t),
				}
				if _, serviceErr := testServices.UploadUserPicture(ctx, picOpts); serviceErr != nil {
					t.Fatal("Failed to upload user picture", "serviceError", serviceErr)
				}

				staffPrms := db.UpdateUserIsStaffParams{
					IsStaff: true,
					ID:      testUser.ID,
				}
				if err := testDb.UpdateUserIsStaff(ctx, staffPrms); err != nil {
					t.Fatal("Failed to update user is staff", "error", err)
				}

				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: http.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.UserResponse{})
				AssertEqual(t, resBody.ID, testUser.ID)
				AssertEqual(t, resBody.FirstName, testUser.FirstName)
				AssertEqual(t, resBody.LastName, testUser.LastName)
				AssertEqual(t, resBody.Location, testUser.Location)
				AssertStringContains(t, resBody.Links.Self.Href, fmt.Sprintf("%s/%d", userPath, testUser.ID))
				AssertStringContains(t,
					resBody.Links.Profile.Href,
					fmt.Sprintf("%s/%d/profile", userPath, testUser.ID),
				)
				AssertStringContains(t,
					resBody.Links.Picture.Href,
					fmt.Sprintf("%s/%d/picture", userPath, testUser.ID),
				)
				AssertEqual(t, resBody.Embedded.Profile.Bio, "Lorem ipsum")
				AssertEqual(t, resBody.Embedded.Profile.GitHub, "https://github.com/johndoe")
				AssertEqual(t, resBody.Embedded.Profile.LinkedIn, "https://www.linkedin.com/in/john-doe")
				AssertEqual(t, resBody.Embedded.Profile.Website, "https://johndoe.com")
				AssertEqual(t,
					resBody.Embedded.Profile.Links.Self.Href,
					fmt.Sprintf("https://api.kiwiscript.com%s/%d/profile", userPath, testUser.ID),
				)
				AssertNotEmpty(t, resBody.Embedded.Picture.EXT)
				AssertNotEmpty(t, resBody.Embedded.Picture.URL)
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if no user is logged in",
			ReqFn: func(t *testing.T) (string, string) {
				return "", ""
			},
			ExpStatus: http.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				assertUnauthorizeError(t, resp)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodGet, mePath, tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}

func TestUpdateMe(t *testing.T) {
	userCleanUp(t)()
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)

	testCases := []TestRequestCase[dtos.UpdateUserBody]{
		{
			Name: "Should return 200 OK if it updated user successfully",
			ReqFn: func(t *testing.T) (dtos.UpdateUserBody, string) {
				userData := GenerateFakeUserData(t)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.UpdateUserBody{
					FirstName: userData.FirstName,
					LastName:  userData.LastName,
					Location:  userData.Location,
				}, accessToken
			},
			ExpStatus: http.StatusOK,
			AssertFn: func(t *testing.T, req dtos.UpdateUserBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.UserResponse{})
				AssertEqual(t, resBody.ID, testUser.ID)
				AssertEqual(t, resBody.FirstName, utils.Capitalized(req.FirstName))
				AssertEqual(t, resBody.LastName, utils.Capitalized(req.LastName))
				AssertEqual(t, resBody.Location, utils.Uppercased(req.Location))
			},
		},
		{
			Name: "Should return 400 BAD REQUEST if first name is too long, last name is empy and location is too short",
			ReqFn: func(t *testing.T) (dtos.UpdateUserBody, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.UpdateUserBody{
					FirstName: strings.Repeat("Some Name ", 50),
					LastName:  "",
					Location:  "B",
				}, accessToken
			},
			ExpStatus: http.StatusBadRequest,
			AssertFn: func(t *testing.T, _ dtos.UpdateUserBody, resp *http.Response) {
				AssertValidationErrorResponse(t, resp, []ValidationErrorAssertion{
					{Param: "firstName", Message: exceptions.StrFieldErrMessageMax},
					{Param: "lastName", Message: exceptions.FieldErrMessageRequired},
					{Param: "location", Message: exceptions.StrFieldErrMessageMin},
				})
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if the user is not logged in",
			ReqFn: func(t *testing.T) (dtos.UpdateUserBody, string) {
				userData := GenerateFakeUserData(t)
				return dtos.UpdateUserBody{
					FirstName: userData.FirstName,
					LastName:  userData.LastName,
					Location:  userData.Location,
				}, ""
			},
			ExpStatus: http.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ dtos.UpdateUserBody, resp *http.Response) {
				AssertUnauthorizedResponse(t, resp)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodPut, mePath, tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}

func TestDeleteMe(t *testing.T) {
	userCleanUp(t)()

	var testUser *db.User
	var password string
	beforeEach := func(t *testing.T) {
		userData := GenerateFakeUserData(t)
		password = userData.Password
		testUser = confirmTestUser(t, CreateTestUser(t, &userData).ID)
	}
	afterEach := func(t *testing.T) {
		userCleanUp(t)()
	}

	testCases := []TestRequestCase[dtos.DeleteUserBody]{
		{
			Name: "Should return 204 NO CONTENT when the user is not staff",
			ReqFn: func(t *testing.T) (dtos.DeleteUserBody, string) {
				beforeEach(t)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.DeleteUserBody{Password: password}, accessToken
			},
			ExpStatus: http.StatusNoContent,
			AssertFn: func(t *testing.T, _ dtos.DeleteUserBody, _ *http.Response) {
				afterEach(t)
			},
		},
		{
			Name: "Should return 400 BAD REQUEST if the password is invalid",
			ReqFn: func(t *testing.T) (dtos.DeleteUserBody, string) {
				beforeEach(t)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.DeleteUserBody{Password: "Wrong Password"}, accessToken
			},
			ExpStatus: http.StatusBadRequest,
			AssertFn: func(t *testing.T, _ dtos.DeleteUserBody, resp *http.Response) {
				AssertValidationErrorWithoutFieldsResponse(t, resp, "'password' does not match")
				afterEach(t)
			},
		},
		{
			Name: "Should return 409 FORBIDDEN if the user is staff",
			ReqFn: func(t *testing.T) (dtos.DeleteUserBody, string) {
				beforeEach(t)
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return dtos.DeleteUserBody{Password: "Wrong Password"}, accessToken
			},
			ExpStatus: http.StatusForbidden,
			AssertFn: func(t *testing.T, _ dtos.DeleteUserBody, resp *http.Response) {
				AssertForbiddenResponse(t, resp)
				afterEach(t)
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if the user is not logged in",
			ReqFn: func(t *testing.T) (dtos.DeleteUserBody, string) {
				beforeEach(t)
				return dtos.DeleteUserBody{Password: password}, ""
			},
			ExpStatus: http.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ dtos.DeleteUserBody, resp *http.Response) {
				AssertUnauthorizedResponse(t, resp)
				afterEach(t)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodDelete, mePath, tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}
