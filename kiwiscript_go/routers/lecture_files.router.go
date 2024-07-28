package routers

import "github.com/kiwiscript/kiwiscript_go/paths"

const lectureFilesPath = paths.LanguagePathV1 +
	"/:languageSlug" +
	paths.SeriesPath +
	"/:seriesSlug" +
	paths.PartsPath +
	"/:seriesPartID" +
	paths.LecturesPath +
	"/:lectureID" +
	paths.FilesPath

func (r *Router) LectureFilesPublicRoutes() {
	lectureFiles := r.router.Group(lectureFilesPath)

	lectureFiles.Get("/", r.controllers.GetLectureFiles)
	lectureFiles.Get("/:fileID", r.controllers.GetLectureFile)
}

func (r *Router) LectureFilesStaffRoutes() {
	lectureFiles := r.router.Group(
		lectureFilesPath,
		r.controllers.AccessClaimsMiddleware,
		r.controllers.StaffUserMiddleware,
	)

	lectureFiles.Post("/", r.controllers.UploadLectureFile)
	lectureFiles.Put("/:fileID", r.controllers.UpdateLectureFile)
	lectureFiles.Delete("/:fileID", r.controllers.DeleteLectureFile)
}
