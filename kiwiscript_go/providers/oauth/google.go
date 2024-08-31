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

package oauth

import (
	"context"
	"encoding/json"
	"github.com/kiwiscript/kiwiscript_go/utils"
	"strings"
)

const googleUserURL string = "https://www.googleapis.com/oauth2/v3/userinfo"

type GoogleUserResponse struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Locale        string `json:"locale"`
	HD            string `json:"hd"`
}

func mapLocaleToLocation(locale string) string {
	switch strings.TrimSpace(locale) {
	case "en_NZ":
		return utils.LocationNZL
	case "en_AU":
		return utils.LocationAUS
	case "en_US", "en_CA", "fr_CA", "es_MX":
		return utils.LocationNAM
	case "bg_BG", "cs_CZ", "da_DK", "de_DE", "el_GR",
		"en_IE", "es_ES", "et_EE", "fi_FI", "fr_FR",
		"hr_HR", "hu_HU", "it_IT", "lt_LT", "lv_LV",
		"mt_MT", "nl_NL", "pl_PL", "pt_PT", "ro_RO",
		"sk_SK", "sl_SI", "sv_SE", "is_IS", "no_NO",
		"de_CH", "fr_CH", "it_CH":
		return utils.LocationEUR
	default:
		return utils.LocationOTH
	}
}

func (ur *GoogleUserResponse) ToUserData() *UserData {
	return &UserData{
		FirstName:  ur.GivenName,
		LastName:   ur.FamilyName,
		Email:      utils.Lowered(ur.Email),
		Location:   mapLocaleToLocation(ur.Locale),
		IsVerified: ur.EmailVerified,
	}
}

func (op *Providers) GetGoogleAuthorizationURL(ctx context.Context, requestID string) (string, string, error) {
	log := op.buildLogger(requestID, "GetGoogleAuthorizationURL")
	log.DebugContext(ctx, "Getting Google authorization url...")

	state, err := GenerateState()
	if err != nil {
		log.ErrorContext(ctx, "Failed to generate state")
		return "", "", err
	}

	log.DebugContext(ctx, "Google authorization url generated successfully")
	return op.google.AuthCodeURL(state), state, nil
}

type GetGoogleAccessTokenOptions struct {
	RequestID string
	Code      string
}

func (op *Providers) GetGoogleAccessToken(ctx context.Context, opts GetGoogleAccessTokenOptions) (string, error) {
	log := op.buildLogger(opts.RequestID, "GetGoogleAccessToken")
	log.DebugContext(ctx, "Getting Google access token...")

	token, err := op.google.Exchange(ctx, opts.Code)
	if err != nil {
		log.ErrorContext(ctx, "Failed to exchange the code for a token", "error", err)
		return "", err
	}

	log.DebugContext(ctx, "Google access token exchanged successfully")
	return token.AccessToken, nil
}

type GetGoogleUserDataOptions struct {
	RequestID string
	Token     string
}

func (op *Providers) GetGoogleUserData(ctx context.Context, opts GetGoogleUserDataOptions) (ToUserData, int, error) {
	log := op.buildLogger(opts.RequestID, "GetGoogleUserData")
	log.DebugContext(ctx, "Getting Google user data...")

	body, status, err := getUserResponse(log, ctx, googleUserURL, opts.Token)
	if err != nil {
		log.ErrorContext(ctx, "Failed to get user data", "error", err)
		return nil, status, err
	}

	userResponse := GoogleUserResponse{}
	if err := json.Unmarshal(body, &userResponse); err != nil {
		log.ErrorContext(ctx, "Failed to parse Google user data", "error", err)
		return nil, status, err
	}

	return &userResponse, status, nil
}
