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

type CreateLanguageOptions struct {
	UserID int32
	Name   string
	Icon   string
}

func (s *Services) FindLanguageBySlug(ctx context.Context, slug string) (*db.Language, *ServiceError) {
	language, err := s.database.FindLanguageBySlug(ctx, slug)

	if err != nil {
		return nil, FromDBError(err)
	}

	return &language, nil
}

type FindLanguageWithProgressBySlugOptions struct {
	UserID       int32
	LanguageSlug string
}

func (s *Services) FindLanguageWithProgressBySlug(ctx context.Context, opts FindLanguageProgressOptions) (*db.FindLanguageBySlugWithLanguageProgressRow, *ServiceError) {
	language, err := s.database.FindLanguageBySlugWithLanguageProgress(ctx, db.FindLanguageBySlugWithLanguageProgressParams{
		UserID: opts.UserID,
		Slug:   opts.LanguageSlug,
	})

	if err != nil {
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

func (s *Services) CreateLanguage(ctx context.Context, options CreateLanguageOptions) (*db.Language, *ServiceError) {
	log := s.log.WithGroup("service.languages.CreateLanguage")
	log.InfoContext(ctx, "create language", "name", options.Name)
	slug := utils.Slugify(options.Name)

	if _, serviceErr := s.FindLanguageBySlug(ctx, slug); serviceErr == nil {
		log.InfoContext(ctx, "language already exists", "slug", slug)
		return nil, NewValidationError("language already exists")
	}

	language, err := s.database.CreateLanguage(ctx, db.CreateLanguageParams{
		AuthorID: options.UserID,
		Name:     options.Name,
		Icon:     options.Icon,
		Slug:     slug,
	})
	if err != nil {
		return nil, FromDBError(err)
	}

	return &language, nil
}

type UpdateLanguageOptions struct {
	Slug string
	Name string
	Icon string
}

func (s *Services) UpdateLanguage(ctx context.Context, options UpdateLanguageOptions) (*db.Language, *ServiceError) {
	log := s.log.WithGroup("service.languages.UpdateLanguage")
	log.InfoContext(ctx, "update language", "slug", options.Slug, "name", options.Name)
	language, serviceErr := s.FindLanguageBySlug(ctx, options.Slug)

	if serviceErr != nil {
		log.InfoContext(ctx, "language not found", "slug", options.Slug)
		return nil, serviceErr
	}

	slug := utils.Slugify(options.Name)
	if _, serviceErr := s.FindLanguageBySlug(ctx, slug); serviceErr == nil {
		log.InfoContext(ctx, "language already exists", "slug", slug)
		return nil, NewValidationError("language already exists")
	}

	updateLanguage, err := s.database.UpdateLanguage(ctx, db.UpdateLanguageParams{
		ID:   language.ID,
		Name: options.Name,
		Icon: options.Icon,
		Slug: slug,
	})
	if err != nil {
		return nil, FromDBError(err)
	}

	return &updateLanguage, nil
}

func (s *Services) DeleteLanguage(ctx context.Context, slug string) *ServiceError {
	log := s.log.WithGroup("service.languages.DeleteLanguage").With("languageSlug", slug)
	log.InfoContext(ctx, "Deleting language...")

	language, serviceErr := s.FindLanguageBySlug(ctx, slug)
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
		return NewDuplicateKeyError("Language has students")
	}

	if err := s.database.DeleteLanguageById(ctx, language.ID); err != nil {
		log.ErrorContext(ctx, "failed to delete language", "error", err)
		return FromDBError(err)
	}

	return nil
}

type FindPaginatedLanguagesOptions struct {
	Search string
	Offset int32
	Limit  int32
}

func (s *Services) FindPaginatedLanguages(ctx context.Context, options FindPaginatedLanguagesOptions) ([]db.Language, int64, *ServiceError) {
	log := s.log.WithGroup("service.languages.GetLanguages")
	log.InfoContext(ctx, "get languages")

	if options.Search == "" {
		languages, err := s.database.FindPaginatedLanguages(ctx, db.FindPaginatedLanguagesParams{
			Offset: options.Offset,
			Limit:  options.Limit,
		})
		if err != nil {
			return nil, 0, FromDBError(err)
		}

		count, err := s.database.CountLanguages(ctx)
		if err != nil {
			return nil, 0, FromDBError(err)
		}

		return languages, count, nil
	}

	search := utils.DbSearch(options.Search)
	languages, err := s.database.FindFilteredPaginatedLanguages(ctx, db.FindFilteredPaginatedLanguagesParams{
		Name:   search,
		Offset: options.Offset,
		Limit:  options.Limit,
	})
	if err != nil {
		return nil, 0, FromDBError(err)
	}

	count, err := s.database.CountFilteredLanguages(ctx, search)
	if err != nil {
		return nil, 0, FromDBError(err)
	}

	return languages, count, nil
}

type FindPaginatedLanguagesWithProgressOptions struct {
	UserID int32
	Offset int32
	Limit  int32
}

func (s *Services) FindPaginatedLanguagesWithProgress(
	ctx context.Context,
	opts FindPaginatedLanguagesWithProgressOptions,
) ([]db.FindPaginatedLanguagesWithLanguageProgressRow, int64, *ServiceError) {
	log := s.log.WithGroup("service.languages.FindPaginatedLanguagesWithProgress")
	log.InfoContext(ctx, "Finding paginated languages with progress...")

	languages, err := s.database.FindPaginatedLanguagesWithLanguageProgress(
		ctx,
		db.FindPaginatedLanguagesWithLanguageProgressParams{
			UserID: opts.UserID,
			Offset: opts.Offset,
			Limit:  opts.Limit,
		},
	)
	if err != nil {
		return nil, 0, FromDBError(err)
	}

	count, err := s.database.CountLanguages(ctx)
	if err != nil {
		return nil, 0, FromDBError(err)
	}

	return languages, count, nil
}

type FindFilteredPaginatedLanguagesWithProgressOptions struct {
	UserID int32
	Search string
	Offset int32
	Limit  int32
}

func (s *Services) FindFilteredPaginatedLanguagesWithProgress(
	ctx context.Context,
	opts FindFilteredPaginatedLanguagesWithProgressOptions,
) ([]db.FindFilteredPaginatedLanguagesWithLanguageProgressRow, int64, *ServiceError) {
	log := s.log.WithGroup("service.languages.FindFilteredPaginatedLanguagesWithProgress")
	log.InfoContext(ctx, "Finding filtered paginated languages with progress...")

	search := utils.DbSearch(opts.Search)
	languages, err := s.database.FindFilteredPaginatedLanguagesWithLanguageProgress(
		ctx,
		db.FindFilteredPaginatedLanguagesWithLanguageProgressParams{
			UserID: opts.UserID,
			Name:   search,
			Offset: opts.Offset,
		},
	)
	if err != nil {
		return nil, 0, FromDBError(err)
	}

	count, err := s.database.CountFilteredLanguages(ctx, search)
	if err != nil {
		return nil, 0, FromDBError(err)
	}

	return languages, count, nil
}
