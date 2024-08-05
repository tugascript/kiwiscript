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

const lessonVideoPath = paths.LanguagePathV1 +
	"/:languageSlug" +
	paths.SeriesPath +
	"/:seriesSlug" +
	paths.SectionsPath +
	"/:seriesPartID" +
	paths.LessonsPath +
	"/:lessonID" +
	paths.VideoPath

func (r *Router) LessonVideoPublicRoutes() {
	lessonVideo := r.router.Group(lessonVideoPath)

	lessonVideo.Get("/", r.controllers.GetLessonVideo)
}

func (r *Router) LessonVideoStaffRoutes() {
	lessonVideo := r.router.Group(
		lessonVideoPath,
		r.controllers.AccessClaimsMiddleware,
		r.controllers.StaffUserMiddleware,
	)

	lessonVideo.Post("/", r.controllers.CreateLessonVideo)
	lessonVideo.Put("/", r.controllers.UpdateLessonVideo)
	lessonVideo.Delete("/", r.controllers.DeleteLessonVideo)
}
