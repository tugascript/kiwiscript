package cc

import (
	"context"
	"time"
)

const blackListPrefix string = "black_list"

type AddBlackListOptions struct {
	ID    string
	Token string
	Exp   time.Time
}

func (c *Cache) AddBlackList(options AddBlackListOptions) error {
	key := blackListPrefix + ":" + options.ID
	return c.storage.Set(key, []byte(options.Token), time.Until(options.Exp))
}

func (c *Cache) IsBlackListed(ctx context.Context, id string) bool {
	key := blackListPrefix + ":" + id

	if _, err := c.storage.Get(key); err != nil {
		return false
	}

	return true
}
