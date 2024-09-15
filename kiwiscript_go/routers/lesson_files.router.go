package routers

import "github.com/kiwiscript/kiwiscript_go/paths"

const lessonFilesPath = paths.LanguagePathV1 +
	"/:languageSlug" +
	paths.SeriesPath +
	"/:seriesSlug" +
	paths.SectionsPath +
	"/:sectionID" +
	paths.LessonsPath +
	"/:lessonID" +
	paths.FilesPath

func (r *Router) LessonFilesPublicRoutes() {
	lessonFiles := r.router.Group(lessonFilesPath)

	lessonFiles.Get("/", r.controllers.GetLessonFiles)
	lessonFiles.Get("/:fileID", r.controllers.GetLessonFile)
}

func (r *Router) LessonFilesStaffRoutes() {
	lessonFiles := r.router.Group(
		lessonFilesPath,
		r.controllers.AccessClaimsMiddleware,
		r.controllers.StaffUserMiddleware,
	)

	lessonFiles.Post("/", r.controllers.UploadLessonFile)
	lessonFiles.Put("/:fileID", r.controllers.UpdateLessonFile)
	lessonFiles.Delete("/:fileID", r.controllers.DeleteLessonFile)
}
