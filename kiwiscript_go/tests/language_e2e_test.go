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
	"net/http"
	"strings"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/kiwiscript/kiwiscript_go/controllers"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/utils"
)

func languagesCleanUp(t *testing.T) func() {
	return func() {
		dbProv := GetTestDatabase(t)
		cacheProv := GetTestCache(t)

		if err := dbProv.DeleteAllLanguages(context.Background()); err != nil {
			t.Fatal("Failed to delete all languages", err)
		}
		if err := cacheProv.ResetCache(); err != nil {
			t.Fatal("Failed to reset cache", err)
		}
	}
}

const baseLanguagesPath = "/api/v1/languages"

type fakeCreateLanguageRequest struct {
	Name string `faker:"oneof: TypeScript, Go, Rust, Python"`
}

var languageIcons = map[string]string{
	"TypeScript": `
		<svg xmlns="http://www.w3.org/2000/svg" preserveAspectRatio="xMidYMid" viewBox="0 0 256 256">
  			<path fill="#3178C6" d="M20 0h216c11.046 0 20 8.954 20 20v216c0 11.046-8.954 20-20 20H20c-11.046 0-20-8.954-20-20V20C0 8.954 8.954 0 20 0Z"/>
  			<path fill="#FFF" d="M150.518 200.475v27.62c4.492 2.302 9.805 4.028 15.938 5.179 6.133 1.151 12.597 1.726 19.393 1.726 6.622 0 12.914-.633 18.874-1.899 5.96-1.266 11.187-3.352 15.678-6.257 4.492-2.906 8.048-6.704 10.669-11.394 2.62-4.689 3.93-10.486 3.93-17.391 0-5.006-.749-9.394-2.246-13.163a30.748 30.748 0 0 0-6.479-10.055c-2.821-2.935-6.205-5.567-10.149-7.898-3.945-2.33-8.394-4.531-13.347-6.602-3.628-1.497-6.881-2.949-9.761-4.359-2.879-1.41-5.327-2.848-7.342-4.316-2.016-1.467-3.571-3.021-4.665-4.661-1.094-1.64-1.641-3.495-1.641-5.567 0-1.899.489-3.61 1.468-5.135s2.362-2.834 4.147-3.927c1.785-1.094 3.973-1.942 6.565-2.547 2.591-.604 5.471-.906 8.638-.906 2.304 0 4.737.173 7.299.518 2.563.345 5.14.877 7.732 1.597a53.669 53.669 0 0 1 7.558 2.719 41.7 41.7 0 0 1 6.781 3.797v-25.807c-4.204-1.611-8.797-2.805-13.778-3.582-4.981-.777-10.697-1.165-17.147-1.165-6.565 0-12.784.705-18.658 2.115-5.874 1.409-11.043 3.61-15.506 6.602-4.463 2.993-7.99 6.805-10.582 11.437-2.591 4.632-3.887 10.17-3.887 16.615 0 8.228 2.375 15.248 7.127 21.06 4.751 5.811 11.963 10.731 21.638 14.759a291.458 291.458 0 0 1 10.625 4.575c3.283 1.496 6.119 3.049 8.509 4.66 2.39 1.611 4.276 3.366 5.658 5.265 1.382 1.899 2.073 4.057 2.073 6.474a9.901 9.901 0 0 1-1.296 4.963c-.863 1.524-2.174 2.848-3.93 3.97-1.756 1.122-3.945 1.999-6.565 2.632-2.62.633-5.687.95-9.2.95-5.989 0-11.92-1.05-17.794-3.151-5.875-2.1-11.317-5.25-16.327-9.451Zm-46.036-68.733H140V109H41v22.742h35.345V233h28.137V131.742Z"/>
		</svg>
	`,
	"Go": `
		<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 207 78">
  			<g fill-rule="evenodd">
    			<path d="M16.2 24.1c-.4 0-.5-.2-.3-.5l2.1-2.7c.2-.3.7-.5 1.1-.5h35.7c.4 0 .5.3.3.6l-1.7 2.6c-.2.3-.7.6-1 .6zM1.1 33.3c-.4 0-.5-.2-.3-.5l2.1-2.7c.2-.3.7-.5 1.1-.5h45.6c.4 0 .6.3.5.6l-.8 2.4c-.1.4-.5.6-.9.6zM25.3 42.5c-.4 0-.5-.3-.3-.6l1.4-2.5c.2-.3.6-.6 1-.6h20c.4 0 .6.3.6.7l-.2 2.4c0 .4-.4.7-.7.7zM129.1 22.3c-6.3 1.6-10.6 2.8-16.8 4.4-1.5.4-1.6.5-2.9-1-1.5-1.7-2.6-2.8-4.7-3.8-6.3-3.1-12.4-2.2-18.1 1.5-6.8 4.4-10.3 10.9-10.2 19 .1 8 5.6 14.6 13.5 15.7 6.8.9 12.5-1.5 17-6.6.9-1.1 1.7-2.3 2.7-3.7H90.3c-2.1 0-2.6-1.3-1.9-3 1.3-3.1 3.7-8.3 5.1-10.9.3-.6 1-1.6 2.5-1.6h36.4c-.2 2.7-.2 5.4-.6 8.1-1.1 7.2-3.8 13.8-8.2 19.6-7.2 9.5-16.6 15.4-28.5 17-9.8 1.3-18.9-.6-26.9-6.6-7.4-5.6-11.6-13-12.7-22.2-1.3-10.9 1.9-20.7 8.5-29.3C71.1 9.6 80.5 3.7 92 1.6c9.4-1.7 18.4-.6 26.5 4.9 5.3 3.5 9.1 8.3 11.6 14.1.6.9.2 1.4-1 1.7z"/>
    			<path fill-rule="nonzero" d="M162.2 77.6c-9.1-.2-17.4-2.8-24.4-8.8-5.9-5.1-9.6-11.6-10.8-19.3-1.8-11.3 1.3-21.3 8.1-30.2 7.3-9.6 16.1-14.6 28-16.7 10.2-1.8 19.8-.8 28.5 5.1 7.9 5.4 12.8 12.7 14.1 22.3 1.7 13.5-2.2 24.5-11.5 33.9-6.6 6.7-14.7 10.9-24 12.8-2.7.5-5.4.6-8 .9zM186 37.2c-.1-1.3-.1-2.3-.3-3.3-1.8-9.9-10.9-15.5-20.4-13.3-9.3 2.1-15.3 8-17.5 17.4-1.8 7.8 2 15.7 9.2 18.9 5.5 2.4 11 2.1 16.3-.6 7.9-4.1 12.2-10.5 12.7-19.1z"/>
  			</g>
		</svg>
	`,
	"Rust": `
		<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 224 224">
  			<path fill="#000" d="m218.46 109.358-9.062-5.614c-.076-.882-.162-1.762-.258-2.642l7.803-7.265a3.107 3.107 0 0 0 .933-2.89 3.093 3.093 0 0 0-1.967-2.312l-9.97-3.715c-.25-.863-.512-1.72-.781-2.58l6.214-8.628a3.114 3.114 0 0 0-.592-4.263 3.134 3.134 0 0 0-1.431-.637l-10.507-1.709a80.869 80.869 0 0 0-1.263-2.353l4.417-9.7a3.12 3.12 0 0 0-.243-3.035 3.106 3.106 0 0 0-2.705-1.385l-10.671.372a85.152 85.152 0 0 0-1.685-2.044l2.456-10.381a3.125 3.125 0 0 0-3.762-3.763l-10.384 2.456a88.996 88.996 0 0 0-2.047-1.684l.373-10.671a3.11 3.11 0 0 0-1.385-2.704 3.127 3.127 0 0 0-3.034-.246l-9.681 4.417c-.782-.429-1.567-.854-2.353-1.265l-1.713-10.506a3.098 3.098 0 0 0-1.887-2.373 3.108 3.108 0 0 0-3.014.35l-8.628 6.213c-.85-.27-1.703-.53-2.56-.778l-3.716-9.97a3.111 3.111 0 0 0-2.311-1.97 3.134 3.134 0 0 0-2.89.933l-7.266 7.802a93.746 93.746 0 0 0-2.643-.258l-5.614-9.082A3.125 3.125 0 0 0 111.97 4c-1.09 0-2.085.56-2.642 1.478l-5.615 9.081a93.32 93.32 0 0 0-2.642.259l-7.266-7.802a3.13 3.13 0 0 0-2.89-.933 3.106 3.106 0 0 0-2.312 1.97l-3.715 9.97c-.857.247-1.71.506-2.56.778L73.7 12.588a3.101 3.101 0 0 0-3.014-.35A3.127 3.127 0 0 0 68.8 14.61l-1.713 10.506c-.79.41-1.575.832-2.353 1.265l-9.681-4.417a3.125 3.125 0 0 0-4.42 2.95l.372 10.67c-.69.553-1.373 1.115-2.048 1.685l-10.383-2.456a3.143 3.143 0 0 0-2.93.832 3.124 3.124 0 0 0-.833 2.93l2.436 10.383a93.897 93.897 0 0 0-1.68 2.043l-10.672-.372a3.138 3.138 0 0 0-2.704 1.385 3.126 3.126 0 0 0-.246 3.035l4.418 9.7c-.43.779-.855 1.563-1.266 2.353l-10.507 1.71a3.097 3.097 0 0 0-2.373 1.886 3.117 3.117 0 0 0 .35 3.013l6.214 8.628a89.12 89.12 0 0 0-.78 2.58l-9.97 3.715a3.117 3.117 0 0 0-1.035 5.202l7.803 7.265c-.098.879-.184 1.76-.258 2.642l-9.062 5.614A3.122 3.122 0 0 0 4 112.021c0 1.092.56 2.084 1.478 2.642l9.062 5.614c.074.882.16 1.762.258 2.642l-7.803 7.265a3.117 3.117 0 0 0 1.034 5.201l9.97 3.716a110 110 0 0 0 .78 2.58l-6.212 8.627a3.112 3.112 0 0 0 .6 4.27c.419.33.916.547 1.443.63l10.507 1.709c.407.792.83 1.576 1.265 2.353l-4.417 9.68a3.126 3.126 0 0 0 2.95 4.42l10.65-.374c.553.69 1.115 1.372 1.685 2.047l-2.435 10.383a3.09 3.09 0 0 0 .831 2.91 3.117 3.117 0 0 0 2.931.83l10.384-2.436a82.268 82.268 0 0 0 2.047 1.68l-.371 10.671a3.11 3.11 0 0 0 1.385 2.704 3.125 3.125 0 0 0 3.034.241l9.681-4.416c.779.432 1.563.854 2.353 1.265l1.713 10.505a3.147 3.147 0 0 0 1.887 2.395 3.111 3.111 0 0 0 3.014-.349l8.628-6.213c.853.271 1.71.535 2.58.783l3.716 9.969a3.112 3.112 0 0 0 2.312 1.967 3.112 3.112 0 0 0 2.89-.933l7.266-7.802c.877.101 1.761.186 2.642.264l5.615 9.061a3.12 3.12 0 0 0 2.642 1.478 3.165 3.165 0 0 0 2.663-1.478l5.614-9.061c.884-.078 1.765-.163 2.643-.264l7.265 7.802a3.106 3.106 0 0 0 2.89.933 3.105 3.105 0 0 0 2.312-1.967l3.716-9.969c.863-.248 1.719-.512 2.58-.783l8.629 6.213a3.12 3.12 0 0 0 4.9-2.045l1.713-10.506c.793-.411 1.577-.838 2.353-1.265l9.681 4.416a3.13 3.13 0 0 0 3.035-.241 3.126 3.126 0 0 0 1.385-2.704l-.372-10.671a81.794 81.794 0 0 0 2.046-1.68l10.383 2.436a3.123 3.123 0 0 0 3.763-3.74l-2.436-10.382a84.588 84.588 0 0 0 1.68-2.048l10.672.374a3.104 3.104 0 0 0 2.704-1.385 3.118 3.118 0 0 0 .244-3.035l-4.417-9.68c.43-.779.852-1.563 1.263-2.353l10.507-1.709a3.08 3.08 0 0 0 2.373-1.886 3.11 3.11 0 0 0-.35-3.014l-6.214-8.627c.272-.857.532-1.717.781-2.58l9.97-3.716a3.109 3.109 0 0 0 1.967-2.311 3.107 3.107 0 0 0-.933-2.89l-7.803-7.265c.096-.88.182-1.761.258-2.642l9.062-5.614a3.11 3.11 0 0 0 1.478-2.642 3.157 3.157 0 0 0-1.476-2.663h-.064zm-60.687 75.337c-3.468-.747-5.656-4.169-4.913-7.637a6.412 6.412 0 0 1 7.617-4.933c3.468.741 5.676 4.169 4.933 7.637a6.414 6.414 0 0 1-7.617 4.933h-.02zm-3.076-20.847c-3.158-.677-6.275 1.334-6.936 4.5l-3.22 15.026c-9.929 4.5-21.055 7.018-32.614 7.018-11.89 0-23.12-2.622-33.234-7.328l-3.22-15.026c-.677-3.158-3.778-5.18-6.936-4.499l-13.273 2.848a80.222 80.222 0 0 1-6.853-8.091h64.61c.731 0 1.218-.132 1.218-.797v-22.91c0-.665-.487-.797-1.218-.797H94.133v-14.469h20.415c1.864 0 9.97.533 12.551 10.898.811 3.179 2.601 13.54 3.818 16.863 1.214 3.715 6.152 11.146 11.415 11.146h32.202c.365 0 .755-.041 1.166-.116a80.56 80.56 0 0 1-7.307 8.587l-13.583-2.911-.113.058zm-89.38 20.537a6.407 6.407 0 0 1-7.617-4.933c-.74-3.467 1.462-6.894 4.934-7.637a6.417 6.417 0 0 1 7.617 4.933c.74 3.468-1.464 6.894-4.934 7.637zm-24.564-99.28a6.438 6.438 0 0 1-3.261 8.484c-3.241 1.438-7.019-.025-8.464-3.261-1.445-3.237.025-7.039 3.262-8.483a6.416 6.416 0 0 1 8.463 3.26zM33.22 102.94l13.83-6.15c2.952-1.311 4.294-4.769 2.972-7.72l-2.848-6.44H58.36v50.362h-22.5a79.158 79.158 0 0 1-3.014-21.672c0-2.869.155-5.697.452-8.483l-.08.103zm60.687-4.892v-14.86h26.629c1.376 0 9.722 1.59 9.722 7.822 0 5.18-6.399 7.038-11.663 7.038h-24.77.082zm96.811 13.375c0 1.973-.072 3.922-.216 5.862h-8.113c-.811 0-1.137.532-1.137 1.327v3.715c0 8.752-4.934 10.671-9.268 11.146-4.129.464-8.691-1.726-9.248-4.252-2.436-13.684-6.482-16.595-12.881-21.672 7.948-5.036 16.204-12.487 16.204-22.498 0-10.753-7.369-17.523-12.385-20.847-7.059-4.644-14.862-5.572-16.968-5.572H52.899c11.374-12.673 26.835-21.673 44.174-24.975l9.887 10.361a5.849 5.849 0 0 0 8.278.19l11.064-10.568c23.119 4.314 42.729 18.721 54.082 38.598l-7.576 17.09c-1.306 2.951.027 6.419 2.973 7.72l14.573 6.48c.255 2.607.383 5.224.384 7.843l-.021.052zM106.912 24.94a6.398 6.398 0 0 1 9.062.209 6.437 6.437 0 0 1-.213 9.082 6.396 6.396 0 0 1-9.062-.21 6.436 6.436 0 0 1 .213-9.083v.002zm75.137 60.476a6.402 6.402 0 0 1 8.463-3.26 6.425 6.425 0 0 1 3.261 8.482 6.402 6.402 0 0 1-8.463 3.261 6.425 6.425 0 0 1-3.261-8.483z"/>
		</svg>
	`,
	"Python": `
		<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 64 64">
  			<path fill="url(#a)" d="M31.885 16c-8.124 0-7.617 3.523-7.617 3.523l.01 3.65h7.752v1.095H21.197S16 23.678 16 31.876c0 8.196 4.537 7.906 4.537 7.906h2.708v-3.804s-.146-4.537 4.465-4.537h7.688s4.32.07 4.32-4.175v-7.019S40.374 16 31.885 16zm-4.275 2.454a1.394 1.394 0 1 1 0 2.79 1.393 1.393 0 0 1-1.395-1.395c0-.771.624-1.395 1.395-1.395z"/>
  			<path fill="url(#b)" d="M32.115 47.833c8.124 0 7.617-3.523 7.617-3.523l-.01-3.65H31.97v-1.095h10.832S48 40.155 48 31.958c0-8.197-4.537-7.906-4.537-7.906h-2.708v3.803s.146 4.537-4.465 4.537h-7.688s-4.32-.07-4.32 4.175v7.019s-.656 4.247 7.833 4.247zm4.275-2.454a1.393 1.393 0 0 1-1.395-1.395 1.394 1.394 0 1 1 1.395 1.395z"/>
  			<defs>
    			<linearGradient id="a" x1="19.075" x2="34.898" y1="18.782" y2="34.658" gradientUnits="userSpaceOnUse">
      				<stop stop-color="#387EB8"/>
      				<stop offset="1" stop-color="#366994"/>
   				</linearGradient>
    			<linearGradient id="b" x1="28.809" x2="45.803" y1="28.882" y2="45.163" gradientUnits="userSpaceOnUse">
      				<stop stop-color="#FFE052"/>
      				<stop offset="1" stop-color="#FFC331"/>
    			</linearGradient>
  			</defs>
		</svg>
	`,
}

func TestCreateLanguage(t *testing.T) {
	generateFakeCreateLanguageRequest := func(t *testing.T) controllers.LanguageRequest {
		var req fakeCreateLanguageRequest

		if err := faker.FakeData(&req); err != nil {
			t.Fatal("Failed to generate fake data", err)
		}

		return controllers.LanguageRequest{
			Name: req.Name,
			Icon: languageIcons[req.Name],
		}
	}

	testCases := []TestRequestCase[controllers.LanguageRequest]{
		{
			Name: "Should return 201 CREATED when creating a language",
			ReqFn: func(t *testing.T) (controllers.LanguageRequest, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				testUser.IsAdmin = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				req := generateFakeCreateLanguageRequest(t)
				return req, accessToken
			},
			ExpStatus: fiber.StatusCreated,
			AssertFn: func(t *testing.T, req controllers.LanguageRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.LanguageResponse{})
				AssertEqual(t, utils.CapitalizedFirst(req.Name), resBody.Name)
				AssertEqual(t, strings.TrimSpace(req.Icon), resBody.Icon)
			},
			DelayMs: 0,
		},
		{
			Name: "Should return 409 FORBIDEN if user is not admin",
			ReqFn: func(t *testing.T) (controllers.LanguageRequest, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				req := generateFakeCreateLanguageRequest(t)
				return req, accessToken
			},
			ExpStatus: fiber.StatusForbidden,
			AssertFn: func(t *testing.T, req controllers.LanguageRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, controllers.StatusForbidden, resBody.Message)
				AssertEqual(t, controllers.StatusForbidden, resBody.Code)
			},
			DelayMs: 0,
		},
		{
			Name: "Should return 401 if user is not authenticated",
			ReqFn: func(t *testing.T) (controllers.LanguageRequest, string) {
				req := generateFakeCreateLanguageRequest(t)
				return req, ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, req controllers.LanguageRequest, resp *http.Response) {
				assertUnauthorizeError(t, resp)
			},
			DelayMs: 0,
		},
		{
			Name: "Should return 400 BAD REQUEST if validation fails",
			ReqFn: func(t *testing.T) (controllers.LanguageRequest, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				testUser.IsAdmin = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				req := controllers.LanguageRequest{
					Name: "Some non valid @#$^%",
					Icon: "not svg",
				}
				return req, accessToken
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req controllers.LanguageRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
				AssertEqual(t, 2, len(resBody.Fields))
				AssertEqual(t, "name", resBody.Fields[0].Param)
				AssertEqual(t, controllers.StrFieldErrMessageExtAlphaNum, resBody.Fields[0].Message)
				AssertEqual(t, "icon", resBody.Fields[1].Param)
				AssertEqual(t, controllers.StrFieldErrMessageSvg, resBody.Fields[1].Message)
			},
			DelayMs: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, MethodPost, baseLanguagesPath, tc)
		})
	}

	t.Cleanup(languagesCleanUp(t))
	t.Cleanup(userCleanUp(t))
}

func TestGetLanguages(t *testing.T) {
	languagesCleanUp(t)()
	reqFn := func(t *testing.T) (string, string) {
		return "", ""
	}

	emptyTestCases := []TestRequestCase[string]{
		{
			Name:      "Should return 200 OK when languages are empty",
			ReqFn:     reqFn,
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.PaginatedResponse[controllers.LanguageResponse]{})
				AssertEqual(t, 0, resBody.Count)
				AssertEqual(t, 0, len(resBody.Results))
				AssertEqual(t, nil, resBody.Links.Next)
				AssertEqual(t, nil, resBody.Links.Prev)
			},
			DelayMs: 0,
			Path:    baseLanguagesPath,
		},
		{
			Name:      "Should return 400 BAD REQUEST if query params are invalid",
			ReqFn:     reqFn,
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
				AssertEqual(t, 3, len(resBody.Fields))
				AssertEqual(t, controllers.RequestValidationLocationQuery, resBody.Location)
				AssertEqual(t, "limit", resBody.Fields[0].Param)
				AssertEqual(t, controllers.IntFieldErrMessageLte, resBody.Fields[0].Message)
				AssertEqual(t, "offset", resBody.Fields[1].Param)
				AssertEqual(t, controllers.IntFieldErrMessageGte, resBody.Fields[1].Message)
				AssertEqual(t, "search", resBody.Fields[2].Param)
				AssertEqual(t, controllers.StrFieldErrMessageExtAlphaNum, resBody.Fields[2].Message)
			},
			DelayMs: 0,
			Path:    baseLanguagesPath + "?limit=1000&offset=-1&search=%21%21%21%21",
		},
	}

	for _, tc := range emptyTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, MethodGet, tc.Path, tc)
		})
	}

	func() {
		testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
		cLangPrms := [4]db.CreateLanguageParams{
			{
				Name:     "Rust",
				Icon:     strings.TrimSpace(languageIcons["Rust"]),
				AuthorID: testUser.ID,
				Slug:     "rust",
			},
			{
				Name:     "TypeScript",
				Icon:     strings.TrimSpace(languageIcons["TypeScript"]),
				AuthorID: testUser.ID,
				Slug:     "typescript",
			},
			{
				Name:     "Go",
				Icon:     strings.TrimSpace(languageIcons["Go"]),
				AuthorID: testUser.ID,
				Slug:     "go",
			},
			{
				Name:     "Python",
				Icon:     strings.TrimSpace(languageIcons["Python"]),
				AuthorID: testUser.ID,
				Slug:     "python",
			},
		}
		testDb := GetTestDatabase(t)

		for _, params := range cLangPrms {
			_, err := testDb.CreateLanguage(context.Background(), params)
			if err != nil {
				t.Fatal("Failed to create language", err)
			}
		}
	}()

	testCases := []TestRequestCase[string]{
		{
			Name:      "Should return 200 OK when languages with the 4 languages",
			ReqFn:     reqFn,
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.PaginatedResponse[controllers.LanguageResponse]{})
				AssertEqual(t, 4, resBody.Count)
				AssertEqual(t, 4, len(resBody.Results))
				AssertEqual(t, nil, resBody.Links.Next)
				AssertEqual(t, nil, resBody.Links.Prev)
				AssertEqual(t, "go", resBody.Results[0].Slug)
				AssertEqual(t, "typescript", resBody.Results[3].Slug)
			},
			DelayMs: 0,
			Path:    baseLanguagesPath,
		},
		{
			Name:      "Should return 200 OK with only Python with a py search",
			ReqFn:     reqFn,
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.PaginatedResponse[controllers.LanguageResponse]{})
				AssertEqual(t, 1, resBody.Count)
				AssertEqual(t, 1, len(resBody.Results))
				AssertEqual(t, nil, resBody.Links.Next)
				AssertEqual(t, nil, resBody.Links.Prev)
				AssertEqual(t, "Python", resBody.Results[0].Name)
			},
			DelayMs: 0,
			Path:    baseLanguagesPath + "?search=py&limit=2",
		},
		{
			Name:      "Should return 200 OK with only go and python with a limit of 2",
			ReqFn:     reqFn,
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.PaginatedResponse[controllers.LanguageResponse]{})
				AssertEqual(t, 4, resBody.Count)
				AssertEqual(t, 2, len(resBody.Results))
				AssertNotEmpty(t, resBody.Links.Next.Href)
				AssertEqual(t, nil, resBody.Links.Prev)
				AssertEqual(t, "go", resBody.Results[0].Slug)
				AssertEqual(t, "python", resBody.Results[1].Slug)
			},
			Path: baseLanguagesPath + "?limit=2",
		},
		{
			Name:      "Should return 200 OK with only typescript with a limit of 1 and offset of 2 and search t",
			ReqFn:     reqFn,
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.PaginatedResponse[controllers.LanguageResponse]{})
				AssertEqual(t, 3, resBody.Count)
				AssertEqual(t, 1, len(resBody.Results))
				AssertEqual(t, nil, resBody.Links.Next)
				AssertNotEmpty(t, resBody.Links.Prev.Href)
				AssertEqual(t, "TypeScript", resBody.Results[0].Name)
			},
			Path: baseLanguagesPath + "?limit=1&offset=2&search=t",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, MethodGet, tc.Path, tc)
		})
	}

	t.Cleanup(languagesCleanUp(t))
	t.Cleanup(userCleanUp(t))
}

func TestGetLanguage(t *testing.T) {
	languagesCleanUp(t)()
	reqFn := func(t *testing.T) (string, string) {
		return "", ""
	}

	emptyTestCases := []TestRequestCase[string]{
		{
			Name:      "Should return 404 NOT FOUND when language is not found",
			ReqFn:     reqFn,
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, controllers.StatusNotFound, resBody.Code)
				AssertEqual(t, "Resource not found", resBody.Message)
			},
			Path: baseLanguagesPath + "/python",
		},
		{
			Name:      "Should return 400 BAD REQUEST if language name is invalid",
			ReqFn:     reqFn,
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
				AssertEqual(t, controllers.RequestValidationLocationParams, resBody.Location)
				AssertEqual(t, 1, len(resBody.Fields))
				AssertEqual(t, "slug", resBody.Fields[0].Param)
				AssertEqual(t, controllers.StrFieldErrMessageSlug, resBody.Fields[0].Message)
			},
			Path: baseLanguagesPath + "/some-name-",
		},
	}

	for _, tc := range emptyTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, MethodGet, tc.Path, tc)
		})
	}

	func(t *testing.T) {
		testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
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
	}(t)

	testCases := []TestRequestCase[string]{
		{
			Name:      "Should return 200 OK when language is found",
			ReqFn:     reqFn,
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.LanguageResponse{})
				AssertEqual(t, "Rust", resBody.Name)
				AssertEqual(t, "rust", resBody.Slug)
				AssertEqual(t, strings.TrimSpace(languageIcons["Rust"]), resBody.Icon)
			},
			Path: baseLanguagesPath + "/rust",
		},
		{
			Name:      "Should return 404 NOT FOUND when language is not found",
			ReqFn:     reqFn,
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, req string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, controllers.StatusNotFound, resBody.Code)
				AssertEqual(t, "Resource not found", resBody.Message)
			},
			Path: baseLanguagesPath + "/python",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, MethodGet, tc.Path, tc)
		})
	}

	t.Cleanup(languagesCleanUp(t))
	t.Cleanup(userCleanUp(t))
}

func TestUpdateLanguage(t *testing.T) {
	languagesCleanUp(t)()
	func() {
		testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
		testDb := GetTestDatabase(t)

		params := db.CreateLanguageParams{
			Name:     "Rustasdasd",
			Icon:     strings.TrimSpace(languageIcons["Rust"]),
			AuthorID: testUser.ID,
			Slug:     "rustasdasd",
		}
		if _, err := testDb.CreateLanguage(context.Background(), params); err != nil {
			t.Fatal("Failed to create language", err)
		}
	}()

	testCases := []TestRequestCase[controllers.LanguageRequest]{
		{
			Name: "Should return 200 OK when updating a language",
			ReqFn: func(t *testing.T) (controllers.LanguageRequest, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				testUser.IsAdmin = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				req := controllers.LanguageRequest{
					Name: "Rust",
					Icon: languageIcons["Rust"],
				}
				return req, accessToken
			},
			ExpStatus: fiber.StatusOK,
			AssertFn: func(t *testing.T, req controllers.LanguageRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.LanguageResponse{})
				AssertEqual(t, "Rust", resBody.Name)
				AssertEqual(t, "rust", resBody.Slug)
				AssertEqual(t, strings.TrimSpace(languageIcons["Rust"]), resBody.Icon)
			},
			Path: baseLanguagesPath + "/rustasdasd",
		},
		{
			Name: "Should return 404 NOT FOUND when language is not found",
			ReqFn: func(t *testing.T) (controllers.LanguageRequest, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				testUser.IsAdmin = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				req := controllers.LanguageRequest{
					Name: "Rust",
					Icon: languageIcons["Rust"],
				}
				return req, accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, req controllers.LanguageRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, controllers.StatusNotFound, resBody.Code)
				AssertEqual(t, "Resource not found", resBody.Message)
			},
			Path: baseLanguagesPath + "/python",
		},
		{
			Name: "Should return 400 BAD REQUEST if language name is invalid",
			ReqFn: func(t *testing.T) (controllers.LanguageRequest, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				testUser.IsAdmin = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				req := controllers.LanguageRequest{
					Name: "Some non valid @#$^%",
					Icon: "not svg",
				}
				return req, accessToken
			},
			ExpStatus: fiber.StatusBadRequest,
			AssertFn: func(t *testing.T, req controllers.LanguageRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestValidationError{})
				AssertEqual(t, 2, len(resBody.Fields))
				AssertEqual(t, "name", resBody.Fields[0].Param)
				AssertEqual(t, controllers.StrFieldErrMessageExtAlphaNum, resBody.Fields[0].Message)
				AssertEqual(t, "icon", resBody.Fields[1].Param)
				AssertEqual(t, controllers.StrFieldErrMessageSvg, resBody.Fields[1].Message)
			},
			Path: baseLanguagesPath + "/rustasdasd",
		},
		{
			Name: "Should return 401 UNAUTHORIZED if user is not authenticated",
			ReqFn: func(t *testing.T) (controllers.LanguageRequest, string) {
				req := controllers.LanguageRequest{
					Name: "Rust",
					Icon: languageIcons["Rust"],
				}
				return req, ""
			},
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, req controllers.LanguageRequest, resp *http.Response) {
				assertUnauthorizeError(t, resp)
			},
			Path: baseLanguagesPath + "/rustasdasd",
		},
		{
			Name: "Should return 403 FORBIDEN if user is not admin",
			ReqFn: func(t *testing.T) (controllers.LanguageRequest, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				req := controllers.LanguageRequest{
					Name: "Rust",
					Icon: languageIcons["Rust"],
				}
				return req, accessToken
			},
			ExpStatus: fiber.StatusForbidden,
			AssertFn: func(t *testing.T, req controllers.LanguageRequest, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, controllers.StatusForbidden, resBody.Message)
				AssertEqual(t, controllers.StatusForbidden, resBody.Code)
			},
			Path: baseLanguagesPath + "/rustasdasd",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, MethodPut, tc.Path, tc)
		})
	}

	t.Cleanup(languagesCleanUp(t))
	t.Cleanup(userCleanUp(t))
}

func TestDeleteLanguage(t *testing.T) {
	languagesCleanUp(t)()
	func() {
		testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
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
	}()

	testCases := []TestRequestCase[string]{
		{
			Name: "Should return 204 NO CONTENT when deleting a language",
			ReqFn: func(t *testing.T) (string, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				testUser.IsAdmin = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNoContent,
			AssertFn:  func(_ *testing.T, _ string, _ *http.Response) {},
			Path:      baseLanguagesPath + "/rust",
		},
		{
			Name: "Should return 404 NOT FOUND when language is not found",
			ReqFn: func(t *testing.T) (string, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				testUser.IsAdmin = true
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusNotFound,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, controllers.StatusNotFound, resBody.Code)
				AssertEqual(t, "Resource not found", resBody.Message)
			},
			Path: baseLanguagesPath + "/python",
		},
		{
			Name:      "Should return 401 UNAUTHORIZED if user is not authenticated",
			ReqFn:     func(t *testing.T) (string, string) { return "", "" },
			ExpStatus: fiber.StatusUnauthorized,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				assertUnauthorizeError(t, resp)
			},
			Path: baseLanguagesPath + "/rust",
		},
		{
			Name: "Should return 403 FORBIDEN if user is not admin",
			ReqFn: func(t *testing.T) (string, string) {
				testUser := confirmTestUser(t, CreateTestUser(t, nil).ID)
				accessToken, _ := GenerateTestAuthTokens(t, testUser)
				return "", accessToken
			},
			ExpStatus: fiber.StatusForbidden,
			AssertFn: func(t *testing.T, _ string, resp *http.Response) {
				resBody := AssertTestResponseBody(t, resp, controllers.RequestError{})
				AssertEqual(t, controllers.StatusForbidden, resBody.Message)
				AssertEqual(t, controllers.StatusForbidden, resBody.Code)
			},
			Path: baseLanguagesPath + "/rustasdasd",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			PerformTestRequestCase(t, MethodDelete, tc.Path, tc)
		})
	}

	t.Cleanup(languagesCleanUp(t))
	t.Cleanup(userCleanUp(t))
}
