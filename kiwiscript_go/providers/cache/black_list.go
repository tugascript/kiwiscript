package cc

import (
	"context"
	"time"
)

const blackListPrefix string = "black_list"

type AddBlackListOptions struct {
	ID  string
	Exp time.Time
}

func (c *Cache) AddBlackList(options AddBlackListOptions) error {
	key := blackListPrefix + ":" + options.ID
	val := []byte(options.ID)
	exp := time.Until(options.Exp)
	return c.storage.Set(key, val, exp)
}

func (c *Cache) IsBlackListed(ctx context.Context, id string) (bool, error) {
	key := blackListPrefix + ":" + id
	valByte, err := c.storage.Get(key)

	if err != nil {
		return false, err
	}
	if valByte == nil {
		return false, nil
	}

	return true, nil
}
