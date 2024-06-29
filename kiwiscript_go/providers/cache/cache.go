package cc

import (
	"github.com/gofiber/storage/redis/v3"
)

type Cache struct {
	storage *redis.Storage
}

func NewCache(storage *redis.Storage) *Cache {
	return &Cache{
		storage: storage,
	}
}
