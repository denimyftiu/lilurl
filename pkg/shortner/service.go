package shortner

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/denimyftiu/lilurl/pkg/storage/cache"
	"github.com/denimyftiu/lilurl/pkg/storage/postgres"
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

func (s ShortnerService) saveCache(id, url string) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_ = s.cache.CreateUrl(ctx, id, url)
}

func (s ShortnerService) Shorten(ctx context.Context, url string) (string, error) {
	if !isValidURL(url) {
		return "", ErrorInvalidURL
	}

	id, err := shortid.Generate()
	if err != nil {
		return "", err
	}

	// Save it to cache asyncronously.
	go s.saveCache(id, url)

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
	if err != nil {
		return false
	}
	if len(x) > 2049 {
		return false
	}
	return true
}

func (s *ShortnerService) Expand(ctx context.Context, token string) (string, error) {
	if !isValidToken(token) {
		return "", ErrorInvalidToken
	}

	// If we recieved the url from cache without errors we just return it.
	if url, err := s.cache.GetUrl(ctx, token); err == nil {
		return url, nil
	}

	if url, err := s.db.GetUrl(ctx, token); err == nil {
		// If we got here that means we did not have it cached.
		// So try to cache it asyncronously and return directly.
		go s.saveCache(token, url)
		return url, nil
	}

	return "", ErrorNotFound
}

func isValidToken(token string) bool {
	if token == "/" || len(token) < 8 {
		return false
	}
	return true
}

var ErrorNotFound = errors.New("token not found")
var ErrorInvalidToken = errors.New("invalid token")
var ErrorInvalidURL = errors.New("invalid url")
