package os

import (
	"context"
	"maps"
	"os"
	"strings"
)

type ConfigProvider struct {
	initial map[string]string
}

func New() *ConfigProvider {
	provider := ConfigProvider{}
	provider.initial, _ = provider.FetchConfig(context.TODO())

	return &provider
}

func (c *ConfigProvider) Config(ctx context.Context) (map[string]string, error) {
	return maps.Clone(c.initial), nil
}

func (c *ConfigProvider) FetchConfig(ctx context.Context) (map[string]string, error) {
	envs := os.Environ()

	cfg := make(map[string]string, len(envs))
	for _, e := range envs {
		pair := strings.SplitN(e, "=", 2)
		val := ""
		if len(pair) > 1 {
			val = pair[1]
		}
		cfg[pair[0]] = val
	}

	return cfg, nil
}
