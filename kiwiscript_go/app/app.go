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

	"github.com/aws/aws-sdk-go-v2/service/s3"
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
	stg "github.com/kiwiscript/kiwiscript_go/providers/object_storage"
	"github.com/kiwiscript/kiwiscript_go/providers/tokens"
	"github.com/kiwiscript/kiwiscript_go/routers"
	"github.com/kiwiscript/kiwiscript_go/services"
)

func CreateApp(
	log *slog.Logger,
	storage *redis.Storage,
	dbConnPool *pgxpool.Pool,
	s3Client *s3.Client,
	mailConfig *EmailConfig,
	tokensConfig *TokensConfig,
	limiterConfig *LimiterConfig,
	s3Bucket,
	backendDomain,
	frontendDomain,
	refreshCookieName,
	cookieSecret string,
) *fiber.App {
	// Build the app
	log.Info("Building the app...")
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
	objStg := stg.NewObjectStorage(s3Client, s3Bucket)
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
	log.Info("Loading validators...")
	vld := validator.New()
	if err := vld.RegisterValidation(svgValidatorTag, isValidSVG); err != nil {
		log.Error("Failed to register svg validator", err)
		panic(err)
	}
	if err := vld.RegisterValidation(extAlphaNumTag, isValidExtAlphaNum); err != nil {
		log.Error("Failed to register extalphanum validator", err)
		panic(err)
	}
	if err := vld.RegisterValidation(slugValidatorTag, isValidSlug); err != nil {
		log.Error("Failed to register slug validator", err)
		panic(err)
	}
	if err := vld.RegisterValidation(markdownValidatorTag, isValidMarkdown); err != nil {
		log.Error("Failed to register markdown validator", err)
		panic(err)
	}
	log.Info("Successfully loaded validators")

	// Build service
	log.Info("Building services...")
	srvs := services.NewServices(log, database, cache, objStg, mailer, tokenProv)
	log.Info("Successfully built services")

	// Build controllers
	log.Info("Building controllers...")
	ctrls := controllers.NewControllers(log, srvs, vld, frontendDomain, backendDomain, refreshCookieName)
	log.Info("Successfully built controllers")

	// Build router
	log.Info("Building router...")
	rtr := routers.NewRouter(app, ctrls)
	log.Info("Successfully built router")

	// Build routes, public routes need to be defined before private ones
	log.Info("Loading public routes...")
	rtr.HealthRoutes()
	rtr.AuthPublicRoutes()
	rtr.LanguagePublicRoutes()
	rtr.SeriesPublicRoutes()
	rtr.SeriesPartPublicRoutes()
	rtr.LecturePublicRoutes()
	rtr.LectureArticlePublicRoutes()
	rtr.LectureVideoPublicRoutes()
	log.Info("Successfully loaded public routes")

	// Private routes
	log.Info("Loading private routes...")
	rtr.AuthPrivateRoutes()
	rtr.LanguagePrivateRoutes()
	rtr.SeriesPrivateRoutes()
	rtr.SeriesPartPrivateRoutes()
	rtr.LecturePrivateRoutes()
	rtr.LectureArticlePrivateRoutes()
	rtr.LectureVideoPrivateRoutes()
	log.Info("Successfully loaded private routes")

	log.Info("Successfully built the app")
	return app
}
