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
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
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

type AddTwoFactorCodeOptions struct {
	RequestID string
	UserID    int32
}

func (c *Cache) AddTwoFactorCode(ctx context.Context, opts AddTwoFactorCodeOptions) (string, error) {
	log := c.buildLogger(opts.RequestID, "AddTwoFactorCode").With("userID", opts.UserID)
	log.DebugContext(ctx, "Adding two factor code...")

	code, err := generateCode()
	if err != nil {
		log.ErrorContext(ctx, "Error generating two factor code", "error", err)
		return "", err
	}

	hashedCode, err := utils.HashPassword(code)
	if err != nil {
		log.ErrorContext(ctx, "Error hashing two factor code", "error", err)
		return "", err
	}

	key := fmt.Sprintf("%s:%d", twoFactorPrefix, opts.UserID)
	val := []byte(hashedCode)
	exp := time.Duration(twoFactorSeconds) * time.Second
	if err := c.storage.Set(key, val, exp); err != nil {
		log.ErrorContext(ctx, "Error setting two factor code", "error", err)
		return "", err
	}

	return code, nil
}

type VerifyTwoFactorCodeOptions struct {
	RequestID string
	UserID    int32
	Code      string
}

func (c *Cache) VerifyTwoFactorCode(ctx context.Context, opts VerifyTwoFactorCodeOptions) (bool, error) {
	log := c.buildLogger(opts.RequestID, "VerifyTwoFactorCode").With("userID", opts.UserID)
	log.DebugContext(ctx, "Verifying two factor code...")
	key := fmt.Sprintf("%s:%d", twoFactorPrefix, opts.UserID)

	valByte, err := c.storage.Get(key)
	if err != nil {
		log.ErrorContext(ctx, "Error verifying two factor code", "error", err)
		return false, err
	}
	if valByte == nil {
		log.DebugContext(ctx, "Two factor code not found")
		return false, nil
	}

	val := string(valByte)
	if !utils.VerifyPassword(opts.Code, val) {
		log.DebugContext(ctx, "Two factor code is invalid")
		return false, nil
	}
	if err := c.storage.Delete(key); err != nil {
		log.ErrorContext(ctx, "Error deleting two factor code", "error", err)
		return true, err
	}

	return true, nil
}
