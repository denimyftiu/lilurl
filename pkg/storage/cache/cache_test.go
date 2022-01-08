package cache

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/google/go-cmp/cmp"
)

func TestBasics(t *testing.T) {
	ctx := context.Background()
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	c := Cache{
		client: redis.NewClient(&redis.Options{Addr: s.Addr()}),
	}

	val := "value"
	must(t, c.CreateUrl(ctx, "key", val))
	got, err := c.GetUrl(ctx, "key")
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(got, val) {
		t.Fatalf("got %v, want %v", got, val)
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
