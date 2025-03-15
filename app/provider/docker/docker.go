package docker

import (
	"context"
	"errors"
	"io"
	"log"
	"sync"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"

	"github.com/rantanevich/homepage/app/config/dynamic"
	"github.com/rantanevich/homepage/app/provider"
)

const DockerAPIVersion = "1.24"

var _ provider.Provider = (*Provider)(nil)

type Provider struct {
	Endpoint string
	Watch    bool
}

func (p *Provider) SetDefaults() {
	p.Endpoint = "unix:///var/run/docker.sock"
	p.Watch = true
}

func (p *Provider) Provide(ctx context.Context, wg *sync.WaitGroup, updateCh chan<- dynamic.Message) error {
	wg.Add(1)

	go func() {
		defer wg.Done()

		operation := func() error {
			dockerClient, err := p.createClient()
			if err != nil {
				return err
			}
			defer dockerClient.Close()

			containers, err := dockerClient.ContainerList(ctx, container.ListOptions{})
			if err != nil {
				return err
			}

			conf := buildConfig(containers)
			if conf != nil {
				updateCh <- dynamic.Message{
					ProviderName: "docker",
					Config:       conf,
				}
			}

			if p.Watch {
				f := filters.NewArgs()
				f.Add("type", "container")
				opts := events.ListOptions{
					Filters: f,
				}

				eventCh, errCh := dockerClient.Events(ctx, opts)
				for {
					select {
					case <-ctx.Done():
						return nil
					case event := <-eventCh:
						if event.Action == "start" || event.Action == "die" {
							log.Printf("[DEBUG] docker provider received: %+v", event)
							containers, err := dockerClient.ContainerList(ctx, container.ListOptions{})
							if err != nil {
								log.Printf("[ERROR] failed to list docker containers: %v", err)
								continue
							}

							conf := buildConfig(containers)
							if conf != nil {
								updateCh <- dynamic.Message{
									ProviderName: "docker",
									Config:       conf,
								}
							}
						}
					case err := <-errCh:
						if errors.Is(err, io.EOF) {
							log.Printf("[DEBUG] docker event stream closed")
						}
						return err
					}
				}
			}

			return nil
		}

		retry.Do(
			operation,
			retry.UntilSucceeded(),
			retry.Delay(5*time.Second),
			retry.MaxDelay(60*time.Second),
		)
	}()

	return nil
}

func (p *Provider) createClient() (*client.Client, error) {
	opts := []client.Opt{
		client.WithHost(p.Endpoint),
		client.WithVersion(DockerAPIVersion),
	}

	dockerClient, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, err
	}
	return dockerClient, nil
}
