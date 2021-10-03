package shortner

import (
	"context"
	"errors"
	"sync"
)

// Url Shortener interface to be implemented.
type Shortner interface {
	// Shorten url to token.
	Shorten(context.Context, string) (string, error)

	// Expand token to url.
	Expand(context.Context, string) (string, error)
}

// Service implementation for testing.
type shortnerSvc struct {
	mu    sync.Mutex
	store map[string]string
}

func NewShortnerSvc() shortnerSvc {
	return shortnerSvc{store: map[string]string{}}
}

func (s *shortnerSvc) Shorten(_ context.Context, url string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store["AAA"] = url
	return "AAA", nil
}

func (s *shortnerSvc) Expand(_ context.Context, token string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	url, ok := s.store[token]
	if ok {
		return url, nil
	}
	return "", ErrorNotFound
}

var ErrorNotFound = errors.New("token not found.")
