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
	"time"
)

const blackListPrefix string = "black_list"

type AddBlackListOptions struct {
	ID  string
	Exp time.Time
}

func (c *Cache) AddBlackList(options AddBlackListOptions) error {
	key := blackListPrefix + ":" + options.ID
	val := []byte(options.ID)
	exp := time.Until(options.Exp)
	return c.storage.Set(key, val, exp)
}

func (c *Cache) IsBlackListed(id string) (bool, error) {
	key := blackListPrefix + ":" + id
	valByte, err := c.storage.Get(key)

	if err != nil {
		return false, err
	}
	if valByte == nil {
		return false, nil
	}

	return true, nil
}
