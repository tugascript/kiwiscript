package app

import (
	"bytes"
	"encoding/xml"
	"github.com/yuin/goldmark"
	"io"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

const svgValidatorTag string = "svg"

// Check if the input is a valid SVG
func isValidSVG(fl validator.FieldLevel) bool {
	input, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	decoder := xml.NewDecoder(strings.NewReader(input))

	for {
		t, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false
		}

		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "svg" {
				return true
			}
			return false
		}
	}

	return false
}

const extAlphaNumTag string = "extalphanum"

func isValidExtAlphaNum(fl validator.FieldLevel) bool {
	input, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	re, err := regexp.Compile(`^[a-zA-Z0-9 #+]+$`)
	if err != nil {
		return false
	}

	return re.MatchString(input)
}

const slugValidatorTag string = "slug"

func isValidSlug(fl validator.FieldLevel) bool {
	input, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	re, err := regexp.Compile(`^[a-z\d]+(?:(\.|-)[a-z\d]+)*$`)
	if err != nil {
		return false
	}

	return re.MatchString(input)
}

const markdownValidatorTag string = "markdown"

func isValidMarkdown(fl validator.FieldLevel) bool {
	input, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	var buf bytes.Buffer
	md := goldmark.New()
	err := md.Convert([]byte(input), &buf)
	return err == nil
}
