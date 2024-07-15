package controllers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/services"
)

func (c *Controllers) CreateSeriesPart(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.GetSingleSeries")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	log.InfoContext(userCtx, "Creating series part...", "languageSlug", languageSlug, "seriesSlug", seriesSlug)

	user, err := c.GetUserClaims(ctx)
	if err != nil || !user.IsStaff {
		log.ErrorContext(userCtx, "User is not staff, should not have reached here")
		return ctx.Status(fiber.StatusForbidden).JSON(NewRequestError(services.NewForbiddenError()))
	}

	params := SeriesParams{
		LanguageSlug: languageSlug,
		SeriesSlug:   seriesSlug,
	}
	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	var request CreateSeriesPartRequest
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	seriesPart, serviceErr := c.services.CreateSeriesPart(userCtx, services.CreateSeriesPartOptions{
		UserID:       user.ID,
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		Title:        request.Title,
		Description:  request.Description,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.Status(fiber.StatusCreated).JSON(
		c.NewSeriesPartResponse(
			seriesPart,
			make([]db.Lecture, 0),
			params.LanguageSlug,
			params.SeriesSlug,
		),
	)
}

func (c *Controllers) FindSeriesPart(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.series.GetSingleSeries")
	userCtx := ctx.UserContext()
	languageSlug := ctx.Params("languageSlug")
	seriesSlug := ctx.Params("seriesSlug")
	seriesPartID := ctx.Params("seriesPartID")
	log.InfoContext(userCtx, "Getting series part...", "languageSlug", languageSlug, "seriesSlug", seriesSlug, "seriesPartID", seriesPartID)

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

	isPublished := false
	user, serviceErr := c.GetUserClaims(ctx)
	if serviceErr != nil || !user.IsStaff {
		isPublished = true
	}

	seriesPart, serviceErr := c.services.FindSeriesPartBySlugAndID(userCtx, services.FindSeriesPartBySlugsAndIDOptions{
		LanguageSlug: params.LanguageSlug,
		SeriesSlug:   params.SeriesSlug,
		IsPublished:  isPublished,
		SeriesPartID: int32(parsedSeriesPartID),
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(c.NewSeriesPartResponseFromDTO(seriesPart, params.LanguageSlug, params.SeriesSlug))
}
