package database

import "context"

// All storage implementations might want to implement this so we can do
// caching for the database (using a middleware pattern).
type ShortnerStorage interface {
	CreateUrl(ctx context.Context, id, url string) error
	GetUrl(ctx context.Context, id string) (url string, err error)
}
