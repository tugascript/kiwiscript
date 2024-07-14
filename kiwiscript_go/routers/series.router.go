package routers

import "github.com/kiwiscript/kiwiscript_go/paths"

const seriesPath = paths.LanguagePathV1 + ":languageSlug" + paths.SeriesPath

func (r *Router) SeriesPublicRoutes() {
	series := r.router.Group(seriesPath)

	series.Get("/", r.controllers.GetPaginatedSeries)
	series.Get("/:seriesSlug", r.controllers.GetSingleSeries)
}

func (r *Router) SeriesPrivateRoutes() {
	series := r.router.Group(
		seriesPath,
		r.controllers.AccessClaimsMiddleware,
		r.controllers.StaffUserMiddleware,
	)

	series.Post("/", r.controllers.CreateSeries)
	series.Put("/:seriesSlug", r.controllers.UpdateSeries)
	series.Delete("/:seriesSlug", r.controllers.DeleteSeries)
}
