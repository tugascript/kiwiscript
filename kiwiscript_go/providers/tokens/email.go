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

package tokens

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
)

const (
	EmailTokenConfirmation string = "confirmation"
	EmailTokenReset        string = "reset"
)

type EmailUserClaims struct {
	ID      int32
	Version int16
}

type emailClaims struct {
	User EmailUserClaims
	Type string
	jwt.RegisteredClaims
}

func (t *Tokens) CreateEmailToken(tokenType string, user db.User) (string, error) {
	now := time.Now()
	iat := jwt.NewNumericDate(now)
	exp := jwt.NewNumericDate(now.Add(time.Second * time.Duration(t.emailData.ttlSec)))
	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, emailClaims{
		Type: tokenType,
		User: EmailUserClaims{
			ID:      user.ID,
			Version: user.Version,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    t.iss,
			Audience:  jwt.ClaimStrings{t.iss},
			Subject:   user.Email,
			IssuedAt:  iat,
			NotBefore: iat,
			ExpiresAt: exp,
			ID:        uuid.NewString(),
		},
	})
	return token.SignedString(t.emailData.privateKey)
}

func (t *Tokens) VerifyEmailToken(token string) (string, EmailUserClaims, error) {
	claims := emailClaims{}
	_, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return t.emailData.publicKey, nil
	})
	return claims.Type, claims.User, err
}
