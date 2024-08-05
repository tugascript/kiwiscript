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

import "github.com/kiwiscript/kiwiscript_go/paths"

const sectionPath = paths.LanguagePathV1 +
	"/:languageSlug" +
	paths.SeriesPath +
	"/:seriesSlug" +
	paths.SectionsPath

func (r *Router) SectionPublicRoutes() {
	section := r.router.Group(sectionPath)

	section.Get("/", r.controllers.GetSections)
	section.Get("/:sectionID", r.controllers.GetSection)
}

func (r *Router) SectionStaffRoutes() {
	section := r.router.Group(
		sectionPath,
		r.controllers.AccessClaimsMiddleware,
		r.controllers.StaffUserMiddleware,
	)

	section.Post("/", r.controllers.CreateSection)
	section.Put("/:sectionID", r.controllers.UpdateSection)
	section.Delete("/:sectionID", r.controllers.DeleteSection)
	section.Patch("/:sectionID/publish", r.controllers.UpdateSectionIsPublished)
}
