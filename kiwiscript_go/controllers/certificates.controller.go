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

package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kiwiscript/kiwiscript_go/dtos"
	"github.com/kiwiscript/kiwiscript_go/exceptions"
	"github.com/kiwiscript/kiwiscript_go/paths"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/services"
)

const certificatesLocation string = "certificates"

func (c *Controllers) GetCertificate(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	certificateID := ctx.Params("certificateID")
	log := c.buildLogger(ctx, requestID, certificatesLocation, "GetCertificate").With(
		"certificateId", certificateID,
	)
	log.InfoContext(userCtx, "Getting certificate...")

	params := dtos.CertificatesPathParams{CertificateID: certificateID}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	parsedCertificateID, err := uuid.FromBytes([]byte(params.CertificateID))
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(exceptions.NewRequestValidationError(
				exceptions.RequestValidationLocationParams,
				[]exceptions.FieldError{{
					Param:   "certificateId",
					Message: exceptions.StrFieldErrMessageUUID,
					Value:   params.CertificateID,
				}},
			))
	}

	certificate, serviceErr := c.services.FindCertificateByID(userCtx, services.FindCertificateByIDOptions{
		RequestID: requestID,
		ID:        parsedCertificateID,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(dtos.NewCertificateResponse(c.backendDomain, certificate.ToCertificateModel()))
}

func (c *Controllers) GetUserCertificates(ctx *fiber.Ctx) error {
	requestID := c.requestID(ctx)
	userCtx := ctx.UserContext()
	log := c.buildLogger(ctx, requestID, certificatesLocation, "GetUserCertificate")
	log.InfoContext(userCtx, "Getting user certificates...")

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil {
		log.ErrorContext(userCtx, "This route is protected should have not reached here")
		return ctx.Status(fiber.StatusUnauthorized).JSON(exceptions.NewRequestError(exceptions.NewUnauthorizedError()))
	}

	queryParams := dtos.PaginationQueryParams{
		Limit:  int32(ctx.QueryInt("offset", dtos.OffsetDefault)),
		Offset: int32(ctx.QueryInt("limit", dtos.LimitDefault)),
	}
	if err := c.validate.StructCtx(userCtx, queryParams); err != nil {
		return c.validateQueryErrorResponse(log, userCtx, err, ctx)
	}

	certificates, count, serviceErr := c.services.FindPaginatedCertificates(
		userCtx,
		services.FindPaginatedCertificatesOptions{
			RequestID: requestID,
			UserID:    user.ID,
			Offset:    queryParams.Offset,
			Limit:     queryParams.Limit,
		},
	)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(
		dtos.NewPaginatedResponse(
			c.backendDomain,
			paths.CertificatesV1,
			&queryParams,
			count,
			certificates,
			func(dto *db.FindPaginatedCertificatesByUserIDRow) *dtos.CertificateResponse {
				return dtos.NewCertificateResponse(
					c.backendDomain,
					dto.ToCertificateModelWithAuthor(user.FirstName, user.LastName),
				)
			},
		),
	)
}
