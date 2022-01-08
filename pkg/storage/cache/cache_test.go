package cache

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/denimyftiu/lilurl/pkg/config"
	"github.com/go-redis/redis/v8"
	"github.com/google/go-cmp/cmp"
)

func TestBasics(t *testing.T) {
	ctx := context.Background()

	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}

	c := Cache{
		client: redis.NewClient(&redis.Options{Addr: s.Addr()}),
	}

	// Create value, get created value
	val := "value"
	must(t, c.CreateUrl(ctx, "key", val))
	got, err := c.GetUrl(ctx, "key")
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(got, val) {
		t.Fatalf("got %v, want %v", got, val)
	}

	// Not existing value
	if _, err = c.GetUrl(ctx, "not_exist"); err == nil {
		t.Fatal("must not exist")
	}

	// Test timeout
	fastCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
	defer cancel()

	_, err = c.GetUrl(fastCtx, "key")

	if err != nil {
		select {
		case <-fastCtx.Done():
		default:
			t.Fatal("faster than 100 millis")
		}
	}

	// Expired
	if ok := s.Del("key"); !ok {
		t.Fail()
	}
	if _, err = c.GetUrl(ctx, "key"); err == nil {
		t.Fatal("must not exist")
	}

	// Coverage
	c.Close()
	if err = c.CreateUrl(ctx, "id", ""); err == nil {
		t.Fatal("No empty value")
	}
}

func TestCloseClient(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}

	c := Cache{
		client: redis.NewClient(&redis.Options{Addr: s.Addr()}),
	}

	if err := c.Close(); err != nil {
		t.Fatal("failed to close client")
	}
}

func TestOpenClient(t *testing.T) {
	ctx := context.Background()

	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}

	host, port, _ := net.SplitHostPort(s.Addr())
	cfg := &config.Config{
		RedisHost: host,
		RedisPort: port,
	}

	c, err := Open(ctx, cfg)
	if err != nil {
		t.Fatalf("cache Open: %s", err.Error())
	}
	defer c.Close()
}

func TestOpenClientTimeout(t *testing.T) {
	ctx := context.Background()

	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}

	host, port, _ := net.SplitHostPort(s.Addr())
	cfg := &config.Config{
		RedisHost: host,
		RedisPort: port,
	}

	// Close to stop Ping
	s.Close()

	_, err = Open(ctx, cfg)
	if err == nil {
		t.Fatal("Ping must fail")
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
