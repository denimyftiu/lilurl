package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/dript0hard/lilurl/pkg/config"
	"github.com/dript0hard/lilurl/pkg/database"
	"github.com/dript0hard/lilurl/pkg/shortner"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Init()
	if err != nil {
		log.Fatal(err.Error())
	}
	cfg.Dump(os.Stdout)

	db, err := database.OpenDB(ctx, cfg)
	if err != nil {
		log.Fatalf("(OpenDB): %s", err.Error())
	}
	defer db.Close()

	ssvc := shortner.NewShortner(shortner.ShortnerConfig{DB: db})
	server := shortner.NewServer(shortner.ServerConfig{Shortner: ssvc})
	handler := server.Install()
	log.Fatal(http.ListenAndServe(":8080", handler))
}
