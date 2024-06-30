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
		config.BackendDomain,
		config.FrontendDomain,
		config.RefreshCookieName,
	)
	log.Info("Finished building the app")

	// Start the app
	log.Info("Starting the app...")
	err = app.Listen(":" + config.Port)
	log.ErrorContext(ctx, "Failed to start the app", "error", err)
}
