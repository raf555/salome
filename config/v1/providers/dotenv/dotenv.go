package dotenv

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"

	"github.com/joho/godotenv"
)

var (
	ErrDotenvFileNotFound = errors.New("dotenv: provided dotenv file not found")
)

// ConfigProvider is a config provider from *.env file.
type ConfigProvider struct {
	filename string
	optional bool
	initial  map[string]string
}

type Option func(*ConfigProvider)

// WithAllowMissing makes a missing file provide no config instead of an error,
// both in New and in later fetches.
func WithAllowMissing() Option {
	return func(c *ConfigProvider) {
		c.optional = true
	}
}

func New(filename string, opts ...Option) (*ConfigProvider, error) {
	provider := ConfigProvider{
		filename: filename,
	}
	for _, opt := range opts {
		opt(&provider)
	}

	if !provider.optional {
		_, err := os.Stat(filename)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil, fmt.Errorf("%w: %w", ErrDotenvFileNotFound, err)
			}
			return nil, fmt.Errorf("os.Stat: %w", err)
		}
	}

	initialCfg, err := provider.FetchConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("provider.FetchConfig: %w", err)
	}

	provider.initial = initialCfg
	return &provider, nil
}

func (c *ConfigProvider) Config(_ context.Context) (map[string]string, error) {
	return maps.Clone(c.initial), nil
}

func (c *ConfigProvider) FetchConfig(_ context.Context) (map[string]string, error) {
	dotenvCfgs, err := c.readConfig()
	if err != nil {
		return nil, fmt.Errorf("d.readConfig: %w", err)
	}

	return dotenvCfgs, nil
}

func (c *ConfigProvider) readConfig() (map[string]string, error) {
	file, err := os.Open(c.filename)
	if err != nil {
		if c.optional && errors.Is(err, fs.ErrNotExist) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("os.Open: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	cfgs, err := godotenv.Parse(file)
	if err != nil {
		return nil, fmt.Errorf("godotenv.Parse: %w", err)
	}

	return cfgs, nil
}
