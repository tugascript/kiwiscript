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

package db

type UserProfileModel struct {
	ID       int32
	Bio      string
	GitHub   string
	LinkedIn string
	Website  string
	UserID   int32
}

type ToUserProfileModel interface {
	ToUserProfileModel() *UserProfileModel
}

func (up *UserProfile) ToUserProfileModel() *UserProfileModel {
	return &UserProfileModel{
		ID:       up.ID,
		Bio:      up.Bio,
		GitHub:   up.Github,
		LinkedIn: up.Linkedin,
		Website:  up.Website,
		UserID:   up.UserID,
	}
}

func (u *FindStaffUserByIdWithProfileAndPictureRow) ToUserProfileModel() *UserProfileModel {
	if u.ProfileID.Valid && u.ProfileBio.Valid &&
		u.ProfileGithub.Valid && u.ProfileLinkedin.Valid &&
		u.ProfileWebsite.Valid {
		return &UserProfileModel{
			ID:       u.ProfileID.Int32,
			Bio:      u.ProfileBio.String,
			GitHub:   u.ProfileGithub.String,
			LinkedIn: u.ProfileLinkedin.String,
			Website:  u.ProfileWebsite.String,
			UserID:   u.ID,
		}
	}

	return nil
}
