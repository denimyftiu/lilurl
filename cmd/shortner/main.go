package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	defer signal.Stop(sigChan)

	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	sig := <-sigChan
	log.Printf("Received termination signal: %s", sig)

	shutDownCtx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	if err := s.Shutdown(shutDownCtx); err != nil {
		log.Fatalf("Could not shut down gracefully: %s", err)
	}

	log.Printf("Terminated gracefully!")
}
