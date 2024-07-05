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
	frontendDomain    string
	refreshCookieName string
}

func NewControllers(log *slog.Logger, services *services.Services, validate *validator.Validate, frontendDomain, refreshCookieName string) *Controllers {
	return &Controllers{
		log:               log,
		services:          services,
		validate:          validate,
		frontendDomain:    frontendDomain,
		refreshCookieName: refreshCookieName,
	}
}
