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

package routers

const languagesPath = "/v1/languages"

func (r *Router) LanguagePublicRoutes() {
	languages := r.router.Group(languagesPath)

	languages.Get("/", r.controllers.GetLanguages)
	languages.Get("/:languageSlug", r.controllers.GetLanguage)
}

func (r *Router) LanguagePrivateRoutes() {
	languages := r.router.Group(
		languagesPath,
		r.controllers.AccessClaimsMiddleware,
		r.controllers.AdminUserMiddleware,
	)

	languages.Post("/", r.controllers.CreateLanguage)
	languages.Put("/:languageSlug", r.controllers.UpdateLanguage)
	languages.Delete("/:languageSlug", r.controllers.DeleteLanguage)
}
