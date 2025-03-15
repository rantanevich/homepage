package provider

import (
	"context"
	"sync"

	"github.com/rantanevich/homepage/app/config/dynamic"
)

type Provider interface {
	Provide(context.Context, *sync.WaitGroup, chan<- dynamic.Message) error
}
