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
	"time"
)

const blackListPrefix string = "black_list"

type AddBlacklistOptions struct {
	RequestID string
	ID        string
	Exp       time.Time
}

func (c *Cache) AddBlacklist(ctx context.Context, opts AddBlacklistOptions) error {
	log := c.buildLogger(opts.RequestID, "AddBlacklist")
	log.DebugContext(ctx, "Blacklisting refresh token", "id", opts.ID, "exp", opts.Exp.Format(time.RFC3339))
	key := blackListPrefix + ":" + opts.ID
	val := []byte(opts.ID)
	exp := time.Until(opts.Exp)
	return c.storage.Set(key, val, exp)
}

type IsBlacklistedOptions struct {
	RequestID string
	ID        string
}

func (c *Cache) IsBlacklisted(ctx context.Context, opts IsBlacklistedOptions) (bool, error) {
	log := c.buildLogger(opts.RequestID, "IsBlacklisted")
	log.DebugContext(ctx, "Checking if refresh token is blacklisted")
	key := blackListPrefix + ":" + opts.ID
	valByte, err := c.storage.Get(key)

	if err != nil {
		log.ErrorContext(ctx, "Error checking if refresh token is blacklisted", "error", err)
		return false, err
	}
	if valByte == nil {
		log.DebugContext(ctx, "Refresh token is not blacklisted")
		return false, nil
	}

	return true, nil
}
