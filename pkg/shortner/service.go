package shortner

import (
	"context"
	"errors"
	"log"
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

	if err = s.cache.CreateUrl(ctx, id, url); err != nil {
		log.Printf("cache set(%q): cant set -> %v", id, url)
		return "", err
	}

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

	getCtx, cancel := context.WithTimeout(ctx, 30*time.Millisecond)
	defer cancel()

	url, err := s.cache.GetUrl(getCtx, token)
	if err != nil {
		select {
		case <-getCtx.Done():
			log.Printf("cache get(%q): context timed out", token)
		default:
			log.Printf("cache get(%q): %v", token, err)
		}
	}

	if url != "" {
		return url, nil
	}

	url, err = s.db.GetUrl(ctx, token)
	if err != nil {
		return "", ErrorNotFound
	}

	return url, nil
}

func isValidToken(token string) bool {
	if token == "/" || len(token) < 5 {
		return false
	}
	return true
}

var ErrorNotFound = errors.New("token not found.")
var ErrorInvalidToken = errors.New("invalid token")
var ErrorInvalidURL = errors.New("invalid url")
