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
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"time"
)

type OAuthUserClaims struct {
	ID      int32
	Version int16
}

type oauthClaims struct {
	User OAuthUserClaims
	jwt.RegisteredClaims
}

func (t *Tokens) CreateOAuthToken(user *db.User) (string, error) {
	now := time.Now()
	iat := jwt.NewNumericDate(now)
	exp := jwt.NewNumericDate(now.Add(time.Second * time.Duration(t.oauthData.ttlSec)))
	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, oauthClaims{
		User: OAuthUserClaims{
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
	return token.SignedString(t.oauthData.privateKey)
}

func (t *Tokens) VerifyOAuthToken(token string) (OAuthUserClaims, error) {
	claims := oauthClaims{}
	_, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return t.oauthData.publicKey, nil
	})
	return claims.User, err
}

func (t *Tokens) GetOAuthTtl() int64 {
	return t.oauthData.ttlSec
}
