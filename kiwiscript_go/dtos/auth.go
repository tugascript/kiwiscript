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

import "github.com/google/uuid"

type SignUpBody struct {
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"firstName" validate:"required,min=2,max=50"`
	LastName  string `json:"lastName" validate:"required,min=2,max=50"`
	Location  string `json:"location" validate:"required,min=3,max=3"`
	Password1 string `json:"password" validate:"required,min=8,max=50"`
	Password2 string `json:"password2" validate:"required,eqfield=Password1"`
}

type SignInBody struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=1"`
}

type ConfirmSignInBody struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,min=1"`
}

type SignOutBody struct {
	RefreshToken string `json:"refreshToken" validate:"required,jwt"`
}

type RefreshBody struct {
	RefreshToken string `json:"refreshToken" validate:"required,jwt"`
}

type ConfirmBody struct {
	ConfirmationToken string `json:"confirmationToken" validate:"required,jwt"`
}

type ForgotPasswordBody struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordBody struct {
	ResetToken string `json:"resetToken" validate:"required,jwt"`
	Password1  string `json:"password" validate:"required,min=8,max=50"`
	Password2  string `json:"password2" validate:"required,eqfield=Password1"`
}

type UpdatePasswordBody struct {
	OldPassword string `json:"oldPassword" validate:"required,min=1"`
	Password1   string `json:"password" validate:"required,min=8,max=50"`
	Password2   string `json:"password2" validate:"required,eqfield=Password1"`
}

type UpdateEmailBody struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=1"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

func NewAuthResponse(accessToken, refreshToken string, expiresIn int64) AuthResponse {
	return AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		TokenType:    "Bearer",
	}
}

type MessageResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

func NewMessageResponse(message string) MessageResponse {
	return MessageResponse{
		ID:      uuid.NewString(),
		Message: message,
	}
}
