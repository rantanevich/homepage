package file

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"

	"github.com/rantanevich/homepage/app/config/dynamic"
	"github.com/rantanevich/homepage/app/provider"
)

const providerName = "file"

var _ provider.Provider = (*Provider)(nil)

type Provider struct {
	Filename string
	Watch    bool
}

func (p *Provider) SetDefaults() {
	p.Filename = ""
	p.Watch = true
}

func (p *Provider) Provide(ctx context.Context, wg *sync.WaitGroup, updateCh chan<- dynamic.Message) error {
	if p.Watch {
		if err := p.addWatcher(ctx, wg, updateCh); err != nil {
			return err
		}
	}

	conf, err := p.loadFileConfig()
	if err != nil {
		return err
	}

	updateCh <- dynamic.Message{
		ProviderName: providerName,
		Config:       conf,
	}

	return nil
}

func (p *Provider) addWatcher(ctx context.Context, wg *sync.WaitGroup, updateCh chan<- dynamic.Message) error {
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
				log.Printf("[DEBUG] file event: %v", event)
				if event.Has(fsnotify.Write) {
					conf, err := p.loadFileConfig()
					if err != nil {
						log.Printf("[ERROR] failed to load file config: %v", err)
						continue
					}
					updateCh <- dynamic.Message{
						ProviderName: providerName,
						Config:       conf,
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("[ERROR] file watcher event error: %v", err)
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
