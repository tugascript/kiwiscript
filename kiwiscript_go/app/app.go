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
	"github.com/kiwiscript/kiwiscript_go/providers/oauth"
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
	oauthProvidersConfig *OAuthProviders,
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
	objStg := stg.NewObjectStorage(log, s3Client, s3Bucket)
	tokenProv := tokens.NewTokens(
		tokens.NewTokenSecretData(tokensConfig.Access.PublicKey, tokensConfig.Access.PrivateKey, tokensConfig.Access.TtlSec),
		tokens.NewTokenSecretData(tokensConfig.Refresh.PublicKey, tokensConfig.Refresh.PrivateKey, tokensConfig.Refresh.TtlSec),
		tokens.NewTokenSecretData(tokensConfig.Email.PublicKey, tokensConfig.Email.PrivateKey, tokensConfig.Email.TtlSec),
		tokens.NewTokenSecretData(tokensConfig.OAuth.PublicKey, tokensConfig.OAuth.PrivateKey, tokensConfig.OAuth.TtlSec),
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
	oauthProviders := oauth.NewOAuthProviders(
		log,
		oauthProvidersConfig.GitHub.ClientID,
		oauthProvidersConfig.GitHub.ClientSecret,
		oauthProvidersConfig.Google.ClientID,
		oauthProvidersConfig.Google.ClientSecret,
		backendDomain,
	)

	// Validators
	log.Info("Loading validators...")
	vld := validator.New()
	if err := vld.RegisterValidation(svgValidatorTag, isValidSVG); err != nil {
		log.Error("Failed to register svg validator", "error", err)
		panic(err)
	}
	if err := vld.RegisterValidation(extAlphaNumTag, isValidExtAlphaNum); err != nil {
		log.Error("Failed to register extalphanum validator", "error", err)
		panic(err)
	}
	if err := vld.RegisterValidation(slugValidatorTag, isValidSlug); err != nil {
		log.Error("Failed to register slug validator", "error", err)
		panic(err)
	}
	if err := vld.RegisterValidation(markdownValidatorTag, isValidMarkdown); err != nil {
		log.Error("Failed to register markdown validator", "error", err)
		panic(err)
	}
	log.Info("Successfully loaded validators")

	// Build service
	log.Info("Building services...")
	srvs := services.NewServices(log, database, cache, objStg, mailer, tokenProv, oauthProviders)
	log.Info("Successfully built services")

	// Build controllers
	log.Info("Building controllers...")
	ctrls := controllers.NewControllers(log, srvs, vld, frontendDomain, backendDomain, refreshCookieName)
	log.Info("Successfully built controllers")

	log.Info("Load user claims...")
	app.Use(ctrls.AccessClaimsMiddleware)
	log.Info("Successfully loaded user claims")

	// Build router
	log.Info("Building router...")
	rtr := routers.NewRouter(app, ctrls)
	log.Info("Successfully built router")

	// Build routes, public routes need to be defined before private ones
	log.Info("Loading public routes...")
	rtr.HealthRoutes()
	rtr.AuthPublicRoutes()
	rtr.OAuthPublicRoutes()
	rtr.LanguagePublicRoutes()
	rtr.SeriesPublicRoutes()
	rtr.SeriesPicturesPublicRoutes()
	rtr.SectionPublicRoutes()
	rtr.LessonsPublicRoutes()
	rtr.LessonArticlePublicRoutes()
	rtr.LessonVideoPublicRoutes()
	rtr.LessonFilesPublicRoutes()
	rtr.CertificatesPublicRoutes()
	log.Info("Successfully loaded public routes")

	// User
	log.Info("Loading user routes...")
	rtr.UsersRoutes()
	log.Info("Successfully loaded user routes")

	// Private routes
	log.Info("Loading private routes...")
	rtr.AuthPrivateRoutes()
	rtr.LanguageProgressPrivateRoutes()
	rtr.SeriesProgressPrivateRoutes()
	rtr.SectionProgressPrivateRoutes()
	rtr.LessonProgressPrivateRoutes()
	rtr.CertificatesPrivateRoutes()
	log.Info("Successfully loaded private routes")

	// Staff routes
	log.Info("Loading staff routes...")
	rtr.SeriesStaffRoutes()
	rtr.SeriesPicturesStaffRoutes()
	rtr.SectionStaffRoutes()
	rtr.LessonsStaffRoutes()
	rtr.LessonArticleStaffRoutes()
	rtr.LessonVideoStaffRoutes()
	rtr.LessonFilesStaffRoutes()
	log.Info("Successfully loaded staff routes")

	// Admin Routes
	log.Info("Loading admin routes...")
	rtr.LanguageAdminRoutes()
	log.Info("Successfully loaded admin routes")

	log.Info("Successfully built the app")
	return app
}
