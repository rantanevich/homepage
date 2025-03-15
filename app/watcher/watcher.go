package watcher

import (
	"context"
	"log/slog"
	"reflect"
	"sync"

	"github.com/rantanevich/homepage/app/config/dynamic"
	"github.com/rantanevich/homepage/app/config/static"
	"github.com/rantanevich/homepage/app/provider"
)

type Watcher struct {
	log *slog.Logger

	wg       sync.WaitGroup
	state    map[string]dynamic.Config
	updateCh chan dynamic.Message

	providers []provider.Provider
	listeners []func(dynamic.Config)
}

func New(conf static.Providers, log *slog.Logger) *Watcher {
	w := &Watcher{
		log:      log.With(slog.String("component", "watcher")),
		updateCh: make(chan dynamic.Message, 10),
		state:    make(map[string]dynamic.Config),
	}

	if conf.File != nil {
		w.providers = append(w.providers, conf.File)
	}

	if conf.Docker != nil {
		w.providers = append(w.providers, conf.Docker)
	}

	return w
}

func (w *Watcher) Start(ctx context.Context) error {
	for _, provider := range w.providers {
		logger := w.log.With(slog.String("provider", provider.ProviderName()))
		logger.Debug("starting provider")
		if err := provider.Provide(ctx, &w.wg, w.updateCh, logger); err != nil {
			return err
		}
	}

	w.wg.Add(1)

	go func() {
		defer w.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case update := <-w.updateCh:
				logger := w.log.With(slog.String("provider", update.ProviderName))

				logger.Debug("received update from provider")
				if update.Config == nil {
					logger.Debug("skipping nil config")
					continue
				}

				if reflect.DeepEqual(w.state[update.ProviderName], update.Config) {
					logger.Debug("skipping unchanged config")
					continue
				}

				w.state[update.ProviderName] = update.Config.DeepCopy()

				conf := mergeConfig(w.state)
				for _, listener := range w.listeners {
					listener(conf)
				}
			}
		}
	}()

	w.log.Info("started")
	return nil
}

func (w *Watcher) Shutdown() {
	close(w.updateCh)
	w.wg.Wait()
	w.log.Info("stopped")
}

func (w *Watcher) AddListener(listener func(dynamic.Config)) {
	w.listeners = append(w.listeners, listener)
}
