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

package services

import (
	"log/slog"

	"github.com/kiwiscript/kiwiscript_go/utils"
)

func (s *Services) buildLogger(requestID, location, function string) *slog.Logger {
	return utils.BuildLogger(s.log, utils.LoggerOptions{
		Layer:     utils.ServicesLogLayer,
		Location:  location,
		Function:  function,
		RequestID: requestID,
	})
}