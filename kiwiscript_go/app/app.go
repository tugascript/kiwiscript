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

package app

import (
	"log/slog"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/storage/redis/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kiwiscript/kiwiscript_go/controllers"
	cc "github.com/kiwiscript/kiwiscript_go/providers/cache"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	"github.com/kiwiscript/kiwiscript_go/providers/email"
	"github.com/kiwiscript/kiwiscript_go/providers/tokens"
	"github.com/kiwiscript/kiwiscript_go/routers"
	"github.com/kiwiscript/kiwiscript_go/services"
)

func CreateApp(
	log *slog.Logger,
	storage *redis.Storage,
	dbConnPool *pgxpool.Pool,
	mailConfig *EmailConfig,
	tokensConfig *TokensConfig,
	limiterConfig *LimiterConfig,
	backendDomain,
	frontendDomain,
	refreshCookieName,
	cookieSecret string,
) *fiber.App {
	// Build the app
	app := fiber.New()

	// Load common middlewares
	log.Info("Loading common middlewares...")
	app.Use(logger.New())
	app.Use(helmet.New())
	app.Use(requestid.New())
	app.Use(limiter.New(limiter.Config{
		Max:               int(limiterConfig.Max),
		Expiration:        time.Duration(limiterConfig.ExpSec) * time.Second,
		LimiterMiddleware: limiter.SlidingWindow{},
		Storage:           storage,
	}))
	app.Use(encryptcookie.New(encryptcookie.Config{
		Key: cookieSecret,
	}))
	log.Info("Finished loading common middlewares")

	database := db.NewDatabase(dbConnPool)
	cache := cc.NewCache(storage)
	tokenProv := tokens.NewTokens(
		tokens.NewTokenSecretData(tokensConfig.Access.PublicKey, tokensConfig.Access.PrivateKey, tokensConfig.Access.TtlSec),
		tokens.NewTokenSecretData(tokensConfig.Refresh.PublicKey, tokensConfig.Refresh.PrivateKey, tokensConfig.Refresh.TtlSec),
		tokens.NewTokenSecretData(tokensConfig.Email.PublicKey, tokensConfig.Email.PrivateKey, tokensConfig.Email.TtlSec),
		"https://"+backendDomain,
	)
	mailer := email.NewMail(
		mailConfig.Username,
		mailConfig.Password,
		mailConfig.Port,
		mailConfig.Host,
		mailConfig.Name,
		frontendDomain,
	)

	// Validators
	vld := validator.New()
	vld.RegisterValidation(svgValidatorTag, isValidSVG)
	vld.RegisterValidation(extAlphaNumTag, isValidExtAlphaNum)
	vld.RegisterValidation(slugValidatorTag, isValidSlug)

	// Build service
	srvs := services.NewServices(database, cache, mailer, tokenProv, log)
	// Build controllers
	ctrls := controllers.NewControllers(log, srvs, vld, frontendDomain, refreshCookieName)
	// Build router
	rtr := routers.NewRouter(app, ctrls)

	// Build routes, public routes need to be defined before private ones
	// Public routes
	rtr.HealthRoutes()
	rtr.AuthPublicRoutes()
	rtr.LanguagePublicRoutes()

	// Private routes
	rtr.AuthPrivateRoutes()
	rtr.LanguagePrivateRoutes()

	return app
}
