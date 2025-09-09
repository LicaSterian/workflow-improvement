package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisCache(
	ctx context.Context,
	addr,
	password string,
	db int,
) (Cache, func() error) {
	rdb := redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db})
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	_ = rdb.Ping(ctx).Err() // if it fails, handler code will surface errors; you can choose to fatal earlier

	return &cache{
			R: rdb,
		}, func() error {
			return rdb.Close()
		}
}

type cache struct {
	R *redis.Client
}

func (c *cache) Get(ctx context.Context, key string) (string, error) {
	return c.R.Get(ctx, key).Result()
}
func (c *cache) Set(ctx context.Context, key, val string) error {
	return c.R.Set(ctx, key, val, 0).Err()
}
