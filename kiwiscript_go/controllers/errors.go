package controllers

import (
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
	Field string `json:"field"`
	Error string `json:"error"`
	Tag   string `json:"tag"`
}

type RequestValidationError struct {
	Code     string       `json:"code"`
	Message  string       `json:"message"`
	Location string       `json:"location"`
	Fields   []FieldError `json:"fields"`
}

const (
	RequestValidationMessage       string = "Invalid request"
	RequestValidationLocationBody  string = "body"
	RequestValidationLocationQuery string = "query"
	RequestValidationLocationParam string = "param"
)

func RequestValidationErrorFromErr(err *validator.ValidationErrors, location string) RequestValidationError {
	fields := make([]FieldError, len(*err))

	for i, field := range *err {
		fields[i] = FieldError{
			Field: field.Field(),
			Error: field.Error(),
			Tag:   field.Tag(),
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
