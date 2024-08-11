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

package main

import (
	"context"
	"fmt"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/storage/redis/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kiwiscript/kiwiscript_go/app"
)

func main() {
	log := app.DefaultLogger()
	ctx := context.Background()
	log.InfoContext(ctx, "Loading configuration...")
	cfg := app.NewConfig(log, "./.env")
	log.InfoContext(ctx, "Finished loading configuration")

	// Set maximum CPU cores
	log.Info("Setting GOMAXPROCS...", "maxProcs", cfg.MaxProcs)
	runtime.GOMAXPROCS(int(cfg.MaxProcs))
	log.Info("Finished setting GOMAXPROCS")

	// Update logger
	log.Info("Updating logger with config...")
	log = app.GetLogger(cfg.Logger.Env, cfg.Logger.Debug)
	log.Info("Finished updating logger")

	// Build storages/models
	log.Info("Building redis connection...")
	storage := redis.New(redis.Config{
		URL: cfg.RedisURL,
	})
	log.Info("Finished building redis connection")

	// Build database connection
	log.Info("Building database connection...")
	dbConnPool, err := pgxpool.New(ctx, cfg.PostgresURL)
	if err != nil {
		log.ErrorContext(ctx, "Failed to connect to database", "error", err)
		panic(err)
	}
	log.Info("Finished building database connection")

	// Build s3 client
	log.Info("Building s3 client...")
	s3Cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(cfg.ObjectStorage.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.ObjectStorage.AccessKey, cfg.ObjectStorage.SecretKey, "")),
	)
	if err != nil {
		log.ErrorContext(ctx, "Failed to load s3 config", "error", err)
		panic(err)
	}
	s3Client := s3.NewFromConfig(s3Cfg, func(o *s3.Options) {
		if cfg.Logger.Env == "production" {
			o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.%s.com", cfg.ObjectStorage.Region, cfg.ObjectStorage.Host))
		}
		o.BaseEndpoint = aws.String("http://" + cfg.ObjectStorage.Host)
	})
	log.Info("Finished building s3 client")

	// Build the fiberApp
	log.Info("Building the fiberApp...")
	fiberApp := app.CreateApp(
		log,
		storage,
		dbConnPool,
		s3Client,
		&cfg.Email,
		&cfg.Tokens,
		&cfg.Limiter,
		&cfg.OAuthProviders,
		cfg.ObjectStorage.Bucket,
		cfg.BackendDomain,
		cfg.FrontendDomain,
		cfg.RefreshCookieName,
		cfg.CookieSecret,
	)
	log.Info("Finished building the fiberApp")

	// Start the fiberApp
	log.Info("Starting the fiberApp...")
	err = fiberApp.Listen(":" + cfg.Port)
	log.ErrorContext(ctx, "Failed to start the fiberApp", "error", err)
}
