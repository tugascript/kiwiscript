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
	"github.com/google/uuid"
	"github.com/kiwiscript/kiwiscript_go/exceptions"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
)

const certificatesLocation string = "certificates"

type FindPaginatedCertificatesOptions struct {
	RequestID string
	UserID    int32
	Offset    int32
	Limit     int32
}

func (s *Services) FindPaginatedCertificates(
	ctx context.Context,
	opts FindPaginatedCertificatesOptions,
) ([]db.FindPaginatedCertificatesByUserIDRow, int64, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, certificatesLocation, "FindPaginatedCertificates").With(
		"userId", opts.UserID,
		"offset", opts.Offset,
		"limit", opts.Limit,
	)
	log.InfoContext(ctx, "Finding paginated certificates...")

	count, err := s.database.CountCertificatesByUserID(ctx, opts.UserID)
	if err != nil {
		log.ErrorContext(ctx, "Failed to count certificates", "error", err)
		return nil, 0, exceptions.FromDBError(err)
	}
	if count == 0 {
		return make([]db.FindPaginatedCertificatesByUserIDRow, 0), 0, nil
	}

	certificates, err := s.database.FindPaginatedCertificatesByUserID(
		ctx,
		db.FindPaginatedCertificatesByUserIDParams{
			UserID: opts.UserID,
			Limit:  opts.Limit,
			Offset: opts.Offset,
		},
	)
	if err != nil {
		log.ErrorContext(ctx, "Failed to find certificates", "error", err)
		return nil, 0, exceptions.FromDBError(err)
	}

	log.InfoContext(ctx, "Certificates found successfully")
	return certificates, count, nil
}

type FindCertificateByIDOptions struct {
	RequestID string
	ID        uuid.UUID
}

func (s *Services) FindCertificateByID(
	ctx context.Context,
	opts FindCertificateByIDOptions,
) (*db.FindCertificateByIDWithUserAndLanguageRow, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, "certificatesLocation", "FindCertificateByID").With(
		"id", opts.ID.String(),
	)
	log.InfoContext(ctx, "Find certificate by id...")

	certificate, err := s.database.FindCertificateByIDWithUserAndLanguage(ctx, opts.ID)
	if err != nil {
		log.WarnContext(ctx, "Certificate not found", "error", err)
		return nil, exceptions.FromDBError(err)
	}

	log.InfoContext(ctx, "Certificate found successfully")
	return &certificate, nil
}
