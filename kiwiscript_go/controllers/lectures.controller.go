package controllers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/kiwiscript/kiwiscript_go/services"
)

func (c *Controllers) CreateLecture(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.CreateLecture")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	seriesPartID := ctx.Params("seriesPartID")
	log.InfoContext(
		userCtx,
		"Creating lecture...",
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"seriesPartID", seriesPartID,
	)

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(NewRequestError(services.NewForbiddenError()))
	}

	params := SeriesPartParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SeriesPartID: seriesPartID,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	parsedSeriesPartID, err := strconv.Atoi(params.SeriesPartID)
	if err != nil {
		return ctx.
			Status(fiber.StatusBadRequest).
			JSON(NewRequestValidationError(RequestValidationLocationParams, []FieldError{{
				Param:   "seriesPartId",
				Message: StrFieldErrMessageNumber,
				Value:   params.SeriesPartID,
			}}))
	}

	var request CreateLectureRequest
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	seriesPartIDi32 := int32(parsedSeriesPartID)
	lecture, serviceErr := c.services.CreateLectures(userCtx, services.CreateLecturesOptions{
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		SeriesPartID: seriesPartIDi32,
		Title:        request.Title,
		Description:  request.Description,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.
		Status(fiber.StatusCreated).
		JSON(
			c.NewLectureResponse(
				lecture,
				nil, nil,
				params.LanguageSlug,
				params.SeriesSlug,
				seriesPartIDi32,
			),
		)
}

func (c *Controllers) GetLecture(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.CreateLecture")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	seriesPartID := ctx.Params("seriesPartID")
	log.InfoContext(
		userCtx,
		"Creating lecture...",
		"languageSlug", languageSlug,
		"seriesSlug", seriesSlug,
		"seriesPartID", seriesPartID,
	)

	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(NewRequestError(services.NewForbiddenError()))
	}

	params := SeriesPartParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
		SeriesPartID: seriesPartID,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	// TODO: finish this function
	return nil
}
