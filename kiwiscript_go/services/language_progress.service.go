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

const languageProgressLocation string = "languageProgress"

type FindLanguageProgressOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
}

func (s *Services) FindLanguageProgressBySlug(ctx context.Context, opts FindLanguageProgressOptions) (*db.LanguageProgress, *exceptions.ServiceError) {
	log := s.log.WithGroup("service.language.FindLanguageProgressByUserIDAndLanguageSlug").With(
		"userID", opts.UserID,
		"languageSlug", opts.LanguageSlug,
	)
	log.InfoContext(ctx, "Finding language progress by user ID and language slug")

	languageProgress, err := s.database.FindLanguageProgressBySlugAndUserID(ctx, db.FindLanguageProgressBySlugAndUserIDParams{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
	})
	if err != nil {
		log.ErrorContext(ctx, "Error finding language progress", "error", err)
		return nil, exceptions.FromDBError(err)
	}

	log.InfoContext(ctx, "Language progress found")
	return &languageProgress, nil
}

type CreateOrUpdateLanguageProgressOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
}

func (s *Services) createLanguageProgress(ctx context.Context, opts CreateOrUpdateLanguageProgressOptions) (*db.LanguageProgress, *exceptions.ServiceError) {
	log := s.log.WithGroup("service.language.CreateLanguageProgress").With("userID", opts.UserID, "languageSlug", opts.LanguageSlug)
	log.InfoContext(ctx, "Creating language progress")

	languageProgress, err := s.database.CreateLanguageProgress(ctx, db.CreateLanguageProgressParams{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
	})
	if err != nil {
		log.ErrorContext(ctx, "Error creating language progress", "error", err)
		return nil, exceptions.FromDBError(err)
	}

	log.InfoContext(ctx, "Language progress created")
	return &languageProgress, nil
}

func (s *Services) CreateOrUpdateLanguageProgress(
	ctx context.Context,
	opts CreateOrUpdateLanguageProgressOptions,
) (*db.Language, *db.LanguageProgress, bool, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, languageProgressLocation, "CreateOrUpdateLanguageProgress").With(
		"userID", opts.UserID,
		"languageSlug", opts.LanguageSlug,
	)
	log.InfoContext(ctx, "Updating language progress...")

	language, serviceErr := s.FindLanguageBySlug(ctx, opts.LanguageSlug)
	if serviceErr != nil {
		log.InfoContext(ctx, "Language not found", "slug", opts.LanguageSlug)
		return nil, nil, false, serviceErr
	}

	languageProgress, serviceErr := s.FindLanguageProgressBySlug(ctx, FindLanguageProgressOptions{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
	})
	if serviceErr != nil {
		languageProgress, serviceErr := s.createLanguageProgress(ctx, opts)
		if serviceErr != nil {
			return nil, nil, false, serviceErr
		}

		return language, languageProgress, true, nil
	}

	if err := s.database.UpdateLanguageProgressViewedAt(ctx, languageProgress.ID); err != nil {
		log.ErrorContext(ctx, "Error updating language progress", "error", err)
		return nil, nil, false, exceptions.FromDBError(err)
	}

	log.InfoContext(ctx, "Language progress updated")
	return language, languageProgress, false, nil
}

type DeleteLanguageProgressOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
}

func (s *Services) DeleteLanguageProgress(ctx context.Context, opts DeleteLanguageProgressOptions) *exceptions.ServiceError {
	log := s.buildLogger(opts.RequestID, languageProgressLocation, "DeleteLanguageProgress").With(
		"userID", opts.UserID,
		"languageSlug", opts.LanguageSlug,
	)
	log.InfoContext(ctx, "Deleting language progress...")

	languageProgress, serviceErr := s.FindLanguageProgressBySlug(ctx, FindLanguageProgressOptions{
		UserID:       opts.UserID,
		LanguageSlug: opts.LanguageSlug,
	})
	if serviceErr != nil {
		log.WarnContext(ctx, "Language progress not found")
		return serviceErr
	}

	if err := s.database.DeleteLanguageProgressByID(ctx, languageProgress.ID); err != nil {
		log.ErrorContext(ctx, "Error deleting language progress", "error", err)
		return exceptions.FromDBError(err)
	}

	log.InfoContext(ctx, "Language progress deleted")
	return nil
}
