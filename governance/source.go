package governance

import (
	"context"
	"sync/atomic"
)

type Source interface {
	Start(ctx context.Context) error
	Registry() Registry
	Close() error
}

type InMemorySource struct {
	registry atomic.Value
}

func NewInMemorySource(registry Registry) *InMemorySource {
	source := &InMemorySource{}
	source.registry.Store(registry.Clone())
	return source
}

func (s *InMemorySource) Start(ctx context.Context) error {
	return ctx.Err()
}

func (s *InMemorySource) Registry() Registry {
	return LoadRegistry(&s.registry)
}

func (s *InMemorySource) Close() error {
	return nil
}

func LoadRegistry(value *atomic.Value) Registry {
	loaded := value.Load()
	if loaded == nil {
		return EmptyRegistry()
	}
	registry, ok := loaded.(Registry)
	if !ok {
		return EmptyRegistry()
	}
	return registry.Clone()
}
