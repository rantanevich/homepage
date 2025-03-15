package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/rantanevich/homepage/app/api"
	"github.com/rantanevich/homepage/app/config/dynamic"
	"github.com/rantanevich/homepage/app/config/static"
	"github.com/rantanevich/homepage/app/watcher"
)

func main() {
	conf := static.Load()
	srv := api.New()

	log.Printf("[DEBUG] conf: %+v", conf)

	w := watcher.New(conf.Providers)
	w.AddListener(func(c dynamic.Config) {
		srv.UpdateIndexPage(c, conf.Title, conf.Logo)
	})

	ctx, stop := signal.NotifyContext(context.TODO(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := w.Start(ctx); err != nil {
			log.Printf("[FATAL] cannot start watcher: %v", err)
			stop()
		}
	}()

	go srv.Run(conf.Port, conf.IconsDir)

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	srv.Shutdown(ctx)
	w.Shutdown()
}
