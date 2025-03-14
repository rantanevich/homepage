package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rantanevich/homepage/app/api"
	"github.com/rantanevich/homepage/app/config"
)

func main() {
	conf, err := config.Load()
	if err != nil {
		log.Fatalf("[FATAL] %v", err)
	}

	srv := api.New()
	if err := srv.RenderIndexPage(conf); err != nil {
		log.Fatalf("[FATAL] cannot render index.html: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go srv.Run(conf.Port, conf.IconsDir)

	sig := <-sigCh
	log.Printf("[INFO] received signal %s", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	srv.Shutdown(ctx)
}
