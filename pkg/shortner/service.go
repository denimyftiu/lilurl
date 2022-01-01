package shortner

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/dript0hard/lilurl/pkg/storage/cache"
	"github.com/dript0hard/lilurl/pkg/storage/postgres"
	"github.com/teris-io/shortid"
)

// Url Shortener interface to be implemented.
type Shortner interface {
	// Shorten url to token.
	Shorten(context.Context, string) (string, error)

	// Expand token to url.
	Expand(context.Context, string) (string, error)
}

type ShortnerConfig struct {
	DB    *postgres.DB
	Cache *cache.Cache
}

// Service implementation.
type ShortnerService struct {
	db    *postgres.DB
	cache *cache.Cache
}

func NewShortner(scfg *ShortnerConfig) *ShortnerService {
	return &ShortnerService{
		db:    scfg.DB,
		cache: scfg.Cache,
	}
}

func (s *ShortnerService) Shorten(ctx context.Context, url string) (string, error) {
	if !isValidURL(url) {
		return "", ErrorInvalidURL
	}

	id, err := shortid.Generate()
	if err != nil {
		return "", err
	}

	go func(id, url string) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		s.cache.CreateUrl(ctx, id, url)
	}(id, url)

	if err = s.db.CreateUrl(ctx, id, url); err != nil {
		return "", err
	}

	return id, nil
}

func isValidURL(x string) bool {
	if x == "" {
		return false
	}
	_, err := url.Parse(x)
	if len(x) > 2049 && err != nil {
		return false
	}
	return true
}

func (s *ShortnerService) Expand(ctx context.Context, token string) (string, error) {
	if !isValidToken(token) {
		return "", ErrorInvalidToken
	}

	url, err := s.cache.GetUrl(ctx, token)
	if err == nil {
		return url, nil
	}

	url, err = s.db.GetUrl(ctx, token)
	if err != nil {
		return "", ErrorNotFound
	}

	// If we got here that means we did not have it cached.
	// So try to cache it asyncronously.
	go func(id, url string) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		s.cache.CreateUrl(ctx, id, url)
	}(token, url)

	return url, nil
}

func isValidToken(token string) bool {
	if token == "/" || len(token) < 8 {
		return false
	}
	return true
}

var ErrorNotFound = errors.New("token not found.")
var ErrorInvalidToken = errors.New("invalid token")
var ErrorInvalidURL = errors.New("invalid url")
