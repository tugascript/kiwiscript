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

const extAuthPath string = paths.AuthPath + "/ext"

func (r *Router) OAuthPublicRoutes() {
	oauthGroup := r.router.Group(extAuthPath)

	oauthGroup.Get("/github", r.controllers.GitHubSignIn)
	oauthGroup.Get("/github/callback", r.controllers.GitHubCallback)
	oauthGroup.Get("/google", r.controllers.GoogleSignIn)
	oauthGroup.Get("/google/callback", r.controllers.GoogleCallback)
	oauthGroup.Post("/token", r.controllers.OAuthToken)
}
