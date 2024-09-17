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
	objStg "github.com/kiwiscript/kiwiscript_go/providers/object_storage"
	"mime/multipart"
)

const userPicturesLocation string = "user_pictures"

type FindUserPictureOptions struct {
	RequestID string
	UserID    int32
}

func (s *Services) FindUserPicture(
	ctx context.Context,
	opts FindUserPictureOptions,
) (*db.UserPicture, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, userPicturesLocation, "FindUserPicture").With(
		"userId", opts.UserID,
	)
	log.InfoContext(ctx, "Finding user picture...")

	picture, err := s.database.FindUserPictureByUserID(ctx, opts.UserID)
	if err != nil {
		log.WarnContext(ctx, "User picture not found", "error", err)
		return nil, exceptions.FromDBError(err)
	}

	return &picture, nil
}

type UploadUserPictureOptions struct {
	RequestID  string
	UserID     int32
	FileHeader *multipart.FileHeader
}

func (s *Services) UploadUserPicture(
	ctx context.Context,
	opts UploadUserPictureOptions,
) (*db.UserPicture, *exceptions.ServiceError) {
	log := s.buildLogger(opts.RequestID, userPicturesLocation, "UploadUserPicture").With(
		"userId", opts.UserID,
	)
	log.InfoContext(ctx, "Uploading user picture...")

	findOpts := FindUserPictureOptions{
		RequestID: opts.RequestID,
		UserID:    opts.UserID,
	}
	if _, serviceErr := s.FindUserPicture(ctx, findOpts); serviceErr == nil {
		return nil, exceptions.NewConflictError("User picture already exists")
	}

	fileID, ext, err := s.objStg.UploadImage(ctx, objStg.UploadImageOptions{
		RequestID: opts.RequestID,
		UserID:    opts.UserID,
		FH:        opts.FileHeader,
	})
	if err != nil {
		if err.Error() == "mime type not supported" {
			log.WarnContext(ctx, "Invalid file type", "error", err)
			return nil, exceptions.NewValidationError("Invalid file type")
		}

		log.ErrorContext(ctx, "Error uploading picture", "error", err)
		return nil, exceptions.NewServerError()
	}

	picture, err := s.database.CreateUserPicture(ctx, db.CreateUserPictureParams{
		ID:     fileID,
		UserID: opts.UserID,
		Ext:    ext,
	})
	if err != nil {
		log.ErrorContext(ctx, "Error creating user picture", "error", err)
		return nil, exceptions.FromDBError(err)
	}

	return &picture, nil
}

type DeleteUserPictureOptions struct {
	RequestID string
	UserID    int32
}

func (s *Services) DeleteUserPicture(
	ctx context.Context,
	opts DeleteUserPictureOptions,
) *exceptions.ServiceError {
	log := s.buildLogger(opts.RequestID, userPicturesLocation, "DeleteUserPicture").With(
		"userId", opts.UserID,
	)
	log.InfoContext(ctx, "Deleting user picture...")

	picture, serviceErr := s.FindUserPicture(ctx, FindUserPictureOptions(opts))
	if serviceErr != nil {
		return serviceErr
	}

	qrs, txn, err := s.database.BeginTx(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return exceptions.FromDBError(err)
	}
	defer func() {
		log.DebugContext(ctx, "Finalizing transaction")
		s.database.FinalizeTx(ctx, txn, err, serviceErr)
	}()

	if err := qrs.DeleteUserPicture(ctx, picture.ID); err != nil {
		log.ErrorContext(ctx, "Failed to delete user picture", "error", err)
		serviceErr = exceptions.FromDBError(err)
		return serviceErr
	}
	if err := s.objStg.DeleteFile(ctx, opts.UserID, picture.ID, picture.Ext); err != nil {
		log.ErrorContext(ctx, "Failed to delete file from object storage", "error", err)
		serviceErr = exceptions.NewServerError()
		return serviceErr
	}

	return nil
}
