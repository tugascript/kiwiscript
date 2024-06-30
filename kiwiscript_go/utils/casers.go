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
