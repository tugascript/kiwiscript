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

package exceptions

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	CodeValidation   string = "VALIDATION"
	CodeConflict     string = "CONFLICT"
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

func NewNotFoundError() *ServiceError {
	return NewError(CodeNotFound, MessageNotFound)
}

func NewValidationError(message string) *ServiceError {
	return NewError(CodeValidation, message)
}

func NewServerError() *ServiceError {
	return NewError(CodeServerError, MessageUnknown)
}

func NewConflictError(message string) *ServiceError {
	return NewError(CodeConflict, message)
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
			return NewError(CodeConflict, MessageDuplicateKey)
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
