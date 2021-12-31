package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/dript0hard/lilurl/pkg/config"
	"github.com/dript0hard/lilurl/pkg/shortner"
	"github.com/dript0hard/lilurl/pkg/storage/cache"
	"github.com/dript0hard/lilurl/pkg/storage/postgres"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Init()
	if err != nil {
		log.Fatal(err.Error())
	}
	cfg.Dump(os.Stdout)

	db, err := postgres.Open(ctx, cfg)
	if err != nil {
		log.Fatalf("(OpenDB): %s", err.Error())
	}
	defer db.Close()

	cache, err := cache.Open(ctx, cfg)
	if err != nil {
		log.Fatalf("(OpenDB): %s", err.Error())
	}
	defer db.Close()

	serviceCfg := &shortner.ShortnerConfig{
		DB:    db,
		Cache: cache,
	}
	ssvc := shortner.NewShortner(serviceCfg)

	serverCfg := &shortner.ServerConfig{Shortner: ssvc}
	server := shortner.NewServer(serverCfg)
	handler := server.Install()

	log.Fatal(http.ListenAndServe(":8080", handler))
}
