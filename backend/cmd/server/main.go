package main

import (
	"context"
	"log"

	"ragserver/backend/internal/app"
	"ragserver/backend/internal/config"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()
	application, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("RagServer listening on %s", cfg.ServerAddr)
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
