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

const lessonsPath = paths.LanguagePathV1 +
	"/:languageSlug" +
	paths.SeriesPath +
	"/:seriesSlug" +
	paths.SectionsPath +
	"/:sectionID" +
	paths.LessonsPath

func (r *Router) LessonPublicRoutes() {
	lesson := r.router.Group(lessonsPath)

	lesson.Get("/", r.controllers.GetLessons)
	lesson.Get("/:lessonID", r.controllers.GetLesson)
}

func (r *Router) LessonStaffRoutes() {
	lesson := r.router.Group(
		lessonsPath,
		r.controllers.AccessClaimsMiddleware,
		r.controllers.StaffUserMiddleware,
	)

	lesson.Post("/", r.controllers.CreateLesson)
	lesson.Put("/:lessonID", r.controllers.UpdateLesson)
	lesson.Delete("/:lessonID", r.controllers.DeleteLesson)
	lesson.Patch("/:lessonID/publish", r.controllers.UpdateLessonIsPublished)
}
