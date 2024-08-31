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

const (
	oauthStatePrefix  string = "oauth_state"
	oauthStateSeconds int    = 120
)

type AddOAuthStateOptions struct {
	RequestID string
	State     string
	Provider  string
}

func (c *Cache) AddOAuthState(ctx context.Context, opts AddOAuthStateOptions) error {
	log := c.buildLogger(opts.RequestID, "AddOAuthState").With(
		"state", opts.State,
		"provider", opts.Provider,
	)
	log.DebugContext(ctx, "Adding OAuth state...")
	return c.storage.Set(
		oauthStatePrefix+":"+opts.State,
		[]byte(opts.Provider),
		time.Duration(oauthStateSeconds)*time.Second,
	)
}

type VerifyOAuthStateOptions struct {
	RequestID string
	State     string
	Provider  string
}

func (c *Cache) VerifyOAuthState(ctx context.Context, opts VerifyOAuthStateOptions) (bool, error) {
	log := c.buildLogger(opts.RequestID, "VerifyOAuthState").With(
		"state", opts.State,
		"provider", opts.Provider,
	)
	log.DebugContext(ctx, "Verifying OAuth state...")
	valByte, err := c.storage.Get(oauthStatePrefix + ":" + opts.State)

	if err != nil {
		log.ErrorContext(ctx, "Error verifying OAuth state", "error", err)
		return false, err
	}
	if valByte == nil {
		log.DebugContext(ctx, "OAuth state not found")
		return false, nil
	}

	return string(valByte) == opts.Provider, nil
}
