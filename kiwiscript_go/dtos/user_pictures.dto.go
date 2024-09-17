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
	"github.com/google/uuid"
	"github.com/kiwiscript/kiwiscript_go/paths"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
)

type UserPictureLinks struct {
	Self    LinkResponse `json:"self"`
	User    LinkResponse `json:"user"`
	Profile LinkResponse `json:"profile"`
}

func newUserPictureLinks(
	backendDomain string,
	userID int32,
) UserPictureLinks {
	return UserPictureLinks{
		Self: LinkResponse{
			Href: fmt.Sprintf(
				"https://%s/api%s/%d%s",
				backendDomain,
				paths.UsersPathV1,
				userID,
				paths.PicturePath,
			),
		},
		User: LinkResponse{
			Href: fmt.Sprintf(
				"https://%s/api%s/%d",
				backendDomain,
				paths.UsersPathV1,
				userID,
			),
		},
		Profile: LinkResponse{
			Href: fmt.Sprintf(
				"https://%s/api%s/%d%s",
				backendDomain,
				paths.UsersPathV1,
				userID,
				paths.ProfilePath,
			),
		},
	}
}

type UserPictureResponse struct {
	ID    uuid.UUID        `json:"id"`
	EXT   string           `json:"ext"`
	URL   string           `json:"url"`
	Links UserPictureLinks `json:"_links"`
}

func NewUserPictureResponse(
	backendDomain string,
	picture *db.UserPictureModel,
) *UserPictureResponse {
	return &UserPictureResponse{
		ID:    picture.ID,
		EXT:   picture.EXT,
		URL:   picture.URL,
		Links: newUserPictureLinks(backendDomain, picture.UserID),
	}
}
