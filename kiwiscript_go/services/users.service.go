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
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kiwiscript/kiwiscript_go/exceptions"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/utils"
	"strings"
)

const usersLocation string = "users"

type CreateUserOptions struct {
	RequestID string
	FirstName string
	LastName  string
	Location  string
	Email     string
	Password  string
	Provider  string
}

func (s *Services) CreateUser(ctx context.Context, opts CreateUserOptions) (*db.User, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, usersLocation, "CreateUser").With(
		"firstName", opts.FirstName,
		"lastName", opts.LastName,
		"location", opts.Location,
		"provider", opts.Provider,
	)
	log.InfoContext(ctx, "Creating user...")
	var provider string
	var password pgtype.Text

	switch opts.Provider {
	case utils.ProviderEmail:
		provider = opts.Provider
		if err := password.Scan(opts.Password); err != nil || opts.Password == "" {
			log.WarnContext(ctx, "Password is invalid")
			return nil, exceptions.NewValidationError("'password' is invalid")
		}
	case utils.ProviderGitHub, utils.ProviderGoogle:
		provider = opts.Provider
	default:
		log.ErrorContext(ctx, "Provider must be 'email', 'github' or 'google'", "provider", opts.Provider)
		return nil, exceptions.NewServerError()
	}

	location := strings.ToUpper(opts.Location)
	if _, ok := utils.Location[location]; !ok {
		location = utils.LocationOTH
	}

	var serviceErr *exceptions.ServiceError
	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return nil, exceptions.FromDBError(err)
	}
	defer func() {
		log.DebugContext(ctx, "Finalizing transaction")
		s.database.FinalizeTx(ctx, txn, err, serviceErr)
	}()

	var user db.User
	if provider == utils.ProviderEmail {
		user, err = qrs.CreateUserWithPassword(ctx, db.CreateUserWithPasswordParams{
			FirstName: opts.FirstName,
			LastName:  opts.LastName,
			Email:     opts.Email,
			Password:  password,
			Location:  location,
		})
	} else {
		user, err = qrs.CreateUserWithoutPassword(ctx, db.CreateUserWithoutPasswordParams{
			FirstName: opts.FirstName,
			LastName:  opts.LastName,
			Email:     opts.Email,
			Location:  location,
		})
	}
	if err != nil {
		log.ErrorContext(ctx, "Failed to create user", "error", err)
		serviceErr = exceptions.FromDBError(err)
		return nil, serviceErr
	}

	params := db.CreateAuthProviderParams{
		Email:    user.Email,
		Provider: provider,
	}
	if err := qrs.CreateAuthProvider(ctx, params); err != nil {
		log.ErrorContext(ctx, "Failed to create auth provider", "error", err)
		serviceErr = exceptions.FromDBError(err)
		return nil, serviceErr
	}

	log.InfoContext(ctx, "Created user successfully")
	return &user, nil
}

type FindUserByEmailOptions struct {
	RequestID string
	Email     string
}

func (s *Services) FindUserByEmail(ctx context.Context, opts FindUserByEmailOptions) (*db.User, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, usersLocation, "FindUserByEmail")
	log.InfoContext(ctx, "Finding user by email...")

	user, err := s.database.FindUserByEmail(ctx, opts.Email)
	if err != nil {
		log.WarnContext(ctx, "User not found")
		return nil, exceptions.FromDBError(err)
	}

	log.InfoContext(ctx, "User found successfully")
	return &user, nil
}

type FindUserByIDOptions struct {
	RequestID string
	ID        int32
}

func (s *Services) FindUserByID(ctx context.Context, opts FindUserByIDOptions) (*db.User, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, usersLocation, "FindUserByID").With("id", opts.ID)
	log.InfoContext(ctx, "Finding user by id...")

	user, err := s.database.FindUserById(ctx, opts.ID)
	if err != nil {
		log.WarnContext(ctx, "User not found")
		return nil, exceptions.FromDBError(err)
	}

	return &user, nil
}

type UpdateUserPasswordOptions struct {
	RequestID string
	ID        int32
	Password  string
}

func (s *Services) UpdateUserPassword(ctx context.Context, opts UpdateUserPasswordOptions) (*db.User, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, usersLocation, "UpdateUserPassword").With("id", opts.ID)
	log.InfoContext(ctx, "Updating user password...")
	var password pgtype.Text
	if err := password.Scan(opts.Password); err != nil || opts.Password == "" {
		log.WarnContext(ctx, "Password is invalid")
		return nil, exceptions.NewValidationError("'password' is invalid")
	}

	user, err := s.database.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		Password: password,
		ID:       opts.ID,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to update user password")
		return nil, exceptions.FromDBError(err)
	}

	return &user, nil
}

type ConfirmUserOptions struct {
	RequestID string
	ID        int32
}

func (s *Services) ConfirmUser(ctx context.Context, opts ConfirmUserOptions) (*db.User, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, usersLocation, "ConfirmUser").With("id", opts.ID)
	log.InfoContext(ctx, "Confirming user...")

	user, err := s.database.ConfirmUser(ctx, opts.ID)
	if err != nil {
		log.ErrorContext(ctx, "Failed to confirm user")
		return nil, exceptions.FromDBError(err)
	}

	log.InfoContext(ctx, "User confirmed successfully")
	return &user, nil
}

type UpdateUserOptions struct {
	RequestID string
	ID        int32
	FirstName string
	LastName  string
	Location  string
}

func (s *Services) UpdateUser(ctx context.Context, opts UpdateUserOptions) (*db.User, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, usersLocation, "UpdateUser").With(
		"id", opts.ID,
		"firstName", opts.FirstName,
		"lastName", opts.LastName,
		"location", opts.Location,
	)
	log.InfoContext(ctx, "Updating user...")

	user, serviceErr := s.FindUserByID(ctx, FindUserByIDOptions{
		RequestID: opts.RequestID,
		ID:        opts.ID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	var err error
	*user, err = s.database.UpdateUser(ctx, db.UpdateUserParams{
		ID:        opts.ID,
		FirstName: opts.FirstName,
		LastName:  opts.LastName,
		Location:  opts.Location,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to update user")
		return nil, exceptions.FromDBError(err)
	}

	log.InfoContext(ctx, "Updated user successfully")
	return user, nil
}

type DeleteUserOptions struct {
	RequestID string
	ID        int32
	Password  string
}

func (s *Services) DeleteUser(ctx context.Context, opts DeleteUserOptions) *exceptions.ServiceError {
	log := s.buildLogger(opts.RequestID, usersLocation, "DeleteUser").With("id", opts.ID)
	log.InfoContext(ctx, "Deleting user...")

	user, serviceErr := s.FindUserByID(ctx, FindUserByIDOptions{
		RequestID: opts.RequestID,
		ID:        opts.ID,
	})
	if serviceErr != nil {
		return serviceErr
	}

	if user.Password.Valid {
		if !utils.VerifyPassword(opts.Password, user.Password.String) {
			log.WarnContext(ctx, "Password does not match")
			return exceptions.NewValidationError("'password' does not match")
		}
	}

	if err := s.database.DeleteUserById(ctx, opts.ID); err != nil {
		log.ErrorContext(ctx, "Failed to delete user", "error", err)
		return exceptions.FromDBError(err)
	}

	log.InfoContext(ctx, "Deleted user successfully")
	return nil
}
