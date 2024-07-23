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

package services

import (
	"log/slog"

	cc "github.com/kiwiscript/kiwiscript_go/providers/cache"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/providers/email"
	objstg "github.com/kiwiscript/kiwiscript_go/providers/object_storage"
	"github.com/kiwiscript/kiwiscript_go/providers/tokens"
)

type Services struct {
	log      *slog.Logger
	database *db.Database
	cache    *cc.Cache
	mail     *email.Mail
	jwt      *tokens.Tokens
	objStg   *objstg.ObjectStorage
}

func NewServices(
	log *slog.Logger,
	database *db.Database,
	cache *cc.Cache,
	objStg *objstg.ObjectStorage,
	mail *email.Mail,
	jwt *tokens.Tokens,
) *Services {
	return &Services{
		database: database,
		cache:    cache,
		objStg:   objStg,
		mail:     mail,
		jwt:      jwt,
		log:      log,
	}
}
