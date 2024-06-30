package cc

import (
	"crypto/rand"
	"math/big"
	"strconv"
	"time"

	"github.com/kiwiscript/kiwiscript_go/utils"
)

const (
	twoFactorPrefix  string = "two_factor"
	twoFactorSeconds int    = 300
)

func generateCode() (string, error) {
	const codeLength = 6
	const digits = "0123456789"
	code := make([]byte, codeLength)

	for i := 0; i < codeLength; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		code[i] = digits[num.Int64()]
	}

	return string(code), nil
}

func (c *Cache) AddTwoFactorCode(userID int32) (string, error) {
	code, err := generateCode()
	if err != nil {
		return code, err
	}

	hashedCode, err := utils.HashPassword(code)
	if err != nil {
		return "", err
	}

	key := twoFactorPrefix + ":" + strconv.Itoa(int(userID))
	val := []byte(hashedCode)
	exp := time.Duration(twoFactorSeconds) * time.Second
	if err := c.storage.Set(key, val, exp); err != nil {
		return "", err
	}

	return code, nil
}

func (c *Cache) VerifyTwoFactorCode(userID int32, code string) (bool, error) {
	key := twoFactorPrefix + ":" + strconv.Itoa(int(userID))
	valByte, err := c.storage.Get(key)

	if err != nil {
		return false, err
	}
	if valByte == nil {
		return false, nil
	}

	val := string(valByte)
	if !utils.VerifyPassword(code, val) {
		return false, nil
	}
	if err := c.storage.Delete(key); err != nil {
		return true, err
	}

	return true, nil
}
