package controllers

import "github.com/kiwiscript/kiwiscript_go/services"

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
