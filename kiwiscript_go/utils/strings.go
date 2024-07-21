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
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var titleCaser = cases.Title(language.English)
var lowerCaser = cases.Lower(language.English)
var upperCaser = cases.Upper(language.English)

func Capitalized(s string) string {
	s = strings.TrimSpace(s)

	if len(s) == 0 {
		return s
	}

	return titleCaser.String(s)
}

func CapitalizedFirst(s string) string {
	s = strings.TrimSpace(s)

	if len(s) == 0 {
		return s
	}

	var formated strings.Builder
	for i, char := range s {
		if i == 0 {
			formated.WriteRune(unicode.ToUpper(char))
		} else {
			formated.WriteRune(char)
		}
	}

	return formated.String()
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
	s = strings.ReplaceAll(s, "+", " plus ")
	s = strings.ReplaceAll(s, "#", " sharp ")
	re := regexp.MustCompile(`[^a-zA-Z0-9\s]+`)
	s = re.ReplaceAllString(s, "")
	s = Lowered(s)
	s = strings.Join(strings.Fields(s), "-")
	return s
}
