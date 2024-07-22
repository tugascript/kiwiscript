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
	"fmt"
	"github.com/google/uuid"
	"time"
)

const fileURLPrefix string = "file_url"

func creteFileURLKey(userID int32, fileID uuid.UUID) string {
	return fmt.Sprintf("%s:%d:%s", fileURLPrefix, userID, fileID.String())
}

type AddFileURLOptions struct {
	UserID int32
	FileID uuid.UUID
	URL    string
}

func (c *Cache) AddFileURL(opts AddFileURLOptions) error {
	key := creteFileURLKey(opts.UserID, opts.FileID)
	val := []byte(opts.URL)
	exp := time.Hour*23 + time.Minute*55
	return c.storage.Set(key, val, exp)
}

type GetFileURLOptions struct {
	UserID int32
	FileID uuid.UUID
}

func (c *Cache) GetFileURL(opts GetFileURLOptions) (string, error) {
	key := creteFileURLKey(opts.UserID, opts.FileID)
	valByte, err := c.storage.Get(key)

	if err != nil {
		return "", err
	}
	if valByte == nil {
		return "", nil
	}

	return string(valByte), nil
}
