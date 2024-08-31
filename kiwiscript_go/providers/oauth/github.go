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

const gitHubUserURL string = "https://api.github.com/user"

type GitHubUserResponse struct {
	AvatarURL         string `json:"avatar_url"`
	EventsURL         string `json:"events_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	GravatarID        string `json:"gravatar_id"`
	HTMLURL           string `json:"html_url"`
	ID                int64  `json:"id"`
	NodeID            string `json:"node_id"`
	Login             string `json:"login"`
	OrganizationsURL  string `json:"organizations_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	ReposURL          string `json:"repos_url"`
	SiteAdmin         bool   `json:"site_admin"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	Type              string `json:"type"`
	URL               string `json:"url"`
	Bio               string `json:"bio"`
	Blog              string `json:"blog"`
	Company           string `json:"company"`
	Email             string `json:"email"`
	Followers         int    `json:"followers"`
	Following         int    `json:"following"`
	Hireable          bool   `json:"hireable"`
	Location          string `json:"location"`
	Name              string `json:"name"`
	PublicGists       int    `json:"public_gists"`
	PublicRepos       int    `json:"public_repos"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

func mapGHLocationToLocation(location string) string {
	lowLoc := utils.Lowered(location)
	locArr := strings.Fields(lowLoc)
	arrLen := len(locArr)
	lastLoc := lowLoc
	if arrLen > 1 {
		lastLoc = locArr[arrLen-1]
	}

	switch lastLoc {
	case "nz", "nzl", "new zealand", "newzealand":
		return utils.LocationNZL
	case "au", "aus", "australia":
		return utils.LocationNZL
	case "us", "usa", "united states", "america", "unitedstates",
		"ca", "can", "canada", "mx", "mex", "mexico":
		return utils.LocationNAM
	case "bg", "bgr", "bulgaria", "cz", "cze", "czech republic", "czechrepublic",
		"dk", "dnk", "denmark", "de", "deu", "germany", "gr", "grc", "greece",
		"ie", "irl", "ireland", "es", "esp", "spain", "ee", "est", "estonia",
		"fi", "fin", "finland", "fr", "fra", "france", "hr", "hrv", "croatia",
		"hu", "hun", "hungary", "it", "ita", "italy", "lt", "ltu", "lithuania",
		"lv", "lva", "latvia", "mt", "mlt", "malta", "nl", "nld", "netherlands",
		"pl", "pol", "poland", "pt", "prt", "portugal", "ro", "rou", "romania",
		"sk", "svk", "slovakia", "si", "svn", "slovenia", "se", "swe", "sweden",
		"is", "isl", "iceland", "no", "nor", "norway", "ch", "che", "switzerland":
		return utils.LocationEUR
	default:
		return utils.LocationOTH
	}
}

func (ur *GitHubUserResponse) ToUserData() *UserData {
	nameSplit := strings.Fields(strings.TrimSpace(ur.Name))
	var firstName, lastName string
	if len(nameSplit) > 1 {
		firstName = nameSplit[0]
		lastName = strings.Join(nameSplit[1:], " ")
	} else {
		firstName = nameSplit[0]
		lastName = nameSplit[0]
	}

	return &UserData{
		FirstName:  firstName,
		LastName:   lastName,
		Email:      utils.Lowered(ur.Email),
		Location:   mapGHLocationToLocation(ur.Location),
		IsVerified: true,
	}
}

func (op *Providers) GetGitHubAuthorizationURL(ctx context.Context, requestID string) (string, string, error) {
	log := op.buildLogger(requestID, "GetGitHubAuthorizationURL")
	log.DebugContext(ctx, "Getting GitHub authorization url...")

	state, err := GenerateState()
	if err != nil {
		log.ErrorContext(ctx, "Failed to generate state")
		return "", "", err
	}

	log.DebugContext(ctx, "GitHub authorization url generated successfully")
	return op.gitHub.AuthCodeURL(state), state, nil
}

type GetGitHubAccessTokenOptions struct {
	RequestID string
	Code      string
}

func (op *Providers) GetGitHubAccessToken(ctx context.Context, opts GetGitHubAccessTokenOptions) (string, error) {
	log := op.buildLogger(opts.RequestID, "GetGitHubAccessToken")
	log.DebugContext(ctx, "Getting GitHub access token...")

	token, err := op.gitHub.Exchange(ctx, opts.Code)
	if err != nil {
		log.ErrorContext(ctx, "Failed to exchange the code for a token", "error", err)
		return "", err
	}

	log.DebugContext(ctx, "GitHub access token exchanged successfully")
	return token.AccessToken, nil
}

type GetGitHubUserDataOptions struct {
	RequestID string
	Token     string
}

func (op *Providers) GetGitHubUserData(ctx context.Context, opts GetGitHubUserDataOptions) (ToUserData, int, error) {
	log := op.buildLogger(opts.RequestID, "GetGitHubUserData")
	log.DebugContext(ctx, "Getting GitHub user data...")

	body, status, err := getUserResponse(log, ctx, gitHubUserURL, opts.Token)
	if err != nil {
		log.ErrorContext(ctx, "Failed to get user data", "error", err)
		return nil, status, err
	}

	userResponse := GitHubUserResponse{}
	if err := json.Unmarshal(body, &userResponse); err != nil {
		log.ErrorContext(ctx, "Failed to parse GitHub user data", "error", err)
		return nil, status, err
	}

	return &userResponse, status, nil
}
