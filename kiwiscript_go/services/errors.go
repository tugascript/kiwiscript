package services

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	CodeValidation   string = "VALIDATION"
	CodeDuplicateKey string = "DUPLICATE_KEY"
	CodeInvalidEnum  string = "INVALID_ENUM"
	CodeNotFound     string = "NOT_FOUND"
	CodeUnknown      string = "UNKNOWN"
	CodeServerError  string = "SERVER_ERROR"
	CodeUnauthorized string = "UNAUTHORIZED"
	CodeForbidden    string = "FORBIDDEN"
)

const (
	MessageDuplicateKey string = "Resource already exists"
	MessageNotFound     string = "Resource not found"
	MessageUnknown      string = "Something went wrong"
	MessageUnauthorized string = "Unauthorized"
	MessageForbidden    string = "Forbidden"
)

type ServiceError struct {
	Code    string
	Message string
}

func NewError(code string, message string) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
	}
}

func NewValidationError(message string) *ServiceError {
	return NewError(CodeValidation, message)
}

func NewServerError(message string) *ServiceError {
	return NewError(CodeServerError, message)
}

func NewDuplicateKeyError(message string) *ServiceError {
	return NewError(CodeDuplicateKey, message)
}

func NewUnauthorizedError() *ServiceError {
	return NewError(CodeUnauthorized, MessageUnauthorized)
}

func NewForbiddenError() *ServiceError {
	return NewError(CodeForbidden, MessageForbidden)
}

func (e *ServiceError) Error() string {
	return e.Message
}

func FromDBError(err error) *ServiceError {
	if errors.Is(err, pgx.ErrNoRows) {
		return NewError(CodeNotFound, MessageNotFound)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return NewError(CodeDuplicateKey, MessageDuplicateKey)
		case "23514":
			return NewError(CodeInvalidEnum, pgErr.Message)
		case "23503":
			return NewError(CodeNotFound, MessageNotFound)
		default:
			return NewError(CodeUnknown, pgErr.Message)
		}
	}

	return NewError(CodeUnknown, MessageUnknown)
}
