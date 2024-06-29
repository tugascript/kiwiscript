package services

import (
	"log/slog"

	cc "github.com/kiwiscript/kiwiscript_go/providers/cache"
	db "github.com/kiwiscript/kiwiscript_go/providers/database"
	email "github.com/kiwiscript/kiwiscript_go/providers/email"
	"github.com/kiwiscript/kiwiscript_go/providers/tokens"
)

type Services struct {
	database *db.Database
	cache    *cc.Cache
	mail     *email.Mail
	jwt      *tokens.Tokens
	log      *slog.Logger
}

func NewServices(database *db.Database, cache *cc.Cache, mail *email.Mail, jwt *tokens.Tokens, log *slog.Logger) *Services {
	return &Services{
		database: database,
		cache:    cache,
		mail:     mail,
		jwt:      jwt,
		log:      log,
	}
}
