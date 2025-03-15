package watcher

import (
	"context"
	"log"
	"reflect"
	"sync"

	"github.com/rantanevich/homepage/app/config/dynamic"
	"github.com/rantanevich/homepage/app/config/static"
	"github.com/rantanevich/homepage/app/provider"
)

type Watcher struct {
	wg sync.WaitGroup

	state    map[string]dynamic.Config
	updateCh chan dynamic.Message

	providers []provider.Provider
	listeners []func(dynamic.Config)
}

func New(conf static.Providers) *Watcher {
	w := &Watcher{
		updateCh: make(chan dynamic.Message, 10),
		state:    make(map[string]dynamic.Config),
	}

	if conf.File != nil {
		w.providers = append(w.providers, conf.File)
	}

	return w
}

func (w *Watcher) Start(ctx context.Context) error {
	for _, provider := range w.providers {
		log.Printf("[DEBUG] starting %v provider: %+v", reflect.TypeOf(provider), provider)
		if err := provider.Provide(ctx, &w.wg, w.updateCh); err != nil {
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
				log.Printf("[DEBUG] received update from %s provider: %+v", update.ProviderName, update.Config)
				if update.Config == nil {
					log.Printf("[DEBUG] skipping nil config")
					continue
				}

				if reflect.DeepEqual(w.state[update.ProviderName], update.Config) {
					log.Printf("[DEBUG] skipping unchanged config")
					continue
				}

				w.state[update.ProviderName] = update.Config.DeepCopy()

				conf := mergeConfig(w.state)
				log.Printf("[DEBUG] final config: %+v", conf)

				for _, listener := range w.listeners {
					listener(conf)
				}
			}
		}
	}()

	return nil
}

func (w *Watcher) Shutdown() {
	close(w.updateCh)
	w.wg.Wait()
}

func (w *Watcher) AddListener(listener func(dynamic.Config)) {
	w.listeners = append(w.listeners, listener)
}
