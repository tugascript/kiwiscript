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
	"github.com/kiwiscript/kiwiscript_go/services"
	"github.com/kiwiscript/kiwiscript_go/utils"
	"net/http"
	"strings"
	"testing"
)

const (
	userPath      = "/api" + paths.UsersPathV1
	mePath        = userPath + "/me"
	myProfilePath = mePath + paths.ProfilePath
	myPicturePath = mePath + paths.PicturePath
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
			Name: "Should return 403 FORBIDDEN if the user is staff",
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

func TestCreateMyProfile(t *testing.T) {
	userCleanUp(t)()
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)

	afterEach := func(t *testing.T) {
		testServices := GetTestServices(t)
		ctx := context.Background()
		requestID := uuid.NewString()

		delOpts := services.DeleteUserProfileOptions{
			RequestID: requestID,
			UserID:    testUser.ID,
		}
		if serviceErr := testServices.DeleteUserProfile(ctx, delOpts); serviceErr != nil {
			t.Log("Failed to delete user profile", "serviceErr", serviceErr)
		}
	}

	testCases := []TestRequestCase[dtos.UserProfileBody]{
		{
			Name: "Should return 201 CREATED when the user is staff and has no profile",
			ReqFn: func(t *testing.T) (dtos.UserProfileBody, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				profileBody := dtos.UserProfileBody{
					Bio:      "Lorem ipsum dolor sit amet",
					GitHub:   "https://github.com/johndoe",
					LinkedIn: "https://www.linkedin.com/in/john-doe",
					Website:  "https://johndoe.com",
				}
				return profileBody, accessToken
			},
			ExpStatus: http.StatusCreated,
			AssertFn: func(t *testing.T, req dtos.UserProfileBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.UserProfileResponse{})
				AssertGreaterThan(t, resBody.ID, 0)
				AssertEqual(t, resBody.Bio, req.Bio)
				AssertEqual(t, resBody.Website, req.Website)
				AssertEqual(t, resBody.LinkedIn, req.LinkedIn)
				AssertEqual(t, resBody.GitHub, req.GitHub)
				AssertStringContains(t,
					resBody.Links.Self.Href,
					fmt.Sprintf("/api%s/%d%s", paths.UsersPathV1, testUser.ID, paths.ProfilePath),
				)
				AssertStringContains(t,
					resBody.Links.User.Href,
					fmt.Sprintf("/api%s/%d", paths.UsersPathV1, testUser.ID),
				)
				afterEach(t)
			},
		},
		{
			Name: "Should return 409 CONFLICT if profile already exists",
			ReqFn: func(t *testing.T) (dtos.UserProfileBody, string) {
				testServices := GetTestServices(t)
				ctx := context.Background()
				requestID := uuid.NewString()
				uProfOpts := services.UserProfileOptions{
					RequestID: requestID,
					UserID:    testUser.ID,
					Bio:       "Lorem ipsum dolor sit amet",
					GitHub:    "https://github.com/johndoe",
					LinkedIn:  "https://www.linkedin.com/in/john-doe",
					Website:   "https://johndoe.com",
				}
				if _, serviceErr := testServices.CreateUserProfile(ctx, uProfOpts); serviceErr != nil {
					t.Fatal("Failed to create user profile", "serviceError", serviceErr)
				}

				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				profileBody := dtos.UserProfileBody{
					Bio:      "Lorem ipsum dolor sit amet",
					GitHub:   "https://github.com/johndoe",
					LinkedIn: "https://www.linkedin.com/in/john-doe",
					Website:  "https://johndoe.com",
				}
				return profileBody, accessToken
			},
			ExpStatus: http.StatusConflict,
			AssertFn: func(t *testing.T, _ dtos.UserProfileBody, resp *http.Response) {
				AssertConflictResponse(t, resp, "User profile already exists")
				afterEach(t)
			},
		},
		{
			Name: "Should return 403 FORBIDDEN if the user is not staff",
			ReqFn: func(t *testing.T) (dtos.UserProfileBody, string) {
				testUser.IsStaff = false
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				profileBody := dtos.UserProfileBody{
					Bio:      "Lorem ipsum dolor sit amet",
					GitHub:   "https://github.com/johndoe",
					LinkedIn: "https://www.linkedin.com/in/john-doe",
					Website:  "https://johndoe.com",
				}
				return profileBody, accessToken
			},
			ExpStatus: http.StatusForbidden,
			AssertFn: func(t *testing.T, _ dtos.UserProfileBody, resp *http.Response) {
				AssertForbiddenResponse(t, resp)
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if the user is not logged in",
			ReqFn: func(t *testing.T) (dtos.UserProfileBody, string) {
				profileBody := dtos.UserProfileBody{
					Bio:      "Lorem ipsum dolor sit amet",
					GitHub:   "https://github.com/johndoe",
					LinkedIn: "https://www.linkedin.com/in/john-doe",
					Website:  "https://johndoe.com",
				}
				return profileBody, ""
			},
			ExpStatus: http.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ dtos.UserProfileBody, resp *http.Response) {
				AssertUnauthorizedResponse(t, resp)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodPost, myProfilePath, tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}

func TestGetMyProfile(t *testing.T) {
	userCleanUp(t)()
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)

	afterEach := func(t *testing.T) {
		testServices := GetTestServices(t)
		ctx := context.Background()
		requestID := uuid.NewString()

		delOpts := services.DeleteUserProfileOptions{
			RequestID: requestID,
			UserID:    testUser.ID,
		}
		if serviceErr := testServices.DeleteUserProfile(ctx, delOpts); serviceErr != nil {
			t.Log("Failed to delete user profile", "serviceErr", serviceErr)
		}
	}

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 200 OK when profile is found and user is staff",
			ReqFn: func(t *testing.T) (string, string) {
				testServices := GetTestServices(t)
				ctx := context.Background()
				requestID := uuid.NewString()

				uProfOpts := services.UserProfileOptions{
					RequestID: requestID,
					UserID:    testUser.ID,
					Bio:       "Lorem ipsum dolor sit amet",
					GitHub:    "https://github.com/johndoe",
					LinkedIn:  "https://www.linkedin.com/in/john-doe",
					Website:   "https://johndoe.com",
				}
				if _, serviceErr := testServices.CreateUserProfile(ctx, uProfOpts); serviceErr != nil {
					t.Fatal("Failed to create user profile", "serviceError", serviceErr)
				}

				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: http.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.UserProfileResponse{})
				AssertGreaterThan(t, resBody.ID, 0)
				AssertEqual(t, resBody.Bio, "Lorem ipsum dolor sit amet")
				AssertEqual(t, resBody.Website, "https://johndoe.com")
				AssertEqual(t, resBody.LinkedIn, "https://www.linkedin.com/in/john-doe")
				AssertEqual(t, resBody.GitHub, "https://github.com/johndoe")
				AssertStringContains(t,
					resBody.Links.Self.Href,
					fmt.Sprintf("/api%s/%d%s", paths.UsersPathV1, testUser.ID, paths.ProfilePath),
				)
				AssertStringContains(t,
					resBody.Links.User.Href,
					fmt.Sprintf("/api%s/%d", paths.UsersPathV1, testUser.ID),
				)
				afterEach(t)
			},
		},
		{
			Name: "Should return 404 NOT FOUND if the staff user has not profile",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: http.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
				afterEach(t)
			},
		},
		{
			Name: "Should return 403 FORBIDDEN if the user is not staff",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = false
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: http.StatusForbidden,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertForbiddenResponse(t, resp)
				afterEach(t)
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if the user is the user is not logged in",
			ReqFn: func(t *testing.T) (string, string) {
				return "", ""
			},
			ExpStatus: http.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertUnauthorizedResponse(t, resp)
				afterEach(t)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodGet, myProfilePath, tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}

func TestUpdateMyProfile(t *testing.T) {
	userCleanUp(t)()
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)

	afterEach := func(t *testing.T) {
		testServices := GetTestServices(t)
		ctx := context.Background()
		requestID := uuid.NewString()

		delOpts := services.DeleteUserProfileOptions{
			RequestID: requestID,
			UserID:    testUser.ID,
		}
		if serviceErr := testServices.DeleteUserProfile(ctx, delOpts); serviceErr != nil {
			t.Log("Failed to delete user profile", "serviceErr", serviceErr)
		}
	}

	testCases := []TestRequestCase[dtos.UserProfileBody]{
		{
			Name: "Should return 200 OK when profile is updated and user is staff",
			ReqFn: func(t *testing.T) (dtos.UserProfileBody, string) {
				testServices := GetTestServices(t)
				ctx := context.Background()
				requestID := uuid.NewString()

				uProfOpts := services.UserProfileOptions{
					RequestID: requestID,
					UserID:    testUser.ID,
					Bio:       "Lorem ipsum dolor sit amet",
					GitHub:    "https://github.com/johndoe",
					LinkedIn:  "https://www.linkedin.com/in/john-doe",
					Website:   "https://johndoe.com",
				}
				if _, serviceErr := testServices.CreateUserProfile(ctx, uProfOpts); serviceErr != nil {
					t.Fatal("Failed to create user profile", "serviceError", serviceErr)
				}

				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				body := dtos.UserProfileBody{
					Bio:      "consectetur adipiscing elit, sed do eiusmod",
					GitHub:   "https://github.com/johndoe2",
					LinkedIn: "https://www.linkedin.com/in/john-doe-2",
					Website:  "https://johndoe2.com",
				}
				return body, accessToken
			},
			ExpStatus: http.StatusOK,
			AssertFn: func(t *testing.T, req dtos.UserProfileBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.UserProfileResponse{})
				AssertGreaterThan(t, resBody.ID, 0)
				AssertEqual(t, resBody.Bio, req.Bio)
				AssertEqual(t, resBody.Website, req.Website)
				AssertEqual(t, resBody.LinkedIn, req.LinkedIn)
				AssertEqual(t, resBody.GitHub, req.GitHub)
				AssertStringContains(t,
					resBody.Links.Self.Href,
					fmt.Sprintf("/api%s/%d%s", paths.UsersPathV1, testUser.ID, paths.ProfilePath),
				)
				AssertStringContains(t,
					resBody.Links.User.Href,
					fmt.Sprintf("/api%s/%d", paths.UsersPathV1, testUser.ID),
				)
				afterEach(t)
			},
		},
		{
			Name: "Should return 400 BAD REQUEST when profile is updated and user is staff",
			ReqFn: func(t *testing.T) (dtos.UserProfileBody, string) {
				testServices := GetTestServices(t)
				ctx := context.Background()
				requestID := uuid.NewString()

				uProfOpts := services.UserProfileOptions{
					RequestID: requestID,
					UserID:    testUser.ID,
					Bio:       "Lorem ipsum dolor sit amet",
					GitHub:    "https://github.com/johndoe",
					LinkedIn:  "https://www.linkedin.com/in/john-doe",
					Website:   "https://johndoe.com",
				}
				if _, serviceErr := testServices.CreateUserProfile(ctx, uProfOpts); serviceErr != nil {
					t.Fatal("Failed to create user profile", "serviceError", serviceErr)
				}

				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				body := dtos.UserProfileBody{
					Bio:      strings.Repeat("consectetur adipiscing elit, sed do eiusmod", 100),
					GitHub:   "github-com",
					LinkedIn: "linkedin-com",
					Website:  "johndoe2-com",
				}
				return body, accessToken
			},
			ExpStatus: http.StatusBadRequest,
			AssertFn: func(t *testing.T, req dtos.UserProfileBody, resp *http.Response) {
				AssertValidationErrorResponse(t, resp, []ValidationErrorAssertion{
					{Param: "bio", Message: exceptions.StrFieldErrMessageMax},
					{Param: "gitHub", Message: exceptions.StrFieldErrMessageUrl},
					{Param: "linkedIn", Message: exceptions.StrFieldErrMessageUrl},
					{Param: "website", Message: exceptions.StrFieldErrMessageUrl},
				})
				afterEach(t)
			},
		},
		{
			Name: "Should return 404 NOT FOUND if the staff user has not profile",
			ReqFn: func(t *testing.T) (dtos.UserProfileBody, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				body := dtos.UserProfileBody{
					Bio:      "consectetur adipiscing elit, sed do eiusmod",
					GitHub:   "https://github.com/johndoe2",
					LinkedIn: "https://www.linkedin.com/in/john-doe-2",
					Website:  "https://johndoe2.com",
				}
				return body, accessToken
			},
			ExpStatus: http.StatusNotFound,
			AssertFn: func(t *testing.T, _ dtos.UserProfileBody, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
				afterEach(t)
			},
		},
		{
			Name: "Should return 403 FORBIDDEN if the user is not staff",
			ReqFn: func(t *testing.T) (dtos.UserProfileBody, string) {
				testUser.IsStaff = false
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				body := dtos.UserProfileBody{
					Bio:      "consectetur adipiscing elit, sed do eiusmod",
					GitHub:   "https://github.com/johndoe2",
					LinkedIn: "https://www.linkedin.com/in/john-doe-2",
					Website:  "https://johndoe2.com",
				}
				return body, accessToken
			},
			ExpStatus: http.StatusForbidden,
			AssertFn: func(t *testing.T, _ dtos.UserProfileBody, resp *http.Response) {
				AssertForbiddenResponse(t, resp)
				afterEach(t)
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if the user is the user is not logged in",
			ReqFn: func(t *testing.T) (dtos.UserProfileBody, string) {
				body := dtos.UserProfileBody{
					Bio:      "consectetur adipiscing elit, sed do eiusmod",
					GitHub:   "https://github.com/johndoe2",
					LinkedIn: "https://www.linkedin.com/in/john-doe-2",
					Website:  "https://johndoe2.com",
				}
				return body, ""
			},
			ExpStatus: http.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ dtos.UserProfileBody, resp *http.Response) {
				AssertUnauthorizedResponse(t, resp)
				afterEach(t)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodPut, myProfilePath, tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}

func TestDeleteMyProfile(t *testing.T) {
	userCleanUp(t)()
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)

	beforeEach := func(t *testing.T) {
		testServices := GetTestServices(t)
		ctx := context.Background()
		requestID := uuid.NewString()

		uProfOpts := services.UserProfileOptions{
			RequestID: requestID,
			UserID:    testUser.ID,
			Bio:       "Lorem ipsum dolor sit amet",
			GitHub:    "https://github.com/johndoe",
			LinkedIn:  "https://www.linkedin.com/in/john-doe",
			Website:   "https://johndoe.com",
		}
		if _, serviceErr := testServices.CreateUserProfile(ctx, uProfOpts); serviceErr != nil {
			t.Fatal("Failed to create user profile", "serviceError", serviceErr)
		}
	}

	afterEach := func(t *testing.T) {
		testServices := GetTestServices(t)
		ctx := context.Background()
		requestID := uuid.NewString()

		delOpts := services.DeleteUserProfileOptions{
			RequestID: requestID,
			UserID:    testUser.ID,
		}
		if serviceErr := testServices.DeleteUserProfile(ctx, delOpts); serviceErr != nil {
			t.Log("Failed to delete user profile", "serviceErr", serviceErr)
		}
	}

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 204 NO CONTENT when the staff user profile is deleted",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: http.StatusNoContent,
			AssertFn: func(t *testing.T, _ string, _ *http.Response) {
				afterEach(t)
			},
		},
		{
			Name: "Should return 404 NOT FOUND when the staff user profile is not found",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: http.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
				afterEach(t)
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
			ExpStatus: http.StatusForbidden,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertForbiddenResponse(t, resp)
				afterEach(t)
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
			ExpStatus: http.StatusForbidden,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertForbiddenResponse(t, resp)
				afterEach(t)
			},
		},
		{
			Name: "Should return 401 UNAUTHORIZED if the user is not staff",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				return "", ""
			},
			ExpStatus: http.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertUnauthorizedResponse(t, resp)
				afterEach(t)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodDelete, myProfilePath, tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}

func TestUploadMyPicture(t *testing.T) {
	userCleanUp(t)()
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)

	beforeEach := func(t *testing.T) {
		testServices := GetTestServices(t)
		requestID := uuid.NewString()
		ctx := context.Background()

		delProfPicOpts := services.DeleteUserPictureOptions{
			RequestID: requestID,
			UserID:    testUser.ID,
		}
		if serviceErr := testServices.DeleteUserPicture(ctx, delProfPicOpts); serviceErr != nil {
			t.Log("Failed to delete user picture", "serviceError", serviceErr)
		}
	}

	testCases := []TestRequestCase[FormFileBody]{
		{
			Name: "Should return 201 CREATED and convert PNG and JPEG",
			ReqFn: func(t *testing.T) (FormFileBody, string) {
				beforeEach(t)
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return ImageUploadForm(t, "png"), accessToken
			},
			ExpStatus: http.StatusCreated,
			AssertFn: func(t *testing.T, _ FormFileBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.UserPictureResponse{})
				AssertNotEmpty(t, resBody.URL)
				AssertStringContains(t, resBody.URL, ".jpeg")
				AssertEqual(t, resBody.EXT, "jpeg")
				AssertStringContains(t,
					resBody.Links.Self.Href,
					fmt.Sprintf("/api%s/%d%s", paths.UsersPathV1, testUser.ID, paths.PicturePath),
				)
				AssertStringContains(t,
					resBody.Links.User.Href,
					fmt.Sprintf("/api%s/%d", paths.UsersPathV1, testUser.ID),
				)
			},
			Path: myPicturePath,
		},
		{
			Name: "Should return 201 CREATED and compress JPEG",
			ReqFn: func(t *testing.T) (FormFileBody, string) {
				beforeEach(t)
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return ImageUploadForm(t, "jpg"), accessToken
			},
			ExpStatus: fiber.StatusCreated,
			AssertFn: func(t *testing.T, _ FormFileBody, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.UserPictureResponse{})
				AssertNotEmpty(t, resBody.URL)
				AssertStringContains(t, resBody.URL, ".jpeg")
				AssertEqual(t, resBody.EXT, "jpeg")
				AssertStringContains(t,
					resBody.Links.Self.Href,
					fmt.Sprintf("/api%s/%d%s", paths.UsersPathV1, testUser.ID, paths.PicturePath),
				)
				AssertStringContains(t,
					resBody.Links.User.Href,
					fmt.Sprintf("/api%s/%d", paths.UsersPathV1, testUser.ID),
				)
			},
			Path: myPicturePath,
		},
		{
			Name: "Should return 403 FORBIDDEN when the user is not staff",
			ReqFn: func(t *testing.T) (FormFileBody, string) {
				beforeEach(t)
				testUser.IsStaff = false
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return ImageUploadForm(t, "jpg"), accessToken
			},
			ExpStatus: fiber.StatusForbidden,
			AssertFn: func(t *testing.T, _ FormFileBody, resp *http.Response) {
				AssertForbiddenResponse(t, resp)
			},
			Path: myPicturePath,
		},
		{
			Name: "Should return 401 UNAUTHORIZED when the user is not logged in",
			ReqFn: func(t *testing.T) (FormFileBody, string) {
				beforeEach(t)
				return ImageUploadForm(t, "jpg"), ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ FormFileBody, resp *http.Response) {
				AssertUnauthorizedResponse(t, resp)
			},
			Path: myPicturePath,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCaseWithForm(t, tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}

func TestGetMyPicture(t *testing.T) {
	userCleanUp(t)()
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)

	beforeEach := func(t *testing.T) {
		testServices := GetTestServices(t)
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
	}

	afterEach := func(t *testing.T) {
		testServices := GetTestServices(t)
		ctx := context.Background()
		requestID := uuid.NewString()

		delPicOpts := services.DeleteUserPictureOptions{
			RequestID: requestID,
			UserID:    testUser.ID,
		}
		if serviceErr := testServices.DeleteUserPicture(ctx, delPicOpts); serviceErr != nil {
			t.Fatal("Failed to delete user picture", "serviceError", serviceErr)
		}
	}

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 200 OK with staff users picture",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: http.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.UserPictureResponse{})
				AssertNotEmpty(t, resBody.URL)
				AssertStringContains(t, resBody.URL, ".jpeg")
				AssertEqual(t, resBody.EXT, "jpeg")
				AssertStringContains(t,
					resBody.Links.Self.Href,
					fmt.Sprintf("/api%s/%d%s", paths.UsersPathV1, testUser.ID, paths.PicturePath),
				)
				AssertStringContains(t,
					resBody.Links.User.Href,
					fmt.Sprintf("/api%s/%d", paths.UsersPathV1, testUser.ID),
				)
				afterEach(t)
			},
		},
		{
			Name: "Should return 404 NOT FOUND if the staff user does not have a picture",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: http.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
		},
		{
			Name: "Should return 403 FORBIDDEN when the user is not staff",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				testUser.IsStaff = false
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusForbidden,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertForbiddenResponse(t, resp)
				afterEach(t)
			},
			Path: myPicturePath,
		},
		{
			Name: "Should return 401 UNAUTHORIZED when the user is not logged in",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				return "", ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertUnauthorizedResponse(t, resp)
				afterEach(t)
			},
			Path: myPicturePath,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodGet, myPicturePath, tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}

func TestDeleteMyPicture(t *testing.T) {
	userCleanUp(t)()
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)

	beforeEach := func(t *testing.T) {
		testServices := GetTestServices(t)
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
	}

	afterEach := func(t *testing.T) {
		testServices := GetTestServices(t)
		ctx := context.Background()
		requestID := uuid.NewString()

		delPicOpts := services.DeleteUserPictureOptions{
			RequestID: requestID,
			UserID:    testUser.ID,
		}
		if serviceErr := testServices.DeleteUserPicture(ctx, delPicOpts); serviceErr != nil {
			t.Log("Failed to delete user picture", "serviceError", serviceErr)
		}
	}

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 204 NO CONTENT when staff users picture is deleted",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: http.StatusNoContent,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				afterEach(t)
			},
		},
		{
			Name: "Should return 404 NOT FOUND if the staff user does not have a picture",
			ReqFn: func(t *testing.T) (string, string) {
				testUser.IsStaff = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: http.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
				afterEach(t)
			},
		},
		{
			Name: "Should return 403 FORBIDDEN when the user is not staff",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				testUser.IsStaff = false
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusForbidden,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertForbiddenResponse(t, resp)
				afterEach(t)
			},
			Path: myPicturePath,
		},
		{
			Name: "Should return 401 UNAUTHORIZED when the user is not logged in",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				return "", ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertUnauthorizedResponse(t, resp)
				afterEach(t)
			},
			Path: myPicturePath,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodDelete, myPicturePath, tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}

func TestGetUser(t *testing.T) {
	userCleanUp(t)()
	staffUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)

	func() {
		testDb := GetTestDatabase(t)
		ctx := context.Background()

		staffPrms := db.UpdateUserIsStaffParams{
			IsStaff: true,
			ID:      staffUser.ID,
		}
		if err := testDb.UpdateUserIsStaff(ctx, staffPrms); err != nil {
			t.Fatal("Failed to update user is staff", "error", err)
		}

		staffUser.IsStaff = true
	}()

	beforeEach := func(t *testing.T) {
		testServices := GetTestServices(t)
		ctx := context.Background()
		requestID := uuid.NewString()

		delProfOpts := services.DeleteUserProfileOptions{
			RequestID: requestID,
			UserID:    staffUser.ID,
		}
		if serviceErr := testServices.DeleteUserProfile(ctx, delProfOpts); serviceErr != nil {
			t.Log("Failed to delete staff user profile", "serviceError", serviceErr)
		}

		delPicOpts := services.DeleteUserPictureOptions{
			RequestID: requestID,
			UserID:    staffUser.ID,
		}
		if serviceErr := testServices.DeleteUserPicture(ctx, delPicOpts); serviceErr != nil {
			t.Log("Failed to delete staff user picture", "serviceError", serviceErr)
		}
	}

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 200 OK with only user data if the user is staff",
			ReqFn: func(t *testing.T) (string, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: http.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.UserResponse{})
				AssertEqual(t, resBody.ID, staffUser.ID)
				AssertEqual(t, resBody.FirstName, staffUser.FirstName)
				AssertEqual(t, resBody.LastName, staffUser.LastName)
				AssertEqual(t, resBody.Location, staffUser.Location)
				AssertEqual(t, resBody.IsAdmin, staffUser.IsAdmin)
				AssertEqual(t, resBody.IsStaff, true)
				AssertEqual(t, resBody.Embedded, nil)
				AssertStringContains(t, resBody.Links.Self.Href, fmt.Sprintf("%s/%d", userPath, staffUser.ID))
			},
			Path: fmt.Sprintf("%s/%d", userPath, staffUser.ID),
		},
		{
			Name: "Should return 200 OK when the user is not staff and its ID",
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
				AssertEqual(t, resBody.IsStaff, false)
				AssertEqual(t, resBody.Embedded, nil)
				AssertStringContains(t, resBody.Links.Self.Href, fmt.Sprintf("%s/%d", userPath, testUser.ID))
			},
			Path: fmt.Sprintf("%s/%d", userPath, testUser.ID),
		},
		{
			Name: "Should return 200 OK with user and profile if the user is staff",
			ReqFn: func(t *testing.T) (string, string) {
				testServices := GetTestServices(t)
				ctx := context.Background()
				requestID := uuid.NewString()

				profOpts := services.UserProfileOptions{
					RequestID: requestID,
					UserID:    staffUser.ID,
					Bio:       "Lorem ipsum",
					GitHub:    "https://github.com/johndoe",
					LinkedIn:  "https://www.linkedin.com/in/john-doe",
					Website:   "https://johndoe.com",
				}
				if _, serviceErr := testServices.CreateUserProfile(ctx, profOpts); serviceErr != nil {
					t.Fatal("Failed to create user profile", "serviceError", serviceErr)
				}

				return "", ""
			},
			ExpStatus: http.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.UserResponse{})
				AssertEqual(t, resBody.ID, staffUser.ID)
				AssertEqual(t, resBody.FirstName, staffUser.FirstName)
				AssertEqual(t, resBody.LastName, staffUser.LastName)
				AssertEqual(t, resBody.Location, staffUser.Location)
				AssertStringContains(t, resBody.Links.Self.Href, fmt.Sprintf("%s/%d", userPath, staffUser.ID))
				AssertStringContains(t,
					resBody.Links.Profile.Href,
					fmt.Sprintf("%s/%d/profile", userPath, staffUser.ID),
				)
				AssertEqual(t, resBody.Links.Picture, nil)
				AssertEqual(t, resBody.Embedded.Profile.Bio, "Lorem ipsum")
				AssertEqual(t, resBody.Embedded.Profile.GitHub, "https://github.com/johndoe")
				AssertEqual(t, resBody.Embedded.Profile.LinkedIn, "https://www.linkedin.com/in/john-doe")
				AssertEqual(t, resBody.Embedded.Profile.Website, "https://johndoe.com")
				AssertEqual(t,
					resBody.Embedded.Profile.Links.Self.Href,
					fmt.Sprintf("https://api.kiwiscript.com%s/%d/profile", userPath, staffUser.ID),
				)
				AssertEqual(t, resBody.Embedded.Picture, nil)
			},
			Path: fmt.Sprintf("%s/%d", userPath, staffUser.ID),
		},
		{
			Name: "Should return 200 OK with user and picture if the user is staff",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				testServices := GetTestServices(t)
				ctx := context.Background()
				requestID := uuid.NewString()

				picOpts := services.UploadUserPictureOptions{
					RequestID:  requestID,
					UserID:     staffUser.ID,
					FileHeader: ImageUploadMock(t),
				}
				if _, serviceErr := testServices.UploadUserPicture(ctx, picOpts); serviceErr != nil {
					t.Fatal("Failed to upload user picture", "serviceError", serviceErr)
				}

				accessToken, _ := GenerateTestAuthTokens(t, staffUser)
				return "", accessToken
			},
			ExpStatus: http.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.UserResponse{})
				AssertEqual(t, resBody.ID, staffUser.ID)
				AssertEqual(t, resBody.FirstName, staffUser.FirstName)
				AssertEqual(t, resBody.LastName, staffUser.LastName)
				AssertEqual(t, resBody.Location, staffUser.Location)
				AssertStringContains(t, resBody.Links.Self.Href, fmt.Sprintf("%s/%d", userPath, staffUser.ID))
				AssertStringContains(t,
					resBody.Links.Picture.Href,
					fmt.Sprintf("%s/%d/picture", userPath, staffUser.ID),
				)
				AssertEqual(t, resBody.Links.Profile, nil)
				AssertNotEmpty(t, resBody.Embedded.Picture.EXT)
				AssertNotEmpty(t, resBody.Embedded.Picture.URL)
				AssertEqual(t,
					resBody.Embedded.Picture.Links.Self.Href,
					fmt.Sprintf("https://api.kiwiscript.com%s/%d/picture", userPath, staffUser.ID),
				)
				AssertEqual(t, resBody.Embedded.Profile, nil)
			},
			Path: fmt.Sprintf("%s/%d", userPath, staffUser.ID),
		},
		{
			Name: "Should return 200 OK with user, profile and picture if the user is staff",
			ReqFn: func(t *testing.T) (string, string) {
				beforeEach(t)
				testServices := GetTestServices(t)
				ctx := context.Background()
				requestID := uuid.NewString()

				profOpts := services.UserProfileOptions{
					RequestID: requestID,
					UserID:    staffUser.ID,
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
					UserID:     staffUser.ID,
					FileHeader: ImageUploadMock(t),
				}
				if _, serviceErr := testServices.UploadUserPicture(ctx, picOpts); serviceErr != nil {
					t.Fatal("Failed to upload user picture", "serviceError", serviceErr)
				}

				return "", ""
			},
			ExpStatus: http.StatusOK,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.UserResponse{})
				AssertEqual(t, resBody.ID, staffUser.ID)
				AssertEqual(t, resBody.FirstName, staffUser.FirstName)
				AssertEqual(t, resBody.LastName, staffUser.LastName)
				AssertEqual(t, resBody.Location, staffUser.Location)
				AssertStringContains(t, resBody.Links.Self.Href, fmt.Sprintf("%s/%d", userPath, staffUser.ID))
				AssertStringContains(t,
					resBody.Links.Profile.Href,
					fmt.Sprintf("%s/%d/profile", userPath, staffUser.ID),
				)
				AssertStringContains(t,
					resBody.Links.Picture.Href,
					fmt.Sprintf("%s/%d/picture", userPath, staffUser.ID),
				)
				AssertEqual(t, resBody.Embedded.Profile.Bio, "Lorem ipsum")
				AssertEqual(t, resBody.Embedded.Profile.GitHub, "https://github.com/johndoe")
				AssertEqual(t, resBody.Embedded.Profile.LinkedIn, "https://www.linkedin.com/in/john-doe")
				AssertEqual(t, resBody.Embedded.Profile.Website, "https://johndoe.com")
				AssertEqual(t,
					resBody.Embedded.Profile.Links.Self.Href,
					fmt.Sprintf("https://api.kiwiscript.com%s/%d/profile", userPath, staffUser.ID),
				)
				AssertNotEmpty(t, resBody.Embedded.Picture.EXT)
				AssertNotEmpty(t, resBody.Embedded.Picture.URL)
			},
			Path: fmt.Sprintf("%s/%d", userPath, staffUser.ID),
		},
		{
			Name: "Should return 404 NOT FOUND if the user is not staff",
			ReqFn: func(t *testing.T) (string, string) {
				return "", ""
			},
			ExpStatus: http.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf("%s/%d", userPath, testUser.ID),
		},
		{
			Name: "Should return 404 NOT FOUND if the user is not found",
			ReqFn: func(t *testing.T) (string, string) {
				return "", ""
			},
			ExpStatus: http.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf("%s/%d", userPath, 987654321),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodGet, tc.Path, tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}

func TestGetUserProfile(t *testing.T) {
	userCleanUp(t)()
	staffUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)

	func() {
		testDb := GetTestDatabase(t)
		testServices := GetTestServices(t)
		ctx := context.Background()
		requestID := uuid.NewString()

		staffPrms := db.UpdateUserIsStaffParams{
			IsStaff: true,
			ID:      staffUser.ID,
		}
		if err := testDb.UpdateUserIsStaff(ctx, staffPrms); err != nil {
			t.Fatal("Failed to update user is staff", "error", err)
		}

		profOpts := services.UserProfileOptions{
			RequestID: requestID,
			UserID:    staffUser.ID,
			Bio:       "Lorem ipsum",
			GitHub:    "https://github.com/johndoe",
			LinkedIn:  "https://www.linkedin.com/in/john-doe",
			Website:   "https://johndoe.com",
		}
		if _, serviceErr := testServices.CreateUserProfile(ctx, profOpts); serviceErr != nil {
			t.Fatal("Failed to create user profile", "serviceError", serviceErr)
		}

		staffUser.IsStaff = true
	}()

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 200 OK when the user is staff and has a profile",
			ReqFn: func(t *testing.T) (string, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: http.StatusOK,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.UserProfileResponse{})
				AssertGreaterThan(t, resBody.ID, 0)
				AssertEqual(t, resBody.Bio, "Lorem ipsum")
				AssertEqual(t, resBody.Website, "https://johndoe.com")
				AssertEqual(t, resBody.LinkedIn, "https://www.linkedin.com/in/john-doe")
				AssertEqual(t, resBody.GitHub, "https://github.com/johndoe")
				AssertStringContains(t,
					resBody.Links.Self.Href,
					fmt.Sprintf("/api%s/%d%s", paths.UsersPathV1, staffUser.ID, paths.ProfilePath),
				)
				AssertStringContains(t,
					resBody.Links.User.Href,
					fmt.Sprintf("/api%s/%d", paths.UsersPathV1, staffUser.ID),
				)
			},
			Path: fmt.Sprintf("%s/%d%s", userPath, staffUser.ID, paths.ProfilePath),
		},
		{
			Name: "Should return 404 when the user is staff and does not have profile",
			ReqFn: func(t *testing.T) (string, string) {
				testServices := GetTestServices(t)
				ctx := context.Background()
				requestID := uuid.NewString()

				delOpts := services.DeleteUserProfileOptions{
					RequestID: requestID,
					UserID:    staffUser.ID,
				}
				if serviceErr := testServices.DeleteUserProfile(ctx, delOpts); serviceErr != nil {
					t.Log("Failed to delete staff user profile", "serviceError", serviceErr)
				}

				return "", ""
			},
			ExpStatus: http.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf("%s/%d%s", userPath, staffUser.ID, paths.ProfilePath),
		},
		{
			Name: "Should return 404 when the user is not staff and does not have profile",
			ReqFn: func(t *testing.T) (string, string) {
				accessToken, _ := GenerateTestAuthTokens(t, staffUser)
				return "", accessToken
			},
			ExpStatus: http.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf("%s/%d%s", userPath, testUser.ID, paths.ProfilePath),
		},
		{
			Name: "Should return 404 when the user does not exist",
			ReqFn: func(t *testing.T) (string, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: http.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf("%s/%d%s", userPath, 987654321, paths.ProfilePath),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodGet, tc.Path, tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}

func TestGetUserPicture(t *testing.T) {
	userCleanUp(t)()
	staffUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
	testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)

	func() {
		testServices := GetTestServices(t)
		ctx := context.Background()
		requestID := uuid.NewString()

		picOpts := services.UploadUserPictureOptions{
			RequestID:  requestID,
			UserID:     staffUser.ID,
			FileHeader: ImageUploadMock(t),
		}
		if _, serviceErr := testServices.UploadUserPicture(ctx, picOpts); serviceErr != nil {
			t.Fatal("Failed to upload user picture", "serviceError", serviceErr)
		}
	}()

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 200 OK with picture if the user is staff and has picture",
			ReqFn: func(t *testing.T) (string, string) {
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: http.StatusOK,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, dtos.UserPictureResponse{})
				AssertNotEmpty(t, resBody.URL)
				AssertStringContains(t, resBody.URL, ".jpeg")
				AssertEqual(t, resBody.EXT, "jpeg")
				AssertStringContains(t,
					resBody.Links.Self.Href,
					fmt.Sprintf("/api%s/%d%s", paths.UsersPathV1, staffUser.ID, paths.PicturePath),
				)
				AssertStringContains(t,
					resBody.Links.User.Href,
					fmt.Sprintf("/api%s/%d", paths.UsersPathV1, staffUser.ID),
				)
			},
			Path: fmt.Sprintf("%s/%d%s", userPath, staffUser.ID, paths.PicturePath),
		},
		{
			Name: "Should return 404 NOT FOUND if the user is staff but does not have a picture",
			ReqFn: func(t *testing.T) (string, string) {
				testServices := GetTestServices(t)
				ctx := context.Background()
				requestID := uuid.NewString()

				delPicOpts := services.DeleteUserPictureOptions{
					RequestID: requestID,
					UserID:    staffUser.ID,
				}
				if serviceErr := testServices.DeleteUserPicture(ctx, delPicOpts); serviceErr != nil {
					t.Log("Failed to delete user picture", "serviceError", serviceErr)
				}

				return "", ""
			},
			ExpStatus: http.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf("%s/%d%s", userPath, staffUser.ID, paths.PicturePath),
		},
		{
			Name: "Should return 404 when the user is not staff and does not have picture",
			ReqFn: func(t *testing.T) (string, string) {
				accessToken, _ := GenerateTestAuthTokens(t, staffUser)
				return "", accessToken
			},
			ExpStatus: http.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf("%s/%d%s", userPath, testUser.ID, paths.PicturePath),
		},
		{
			Name: "Should return 404 when the user does not exist",
			ReqFn: func(t *testing.T) (string, string) {
				return "", ""
			},
			ExpStatus: http.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				AssertNotFoundResponse(t, resp)
			},
			Path: fmt.Sprintf("%s/%d%s", userPath, 987654321, paths.PicturePath),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, http.MethodGet, tc.Path, tc)
		})
	}

	t.Cleanup(userCleanUp(t))
}
