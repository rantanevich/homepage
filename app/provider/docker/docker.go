package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/docker/cli/cli/connhelper"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"

	"github.com/rantanevich/homepage/app/config/dynamic"
	"github.com/rantanevich/homepage/app/provider"
)

var _ provider.Provider = (*Provider)(nil)

type Provider struct {
	Endpoint          string
	Watch             bool
	HTTPClientTimeout time.Duration
}

func (p *Provider) SetDefaults() {
	p.Endpoint = "unix:///var/run/docker.sock"
	p.Watch = true
}

func (p *Provider) ProviderName() string {
	return "docker"
}

func (p *Provider) Provide(ctx context.Context, wg *sync.WaitGroup, updateCh chan<- dynamic.Message, log *slog.Logger) error {
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
				fmt.Printf("err: %v\n", err)
				return err
			}

			conf := buildConfig(containers)
			if conf != nil {
				updateCh <- dynamic.Message{
					ProviderName: p.ProviderName(),
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
							log.Debug(
								"received event",
								slog.String("event_from", event.Actor.Attributes["name"]),
								slog.String("event_action", string(event.Action)),
							)
							containers, err := dockerClient.ContainerList(ctx, container.ListOptions{})
							if err != nil {
								log.Error("failed to list containers", slog.String("error", err.Error()))
								continue
							}

							conf := buildConfig(containers)
							if conf != nil {
								updateCh <- dynamic.Message{
									ProviderName: p.ProviderName(),
									Config:       conf,
								}
							}
						}
					case err := <-errCh:
						if errors.Is(err, io.EOF) {
							log.Debug("event stream closed")
						}
						return err
					}
				}
			}

			return nil
		}

		retry.Do(
			operationWithRecover(operation, log),
			retry.UntilSucceeded(),
			retry.Delay(5*time.Second),
			retry.MaxDelay(60*time.Second),
			retry.Context(ctx),
			retry.OnRetry(func(attempt uint, err error) {
				log.Error(
					"provider error, retrying",
					slog.Int("attempt", int(attempt)),
					slog.String("error", err.Error()),
				)
			}),
		)
	}()

	return nil
}

func (p *Provider) createClient() (*client.Client, error) {
	opts, err := p.getClientOpts()
	if err != nil {
		return nil, err
	}

	dockerClient, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, err
	}
	return dockerClient, nil
}

func (p *Provider) getClientOpts() ([]client.Opt, error) {
	helper, err := connhelper.GetConnectionHelper(p.Endpoint)
	if err != nil {
		return nil, err
	}

	opts := []client.Opt{
		client.WithAPIVersionNegotiation(),
	}

	if helper != nil {
		// https://github.com/docker/cli/blob/ebca1413117a3fcb81c89d6be226dcec74e5289f/cli/context/docker/load.go#L112-L123
		httpClient := &http.Client{
			Transport: &http.Transport{
				DialContext: helper.Dialer,
			},
		}

		opts = append(
			opts,
			client.WithHTTPClient(httpClient),
			client.WithTimeout(time.Duration(p.HTTPClientTimeout)),
			client.WithHost(helper.Host), // To avoid 400 Bad Request: malformed Host header daemon error
			client.WithDialContext(helper.Dialer),
		)
		return opts, nil
	}

	opts = append(opts, client.WithHost(p.Endpoint))
	return opts, nil
}
