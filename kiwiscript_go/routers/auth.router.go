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

func (r *Router) AuthPublicRoutes() {
	auth := r.router.Group(paths.AuthPath)

	auth.Post("/register", r.controllers.SignUp)
	auth.Post("/confirm-email", r.controllers.ConfirmEmail)
	auth.Post("/login", r.controllers.SignIn)
	auth.Post("/login/confirm", r.controllers.ConfirmSignIn)
	auth.Post("/refresh", r.controllers.Refresh)
	auth.Post("/forgot-password", r.controllers.ForgotPassword)
	auth.Post("/reset-password", r.controllers.ResetPassword)
}

func (r *Router) AuthPrivateRoutes() {
	auth := r.router.Group(paths.AuthPath, r.controllers.UserMiddleware)

	auth.Post("/logout", r.controllers.SignOut)
	auth.Post("/update-password", r.controllers.UpdatePassword)
	auth.Post("/update-email", r.controllers.UpdateEmail)
}
