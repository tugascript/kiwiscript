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

package utils

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var titleCaser cases.Caser = cases.Title(language.English)
var lowerCaser cases.Caser = cases.Lower(language.English)
var upperCaser cases.Caser = cases.Upper(language.English)

func Capitalized(s string) string {
	s = strings.TrimSpace(s)

	if len(s) == 0 {
		return s
	}

	return titleCaser.String(s)
}

func Lowered(s string) string {
	s = strings.TrimSpace(s)

	if len(s) == 0 {
		return s
	}

	return lowerCaser.String(s)
}

func Uppercased(s string) string {
	s = strings.TrimSpace(s)

	if len(s) == 0 {
		return s
	}

	return upperCaser.String(s)
}

func DbSearch(s string) string {
	return "%" + Lowered(s) + "%"
}

func Slugify(s string) string {
	s = Lowered(s)
	s = strings.Join(strings.Fields(s), "-")
	return s
}
