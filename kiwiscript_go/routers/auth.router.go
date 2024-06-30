package routers

const authPath = "/auth"

func (r *Router) AuthPublicRoutes() {
	auth := r.router.Group(authPath)

	auth.Post("/register", r.controllers.SignUp)
	auth.Post("/login", r.controllers.SignIn)
	auth.Post("/login/confirm", r.controllers.ConfirmSignIn)
	auth.Post("/confirm-email", r.controllers.ConfirmEmail)
	auth.Post("/forgot-password", r.controllers.ForgotPassword)
	auth.Post("/reset-password", r.controllers.ResetPassword)
}

func (r *Router) AuthPrivateRoutes() {
	auth := r.router.Group(authPath, r.controllers.AccessClaimsMiddleware)

	auth.Post("/logout", r.controllers.SignOut)
	auth.Post("/refresh", r.controllers.Refresh)
	auth.Post("/update-password", r.controllers.UpdatePassword)
	auth.Post("/update-email", r.controllers.UpdateEmail)
}
