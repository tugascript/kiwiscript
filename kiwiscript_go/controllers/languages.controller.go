package controllers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/kiwiscript/kiwiscript_go/services"
	"github.com/kiwiscript/kiwiscript_go/utils"
)

func isValidSVG(svg string) bool {
	return strings.HasPrefix(svg, "<svg") && strings.HasSuffix(svg, "</svg>")
}

func (c *Controllers) CreateLanguage(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.languages.CreateLanguage")
	userCtx := ctx.UserContext()
	var request CreateLanguageRequest
	user, err := c.GetUserClaims(ctx)

	if err != nil || !user.IsAdmin {
		return ctx.Status(fiber.StatusForbidden).JSON(NewRequestError(services.NewForbiddenError()))
	}
	if err := ctx.BodyParser(&request); err != nil {
		return c.parseRequestErrorResponse(log, userCtx, err, ctx)
	}
	if err := c.validate.StructCtx(userCtx, request); err != nil {
		return c.validateRequestErrorResponse(log, userCtx, err, ctx)
	}

	icon := strings.TrimSpace(request.Icon)
	if !isValidSVG(icon) {
		return ctx.Status(fiber.StatusBadRequest).JSON(NewRequestValidationError(RequestValidationLocationBody, []FieldError{{
			Param:   "icon",
			Message: "icon must be a valid SVG",
			Value:   icon,
		}}))
	}

	language, serviceErr := c.services.CreateLanguage(userCtx, services.CreateLanguageOptions{
		UserID: user.ID,
		Name:   utils.Slugify(request.Name),
		Icon:   icon,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.Status(fiber.StatusCreated).JSON(NewLanguageResponse(language))
}

func (c *Controllers) GetLanguage(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.languages.GetLanguage")
	userCtx := ctx.UserContext()
	name := ctx.Params("name")
	log.InfoContext(userCtx, "get language", "name", name)
	params := GetLanguageParams{name}

	if err := c.validate.StructCtx(userCtx, params); err != nil {
		return c.validateParamsErrorResponse(log, userCtx, err, ctx)
	}

	language, serviceErr := c.services.FindLanguageByName(userCtx, name)
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(NewLanguageResponse(language))
}

func (c *Controllers) GetLanguages(ctx *fiber.Ctx) error {
	log := c.log.WithGroup("controllers.languages.GetLanguages")
	userCtx := ctx.UserContext()
	log.InfoContext(userCtx, "get languages")
	query_params := GetLanguagesQueryParams{
		Offset: int32(ctx.QueryInt("offset", OffsetDefault)),
		Limit:  int32(ctx.QueryInt("limit", LimitDefault)),
		Search: ctx.Query("search"),
	}

	if err := c.validate.StructCtx(userCtx, query_params); err != nil {
		return c.validateQueryErrorResponse(log, userCtx, err, ctx)
	}

	languages, count, serviceErr := c.services.FindPaginatedLanguages(userCtx, services.FindPaginatedLanguagesOptions{
		Search: query_params.Search,
		Offset: query_params.Offset,
		Limit:  query_params.Limit,
	})
	if serviceErr != nil {
		return c.serviceErrorResponse(serviceErr, ctx)
	}

	return ctx.JSON(NewPaginatedResponse(
		c.frontendDomain,
		query_params.Limit,
		query_params.Offset,
		count,
		languages,
		NewLanguageResponse,
	))
}
