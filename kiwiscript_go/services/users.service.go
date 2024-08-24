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
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/utils"
	"strings"
)

type CreateUserOptions struct {
	FirstName string
	LastName  string
	Location  string
	Email     string
	Password  string
	Provider  string
}

func (s *Services) CreateUser(ctx context.Context, options CreateUserOptions) (*db.User, *ServiceError) {
	log := s.log.WithGroup("services.users.CreateUser").With(
		"firstName", options.FirstName,
		"lastName", options.LastName,
		"location", options.Location,
		"provider", options.Provider,
	)
	log.InfoContext(ctx, "Creating user...")
	var provider string
	var password pgtype.Text

	switch options.Provider {
	case utils.ProviderEmail:
		provider = options.Provider
		if err := password.Scan(options.Password); err != nil || options.Password == "" {
			log.WarnContext(ctx, "Password is invalid")
			return nil, NewValidationError("'password' is invalid")
		}
	case utils.ProviderGitHub, utils.ProviderGoogle:
		provider = options.Provider
	default:
		log.ErrorContext(ctx, "Provider must be 'email', 'github' or 'google'", "provider", options.Provider)
		return nil, NewServerError()
	}

	location := strings.ToUpper(options.Location)
	if _, ok := utils.Location[location]; !ok {
		location = utils.LocationOTH
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		return nil, FromDBError(err)
	}
	defer s.database.FinalizeTx(ctx, txn, err, nil)

	var user db.User
	if provider == utils.ProviderEmail {
		user, err = qrs.CreateUserWithPassword(ctx, db.CreateUserWithPasswordParams{
			FirstName: options.FirstName,
			LastName:  options.LastName,
			Email:     options.Email,
			Password:  password,
			Location:  location,
		})
	} else {
		user, err = qrs.CreateUserWithoutPassword(ctx, db.CreateUserWithoutPasswordParams{
			FirstName: options.FirstName,
			LastName:  options.LastName,
			Email:     options.Email,
			Location:  location,
		})
	}
	if err != nil {
		log.ErrorContext(ctx, "Failed to create user", "error", err)
		return nil, FromDBError(err)
	}

	params := db.CreateAuthProviderParams{
		Email:    user.Email,
		Provider: provider,
	}
	if err = qrs.CreateAuthProvider(ctx, params); err != nil {
		log.ErrorContext(ctx, "Failed to create auth provider", "error", err)
		return nil, FromDBError(err)
	}

	log.InfoContext(ctx, "Created user successfully")
	return &user, nil
}

func (s *Services) FindUserByEmail(ctx context.Context, email string) (*db.User, *ServiceError) {
	log := s.log.WithGroup("services.users.FindUserByEmail")
	log.InfoContext(ctx, "Finding user by email...")
	user, err := s.database.FindUserByEmail(ctx, email)

	if err != nil {
		log.WarnContext(ctx, "User not found")
		return nil, FromDBError(err)
	}

	log.InfoContext(ctx, "User found successfully")
	return &user, nil
}

func (s *Services) FindUserByID(ctx context.Context, id int32) (*db.User, *ServiceError) {
	log := s.log.WithGroup("services.users.FindUserByID").With("id", id)
	log.InfoContext(ctx, "Finding user by id...")
	user, err := s.database.FindUserById(ctx, id)

	if err != nil {
		log.WarnContext(ctx, "User not found")
		return nil, FromDBError(err)
	}

	return &user, nil
}

type UpdateUserPasswordOptions struct {
	ID       int32
	Password string
}

func (s *Services) UpdateUserPassword(ctx context.Context, opts UpdateUserPasswordOptions) (*db.User, *ServiceError) {
	log := s.log.WithGroup("services.users.UpdateUserPassword").With("id", opts.ID)
	log.InfoContext(ctx, "Updating user password...")
	var password pgtype.Text
	if err := password.Scan(opts.Password); err != nil || opts.Password == "" {
		log.WarnContext(ctx, "Password is invalid")
		return nil, NewValidationError("'password' is invalid")
	}

	user, err := s.database.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		Password: password,
		ID:       opts.ID,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to update user password")
		return nil, FromDBError(err)
	}

	return &user, nil
}

func (s *Services) ConfirmUser(ctx context.Context, id int32) (*db.User, *ServiceError) {
	log := s.log.WithGroup("services.users.ConfirmUser").With("id", id)
	log.InfoContext(ctx, "Confirming user...")

	user, err := s.database.ConfirmUser(ctx, id)
	if err != nil {
		log.ErrorContext(ctx, "Failed to confirm user")
		return nil, FromDBError(err)
	}

	log.InfoContext(ctx, "User confirmed successfully")
	return &user, nil
}

func (s *Services) DeleteUser(ctx context.Context, id int32) *ServiceError {
	log := s.log.WithGroup("services.users.DeleteUser").With("id", id)
	log.InfoContext(ctx, "Deleting user...")

	err := s.database.DeleteUserById(ctx, id)
	if err != nil {
		log.ErrorContext(ctx, "Failed to delete user", "error", err)
		return FromDBError(err)
	}

	log.InfoContext(ctx, "Deleted user successfully")
	return nil
}
