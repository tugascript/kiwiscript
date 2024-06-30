package controllers

import (
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/kiwiscript/kiwiscript_go/services"
)

type Controllers struct {
	log               *slog.Logger
	services          *services.Services
	validate          *validator.Validate
	refreshCookieName string
}

func NewControllers(log *slog.Logger, services *services.Services, validate *validator.Validate, refreshCookieName string) *Controllers {
	return &Controllers{
		log:               log,
		services:          services,
		validate:          validate,
		refreshCookieName: refreshCookieName,
	}
}
