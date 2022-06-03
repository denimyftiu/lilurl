package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/denimyftiu/lilurl/pkg/config"
	"github.com/denimyftiu/lilurl/pkg/shortner"
	"github.com/denimyftiu/lilurl/pkg/storage/cache"
	"github.com/denimyftiu/lilurl/pkg/storage/postgres"
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
	defer cache.Close()

	serviceCfg := &shortner.ShortnerConfig{
		DB:    db,
		Cache: cache,
	}
	ssvc := shortner.NewShortner(serviceCfg)

	serverCfg := &shortner.ServerConfig{Shortner: ssvc}
	server := shortner.NewServer(serverCfg)
	serveMux := server.Install()

	s := http.Server{
		Addr:         *hostAddr,
		Handler:      serveMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Serving on http://%s", *hostAddr)
		if err := s.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	sigChan := make(chan os.Signal)
	defer close(sigChan)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	sig := <-sigChan
	log.Printf("Recieved termination signal: %s", sig)

	shutDownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Shutdown(shutDownCtx); err != nil {
		log.Fatal(err)
	}
	log.Printf("Terminated gracefully!")
}
