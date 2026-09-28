package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/racecoach/workers/internal/config"
	"github.com/racecoach/workers/internal/fitsummary"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := fitsummary.Run(ctx, cfg); err != nil {
		log.Fatal(err)
	}
}
