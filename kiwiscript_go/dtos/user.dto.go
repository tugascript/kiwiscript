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

type UserPathParams struct {
	UserID string `validate:"required,number,min=1"`
}

type DeleteUserBody struct {
	Password string `json:"password" validate:"required,min=1"`
}

type UserResponse struct {
	ID        int32            `json:"id"`
	FirstName string           `json:"firstName"`
	LastName  string           `json:"lastName"`
	Location  string           `json:"location"`
	IsAdmin   bool             `json:"isAdmin"`
	IsStaff   bool             `json:"isStaff"`
	Links     SelfLinkResponse `json:"_links"`
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
				Href: fmt.Sprintf("https://%s%s/%d", backendDomain, paths.UsersPathV1, user.ID),
			},
		},
	}
}
