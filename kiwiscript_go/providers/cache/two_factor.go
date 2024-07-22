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

import (
	"crypto/rand"
	"math/big"
	"strconv"
	"time"

	"github.com/kiwiscript/kiwiscript_go/utils"
)

const (
	twoFactorPrefix  string = "two_factor"
	twoFactorSeconds int    = 300
)

func generateCode() (string, error) {
	const codeLength = 6
	const digits = "0123456789"
	code := make([]byte, codeLength)

	for i := 0; i < codeLength; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		code[i] = digits[num.Int64()]
	}

	return string(code), nil
}

func (c *Cache) AddTwoFactorCode(userID int32) (string, error) {
	code, err := generateCode()
	if err != nil {
		return code, err
	}

	hashedCode, err := utils.HashPassword(code)
	if err != nil {
		return "", err
	}

	key := twoFactorPrefix + ":" + strconv.Itoa(int(userID))
	val := []byte(hashedCode)
	exp := time.Duration(twoFactorSeconds) * time.Second
	if err := c.storage.Set(key, val, exp); err != nil {
		return "", err
	}

	return code, nil
}

func (c *Cache) VerifyTwoFactorCode(userID int32, code string) (bool, error) {
	key := twoFactorPrefix + ":" + strconv.Itoa(int(userID))
	valByte, err := c.storage.Get(key)

	if err != nil {
		return false, err
	}
	if valByte == nil {
		return false, nil
	}

	val := string(valByte)
	if !utils.VerifyPassword(code, val) {
		return false, nil
	}
	if err := c.storage.Delete(key); err != nil {
		return true, err
	}

	return true, nil
}
