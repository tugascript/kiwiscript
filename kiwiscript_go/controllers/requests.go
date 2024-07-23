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

package controllers

type SignUpRequest struct {
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"firstName" validate:"required,min=2,max=50"`
	LastName  string `json:"lastName" validate:"required,min=2,max=50"`
	Location  string `json:"location" validate:"required,min=3,max=3"`
	BirthDate string `json:"birthDate" validate:"required"`
	Password1 string `json:"password" validate:"required,min=8,max=50"`
	Password2 string `json:"password2" validate:"required,eqfield=Password1"`
}

type SignInRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=1"`
}

type ConfirmSignInRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,min=1"`
}

type SignOutRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required,jwt"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required,jwt"`
}

type ConfirmRequest struct {
	ConfirmationToken string `json:"confirmationToken" validate:"required,jwt"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	ResetToken string `json:"resetToken" validate:"required,jwt"`
	Password1  string `json:"password" validate:"required,min=8,max=50"`
	Password2  string `json:"password2" validate:"required,eqfield=Password1"`
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"oldPassword" validate:"required,min=1"`
	Password1   string `json:"password" validate:"required,min=8,max=50"`
	Password2   string `json:"password2" validate:"required,eqfield=Password1"`
}

type UpdateEmailRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=1"`
}

type LanguageRequest struct {
	Name string `json:"name" validate:"required,min=2,max=50,extalphanum"`
	Icon string `json:"icon" validate:"required,svg"`
}

type CreateSeriesRequest struct {
	Title       string   `json:"title" validate:"require,min=2,max=100"`
	Description string   `json:"description" validate:"required,min=2"`
	Tags        []string `json:"tags" validate:"required,max=5,unique,dive,min=2,max=50,slug"`
}

type UpdateSeriesRequest struct {
	Title       string `json:"title" validate:"require,min=2,max=100"`
	Description string `json:"description" validate:"required,min=2"`
	Position    int16  `json:"position" validate:"required,gte=1"`
}

type AddTagToSeriesRequest struct {
	Tag string `json:"tag" validate:"required,min=2,max=50,slug"`
}

type UpdateIsPublishedRequest struct {
	IsPublished bool `json:"is_published" validate:"required"`
}

type CreateSeriesPartRequest struct {
	Title       string `json:"title" validate:"required,min=2,max=250"`
	Description string `json:"description" validate:"required,min=2"`
}

type UpdateSeriesPartRequest struct {
	Title       string `json:"title" validate:"required,min=2,max=250"`
	Description string `json:"description" validate:"required,min=2"`
	Position    int16  `json:"position" validate:"required,gte=1"`
}

type CreateLectureRequest struct {
	Title string `json:"title" validate:"required,min=2,max=250"`
}

type UpdateLectureRequest struct {
	Title    string `json:"title" validate:"required,min=2,max=250"`
	Position int16  `json:"position" validate:"required,gte=1"`
}

type LectureArticleRequest struct {
	Content string `json:"content" validate:"required,min=1,markdown"`
}

type LectureVideoRequest struct {
	URL       string `json:"url" validate:"required,url"`
	WatchTime int32  `json:"watchTime" validate:"required,number,min=1"`
}

type LectureFileRequest struct {
	Name string `validate:"required,min=2,max=250,extalphanum"`
}
