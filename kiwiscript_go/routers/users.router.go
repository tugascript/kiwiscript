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

const (
	myProfilePath   = paths.MePath + paths.ProfilePath
	myPicturePath   = paths.MePath + paths.PicturePath
	userIDPath      = "/:userID"
	userProfilePath = userIDPath + paths.ProfilePath
	userPicturePath = userIDPath + paths.PicturePath
)

func (r *Router) UsersRoutes() {
	users := r.router.Group(paths.UsersPathV1)

	users.Get(paths.MePath, r.controllers.GetMe)
	users.Put(paths.MePath, r.controllers.UpdateCurrentAccount)
	users.Delete(paths.MePath, r.controllers.DeleteCurrentAccount)

	users.Get(myProfilePath, r.controllers.GetMyProfile)
	users.Post(myProfilePath, r.controllers.CreateUserProfile)
	users.Put(myProfilePath, r.controllers.UpdateUserProfile)
	users.Delete(myProfilePath, r.controllers.DeleteUserProfile)

	users.Get(myPicturePath, r.controllers.GetMyPicture)
	users.Post(myPicturePath, r.controllers.UploadUserPicture)
	users.Delete(myPicturePath, r.controllers.DeleteUserPicture)

	users.Get(userIDPath, r.controllers.GetUser)
	users.Get(userProfilePath, r.controllers.GetUserProfile)
	users.Get(userPicturePath, r.controllers.GetUserPicture)
}
