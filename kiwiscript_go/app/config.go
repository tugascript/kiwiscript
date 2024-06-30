package app

import (
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type LoggerConfig struct {
	Env   string
	Debug bool
}

type SingleJwtConfig struct {
	PublicKey  string
	PrivateKey string
	TtlSec     int64
}

type TokensConfig struct {
	Access  SingleJwtConfig
	Refresh SingleJwtConfig
	Email   SingleJwtConfig
}

type EmailConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Name     string
}

type LimiterConfig struct {
	Max    int64
	ExpSec int64
}

type AppConfig struct {
	MaxProcs          int64
	Port              string
	RefreshCookieName string
	CookieSecret      string
	RedisURL          string
	PostgresURL       string
	FrontendDomain    string
	BackendDomain     string
	Logger            LoggerConfig
	Email             EmailConfig
	Tokens            TokensConfig
	Limiter           LimiterConfig
}

var variables = [26]string{
	"PORT",
	"ENV",
	"DEBUG",
	"MAX_PROCS",
	"DATABASE_URL",
	"REDIS_URL",
	"FRONTEND_DOMAIN",
	"BACKEND_DOMAIN",
	"COOKIE_SECRET",
	"REFRESH_COOKIE_NAME",
	"JWT_ACCESS_PUBLIC_KEY",
	"JWT_ACCESS_PRIVATE_KEY",
	"JWT_ACCESS_TTL_SEC",
	"JWT_REFRESH_PUBLIC_KEY",
	"JWT_REFRESH_PRIVATE_KEY",
	"JWT_REFRESH_TTL_SEC",
	"JWT_EMAIL_PUBLIC_KEY",
	"JWT_EMAIL_PRIVATE_KEY",
	"JWT_EMAIL_TTL_SEC",
	"EMAIL_HOST",
	"EMAIL_PORT",
	"EMAIL_USERNAME",
	"EMAIL_PASSWORD",
	"EMAIL_NAME",
	"LIMITER_MAX",
	"LIMITER_EXP_SEC",
}

var numerics = [6]string{
	"MAX_PROCS",
	"JWT_ACCESS_TTL_SEC",
	"JWT_REFRESH_TTL_SEC",
	"JWT_EMAIL_TTL_SEC",
	"LIMITER_MAX",
	"LIMITER_EXP_SEC",
}

func NewConfig(log *slog.Logger, envPath string) *AppConfig {
	err := godotenv.Load(envPath)
	if err != nil {
		log.Error("Error loading .env file")
	}

	variablesMap := make(map[string]string)
	for _, variable := range variables {
		value := os.Getenv(variable)
		if value == "" {
			log.Error(variable + " is not set")
			panic(variable + " is not set")
		}
		variablesMap[variable] = value
	}

	intMap := make(map[string]int64)
	for _, numeric := range numerics {
		value, err := strconv.ParseInt(variablesMap[numeric], 10, 0)
		if err != nil {
			log.Error(numeric + " is not an integer")
			panic(numeric + " is not an integer")
		}
		intMap[numeric] = value
	}
	return &AppConfig{
		MaxProcs:          intMap["MAX_PROCS"],
		Port:              variablesMap["PORT"],
		PostgresURL:       variablesMap["DATABASE_URL"],
		RedisURL:          variablesMap["REDIS_URL"],
		FrontendDomain:    variablesMap["FRONTEND_DOMAIN"],
		BackendDomain:     variablesMap["BACKEND_DOMAIN"],
		CookieSecret:      variablesMap["COOKIE_SECRET"],
		RefreshCookieName: variablesMap["REFRESH_COOKIE_NAME"],
		Logger: LoggerConfig{
			Env:   strings.ToLower(variablesMap["ENV"]),
			Debug: strings.ToLower(variablesMap["DEBUG"]) == "true",
		},
		Email: EmailConfig{
			Host:     variablesMap["EMAIL_HOST"],
			Port:     variablesMap["EMAIL_PORT"],
			Username: variablesMap["EMAIL_USERNAME"],
			Password: variablesMap["EMAIL_PASSWORD"],
			Name:     variablesMap["EMAIL_NAME"],
		},
		Tokens: TokensConfig{
			Access: SingleJwtConfig{
				PublicKey:  variablesMap["JWT_ACCESS_PUBLIC_KEY"],
				PrivateKey: variablesMap["JWT_ACCESS_PRIVATE_KEY"],
				TtlSec:     intMap["JWT_ACCESS_TTL_SEC"],
			},
			Refresh: SingleJwtConfig{
				PublicKey:  variablesMap["JWT_REFRESH_PUBLIC_KEY"],
				PrivateKey: variablesMap["JWT_REFRESH_PRIVATE_KEY"],
				TtlSec:     intMap["JWT_REFRESH_TTL_SEC"],
			},
			Email: SingleJwtConfig{
				PublicKey:  variablesMap["JWT_EMAIL_PUBLIC_KEY"],
				PrivateKey: variablesMap["JWT_EMAIL_PRIVATE_KEY"],
				TtlSec:     intMap["JWT_EMAIL_TTL_SEC"],
			},
		},
		Limiter: LimiterConfig{
			Max:    intMap["LIMITER_MAX"],
			ExpSec: intMap["LIMITER_EXP_SEC"],
		},
	}
}
