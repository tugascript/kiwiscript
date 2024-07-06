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
	"runtime"

	"github.com/gofiber/storage/redis/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kiwiscript/kiwiscript_go/app"
)

func main() {
	log := app.DefaultLogger()
	log.Info("Loading configuration...")
	config := app.NewConfig(log, "./.env")
	log.Info("Finished loading configuration")

	// Set maximum CPU cores
	log.Info("Setting GOMAXPROCS...", "maxProcs", config.MaxProcs)
	runtime.GOMAXPROCS(int(config.MaxProcs))
	log.Info("Finished setting GOMAXPROCS")

	// Update logger
	log.Info("Updating logger with config...")
	log = app.GetLogger(config.Logger.Env, config.Logger.Debug)
	log.Info("Finished updating logger")

	// Build storages/models
	log.Info("Building redis connection...")
	storage := redis.New(redis.Config{
		URL: config.RedisURL,
	})
	log.Info("Finished building redis connection")

	// Build database connection
	log.Info("Building database connection...")
	ctx := context.Background()
	dbConnPool, err := pgxpool.New(ctx, config.PostgresURL)
	if err != nil {
		log.ErrorContext(ctx, "Failed to connect to database", "error", err)
		panic(err)
	}
	log.Info("Finished building database connection")

	// Build the app
	log.Info("Building the app...")
	app := app.CreateApp(
		log,
		storage,
		dbConnPool,
		&config.Email,
		&config.Tokens,
		&config.Limiter,
		config.BackendDomain,
		config.FrontendDomain,
		config.RefreshCookieName,
		config.CookieSecret,
	)
	log.Info("Finished building the app")

	// Start the app
	log.Info("Starting the app...")
	err = app.Listen(":" + config.Port)
	log.ErrorContext(ctx, "Failed to start the app", "error", err)
}
