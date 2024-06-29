package cc

import (
	"strconv"
	"time"
)

const (
	twoFactorPrefix  string = "two_factor"
	twoFactorSeconds int    = 300
)

type AddTwoFactorCodeOptions struct {
	UserID int32
	Code   string
}

func (c *Cache) AddTwoFactorCode(options AddTwoFactorCodeOptions) error {
	key := twoFactorPrefix + ":" + strconv.Itoa(int(options.UserID))
	return c.storage.Set(key, []byte(options.Code), time.Duration(twoFactorSeconds)*time.Second)
}

func (c *Cache) GetTwoFactorCode(userID int32) (string, error) {
	key := twoFactorPrefix + ":" + strconv.Itoa(int(userID))
	valByte, err := c.storage.Get(key)

	if err != nil {
		return "", err
	}

	return string(valByte), nil
}

func (c *Cache) DeleteTwoFactorCode(userID int32) error {
	key := twoFactorPrefix + ":" + strconv.Itoa(int(userID))
	return c.storage.Delete(key)
}
