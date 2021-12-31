package cache

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/dript0hard/lilurl/pkg/config"
	"github.com/go-redis/redis/v8"
)

type Cache struct {
	client *redis.Client
}

func (c *Cache) Close() error {
	return c.client.Close()
}

func (c Cache) CreateUrl(ctx context.Context, id, url string) error {
	log.Printf("(Redis Set) Id: %s, Url: %s", id, url)
	if err := c.client.Set(ctx, id, url, 3*time.Minute).Err(); err != nil {
		log.Printf("(Redis Set): %s", err.Error())
		return err
	}
	return nil
}

func (c Cache) GetUrl(ctx context.Context, id string) (string, error) {
	url, err := c.client.Get(ctx, id).Result()
	if err != nil {
		log.Printf("(Redis Get): %s", err.Error())
		return "", err
	}
	log.Printf("(Redis Retrieving) Id: %s, Url: %s", id, url)
	return url, nil
}

func Open(ctx context.Context, c *config.Config) (*Cache, error) {
	opts := &redis.Options{
		Addr:     net.JoinHostPort(c.RedisHost, c.RedisPort),
		DB:       c.RedisDB,
		Password: c.RedisPassword,
	}

	client := redis.NewClient(opts)
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Cache{client: client}, nil
}
