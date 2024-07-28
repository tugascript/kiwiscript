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

const lectureVideoPath = paths.LanguagePathV1 +
	"/:languageSlug" +
	paths.SeriesPath +
	"/:seriesSlug" +
	paths.PartsPath +
	"/:seriesPartID" +
	paths.LecturesPath +
	"/:lectureID" +
	paths.VideoPath

func (r *Router) LectureVideoPublicRoutes() {
	lectureVideo := r.router.Group(lectureVideoPath)

	lectureVideo.Get("/", r.controllers.GetLectureVideo)
}

func (r *Router) LectureVideoStaffRoutes() {
	lectureVideo := r.router.Group(
		lectureVideoPath,
		r.controllers.AccessClaimsMiddleware,
		r.controllers.StaffUserMiddleware,
	)

	lectureVideo.Post("/", r.controllers.CreateLectureVideo)
	lectureVideo.Put("/", r.controllers.UpdateLectureVideo)
	lectureVideo.Delete("/", r.controllers.DeleteLectureVideo)
}
