package app

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

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

type AppConfig struct {
	Port              string
	RefreshCookieName string
	CookieSecret      string
	RedisURL          string
	PostgresURL       string
	FrontendDomain    string
	BackendDomain     string
	Email             EmailConfig
	Tokens            TokensConfig
}

var variables = [21]string{
	"PORT",
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
}

var ttls = [3]string{
	"JWT_ACCESS_TTL_SEC",
	"JWT_REFRESH_TTL_SEC",
	"JWT_EMAIL_TTL_SEC",
}

func NewConfig(log *slog.Logger) *AppConfig {
	err := godotenv.Load()
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
	for _, ttl := range ttls {
		value, err := strconv.ParseInt(variablesMap[ttl], 10, 0)
		if err != nil {
			log.Error(ttl + " is not an integer")
			panic(ttl + " is not an integer")
		}
		intMap[ttl] = value
	}
	return &AppConfig{
		Port:              variablesMap["PORT"],
		PostgresURL:       variablesMap["DATABASE_URL"],
		RedisURL:          variablesMap["REDIS_URL"],
		FrontendDomain:    variablesMap["FRONTEND_DOMAIN"],
		BackendDomain:     variablesMap["BACKEND_DOMAIN"],
		CookieSecret:      variablesMap["COOKIE_SECRET"],
		RefreshCookieName: variablesMap["REFRESH_COOKIE_NAME"],
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
	}
}
