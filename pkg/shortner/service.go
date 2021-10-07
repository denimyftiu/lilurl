package shortner

import (
	"context"
	"errors"
	"net/url"
	"sync"

	"github.com/teris-io/shortid"
)

// Url Shortener interface to be implemented.
type Shortner interface {
	// Shorten url to token.
	Shorten(context.Context, string) (string, error)

	// Expand token to url.
	Expand(context.Context, string) (string, error)
}

var ErrorNotFound = errors.New("token not found.")
var ErrorInvalidToken = errors.New("invalid token")
var ErrorInvalidURL = errors.New("invalid url")

// Service implementation for testing.
type shortnerSvc struct {
	mu    sync.Mutex
	store map[string]string
}

func NewShortnerSvc() shortnerSvc {
	return shortnerSvc{store: map[string]string{}}
}

func (s *shortnerSvc) Shorten(_ context.Context, url string) (string, error) {
	if !isValidURL(url) {
		return "", ErrorInvalidURL
	}
	token, err := shortid.Generate()
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[token] = url
	return token, nil
}

func isValidURL(x string) bool {
	_, err := url.Parse(x)
	if len(x) > 2049 && err != nil {
		return false
	}
	return true
}

func (s *shortnerSvc) Expand(_ context.Context, token string) (string, error) {
	if !isValidToken(token) {
		return "", ErrorInvalidToken
	}

	s.mu.Lock()
	url, ok := s.store[token]
	s.mu.Unlock()

	if ok {
		return url, nil
	}
	return "", ErrorNotFound
}

func isValidToken(token string) bool {
	if token == "/" || len(token) < 5 {
		return false
	}
	return true
}
