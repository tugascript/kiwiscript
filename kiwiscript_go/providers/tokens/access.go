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

type AccessUserClaims struct {
	ID        int32
	Version   int16
	FirstName string
	LastName  string
	IsAdmin   bool
	IsStaff   bool
}

type tokenClaims struct {
	User AccessUserClaims
	jwt.RegisteredClaims
}

func (t *Tokens) CreateAccessToken(user *db.User) (string, error) {
	now := time.Now()
	iat := jwt.NewNumericDate(now)
	exp := jwt.NewNumericDate(now.Add(time.Second * time.Duration(t.accessData.ttlSec)))
	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, tokenClaims{
		User: AccessUserClaims{
			ID:        user.ID,
			Version:   user.Version,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			IsAdmin:   user.IsAdmin,
			IsStaff:   user.IsStaff,
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
	return token.SignedString(t.accessData.privateKey)
}

func (t *Tokens) VerifyAccessToken(token string) (AccessUserClaims, error) {
	claims := tokenClaims{}
	_, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return t.accessData.publicKey, nil
	})
	return claims.User, err
}

func (t *Tokens) GetAccessTtl() int64 {
	return t.accessData.ttlSec
}
