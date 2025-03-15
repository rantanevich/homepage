package provider

import (
	"context"
	"log/slog"
	"sync"

	"github.com/rantanevich/homepage/app/config/dynamic"
)

type Provider interface {
	Provide(context.Context, *sync.WaitGroup, chan<- dynamic.Message, *slog.Logger) error
	ProviderName() string
}
