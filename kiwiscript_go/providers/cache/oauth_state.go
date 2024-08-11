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

package cc

import "time"

const (
	oauthStatePrefix  string = "oauth_state"
	oauthEmailPrefix  string = "oauth_email"
	oauthStateSeconds int    = 120
)

func (c *Cache) AddOAuthState(state, provider string) error {
	return c.storage.Set(
		oauthStatePrefix+":"+state,
		[]byte(provider),
		time.Duration(oauthStateSeconds)*time.Second,
	)
}

func (c *Cache) VerifyOAuthState(state, provider string) (bool, error) {
	valByte, err := c.storage.Get(oauthStatePrefix + ":" + state)

	if err != nil {
		return false, err
	}
	if valByte == nil {
		return false, nil
	}

	return string(valByte) == provider, nil
}

func (c *Cache) AddOAuthEmail(code, email string) error {
	return c.storage.Set(
		oauthEmailPrefix+":"+code,
		[]byte(email),
		time.Duration(oauthStateSeconds)*time.Second,
	)
}

func (c *Cache) GetOAuthEmail(code string) (string, error) {
	valByte, err := c.storage.Get(oauthEmailPrefix + ":" + code)
	if err != nil {
		return "", err
	}

	if valByte == nil {
		return "", nil
	}

	return string(valByte), nil
}
