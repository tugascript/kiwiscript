package controllers

import "github.com/gofiber/fiber/v2"

func (c *Controllers) HealthCheck(ctx *fiber.Ctx) error {
	return ctx.SendStatus(fiber.StatusOK)
}
