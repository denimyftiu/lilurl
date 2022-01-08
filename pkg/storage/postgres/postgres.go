package postgres

import (
	"context"
	"log"

	"github.com/denimyftiu/lilurl/pkg/config"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
}

func (db *DB) Close() {
	db.pool.Close()
}

func (db DB) CreateUrl(ctx context.Context, id, url string) error {
	log.Printf("postgres set(%q): %s", id, url)
	tx, err := db.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Printf("postgres beginTxError: %s", err.Error())
		return err
	}

	if ct, err := tx.Exec(ctx, "INSERT INTO urls (id, url) VALUES ($1, $2)", id, url); err != nil {
		log.Printf("postgres commandTag: %s", ct.String())
		log.Printf("postgres set(%q): %s", id, err.Error())
		tx.Rollback(ctx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("postgres commit: %s", err.Error())
		tx.Rollback(ctx)
		return err
	}
	return nil
}

func (db DB) GetUrl(ctx context.Context, id string) (string, error) {
	var url string
	if err := db.pool.QueryRow(ctx, "SELECT url FROM urls WHERE id = $1", id).Scan(&url); err != nil {
		log.Printf("postgres get(%q): %s", id, err.Error())
		return "", err
	}
	log.Printf("postgres get(%q): %s", id, url)
	return url, nil
}

func Open(ctx context.Context, cfg *config.Config) (*DB, error) {
	pool, err := pgxpool.Connect(ctx, cfg.DBConnURI())
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return &DB{pool: pool}, nil
}
