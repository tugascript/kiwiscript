package routers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/kiwiscript/kiwiscript_go/controllers"
)

type Router struct {
	router      fiber.Router
	controllers *controllers.Controllers
}

func NewRouter(app *fiber.App, controllers *controllers.Controllers) *Router {

	return &Router{
		router:      app.Group("/api"),
		controllers: controllers,
	}
}
