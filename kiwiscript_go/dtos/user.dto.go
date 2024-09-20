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

type UserPathParams struct {
	UserID string `validate:"required,number,min=1"`
}

type DeleteUserBody struct {
	Password string `json:"password,omitempty" validate:"min=1"`
}

type UpdateUserBody struct {
	FirstName string `json:"firstName" validate:"required,min=2,max=50"`
	LastName  string `json:"lastName" validate:"required,min=2,max=50"`
	Location  string `json:"location" validate:"required,min=3,max=3"`
}

type UserPictureEmbedded struct {
	ID    uuid.UUID        `json:"id"`
	EXT   string           `json:"ext"`
	URL   string           `json:"url"`
	Links SelfLinkResponse `json:"_links"`
}

type UserProfileEmbedded struct {
	Bio      string           `json:"bio"`
	GitHub   string           `json:"github"`
	LinkedIn string           `json:"linkedin"`
	Website  string           `json:"website"`
	Links    SelfLinkResponse `json:"_links"`
}

type UserEmbedded struct {
	Picture *UserPictureEmbedded `json:"picture,omitempty"`
	Profile *UserProfileEmbedded `json:"profile,omitempty"`
}

func newFullSeriesEmbedded(
	backendDomain string,
	userID int32,
	profile *db.UserProfileModel,
	picture *db.UserPictureModel,
) *UserEmbedded {
	if profile != nil && picture != nil {
		return nil
	}

	var pictureEmbedded *UserPictureEmbedded
	if picture != nil {
		pictureEmbedded = &UserPictureEmbedded{
			ID:  picture.ID,
			EXT: picture.EXT,
			URL: fmt.Sprintf("https://kiwiscript.com%s", picture.URL),
			Links: SelfLinkResponse{
				Self: LinkResponse{
					Href: fmt.Sprintf(
						"https://%s/api%s/%d%s",
						backendDomain,
						paths.UsersPathV1,
						userID,
						paths.PicturePath,
					),
				},
			},
		}
	}

	var profileEmbedded *UserProfileEmbedded
	if profile != nil {
		profileEmbedded = &UserProfileEmbedded{
			Bio:      profile.Bio,
			GitHub:   profile.GitHub,
			LinkedIn: profile.LinkedIn,
			Website:  profile.Website,
			Links: SelfLinkResponse{
				Self: LinkResponse{
					Href: fmt.Sprintf(
						"https://%s/api%s/%d%s",
						backendDomain,
						paths.UsersPathV1,
						userID,
						paths.ProfilePath,
					),
				},
			},
		}
	}

	return &UserEmbedded{
		Picture: pictureEmbedded,
		Profile: profileEmbedded,
	}
}

type UserResponse struct {
	ID        int32            `json:"id"`
	FirstName string           `json:"firstName"`
	LastName  string           `json:"lastName"`
	Location  string           `json:"location"`
	IsAdmin   bool             `json:"isAdmin"`
	IsStaff   bool             `json:"isStaff"`
	Links     SelfLinkResponse `json:"_links"`
	Embedded  *UserEmbedded    `json:"_embedded"`
}

func NewUserResponse(backendDomain string, user *db.UserModel) *UserResponse {
	return &UserResponse{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Location:  user.Location,
		IsAdmin:   user.IsAdmin,
		IsStaff:   user.IsStaff,
		Links: SelfLinkResponse{
			Self: LinkResponse{
				Href: fmt.Sprintf("https://%s/api%s/%d", backendDomain, paths.UsersPathV1, user.ID),
			},
		},
	}
}

func NewUserResponseWithEmbedded(
	backendDomain string,
	user *db.UserModel,
	profile *db.UserProfileModel,
	picture *db.UserPictureModel,
) *UserResponse {
	return &UserResponse{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Location:  user.Location,
		IsAdmin:   user.IsAdmin,
		IsStaff:   user.IsStaff,
		Links: SelfLinkResponse{
			Self: LinkResponse{
				Href: fmt.Sprintf("https://%s/api%s/%d", backendDomain, paths.UsersPathV1, user.ID),
			},
		},
		Embedded: newFullSeriesEmbedded(backendDomain, user.ID, profile, picture),
	}
}
