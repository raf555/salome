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
	initial  map[string]string
}

func New(filename string) (*ConfigProvider, error) {
	_, err := os.Stat(filename)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("%w: %w", ErrDotenvFileNotFound, err)
		}
		return nil, fmt.Errorf("os.Stat: %w", err)
	}

	provider := ConfigProvider{
		filename: filename,
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
		return nil, fmt.Errorf("os.Open: %w", err)
	}
	defer file.Close()

	cfgs, err := godotenv.Parse(file)
	if err != nil {
		return nil, fmt.Errorf("godotenv.Parse: %w", err)
	}

	return cfgs, nil
}
