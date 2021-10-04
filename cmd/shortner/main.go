package main

import (
	"net/http"

	"github.com/dript0hard/lilurl/pkg/shortner"
)

func main() {
	svc := shortner.NewShortnerSvc()
	server := shortner.NewServer(&svc)
	handler := server.Install()
	http.ListenAndServe(":8080", handler)
}
