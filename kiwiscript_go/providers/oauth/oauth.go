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
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/kiwiscript/kiwiscript_go/paths"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"io"
	"log/slog"
	"net/http"
)

var gitHubScopes = [2]string{
	"user:email",
	"read:user",
}

var googleScopes = [2]string{
	"https://www.googleapis.com/auth/userinfo.email",
	"https://www.googleapis.com/auth/userinfo.profile",
}

type Providers struct {
	gitHub oauth2.Config
	google oauth2.Config
	log    *slog.Logger
}

func NewOAuthProviders(
	log *slog.Logger,
	gitHubID,
	gitHubSecret,
	googleID,
	googleSecret,
	backendDomain string,
) *Providers {
	redirectUrl := fmt.Sprintf("https://%s%s/ext/", backendDomain, paths.AuthPath)

	return &Providers{
		gitHub: oauth2.Config{
			ClientID:     gitHubID,
			ClientSecret: gitHubSecret,
			Endpoint:     github.Endpoint,
			Scopes:       gitHubScopes[:],
			RedirectURL:  redirectUrl + "github/callback",
		},
		google: oauth2.Config{
			ClientID:     googleID,
			ClientSecret: googleSecret,
			Endpoint:     google.Endpoint,
			Scopes:       googleScopes[:],
			RedirectURL:  redirectUrl + "google/callback",
		},
		log: log,
	}
}

func GenerateState() (string, error) {
	bytes := make([]byte, 16)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func getUserResponse(log *slog.Logger, ctx context.Context, url, token string) ([]byte, int, error) {
	log.DebugContext(ctx, "Getting user data...", "url", url)

	log.DebugContext(ctx, "Building user data request")
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.ErrorContext(ctx, "Failed to build user data request")
		return nil, 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	log.DebugContext(ctx, "Requesting user data...")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.ErrorContext(ctx, "Failed to request the user data")
		return nil, 0, err
	}

	if res.StatusCode != http.StatusOK {
		log.ErrorContext(ctx, "Responded with a non 200 OK status", "status", res.StatusCode)
		return nil, res.StatusCode, errors.New("status code is not 200 OK")
	}

	log.DebugContext(ctx, "Reading the body")
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.ErrorContext(ctx, "Failed to read the body", "error", err)
		return nil, 0, err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.ErrorContext(ctx, "Failed to close response body", "error", err)
		}
	}()

	return body, res.StatusCode, nil
}

type UserData struct {
	FirstName  string
	LastName   string
	Email      string
	Location   string
	IsVerified bool
}

type ToUserData interface {
	ToUserData() *UserData
}
