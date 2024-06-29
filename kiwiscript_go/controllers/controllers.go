package controllers

import (
	"github.com/go-playground/validator/v10"
	"github.com/kiwiscript/kiwiscript_go/services"
)

type Controllers struct {
	services          *services.Services
	validate          *validator.Validate
	refreshCookieName string
}

func NewControllers(services *services.Services, validate *validator.Validate, refreshCookieName string) *Controllers {
	return &Controllers{
		services:          services,
		validate:          validate,
		refreshCookieName: refreshCookieName,
	}
}
