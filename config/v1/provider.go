package config

import "context"

type Provider interface {
	// Config provides a config that is initially read by the provider.
	Config(ctx context.Context) (map[string]string, error)
	// FetchConfig provides a config that is read by provider immediately upon call.
	FetchConfig(ctx context.Context) (map[string]string, error)
}

type DynamicConfigManager interface {
	// GetConfig provides a config from the provided key.
	// It may return nil if not found.
	GetConfig(key any) any
	// RegisterConfig registers a config to be dynamically updated.
	// Factory must provide a pointer to zero config struct to be used for parsing configs..
	RegisterConfig(key any, factory func() any) error
}
