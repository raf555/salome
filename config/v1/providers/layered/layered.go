// Package layered merges several config providers into one.
package layered

import (
	"context"
	"errors"
	"fmt"
	"maps"

	config "github.com/raf555/salome/config/v1"
)

var ErrNoProviders = errors.New("layered: no providers given")

// ConfigProvider merges configs from its providers in order.
// When a key exists in more than one provider, the later provider wins.
type ConfigProvider struct {
	providers []config.Provider
	initial   map[string]string
}

func New(providers ...config.Provider) (*ConfigProvider, error) {
	if len(providers) == 0 {
		return nil, ErrNoProviders
	}

	provider := &ConfigProvider{
		providers: providers,
	}

	initial, err := provider.merge(func(p config.Provider) (map[string]string, error) {
		return p.Config(context.TODO())
	})
	if err != nil {
		return nil, err
	}

	provider.initial = initial
	return provider, nil
}

func (c *ConfigProvider) Config(_ context.Context) (map[string]string, error) {
	return maps.Clone(c.initial), nil
}

func (c *ConfigProvider) FetchConfig(ctx context.Context) (map[string]string, error) {
	return c.merge(func(p config.Provider) (map[string]string, error) {
		return p.FetchConfig(ctx)
	})
}

func (c *ConfigProvider) merge(read func(config.Provider) (map[string]string, error)) (map[string]string, error) {
	out := make(map[string]string)
	for i, p := range c.providers {
		cfg, err := read(p)
		if err != nil {
			return nil, fmt.Errorf("provider %d: %w", i, err)
		}
		maps.Copy(out, cfg)
	}
	return out, nil
}
