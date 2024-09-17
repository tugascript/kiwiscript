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

package dtos

import (
	"fmt"
	"github.com/kiwiscript/kiwiscript_go/paths"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
)

type UserProfileBody struct {
	Bio      string `json:"bio" validate:"required,min=1,max=1000"`
	GitHub   string `json:"github" validate:"required,url"`
	LinkedIn string `json:"linkedin" validate:"required,url"`
	Website  string `json:"website" validate:"required,url"`
}

type UserProfileLinks struct {
	Self LinkResponse `json:"self"`
	User LinkResponse `json:"user"`
}

type UserProfileResponse struct {
	ID       int32            `json:"id"`
	Bio      string           `json:"bio"`
	GitHub   string           `json:"github"`
	LinkedIn string           `json:"linkedin"`
	Website  string           `json:"website"`
	Links    UserProfileLinks `json:"_links"`
}

func NewUserProfileResponse(backendDomain string, profile *db.UserProfile) *UserProfileResponse {
	return &UserProfileResponse{
		ID:       profile.ID,
		Bio:      profile.Bio,
		GitHub:   profile.Github,
		LinkedIn: profile.Linkedin,
		Website:  profile.Website,
		Links: UserProfileLinks{
			Self: LinkResponse{
				Href: fmt.Sprintf(
					"https://%s/api%s/%d%s",
					backendDomain,
					paths.UsersPathV1,
					profile.UserID,
					paths.ProfilePath,
				),
			},
			User: LinkResponse{
				Href: fmt.Sprintf(
					"https://%s/api%s/%d",
					backendDomain,
					paths.UsersPathV1,
					profile.UserID,
				),
			},
		},
	}
}
