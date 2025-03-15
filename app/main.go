package main

import (
	"context"
	"log/slog"
	"os"
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
	log := setupLogger(conf.LogLevel)

	srv := api.New(log)

	w := watcher.New(conf.Providers, log)
	w.AddListener(func(c dynamic.Config) {
		srv.UpdateIndexPage(c, conf.Title, conf.Logo)
	})

	ctx, stop := signal.NotifyContext(context.TODO(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := w.Start(ctx); err != nil {
			log.Error("failed to start watcher", slog.String("error", err.Error()))
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

func setupLogger(level string) *slog.Logger {
	var logLevel slog.Level

	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
}
