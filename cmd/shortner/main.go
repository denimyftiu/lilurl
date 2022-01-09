package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/denimyftiu/lilurl/pkg/config"
	"github.com/denimyftiu/lilurl/pkg/shortner"
	"github.com/denimyftiu/lilurl/pkg/storage/cache"
	"github.com/denimyftiu/lilurl/pkg/storage/postgres"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var (
	hostAddr = flag.String("host", "localhost:8080", "Host address for the server")
)

func main() {
	flag.Parse()

	ctx := context.Background()
	cfg := config.Init()
	cfg.Dump(os.Stdout)

	db, err := postgres.Open(ctx, cfg)
	if err != nil {
		log.Fatalf("postgres Open: %s", err.Error())
	}
	defer db.Close()

	cache, err := cache.Open(ctx, cfg)
	if err != nil {
		log.Fatalf("cache Open: %s", err.Error())
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

	h2s := &http2.Server{}
	s := http.Server{
		Addr:    *hostAddr,
		Handler: h2c.NewHandler(handler, h2s),
	}

	log.Printf("Serving on http://%s", *hostAddr)
	log.Fatal(s.ListenAndServe())
}
