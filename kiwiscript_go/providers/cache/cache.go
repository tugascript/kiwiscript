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
	"github.com/gofiber/storage/redis/v3"
	"github.com/kiwiscript/kiwiscript_go/utils"
	"log/slog"
)

type Cache struct {
	log     *slog.Logger
	storage *redis.Storage
}

func NewCache(log *slog.Logger, storage *redis.Storage) *Cache {
	return &Cache{
		log:     log,
		storage: storage,
	}
}

func (c *Cache) ResetCache() error {
	return c.storage.Reset()
}

func (c *Cache) buildLogger(requestID, function string) *slog.Logger {
	return utils.BuildLogger(c.log, utils.LoggerOptions{
		Layer:     utils.ProvidersLogLayer,
		Location:  "cache",
		Function:  function,
		RequestID: requestID,
	})
}
