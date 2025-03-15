package file

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"

	"github.com/rantanevich/homepage/app/config/dynamic"
	"github.com/rantanevich/homepage/app/provider"
)

var _ provider.Provider = (*Provider)(nil)

type Provider struct {
	Filename string
	Watch    bool
}

func (p *Provider) SetDefaults() {
	p.Filename = ""
	p.Watch = true
}

func (p *Provider) ProviderName() string {
	return "file"
}

func (p *Provider) Provide(ctx context.Context, wg *sync.WaitGroup, updateCh chan<- dynamic.Message, log *slog.Logger) error {
	if p.Watch {
		if err := p.addWatcher(ctx, wg, updateCh, log); err != nil {
			return err
		}
	}

	conf, err := p.loadFileConfig()
	if err != nil {
		return err
	}

	updateCh <- dynamic.Message{
		ProviderName: p.ProviderName(),
		Config:       conf,
	}

	return nil
}

func (p *Provider) addWatcher(ctx context.Context, wg *sync.WaitGroup, updateCh chan<- dynamic.Message, log *slog.Logger) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	if err := watcher.Add(p.Filename); err != nil {
		return err
	}

	wg.Add(1)

	go func() {
		defer wg.Done()
		defer watcher.Close()

		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Debug(
					"received event",
					slog.String("event_from", event.Name),
					slog.String("event_action", event.Op.String()),
				)
				if event.Has(fsnotify.Write) {
					conf, err := p.loadFileConfig()
					if err != nil {
						log.Error(
							"cannot load file config",
							slog.String("filename", p.Filename),
							slog.String("error", err.Error()),
						)
						continue
					}
					updateCh <- dynamic.Message{
						ProviderName: p.ProviderName(),
						Config:       conf,
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Error(
					"received error event",
					slog.String("error", err.Error()),
				)
			}
		}
	}()

	return nil
}

func (p *Provider) loadFileConfig() (dynamic.Config, error) {
	var config dynamic.Config

	file, err := os.Open(p.Filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if err := yaml.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}
	return config, nil
}
