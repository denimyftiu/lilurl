package database

import (
	"context"
	"log"

	"github.com/dript0hard/lilurl/pkg/config"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type ShortnerStorage interface {
	CreateUrl(ctx context.Context, id, url string) error
	GetUrl(ctx context.Context, id string) (url string, err error)
}

type DB struct {
	pool *pgxpool.Pool
}

func (db *DB) Close() {
	db.pool.Close()
}

func (db DB) CreateUrl(ctx context.Context, id, url string) error {
	log.Printf("(Inserting) Id: %s, Url: %s", id, url)
	tx, err := db.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Printf("(BeginTx): %s", err.Error())
		return err
	}

	if ct, err := tx.Exec(ctx, "INSERT INTO urls (id, url) VALUES ($1, $2)", id, url); err != nil {
		log.Printf("sql: %s", ct.String())
		log.Printf("(Exec): %s", err.Error())
		tx.Rollback(ctx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("(Commit): %s", err.Error())
		tx.Rollback(ctx)
		return err
	}
	return nil
}

func (db DB) GetUrl(ctx context.Context, id string) (string, error) {
	var url string
	if err := db.pool.QueryRow(ctx, "SELECT url FROM urls WHERE id = $1", id).Scan(&url); err != nil {
		log.Printf("(Scan): %s", err.Error())
		return "", err
	}
	log.Printf("(Retrieving) Url: %s", url)
	return url, nil
}

func OpenDB(ctx context.Context, cfg *config.Config) (*DB, error) {
	pool, err := pgxpool.Connect(ctx, cfg.DBConnURI())
	if err != nil {
		return nil, err
	}

	return &DB{pool: pool}, nil
}
