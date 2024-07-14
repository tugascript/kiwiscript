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

package controllers

import (
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/kiwiscript/kiwiscript_go/services"
)

type Controllers struct {
	log               *slog.Logger
	services          *services.Services
	validate          *validator.Validate
	frontendDomain    string
	backendDomain     string
	refreshCookieName string
}

func NewControllers(log *slog.Logger, services *services.Services, validate *validator.Validate, frontendDomain, backendDomain, refreshCookieName string) *Controllers {
	return &Controllers{
		log:               log,
		services:          services,
		validate:          validate,
		frontendDomain:    frontendDomain,
		backendDomain:     backendDomain,
		refreshCookieName: refreshCookieName,
	}
}
