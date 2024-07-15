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

package controllers

import (
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/kiwiscript/kiwiscript_go/services"
)

const (
	StatusDuplicateKey string = "Conflict"
	StatusInvalidEnum  string = "BadRequest"
	StatusNotFound     string = "NotFound"
	StatusUnknown      string = "InternalServerError"
	StatusUnauthorized string = "Unauthorized"
	StatusForbidden    string = "Forbidden"
	StatusValidation   string = "Validation"
)

type RequestError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewRequestError(err *services.ServiceError) RequestError {
	switch err.Code {
	case services.CodeDuplicateKey:
		return RequestError{
			Code:    StatusDuplicateKey,
			Message: err.Message,
		}
	case services.CodeInvalidEnum:
		return RequestError{
			Code:    StatusInvalidEnum,
			Message: err.Message,
		}
	case services.CodeNotFound:
		return RequestError{
			Code:    StatusNotFound,
			Message: err.Message,
		}
	case services.CodeValidation:
		return RequestError{
			Code:    StatusValidation,
			Message: err.Message,
		}
	case services.CodeUnknown:
		return RequestError{
			Code:    StatusUnknown,
			Message: StatusUnknown,
		}
	case services.CodeUnauthorized:
		return RequestError{
			Code:    StatusUnauthorized,
			Message: StatusUnauthorized,
		}
	case services.CodeForbidden:
		return RequestError{
			Code:    StatusForbidden,
			Message: StatusForbidden,
		}
	default:
		return RequestError{
			Code:    StatusUnknown,
			Message: err.Message,
		}
	}
}

type FieldError struct {
	Param   string      `json:"param"`
	Message string      `json:"message"`
	Value   interface{} `json:"value"`
}

type RequestValidationError struct {
	Code     string       `json:"code"`
	Message  string       `json:"message"`
	Location string       `json:"location"`
	Fields   []FieldError `json:"fields"`
}

const (
	RequestValidationMessage        string = "Invalid request"
	RequestValidationLocationBody   string = "body"
	RequestValidationLocationQuery  string = "query"
	RequestValidationLocationParams string = "params"
)

func toSnakeCase(camel string) string {
	var result strings.Builder
	for i, char := range camel {
		if unicode.IsUpper(char) {
			// Add an underscore before uppercase letters (except the first letter)
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(char))
		} else {
			result.WriteRune(char)
		}
	}
	return result.String()
}

const (
	fieldErrTagEqField  string = "eqfield"
	fieldErrTagRequired string = "required"

	strFieldErrTagMin      string = "min"
	strFieldErrTagMax      string = "max"
	strFieldErrTagEmail    string = "email"
	strFieldErrTagJWT      string = "jwt"
	strFieldErrTagUrl      string = "url"
	strFieldErrTagSvg      string = "svg"
	strFieldErrExtAlphaNum string = "extalphanum"
	strFieldErrSlug        string = "slug"
	strFieldNumber         string = "number"

	intFieldErrTagGte string = "gte"
	intFieldErrTagLte string = "lte"

	FieldErrMessageInvalid  string = "must be valid"
	FieldErrMessageRequired string = "must be provided"
	FieldErrMessageEqField  string = "does not match equivalent field"

	StrFieldErrMessageEmail       string = "must be a valid email"
	StrFieldErrMessageMin         string = "must be longer"
	StrFieldErrMessageMax         string = "must be shorter"
	StrFieldErrMessageJWT         string = "must be a valid JWT token"
	StrFieldErrMessageUrl         string = "must be a valid URL"
	StrFieldErrMessageSvg         string = "must be a valid SVG"
	StrFieldErrMessageExtAlphaNum string = "must only contain letters, numbers, spaces, plus and hashtags"
	StrFieldErrMessageSlug        string = "must be a valid slug"
	StrFieldErrMessageNumber      string = "must be a number"

	IntFieldErrMessageLte string = "must be less"
	IntFieldErrMessageGte string = "must be greater"
)

func selectStrErrMessage(tag string) string {
	switch tag {
	case fieldErrTagRequired:
		return FieldErrMessageRequired
	case strFieldErrTagEmail:
		return StrFieldErrMessageEmail
	case strFieldErrTagMin:
		return StrFieldErrMessageMin
	case strFieldErrTagMax:
		return StrFieldErrMessageMax
	case fieldErrTagEqField:
		return FieldErrMessageEqField
	case strFieldErrTagJWT:
		return StrFieldErrMessageJWT
	case strFieldErrTagUrl:
		return StrFieldErrMessageUrl
	case strFieldErrTagSvg:
		return StrFieldErrMessageSvg
	case strFieldErrExtAlphaNum:
		return StrFieldErrMessageExtAlphaNum
	case strFieldErrSlug:
		return StrFieldErrMessageSlug
	case strFieldNumber:
		return StrFieldErrMessageNumber
	default:
		return FieldErrMessageInvalid
	}
}

func selectIntErrMessage(tag string) string {
	switch tag {
	case fieldErrTagRequired:
		return FieldErrMessageRequired
	case intFieldErrTagLte:
		return IntFieldErrMessageLte
	case intFieldErrTagGte:
		return IntFieldErrMessageGte
	default:
		return FieldErrMessageInvalid
	}
}

func buildFieldErrorMessage(tag string, val interface{}) string {
	switch val.(type) {
	case string:
		return selectStrErrMessage(tag)
	case int, int32, int64:
		return selectIntErrMessage(tag)
	default:
		return FieldErrMessageInvalid
	}
}

func RequestValidationErrorFromErr(err *validator.ValidationErrors, location string) RequestValidationError {
	fields := make([]FieldError, len(*err))

	for i, field := range *err {
		value := field.Value()
		fields[i] = FieldError{
			Value:   value,
			Param:   toSnakeCase(field.Field()),
			Message: buildFieldErrorMessage(field.Tag(), value),
		}
	}

	return RequestValidationError{
		Code:     StatusValidation,
		Message:  RequestValidationMessage,
		Fields:   fields,
		Location: location,
	}
}

func NewRequestValidationError(location string, fields []FieldError) RequestValidationError {
	return RequestValidationError{
		Code:     StatusValidation,
		Message:  RequestValidationMessage,
		Fields:   fields,
		Location: location,
	}
}

type EmptyRequestValidationError struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Location string `json:"location"`
}

func NewEmptyRequestValidationError(location string) EmptyRequestValidationError {
	return EmptyRequestValidationError{
		Code:     StatusValidation,
		Message:  RequestValidationMessage,
		Location: location,
	}
}

func NewRequestErrorStatus(code string) int {
	switch code {
	case services.CodeDuplicateKey:
		return 409
	case services.CodeInvalidEnum, services.CodeValidation:
		return 400
	case services.CodeNotFound:
		return 404
	case services.CodeForbidden:
		return 403
	case services.CodeUnauthorized:
		return 401
	case services.CodeUnknown:
		return 500
	default:
		return 500
	}
}
