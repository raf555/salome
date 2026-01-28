package osdotenv

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"maps"

	"github.com/raf555/salome/config/v1/providers/dotenv"
	"github.com/raf555/salome/config/v1/providers/os"
)

// ConfigProvider provides config from both OS and .env file (if present).
// Config from .env file takes precedence if exists in both places.
type ConfigProvider struct {
	osCfgProvider     *os.ConfigProvider
	dotenvCfgProvider *dotenv.ConfigProvider

	initial map[string]string
}

func New(filename string) (*ConfigProvider, error) {
	dotenvProvider, err := dotenv.New(filename)
	if err != nil && !errors.Is(err, dotenv.ErrDotenvFileNotFound) {
		return nil, fmt.Errorf("dotenv.New: %w", err)
	}

	osCfgProvider := os.New()

	provider := &ConfigProvider{
		dotenvCfgProvider: dotenvProvider,
		osCfgProvider:     osCfgProvider,
	}

	provider.initial, err = provider.FetchConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("provider.FetchConfig: %w", err)
	}

	return provider, nil
}

func (c *ConfigProvider) Config(ctx context.Context) (map[string]string, error) {
	return maps.Clone(c.initial), nil
}

func (c *ConfigProvider) FetchConfig(ctx context.Context) (map[string]string, error) {
	osEnv, err := c.osCfgProvider.FetchConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("c.osCfgProvider.FetchConfig: %w", err)
	}

	out := maps.Clone(osEnv)

	if c.dotenvCfgProvider != nil {
		dotEnv, err := c.dotenvCfgProvider.FetchConfig(ctx)
		if err != nil && !errors.Is(err, fs.ErrNotExist) { // if file not present, omit error
			return nil, fmt.Errorf("c.dotenvCfgProvider.FetchConfig: %w", err)
		}

		maps.Copy(out, dotEnv)
	}

	return out, nil
}
