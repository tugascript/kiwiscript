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

	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/utils"
)

const languagesLocation string = "languages"

func (s *Services) FindLanguageBySlug(ctx context.Context, slug string) (*db.Language, *ServiceError) {
	language, err := s.database.FindLanguageBySlug(ctx, slug)

	if err != nil {
		return nil, FromDBError(err)
	}

	return &language, nil
}

type FindLanguageWithProgressBySlugOptions struct {
	RequestID    string
	UserID       int32
	LanguageSlug string
}

func (s *Services) FindLanguageWithProgressBySlug(
	ctx context.Context,
	opts FindLanguageWithProgressBySlugOptions,
) (*db.FindLanguageBySlugWithLanguageProgressRow, *ServiceError) {
	log := s.buildLogger(opts.RequestID, languagesLocation, "FindLanguageWithProgressBySlug").With(
		"userId", opts.UserID,
		"slug", opts.LanguageSlug,
	)
	log.InfoContext(ctx, "Finding language with progress by slug...")

	language, err := s.database.FindLanguageBySlugWithLanguageProgress(
		ctx,
		db.FindLanguageBySlugWithLanguageProgressParams{
			UserID: opts.UserID,
			Slug:   opts.LanguageSlug,
		},
	)
	if err != nil {
		log.WarnContext(ctx, "Language not found", "error", err)
		return nil, FromDBError(err)
	}

	return &language, nil
}

func (s *Services) FindLanguageByID(ctx context.Context, id int32) (db.Language, *ServiceError) {
	language, err := s.database.FindLanguageById(ctx, id)

	if err != nil {
		return language, FromDBError(err)
	}

	return language, nil
}

type CreateLanguageOptions struct {
	RequestID string
	UserID    int32
	Name      string
	Icon      string
}

func (s *Services) CreateLanguage(ctx context.Context, opts CreateLanguageOptions) (*db.Language, *ServiceError) {
	log := s.buildLogger(opts.RequestID, languagesLocation, "CreateLanguage").With(
		"userId", opts.UserID,
		"name", opts.Name,
	)
	log.InfoContext(ctx, "Creating language...")

	slug := utils.Slugify(opts.Name)
	if _, serviceErr := s.FindLanguageBySlug(ctx, slug); serviceErr == nil {
		log.InfoContext(ctx, "language already exists", "slug", slug)
		return nil, NewValidationError("language already exists")
	}

	language, err := s.database.CreateLanguage(ctx, db.CreateLanguageParams{
		AuthorID: opts.UserID,
		Name:     opts.Name,
		Icon:     opts.Icon,
		Slug:     slug,
	})
	if err != nil {
		return nil, FromDBError(err)
	}

	return &language, nil
}

type UpdateLanguageOptions struct {
	RequestID string
	Slug      string
	Name      string
	Icon      string
}

func (s *Services) UpdateLanguage(ctx context.Context, opts UpdateLanguageOptions) (*db.Language, *ServiceError) {
	log := s.buildLogger(opts.RequestID, languagesLocation, "UpdateLanguage").With(
		"slug", opts.Slug,
		"name", opts.Name,
	)
	log.InfoContext(ctx, "Updating language...")
	language, serviceErr := s.FindLanguageBySlug(ctx, opts.Slug)

	if serviceErr != nil {
		log.InfoContext(ctx, "language not found", "slug", opts.Slug)
		return nil, serviceErr
	}

	slug := utils.Slugify(opts.Name)
	if _, serviceErr := s.FindLanguageBySlug(ctx, slug); serviceErr == nil {
		log.InfoContext(ctx, "language already exists", "slug", slug)
		return nil, NewValidationError("language already exists")
	}

	updateLanguage, err := s.database.UpdateLanguage(ctx, db.UpdateLanguageParams{
		ID:   language.ID,
		Name: opts.Name,
		Icon: opts.Icon,
		Slug: slug,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to update language", "error", err)
		return nil, FromDBError(err)
	}

	return &updateLanguage, nil
}

type DeleteLanguageOptions struct {
	RequestID string
	Slug      string
}

func (s *Services) DeleteLanguage(ctx context.Context, opts DeleteLanguageOptions) *ServiceError {
	log := s.buildLogger(opts.RequestID, languagesLocation, "DeleteLanguage").With(
		"slug", opts.Slug,
	)
	log.InfoContext(ctx, "Deleting language...")

	language, serviceErr := s.FindLanguageBySlug(ctx, opts.Slug)
	if serviceErr != nil {
		log.WarnContext(ctx, "Language not found")
		return serviceErr
	}

	progressCount, err := s.database.CountLanguageProgressBySlug(ctx, language.Slug)
	if err != nil {
		log.ErrorContext(ctx, "Failed to count language progress", "error", err)
		return FromDBError(err)
	}

	if progressCount > 0 {
		log.WarnContext(ctx, "Language has students", "studentsCount", progressCount)
		return NewConflictError("Language has students")
	}

	if err := s.database.DeleteLanguageById(ctx, language.ID); err != nil {
		log.ErrorContext(ctx, "failed to delete language", "error", err)
		return FromDBError(err)
	}

	return nil
}

type FindPaginatedLanguagesOptions struct {
	RequestID string
	Search    string
	Offset    int32
	Limit     int32
}

func (s *Services) FindPaginatedLanguages(
	ctx context.Context,
	opts FindPaginatedLanguagesOptions,
) ([]db.Language, int64, *ServiceError) {
	log := s.buildLogger(opts.RequestID, languagesLocation, "GetLanguages").With(
		"search", opts.Search,
		"offset", opts.Offset,
		"limit", opts.Limit,
	)
	log.InfoContext(ctx, "Finding paginated languages...")

	if opts.Search == "" {
		count, err := s.database.CountLanguages(ctx)
		if err != nil {
			log.ErrorContext(ctx, "Failed to count languages", "error", err)
			return nil, 0, FromDBError(err)
		}

		if count == 0 {
			log.DebugContext(ctx, "No languages found", "count", count)
			return make([]db.Language, 0), 0, nil
		}

		languages, err := s.database.FindPaginatedLanguages(ctx, db.FindPaginatedLanguagesParams{
			Offset: opts.Offset,
			Limit:  opts.Limit,
		})
		if err != nil {
			log.ErrorContext(ctx, "Failed to find paginated languages", "error", err)
			return nil, 0, FromDBError(err)
		}

		return languages, count, nil
	}

	search := utils.DbSearch(opts.Search)
	count, err := s.database.CountFilteredLanguages(ctx, search)
	if err != nil {
		log.ErrorContext(ctx, "Failed to count languages", "error", err)
		return nil, 0, FromDBError(err)
	}

	if count == 0 {
		log.DebugContext(ctx, "No languages found", "count", count)
		return make([]db.Language, 0), 0, nil
	}

	languages, err := s.database.FindFilteredPaginatedLanguages(ctx, db.FindFilteredPaginatedLanguagesParams{
		Name:   search,
		Offset: opts.Offset,
		Limit:  opts.Limit,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to find filtered languages", "error", err)
		return nil, 0, FromDBError(err)
	}

	return languages, count, nil
}

type FindPaginatedLanguagesWithProgressOptions struct {
	RequestID string
	UserID    int32
	Offset    int32
	Limit     int32
}

func (s *Services) FindPaginatedLanguagesWithProgress(
	ctx context.Context,
	opts FindPaginatedLanguagesWithProgressOptions,
) ([]db.FindPaginatedLanguagesWithLanguageProgressRow, int64, *ServiceError) {
	log := s.buildLogger(opts.RequestID, languagesLocation, "FindPaginatedLanguagesWithProgress").With(
		"userId", opts.UserID,
		"offset", opts.Offset,
		"limit", opts.Limit,
	)
	log.InfoContext(ctx, "Finding paginated languages with progress...")

	count, err := s.database.CountLanguages(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to count languages", "error", err)
		return nil, 0, FromDBError(err)
	}

	if count == 0 {
		log.DebugContext(ctx, "No languages found", "count", count)
		return make([]db.FindPaginatedLanguagesWithLanguageProgressRow, 0), 0, nil
	}

	languages, err := s.database.FindPaginatedLanguagesWithLanguageProgress(
		ctx,
		db.FindPaginatedLanguagesWithLanguageProgressParams{
			UserID: opts.UserID,
			Offset: opts.Offset,
			Limit:  opts.Limit,
		},
	)
	if err != nil {
		log.ErrorContext(ctx, "Failed to find languages", "error", err)
		return nil, 0, FromDBError(err)
	}

	return languages, count, nil
}

type FindFilteredPaginatedLanguagesWithProgressOptions struct {
	RequestID string
	UserID    int32
	Search    string
	Offset    int32
	Limit     int32
}

func (s *Services) FindFilteredPaginatedLanguagesWithProgress(
	ctx context.Context,
	opts FindFilteredPaginatedLanguagesWithProgressOptions,
) ([]db.FindFilteredPaginatedLanguagesWithLanguageProgressRow, int64, *ServiceError) {
	log := s.buildLogger(opts.RequestID, languagesLocation, "FindFilteredPaginatedLanguagesWithProgress").With(
		"userId", opts.UserID,
		"search", opts.Search,
		"offset", opts.Offset,
		"limit", opts.Limit,
	)
	log.InfoContext(ctx, "Finding filtered paginated languages with progress...")

	search := utils.DbSearch(opts.Search)
	count, err := s.database.CountFilteredLanguages(ctx, search)
	if err != nil {
		log.ErrorContext(ctx, "Failed count languages", "error", err)
		return nil, 0, FromDBError(err)
	}

	if count == 0 {
		log.DebugContext(ctx, "No languages found", "count", count)
		return make([]db.FindFilteredPaginatedLanguagesWithLanguageProgressRow, 0), 0, nil
	}

	languages, err := s.database.FindFilteredPaginatedLanguagesWithLanguageProgress(
		ctx,
		db.FindFilteredPaginatedLanguagesWithLanguageProgressParams{
			UserID: opts.UserID,
			Name:   search,
			Offset: opts.Offset,
			Limit:  opts.Limit,
		},
	)
	if err != nil {
		log.ErrorContext(ctx, "Failed to find languages", "error", err)
		return nil, 0, FromDBError(err)
	}

	return languages, count, nil
}
