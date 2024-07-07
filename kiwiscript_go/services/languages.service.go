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

func (s *Services) FindLanguageBySlug(ctx context.Context, slug string) (db.Language, *ServiceError) {
	language, err := s.database.FindLanguageBySlug(ctx, slug)

	if err != nil {
		return language, FromDBError(err)
	}

	return language, nil
}

func (s *Services) FindLanguageByID(ctx context.Context, id int32) (db.Language, *ServiceError) {
	language, err := s.database.FindLanguageById(ctx, id)

	if err != nil {
		return language, FromDBError(err)
	}

	return language, nil
}

func (s *Services) CreateLanguage(ctx context.Context, options CreateLanguageOptions) (db.Language, *ServiceError) {
	log := s.log.WithGroup("service.languages.CreateLanguage")
	log.InfoContext(ctx, "create language", "name", options.Name)
	var language db.Language
	slug := utils.Slugify(options.Name)

	if _, serviceErr := s.FindLanguageBySlug(ctx, slug); serviceErr == nil {
		log.InfoContext(ctx, "language already exists", "slug", slug)
		return language, NewValidationError("language already exists")
	}

	language, err := s.database.CreateLanguage(ctx, db.CreateLanguageParams{
		AuthorID: options.UserID,
		Name:     options.Name,
		Icon:     options.Icon,
		Slug:     slug,
	})
	if err != nil {
		return language, FromDBError(err)
	}

	return language, nil
}

type UpdateLanguageOptions struct {
	Slug string
	Name string
	Icon string
}

func (s *Services) UpdateLanguage(ctx context.Context, options UpdateLanguageOptions) (db.Language, *ServiceError) {
	log := s.log.WithGroup("service.languages.UpdateLanguage")
	log.InfoContext(ctx, "update language", "slug", options.Slug, "name", options.Name)
	language, serviceErr := s.FindLanguageBySlug(ctx, options.Slug)

	if serviceErr != nil {
		log.InfoContext(ctx, "language not found", "slug", options.Slug)
		return language, serviceErr
	}

	slug := utils.Slugify(options.Name)
	if _, serviceErr := s.FindLanguageBySlug(ctx, slug); serviceErr == nil {
		log.InfoContext(ctx, "language already exists", "slug", slug)
		return language, NewValidationError("language already exists")
	}

	language, err := s.database.UpdateLanguage(ctx, db.UpdateLanguageParams{
		ID:   language.ID,
		Name: options.Name,
		Icon: options.Icon,
		Slug: slug,
	})
	if err != nil {
		return language, FromDBError(err)
	}

	return language, nil
}

func (s *Services) DeleteLanguage(ctx context.Context, slug string) *ServiceError {
	log := s.log.WithGroup("service.languages.DeleteLanguage")
	log.InfoContext(ctx, "delete language", "slug", slug)
	language, serviceErr := s.FindLanguageBySlug(ctx, slug)

	// TODO add restrictions to see if the language is used in any course
	if serviceErr != nil {
		log.InfoContext(ctx, "language not found", "slug", slug)
		return serviceErr
	}

	err := s.database.DeleteLanguageById(ctx, language.ID)
	if err != nil {
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
	var count int64

	if options.Search == "" {
		languages, err := s.database.FindPaginatedLanguages(ctx, db.FindPaginatedLanguagesParams{
			Offset: options.Offset,
			Limit:  options.Limit,
		})
		if err != nil {
			return languages, count, FromDBError(err)
		}

		count, err = s.database.CountLanguages(ctx)
		if err != nil {
			return languages, count, FromDBError(err)
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
		return languages, count, FromDBError(err)
	}

	count, err = s.database.CountFilteredLanguages(ctx, search)
	if err != nil {
		return languages, count, FromDBError(err)
	}

	return languages, count, nil
}
