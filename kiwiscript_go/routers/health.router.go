package routers

const healthPath = "/health"

func (r *Router) HealthRoutes() {
	health := r.router.Group(healthPath)

	health.Get("/", r.controllers.HealthCheck)
}
