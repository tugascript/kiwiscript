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

const oauthEmailPrefix string = "oauth_email"

type AddOAuthEmailOptions struct {
	RequestID       string
	Code            string
	Email           string
	DurationSeconds int64
}

func (c *Cache) AddOAuthEmail(ctx context.Context, opts AddOAuthEmailOptions) error {
	log := c.buildLogger(opts.RequestID, "AddOAuthEmail").With("code", opts.Code)
	log.DebugContext(ctx, "Adding OAuth email...")
	return c.storage.Set(
		oauthEmailPrefix+":"+opts.Code,
		[]byte(opts.Email),
		time.Duration(opts.DurationSeconds)*time.Second,
	)
}

type GetOAuthEmailOptions struct {
	RequestID string
	Code      string
}

func (c *Cache) GetOAuthEmail(ctx context.Context, opts GetOAuthEmailOptions) (string, error) {
	log := c.buildLogger(opts.RequestID, "GetOAuthEmail").With("code", opts.Code)
	log.DebugContext(ctx, "Getting OAuth email...")

	valByte, err := c.storage.Get(oauthEmailPrefix + ":" + opts.Code)
	if err != nil {
		log.ErrorContext(ctx, "Error getting OAuth email", "error", err)
		return "", err
	}

	if valByte == nil {
		log.DebugContext(ctx, "OAuth email not found")
		return "", nil
	}

	return string(valByte), nil
}
