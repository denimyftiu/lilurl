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
	log.Printf("cache set(%q): %s", id, url)
	if err := c.client.Set(ctx, id, url, 3*time.Minute).Err(); err != nil {
		log.Printf("cache set(%q): %s", id, err.Error())
		return err
	}
	return nil
}

func (c Cache) GetUrl(ctx context.Context, id string) (string, error) {
	// Set small timeout for fast fallback to postgres
	getCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	url, err := c.client.Get(getCtx, id).Result()
	if err != nil {
		select {
		case <-getCtx.Done():
			log.Printf("cache get(%q): context timed out", id)
		default:
			log.Printf("cache get(%q): %s", id, err.Error())
		}
		return "", err
	}
	log.Printf("cache get(%q): %s", id, url)
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
