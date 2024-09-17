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
	"github.com/kiwiscript/kiwiscript_go/exceptions"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
)

const userProfilesLocation string = "user_profiles"

type FindUserProfileOptions struct {
	RequestID string
	UserID    int32
}

func (s *Services) FindUserProfile(
	ctx context.Context,
	opts FindUserProfileOptions,
) (*db.UserProfile, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, userProfilesLocation, "FindUserProfile").With(
		"userID", opts.UserID,
	)
	log.InfoContext(ctx, "Finding user profile...")

	profile, err := s.database.FindUserProfileByUserID(ctx, opts.UserID)
	if err != nil {
		log.WarnContext(ctx, "Failed to find user profile", "error", err)
		return nil, exceptions.FromDBError(err)
	}

	return &profile, nil
}

type UserProfileOptions struct {
	RequestID string
	UserID    int32
	Bio       string
	GitHub    string
	LinkedIn  string
	Website   string
}

func (s *Services) CreateUserProfile(
	ctx context.Context,
	opts UserProfileOptions,
) (*db.UserProfile, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, userProfilesLocation, "CreateUserProfile").With(
		"userID", opts.UserID,
		"github", opts.GitHub,
		"linkedin", opts.LinkedIn,
		"website", opts.Website,
	)
	log.InfoContext(ctx, "Creating user profile...")

	if _, serviceErr := s.FindUserProfile(ctx, FindUserProfileOptions{RequestID: opts.RequestID, UserID: opts.UserID}); serviceErr == nil {
		log.WarnContext(ctx, "User profile already exists")
		return nil, exceptions.NewConflictError("User profile already exists")
	}

	profile, err := s.database.CreateUserProfile(ctx, db.CreateUserProfileParams{
		UserID:   opts.UserID,
		Bio:      opts.Bio,
		Github:   opts.GitHub,
		Linkedin: opts.LinkedIn,
		Website:  opts.Website,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to create user profile", "error", err)
		return nil, exceptions.FromDBError(err)
	}

	return &profile, nil
}

func (s *Services) UpdateUserProfile(
	ctx context.Context,
	opts UserProfileOptions,
) (*db.UserProfile, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, userProfilesLocation, "UpdateUserProfile").With(
		"userID", opts.UserID,
		"github", opts.GitHub,
		"linkedin", opts.LinkedIn,
		"website", opts.Website,
	)
	log.InfoContext(ctx, "Updating user profile...")

	profile, serviceErr := s.FindUserProfile(ctx, FindUserProfileOptions{
		RequestID: opts.RequestID,
		UserID:    opts.UserID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	var err error
	*profile, err = s.database.UpdateUserProfile(ctx, db.UpdateUserProfileParams{
		ID:       profile.ID,
		Bio:      opts.Bio,
		Github:   opts.GitHub,
		Linkedin: opts.LinkedIn,
		Website:  opts.Website,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to update user profile", "error", err)
		return nil, exceptions.FromDBError(err)
	}

	return profile, nil
}

type DeleteUserProfileOptions struct {
	RequestID string
	UserID    int32
}

func (s *Services) DeleteUserProfile(
	ctx context.Context,
	opts DeleteUserProfileOptions,
) *exceptions.ServiceError {
	log := s.buildLogger(opts.RequestID, userProfilesLocation, "DeleteUserProfile").With(
		"userId", opts.UserID,
	)
	log.InfoContext(ctx, "Deleting user profile...")

	profile, serviceErr := s.FindUserProfile(ctx, FindUserProfileOptions{
		RequestID: opts.RequestID,
		UserID:    opts.UserID,
	})
	if serviceErr != nil {
		return serviceErr
	}

	if err := s.database.DeleteUserProfile(ctx, profile.ID); err != nil {
		log.ErrorContext(ctx, "Failed to delete user profile", "error", err)
		return exceptions.FromDBError(err)
	}

	return nil
}
